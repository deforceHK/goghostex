package binance

import (
	"encoding/json"
	"net/http"
	"net/url"
	"testing"
	"time"

	. "github.com/deforceHK/goghostex"
)

const (
	API_KEY       = ""
	API_SECRETKEY = ""
	PROXY_URL     = "socks5://127.0.0.1:1090"
)

/**
 *
 * The func of market unit test step is:
 * 1. Get the BNBBTC ticker
 * 2. Get the BNBBTC depth
 * 3. Get the BNBBTC 1d 1m kline
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
		ApiKey:        API_KEY,
		ApiSecretKey:  API_SECRETKEY,
		ApiPassphrase: "",
		Location:      time.Now().Location(),
	}
	bn := New(config)

	// ticker unit test
	if ticker, resp, err := bn.Spot.GetTicker(
		Pair{Basis: BTC, Counter: USDT},
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
	if depth, resp, err := bn.Spot.GetDepth(
		Pair{Basis: BTC, Counter: USDT},
		20,
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
		depth1, _, _ := bn.Spot.GetDepth(
			Pair{Basis: BTC, Counter: USDT},
			20,
		)

		if depth1.Sequence < depth.Sequence {
			t.Error("later request get smaller sequence!!")
			return
		}

		if err := depth.Verify(); err != nil {
			t.Error(err)
			return
		}

		if err := depth1.Verify(); err != nil {
			t.Error(err)
			return
		}
	}

	// klines unit test
	if minKlines, resp, err := bn.Spot.GetKlineRecords(
		Pair{Basis: BTC, Counter: USDT},
		KLINE_PERIOD_1MIN,
		10,
		int(time.Now().Add(-2*24*time.Hour).UnixNano()),
		//int(time.Now().Add(-1*time.Hour).UnixNano()),
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

	if dayKlines, resp, err := bn.Spot.GetKlineRecords(
		Pair{Basis: BTC, Counter: USDT},
		KLINE_PERIOD_1DAY,
		10,
		int(time.Now().Add(-11*24*time.Hour).UnixNano()),
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

	if body, err := bn.Spot.ExchangeInfo(); err != nil {
		t.Log(err)
		return
	} else {
		info := map[string]json.RawMessage{}
		if err := json.Unmarshal(body, &info); err != nil {
			panic(err)
		} else {
			for key := range info {
				t.Log(key)
			}
			t.Log(string(info["rateLimits"]))
		}
	}
}

/**
 *
 * The func of order unit test step is:
 * 1. Get BNBBTC ticker.
 * 2. Get the account, and find have the enough crypto.
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
		ApiKey:        API_KEY,
		ApiSecretKey:  API_SECRETKEY,
		ApiPassphrase: "",
		Location:      time.Now().Location(),
	}

	bn := New(config)
	if account, _, err := bn.Spot.GetAccount(); err != nil {
		t.Error(err)
		return
	} else {
		for currency, subAccount := range account.SubAccounts {
			if currency == BNB.Symbol && subAccount.Amount < 1 {
				t.Error("You don not has 1 BNB to order. ")
				return
			}
		}
	}

	testPrice := 0.0
	// ticker unit test
	if ticker, _, err := bn.Spot.GetTicker(
		Pair{Basis: BNB, Counter: BTC},
	); err != nil {
		t.Error(err)
		return
	} else {
		testPrice = ticker.Sell * 1.1
	}

	normalOrder := Order{
		Pair:      Pair{Basis: BNB, Counter: BTC},
		Price:     testPrice,
		Amount:    1,
		Side:      SELL,
		OrderType: NORMAL,
	}

	if resp, err := bn.Spot.PlaceOrder(&normalOrder); err != nil {
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
		if resp, err := bn.Spot.GetOrder(&normalOrder); err != nil {
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

	if orders, resp, err := bn.Spot.GetUnFinishOrders(
		Pair{Basis: BNB, Counter: BTC},
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
	}

	if resp, err := bn.Spot.CancelOrder(&normalOrder); err != nil {
		t.Error(err)
		return
	} else {
		standard, err := json.Marshal(normalOrder)
		if err != nil {
			t.Error(err)
			return
		}
		t.Log("Cancel Order standard struct:")
		t.Log(string(standard))
		t.Log("Cancel Order remote api response: ")
		t.Log(string(resp))
	}

	fokOrder := Order{
		Pair:      Pair{Basis: BNB, Counter: BTC},
		Price:     testPrice,
		Amount:    1,
		Side:      SELL,
		OrderType: FOK,
	}

	if resp, err := bn.Spot.PlaceOrder(&fokOrder); err != nil {
		t.Error(err)
		return
	} else {
		standard, err := json.Marshal(fokOrder)
		if err != nil {
			t.Error(err)
			return
		}
		t.Log("FOK Order standard struct:")
		t.Log(string(standard))
		t.Log("FOK Order remote api response: ")
		t.Log(string(resp))
	}

	iocOrder := Order{
		Pair:      Pair{Basis: BNB, Counter: BTC},
		Price:     testPrice,
		Amount:    1,
		Side:      SELL,
		OrderType: IOC,
	}

	if resp, err := bn.Spot.PlaceOrder(&iocOrder); err != nil {
		t.Error(err)
		return
	} else {
		standard, err := json.Marshal(iocOrder)
		if err != nil {
			t.Error(err)
			return
		}
		t.Log("IOC Order standard struct:")
		t.Log(string(standard))
		t.Log("IOC Order remote api response: ")
		t.Log(string(resp))
	}

}
