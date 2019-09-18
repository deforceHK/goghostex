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
	FUTURE_API_KEY        = "a127cc13-2c21-4b19-9a3b-7be62ca8a6f1"
	FUTURE_API_SECRETKEY  = "B7318B036B1C5C37BEA45DC3B12AD804"
	FUTURE_API_PASSPHRASE = "strengthening"
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
	if ticker, body, err := ok.Future.GetFutureTicker(
		CurrencyPair{BTC, USD},
		THIS_WEEK_CONTRACT,
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

	if depth, body, err := ok.Future.GetFutureDepth(
		CurrencyPair{BTC, USD},
		THIS_WEEK_CONTRACT,
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

		if depth1, _, err := ok.Future.GetFutureDepth(
			CurrencyPair{BTC, USD},
			THIS_WEEK_CONTRACT,
			20,
		); err != nil {
			t.Error(err)
			return
		} else {
			if depth1.Sequence <= depth.Sequence {
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

	if minKline, body, err := ok.Future.GetFutureKlineRecords(
		THIS_WEEK_CONTRACT,
		CurrencyPair{BTC, USD},
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

	if dayKline, body, err := ok.Future.GetFutureKlineRecords(
		THIS_WEEK_CONTRACT,
		CurrencyPair{BTC, USD},
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

	if Contracts, body, err := ok.Future.GetFutureContractInfo(); err != nil {
		t.Error(err)
		return
	} else {
		t.Log(Contracts)

		t.Log("contract info remote api response: ")
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
	if account, resp, err := ok.Future.GetFutureAccount(); err != nil {
		t.Error(err)
		return
	} else {

		t.Log("Future account standard struct: ")
		t.Log(*account)

		t.Log("Future account remote api struct: ")
		t.Log(string(resp))

		if account.SubAccount[ETC].BalanceAvail <= 1 {
			t.Error("You do not have enough ETC for test. ")
			return
		}
	}

	//ticker, _, err := ok.Future.GetFutureTicker(
	//	CurrencyPair{ETC, USD},
	//	THIS_WEEK_CONTRACT,
	//)
	//if err != nil {
	//	t.Error(err)
	//	return
	//}
	//
	//order := FutureOrder{
	//	Cid:          UUID(),
	//	Price:        ticker.Last * 1.03,
	//	Amount:       1,
	//	PlaceType:    NORMAL,
	//	Type:         OPEN_SHORT,
	//	LeverRate:    20,
	//	Currency:     CurrencyPair{ETC, USD},
	//	ContractType: THIS_WEEK_CONTRACT,
	//	MatchPrice:   0,
	//}
	//preCid := order.Cid
	//
	//if resp, err := ok.Future.PlaceFutureOrder(&order); err != nil {
	//	t.Error(err)
	//	return
	//} else {
	//	if standard, err := json.Marshal(order); err != nil {
	//		t.Error(err)
	//		return
	//	} else {
	//		t.Log("Place Order standard struct: ")
	//		t.Log(string(standard))
	//
	//		t.Log("Place Order remote api struct: ")
	//		t.Log(string(resp))
	//	}
	//
	//	if preCid != order.Cid {
	//		t.Error("The cid is not same in the api. ")
	//		return
	//	}
	//}
	//
	//if resp, err := ok.Future.GetFutureOrder(&order); err != nil {
	//	t.Error(err)
	//	return
	//} else {
	//
	//	if standard, err := json.Marshal(order); err != nil {
	//		t.Error(err)
	//		return
	//	} else {
	//		t.Log("Get Order standard struct: ")
	//		t.Log(string(standard))
	//
	//		t.Log("Get Order remote api struct: ")
	//		t.Log(string(resp))
	//	}
	//}
	//
	//if resp, err := ok.Future.CancelFutureOrder(&order); err != nil {
	//	t.Error(err)
	//	return
	//} else {
	//	if standard, err := json.Marshal(order); err != nil {
	//		t.Error(err)
	//		return
	//	} else {
	//		t.Log("Cancel Order standard struct: ")
	//		t.Log(string(standard))
	//
	//		t.Log("Cancel Order remote api struct: ")
	//		t.Log(string(resp))
	//	}
	//}
	//
	//if resp, err := ok.Future.GetFutureOrder(&order); err != nil {
	//	t.Error(err)
	//	return
	//} else {
	//
	//	if standard, err := json.Marshal(order); err != nil {
	//		t.Error(err)
	//		return
	//	} else {
	//		t.Log("Get Order after standard struct: ")
	//		t.Log(string(standard))
	//
	//		t.Log("Get Order after remote api struct: ")
	//		t.Log(string(resp))
	//	}
	//
	//	if order.Status != ORDER_CANCEL {
	//		t.Error("The order must bi canceled. ")
	//		return
	//	}
	//}
	//
	//onlyMakerOrder := FutureOrder{
	//	Cid:          UUID(),
	//	Price:        ticker.Last * 0.99,
	//	Amount:       1,
	//	PlaceType:    ONLY_MAKER,
	//	Type:         OPEN_SHORT,
	//	LeverRate:    20,
	//	Currency:     CurrencyPair{ETC, USD},
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
	//if resp, err := ok.Future.GetFutureOrder(&onlyMakerOrder); err != nil {
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
