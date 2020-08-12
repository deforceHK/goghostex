package okex

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"testing"
	"time"

	. "github.com/strengthening/goghostex"
)

const (
	MARGIN_API_KEY        = ""
	MARGIN_API_SECRETKEY  = ""
	MARGIN_API_PASSPHRASE = ""
)

func TestMargin_MarketAPI(t *testing.T) {

	config := &APIConfig{
		Endpoint: ENDPOINT,
		HttpClient: &http.Client{
			Transport: &http.Transport{
				Proxy: func(req *http.Request) (*url.URL, error) {
					return url.Parse(PROXY_URL)
				},
			},
		},
		ApiKey:        MARGIN_API_KEY,
		ApiSecretKey:  MARGIN_API_SECRETKEY,
		ApiPassphrase: MARGIN_API_PASSPHRASE,
		Location:      time.Now().Location(),
	}

	ok := New(config)
	if resp, err := ok.Margin.GetMarginInfo(
		Pair{Basis: BTC, Counter: USDT},
	); err != nil {
		t.Error(err)
		return
	} else {
		fmt.Println(string(resp))
	}
}

/**
 *
 * The func of loan unit test step is:
 * 1. Get the account, and find have the enough crypto to loan 1 usdt.
 * 2. Loan 1 usdt in btc_usdt currency pair.
 * 3. Get the Loan info, and get the loan fee
 * 4. Repay the 1 usdt and loan fee
 *
 **/
func TestMargin_LoanAPI(t *testing.T) {
	config := &APIConfig{
		Endpoint: ENDPOINT,
		HttpClient: &http.Client{
			Transport: &http.Transport{
				Proxy: func(req *http.Request) (*url.URL, error) {
					return url.Parse(PROXY_URL)
				},
			},
		},
		ApiKey:        MARGIN_API_KEY,
		ApiSecretKey:  MARGIN_API_SECRETKEY,
		ApiPassphrase: MARGIN_API_PASSPHRASE,
		Location:      time.Now().Location(),
	}

	ok := New(config)
	if account, resp, err := ok.Margin.GetMarginAccount(
		Pair{Basis: BTC, Counter: USDT},
	); err != nil {
		t.Error(err)
		return
	} else {
		t.Log("")
		t.Log(*account)
		t.Log("Remote api return the response: ")
		t.Log(string(resp))
	}

	loan := LoanRecord{
		Pair:     Pair{Basis: BTC, Counter: USDT},
		Currency: USDT,
		Amount:   1,
	}

	if resp, err := ok.Margin.Loan(&loan); err != nil {
		t.Error(err)
		return
	} else {
		if standard, err := json.Marshal(loan); err != nil {
			t.Error(err)
			return
		} else {
			t.Log("standard struct:")
			t.Log(string(standard))
			t.Log("remote api return: ")
			t.Log(string(resp))
		}
	}

	if resp, err := ok.Margin.GetOneLoan(&loan); err != nil {
		t.Error(err)
		return
	} else {
		if standard, err := json.Marshal(loan); err != nil {
			t.Error(err)
			return
		} else {
			t.Log("The standard loan record is: ")
			t.Log(string(standard))
			t.Log("The remote api return is: ")
			t.Log(string(resp))
		}
	}

	if resp, err := ok.Margin.Repay(&loan); err != nil {
		t.Error(err)
		return
	} else {
		if standard, err := json.Marshal(loan); err != nil {
			t.Error(err)
			return
		} else {
			t.Log("The standard loan record is: ")
			t.Log(string(standard))
			t.Log("The remote api return is: ")
			t.Log(string(resp))
		}
	}
}

/**
 *
 * The func of order unit test step is:
 * 1. Get the account, and find have the enough crypto.
 * 2. Get BTC-USDT ticker.
 * 2. Place order without deal.
 * 3. Get the unfinished orders info, and find the order in step 1.
 * 4. Get the order info.
 * 5. Cancel the Limit Order
 *
 **/

func TestMargin_TradeAPI(t *testing.T) {

	config := &APIConfig{
		Endpoint: ENDPOINT,
		HttpClient: &http.Client{
			Transport: &http.Transport{
				Proxy: func(req *http.Request) (*url.URL, error) {
					return url.Parse(PROXY_URL)
				},
			},
		},
		ApiKey:        MARGIN_API_KEY,
		ApiSecretKey:  MARGIN_API_SECRETKEY,
		ApiPassphrase: MARGIN_API_PASSPHRASE,
		Location:      time.Now().Location(),
	}

	ok := New(config)
	if _, resp, err := ok.Margin.GetMarginAccount(
		Pair{Basis: BTC, Counter: USDT},
	); err != nil {
		t.Error(err)
		return
	} else {
		fmt.Println(string(resp))
	}

	testPrice := 0.0
	if ticker, resp, err := ok.Margin.GetMarginTicker(
		Pair{Basis: BTC, Counter: USDT},
	); err != nil {
		t.Error(err)
		return
	} else {
		fmt.Println(ticker)
		fmt.Println(string(resp))

		testPrice = ticker.Sell * 1.1
	}

	normalOrder := Order{
		Pair:      Pair{Basis: BTC, Counter: USDT},
		Price:     testPrice,
		Amount:    0.005,
		Side:      SELL,
		OrderType: NORMAL,
	}

	if resp, err := ok.Margin.PlaceMarginOrder(&normalOrder); err != nil {
		t.Error(err)
		return
	} else {
		standard, err := json.Marshal(normalOrder)
		if err != nil {
			t.Error(err)
			return
		}
		t.Log("Place margin order standard struct:")
		t.Log(string(standard))
		t.Log("Place margin order remote api response: ")
		t.Log(string(resp))
	}

	if _, err := ok.Margin.GetMarginOneOrder(&normalOrder); err != nil {
		t.Error(err)
		return
	} else {
		if normalOrder.Status != ORDER_UNFINISH {
			t.Error("The order status must be unfinished")
			return
		}
	}

	if orders, _, err := ok.Margin.GetMarginUnFinishOrders(Pair{Basis: BTC, Counter: USDT}); err != nil {
		t.Error(err)
		return
	} else {
		isFound := false
		for _, order := range orders {
			if order.OrderId == normalOrder.OrderId {
				isFound = true
			}
		}

		if !isFound {
			t.Error("The Unfinished order can not find the order! ")
			return
		}
	}

	if resp, err := ok.Margin.CancelMarginOrder(&normalOrder); err != nil {
		t.Error(err)
		return
	} else {
		standard, err := json.Marshal(normalOrder)
		if err != nil {
			t.Error(err)
			return
		}
		t.Log("Cancel margin order standard struct:")
		t.Log(string(standard))

		t.Log("Cancel margin order remote api response: ")
		t.Log(string(resp))
	}

	if resp, err := ok.Margin.GetMarginOneOrder(&normalOrder); err != nil {
		t.Error(err)
		return
	} else {
		if normalOrder.Status != ORDER_CANCEL {
			t.Error("The order status must be canceled. ")
			return
		}

		standard, err := json.Marshal(normalOrder)
		if err != nil {
			t.Error(err)
			return
		}

		t.Log("Get one order standard struct:")
		t.Log(string(standard))

		t.Log("Get one order remote api response: ")
		t.Log(string(resp))
	}
}
