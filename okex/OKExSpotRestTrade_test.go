package okex

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"testing"
	"time"

	. "github.com/deforceHK/goghostex"
)

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
			if currency == BTC.Symbol && subAccount.Amount < 0.005 {
				t.Error("You don not has 0.005 BTC to order. ")
				return
			}
		}
	}

	testPrice := 0.0
	// ticker unit test
	if ticker, _, err := ok.Spot.GetTicker(
		Pair{Basis: BTC, Counter: USDT},
	); err != nil {
		t.Error(err)
		return
	} else {
		testPrice = ticker.Sell * 1.1
	}

	normalOrder := Order{
		Pair:      Pair{Basis: BTC, Counter: USDT},
		Price:     testPrice,
		Amount:    0.005,
		Side:      SELL,
		OrderType: NORMAL,
	}

	if resp, err := ok.Spot.PlaceOrder(&normalOrder); err != nil {
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
		if resp, err := ok.Spot.GetOrder(&normalOrder); err != nil {
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
		Pair{Basis: BTC, Counter: USDT},
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
