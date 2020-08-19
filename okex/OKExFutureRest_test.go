package okex

import (
	"encoding/json"
	"net/http"
	"net/url"
	"testing"
	"time"

	. "github.com/strengthening/goghostex"
)

const (
	FUTURE_API_KEY        = ""
	FUTURE_API_SECRETKEY  = ""
	FUTURE_API_PASSPHRASE = ""
)

/**
 *
 * The func of market unit test step is:
 * 1. Get the BTC_USD this_week ticker
 * 2. Get the BTC_USD this_week depth and the later depth's sequence is bigger
 * 3. Get the BTC_USD this_week 1d 1m kline
 *
 **/

func TestFuture_MarketAPI(t *testing.T) {

	config := &APIConfig{
		Endpoint: ENDPOINT,
		HttpClient: &http.Client{
			Transport: &http.Transport{
				Proxy: func(req *http.Request) (*url.URL, error) {
					return url.Parse(PROXY_URL)
				},
			},
		},
		ApiKey:        FUTURE_API_KEY,
		ApiSecretKey:  FUTURE_API_SECRETKEY,
		ApiPassphrase: FUTURE_API_PASSPHRASE,
		Location:      time.Now().Location(),
	}

	ok := New(config)
	if ticker, body, err := ok.Future.GetTicker(
		Pair{Basis: BTC, Counter: USD},
		NEXT_QUARTER_CONTRACT,
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
		t.Log(string(body))
	}

	if depth, body, err := ok.Future.GetDepth(
		NewPair("btc_usd", "_"),
		NEXT_QUARTER_CONTRACT,
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

		t.Log("Depth standard struct: ")
		t.Log(string(standard))

		t.Log("Depth remote api response: ")
		t.Log(string(body))

		if depth1, _, err := ok.Future.GetDepth(
			Pair{Basis: BTC, Counter: USD},
			NEXT_QUARTER_CONTRACT,
			20,
		); err != nil {
			t.Error(err)
			return
		} else {
			if depth1.Sequence < depth.Sequence {
				t.Error("The sequence not work. ")
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
	}

	if minKline, body, err := ok.Future.GetKlineRecords(
		NEXT_QUARTER_CONTRACT,
		Pair{Basis: BTC, Counter: USD},
		KLINE_PERIOD_1MIN,
		20,
		int(time.Now().Add(-24*time.Hour).UnixNano()/1000000),
	); err != nil {
		t.Error(err)
		return
	} else {
		standard, err := json.Marshal(minKline)
		if err != nil {
			t.Error(err)
			return
		}

		t.Log("minKline standard struct: ")
		t.Log(string(standard))

		t.Log("minKline remote api response: ")
		t.Log(string(body))

		for _, kline := range minKline {
			if kline.Timestamp < 1000000000000 {
				t.Error("The timestamp must be 13 number. ")
				return
			}
		}
	}

	if dayKline, body, err := ok.Future.GetKlineRecords(
		NEXT_QUARTER_CONTRACT,
		Pair{Basis: BTC, Counter: USD},
		KLINE_PERIOD_1DAY,
		20,
		int(time.Now().Add(-20*24*time.Hour).UnixNano()/1000000),
	); err != nil {
		t.Error(err)
		return
	} else {
		standard, err := json.Marshal(dayKline)
		if err != nil {
			t.Error(err)
			return
		}

		t.Log("dayKline standard struct: ")
		t.Log(string(standard))

		t.Log("dayKline remote api response: ")
		t.Log(string(body))

		for _, kline := range dayKline {
			if kline.Timestamp < 1000000000000 {
				t.Error("The timestamp must be 13 number. ")
				return
			}
		}
	}

	Contracts := ok.Future.GetFutureContract(
		Pair{Basis: BTC, Counter: USD},
		NEXT_QUARTER_CONTRACT,
	)
	t.Log(Contracts)

	if marketPrice, body, err := ok.Future.GetFutureMarkPrice(
		Pair{Basis: BTC, Counter: USD},
		THIS_WEEK_CONTRACT,
	); err != nil {
		t.Log(marketPrice)

		t.Log("the remote api response:")
		t.Log(string(body))
	}

}

/**
 *
 * The func of order unit test step is:
 * 1. Get the account, and find have the enough crypto.
 * 2. Get BTC-USD this_week ticker.
 * 2. Order the open_long without deal.
 * 3. Get the unfinished orders info, and find the order in step 1.
 * 4. Get the order info.
 * 5. Cancel the open_long Order
 *
 **/

func TestFuture_TradeAPI(t *testing.T) {

	config := &APIConfig{
		Endpoint: ENDPOINT,
		HttpClient: &http.Client{
			Transport: &http.Transport{
				Proxy: func(req *http.Request) (*url.URL, error) {
					return url.Parse(PROXY_URL)
				},
			},
		},
		ApiKey:        FUTURE_API_KEY,
		ApiSecretKey:  FUTURE_API_SECRETKEY,
		ApiPassphrase: FUTURE_API_PASSPHRASE,
		Location:      time.Now().Location(),
	}

	ok := New(config)
	if account, resp, err := ok.Future.GetAccount(); err != nil {
		t.Error(err)
		return
	} else {

		t.Log("Future account standard struct: ")
		t.Log(*account)

		t.Log("Future account remote api struct: ")
		t.Log(string(resp))

		if account.SubAccount[LTC].BalanceAvail <= 1 {
			t.Error("You do not have enough LTC for test. ")
			return
		}
	}

	ticker, _, err := ok.Future.GetTicker(
		Pair{Basis: LTC, Counter: USD},
		THIS_WEEK_CONTRACT,
	)
	if err != nil {
		t.Error(err)
		return
	}

	order := FutureOrder{
		Cid:          UUID(),
		Price:        ticker.Last * 1.03,
		Amount:       1,
		PlaceType:    NORMAL,
		Type:         OPEN_SHORT,
		LeverRate:    20,
		Pair:         Pair{Basis: LTC, Counter: USD},
		ContractType: THIS_WEEK_CONTRACT,
		MatchPrice:   0,
	}
	preCid := order.Cid

	if resp, err := ok.Future.PlaceOrder(&order); err != nil {
		t.Error(err)
		return
	} else {
		if standard, err := json.Marshal(order); err != nil {
			t.Error(err)
			return
		} else {
			t.Log("Place Order standard struct: ")
			t.Log(string(standard))

			t.Log("Place Order remote api struct: ")
			t.Log(string(resp))
		}

		if preCid != order.Cid {
			t.Error("The cid is not same in the api. ")
			return
		}

		if order.OrderId == "" {
			t.Error("The order_id can not be empty string. ")
			return
		}
	}

	if resp, err := ok.Future.GetOrder(&order); err != nil {
		t.Error(err)
		return
	} else {

		if standard, err := json.Marshal(order); err != nil {
			t.Error(err)
			return
		} else {
			t.Log("Get Order standard struct: ")
			t.Log(string(standard))

			t.Log("Get Order remote api struct: ")
			t.Log(string(resp))
		}
	}

	if resp, err := ok.Future.CancelOrder(&order); err != nil {
		t.Error(err)
		return
	} else {
		if standard, err := json.Marshal(order); err != nil {
			t.Error(err)
			return
		} else {
			t.Log("Cancel Order standard struct: ")
			t.Log(string(standard))

			t.Log("Cancel Order remote api struct: ")
			t.Log(string(resp))
		}
	}

	for i := 0; i < 3; i++ {
		if resp, err := ok.Future.GetOrder(&order); err != nil {
			t.Error(err)
			return
		} else {

			if standard, err := json.Marshal(order); err != nil {
				t.Error(err)
				return
			} else {
				t.Log("Get Order after standard struct: ")
				t.Log(string(standard))

				t.Log("Get Order after remote api struct: ")
				t.Log(string(resp))
			}

			if order.Status == ORDER_CANCEL {
				break
			}
		}
	}

	if order.Status != ORDER_CANCEL {
		t.Error("The order must be canceled. ")
		return
	}

	//onlyMakerOrder := FutureOrder{
	//	Cid:          UUID(),
	//	Price:        ticker.Last * 0.99,
	//	Amount:       1,
	//	PlaceType:    ONLY_MAKER,
	//	Type:         OPEN_SHORT,
	//	LeverRate:    20,
	//	Pair:     Pair{LTC, USD},
	//	ContractType: THIS_WEEK_CONTRACT,
	//	MatchPrice:   0,
	//}
	//
	//if resp, err := ok.Future.PlaceFutureOrder(&onlyMakerOrder); err != nil {
	//	t.Error(err)
	//	return
	//} else {
	//
	//	if standard, err := json.Marshal(onlyMakerOrder); err != nil {
	//		t.Error(err)
	//		return
	//	} else {
	//		t.Log("Place only maker Order standard struct: ")
	//		t.Log(string(standard))
	//
	//		t.Log("Place only maker Order remote api struct: ")
	//		t.Log(string(resp))
	//	}
	//}
	//
	//if resp, err := ok.Future.GetOrder(&onlyMakerOrder); err != nil {
	//	t.Error(err)
	//	return
	//} else {
	//
	//	if standard, err := json.Marshal(onlyMakerOrder); err != nil {
	//		t.Error(err)
	//		return
	//	} else {
	//		t.Log("Get only maker Order after standard struct: ")
	//		t.Log(string(standard))
	//
	//		t.Log("Get only maker Order after remote api struct: ")
	//		t.Log(string(resp))
	//	}
	//
	//	if onlyMakerOrder.Status != ORDER_CANCEL {
	//		t.Error("The only maker order must bi canceled. ")
	//		return
	//	}
	//}

}
