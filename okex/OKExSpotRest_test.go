package okex

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"testing"
	"time"

	. "github.com/strengthening/goghostex"
)

const (
	SPOT_API_KEY        = ""
	SPOT_API_SECRETKEY  = ""
	SPOT_API_PASSPHRASE = ""
	PROXY_URL           = "socks5://127.0.0.1:1090"
)

/**
 *
 * The func of market unit test step is:
 * 1. Get the BTC_USDT ticker
 * 2. Get the BTC_USDT depth
 * 3. Get the BTC_USDT 1d 1m kline
 *
 **/

func TestSpot_MarketAPI(t *testing.T) {

	config := &APIConfig{
		Endpoint: ENDPOINT,
		HttpClient: &http.Client{
			Transport: &http.Transport{
				Proxy: func(req *http.Request) (*url.URL, error) {
					return url.Parse(PROXY_URL)
				},
			},
		},
		ApiKey:        SPOT_API_KEY,
		ApiSecretKey:  SPOT_API_SECRETKEY,
		ApiPassphrase: SPOT_API_PASSPHRASE,
		Location:      time.Now().Location(),
	}

	ok := New(config)
	// ticker unit test
	if ticker, resp, err := ok.Spot.GetTicker(
		CurrencyPair{BTC, USDT},
	); err != nil {
		t.Error(err)
		return
	} else {
		standard, err := json.Marshal(ticker)
		if err != nil {
			t.Error(err)
			return
		}

		t.Log("Ticker standard struct: ")
		t.Log(string(standard))
		t.Log("Ticker remote api response: ")
		t.Log(string(resp))
	}

	// depth unit test
	if depth, resp, err := ok.Spot.GetDepth(
		20,
		CurrencyPair{BTC, USDT},
	); err != nil {
		t.Error(err)
		return
	} else {
		standard, err := json.Marshal(depth)
		if err != nil {
			t.Error(err)
			return
		}

		t.Log("Depth standard struct:")
		t.Log(string(standard))
		t.Log("Depth remote api response: ")
		t.Log(string(resp))

		// make sure the later request get bigger sequence
		depth1, _, _ := ok.Spot.GetDepth(
			20,
			CurrencyPair{BTC, USDT},
		)

		if depth1.Sequence <= depth.Sequence {
			t.Error("later request get smaller sequence!!")
			return
		}

		if err := depth.Check(); err != nil {
			t.Error(err)
			return
		}

		if err := depth1.Check(); err != nil {
			t.Error(err)
			return
		}
	}

	// klines unit test
	if minKlines, resp, err := ok.Spot.GetKlineRecords(
		CurrencyPair{BTC, USDT},
		KLINE_PERIOD_1MIN,
		10,
		int(time.Now().Add(-24*time.Hour).UnixNano()),
	); err != nil {
		t.Error(err)
		return
	} else {
		standard, err := json.Marshal(minKlines)
		if err != nil {
			t.Error(err)
			return
		}
		t.Log("min kline standard struct:")
		t.Log(string(standard))
		t.Log("min kline remote api response: ")
		t.Log(string(resp))
	}

	if dayKlines, resp, err := ok.Spot.GetKlineRecords(
		CurrencyPair{BTC, USDT},
		KLINE_PERIOD_1DAY,
		10,
		int(time.Now().Add(-20*24*time.Hour).UnixNano()),
	); err != nil {
		t.Error(err)
		return
	} else {
		standard, err := json.Marshal(dayKlines)
		if err != nil {
			t.Error(err)
			return
		}
		t.Log("day kline standard struct:")
		t.Log(string(standard))
		t.Log("day kline remote api response: ")
		t.Log(string(resp))
	}
}

/**
 *
 * The func of order unit test step is:
 * 1. Get the account, and find have the enough crypto.
 * 2. Get BTC-USDT ticker.
 * 2. Order the Limit Sell/Buy without deal.
 * 3. Get the unfinished orders info, and find the order in step 1.
 * 4. Get the order info.
 * 5. Cancel the Limit Order
 *
 **/

func TestSpot_TradeAPI(t *testing.T) {
	config := &APIConfig{
		Endpoint: ENDPOINT,
		HttpClient: &http.Client{
			Transport: &http.Transport{
				Proxy: func(req *http.Request) (*url.URL, error) {
					return url.Parse(PROXY_URL)
				},
			},
		},
		ApiKey:        SPOT_API_KEY,
		ApiSecretKey:  SPOT_API_SECRETKEY,
		ApiPassphrase: SPOT_API_PASSPHRASE,
		Location:      time.Now().Location(),
	}

	ok := New(config)
	if account, resp, err := ok.Spot.GetAccount(); err != nil {
		t.Error(err)
		return
	} else {
		t.Log("Account standard struct:")
		t.Log(*account)
		t.Log("Account remote api response: ")
		t.Log(string(resp))

		for currency, subAccount := range account.SubAccounts {
			if currency == BTC && subAccount.Amount < 0.005 {
				t.Error("You don not has 0.005 BTC to order. ")
				return
			}
		}
	}

	testPrice := 0.0
	// ticker unit test
	if ticker, _, err := ok.Spot.GetTicker(
		CurrencyPair{BTC, USDT},
	); err != nil {
		t.Error(err)
		return
	} else {
		testPrice = ticker.Sell * 1.1
	}

	normalOrder := Order{
		Currency:  CurrencyPair{BTC, USDT},
		Price:     testPrice,
		Amount:    0.005,
		Side:      SELL,
		OrderType: NORMAL,
	}

	if resp, err := ok.Spot.LimitSell(&normalOrder); err != nil {
		t.Error(err)
		return
	} else {
		standard, err := json.Marshal(normalOrder)
		if err != nil {
			t.Error(err)
			return
		}
		t.Log("Order standard struct:")
		t.Log(string(standard))
		t.Log("Order remote api response: ")
		t.Log(string(resp))
	}

	for i := 0; i < 3; i++ {
		if resp, err := ok.Spot.GetOneOrder(&normalOrder); err != nil {
			t.Error(err)
			return
		} else if i == 0 {
			standard, err := json.Marshal(normalOrder)
			if err != nil {
				t.Error(err)
				return
			}
			t.Log("Order standard struct:")
			t.Log(string(standard))
			t.Log("Order remote api response: ")
			t.Log(string(resp))
		}
	}

	if orders, resp, err := ok.Spot.GetUnFinishOrders(
		CurrencyPair{BTC, USDT},
	); err != nil {
		t.Error(err)
		return
	} else {
		standard, err := json.Marshal(orders)
		if err != nil {
			t.Error(err)
			return
		}
		t.Log("UnFinished Order standard struct:")
		t.Log(string(standard))
		t.Log("UnFinished Order remote api response: ")
		t.Log(string(resp))

		isFind := false
		for _, order := range orders {
			if order.Cid == normalOrder.Cid && order.OrderId == normalOrder.OrderId {
				isFind = true
				break
			}
		}
		if !isFind {
			t.Error(errors.New("Can not find the order in unfinished orders! "))
			return
		}
	}

	if resp, err := ok.Spot.CancelOrder(&normalOrder); err != nil {
		t.Error(err)
		return
	} else {
		standard, err := json.Marshal(normalOrder)
		if err != nil {
			t.Error(err)
			return
		}

		t.Log("Cancel order standard struct:")
		t.Log(string(standard))
		t.Log("Cancel order remote api response: ")
		t.Log(string(resp))
	}

}
