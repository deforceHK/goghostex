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
 * go test -v ./okex/... -count=1 -run=TestFuture_MarketAPI
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
	//if ticker, body, err := ok.Future.GetTicker(
	//	Pair{Basis: BTC, Counter: USD},
	//	NEXT_QUARTER_CONTRACT,
	//); err != nil {
	//	t.Error(err)
	//	return
	//} else {
	//	standard, err := json.Marshal(ticker)
	//	if err != nil {
	//		t.Error(err)
	//		return
	//	}
	//
	//	t.Log("Ticker standard struct: ")
	//	t.Log(string(standard))
	//
	//	t.Log("Ticker remote api response: ")
	//	t.Log(string(body))
	//}
	//
	//if depth, body, err := ok.Future.GetDepth(
	//	NewPair("btc_usd", "_"),
	//	NEXT_QUARTER_CONTRACT,
	//	20,
	//); err != nil {
	//	t.Error(err)
	//	return
	//} else {
	//	standard, err := json.Marshal(depth)
	//	if err != nil {
	//		t.Error(err)
	//		return
	//	}
	//
	//	t.Log("Depth standard struct: ")
	//	t.Log(string(standard))
	//
	//	t.Log("Depth remote api response: ")
	//	t.Log(string(body))
	//
	//	if depth1, _, err := ok.Future.GetDepth(
	//		Pair{Basis: BTC, Counter: USD},
	//		NEXT_QUARTER_CONTRACT,
	//		20,
	//	); err != nil {
	//		t.Error(err)
	//		return
	//	} else {
	//		if depth1.Sequence < depth.Sequence {
	//			t.Error("The sequence not work. ")
	//			return
	//		}
	//
	//		if err := depth.Verify(); err != nil {
	//			t.Error(err)
	//			return
	//		}
	//
	//		if err := depth1.Verify(); err != nil {
	//			t.Error(err)
	//			return
	//		}
	//	}
	//}
	//
	//if highest, lowest, err := ok.Future.GetLimit(Pair{BTC, USD}, THIS_WEEK_CONTRACT); err != nil {
	//	t.Error(err)
	//} else {
	//	t.Log(highest, lowest)
	//}
	//
	//if minKline, body, err := ok.Future.GetKlineRecords(
	//	QUARTER_CONTRACT,
	//	Pair{Basis: BTC, Counter: USD},
	//	KLINE_PERIOD_1MIN,
	//	300,
	//	0,
	//	//int(time.Now().Add(-24*time.Hour).UnixNano()/1000000),
	//); err != nil {
	//	t.Error(err)
	//	return
	//} else {
	//	standard, err := json.Marshal(minKline)
	//	if err != nil {
	//		t.Error(err)
	//		return
	//	}
	//
	//	t.Log("minKline standard struct: ")
	//	t.Log(string(standard))
	//
	//	t.Log("minKline remote api response: ")
	//	t.Log(string(body))
	//
	//	for _, kline := range minKline {
	//		if kline.Timestamp < 1000000000000 {
	//			t.Error("The timestamp must be 13 number. ")
	//			return
	//		}
	//	}
	//}

	if minCandles, body, err := ok.Future.GetCandles(
		1735286400000,
		"btc_usd",
		KLINE_PERIOD_1MIN,
		300,
		0,
	); err != nil {
		t.Error(err)
		return
	} else {
		standard, err := json.Marshal(minCandles)
		if err != nil {
			t.Error(err)
			return
		}

		t.Log("minKline standard struct: ")
		t.Log(string(standard))

		t.Log("minKline remote api response: ")
		t.Log(string(body))

		for _, kline := range minCandles {
			if kline.Timestamp < 1000000000000 {
				t.Error("The timestamp must be 13 number. ")
				return
			}
		}
	}

	//if dayKline, body, err := ok.Future.GetKlineRecords(
	//	NEXT_QUARTER_CONTRACT,
	//	Pair{Basis: BTC, Counter: USD},
	//	KLINE_PERIOD_1DAY,
	//	20,
	//	//int(time.Now().Add(-20*24*time.Hour).UnixNano()/1000000),
	//	0,
	//); err != nil {
	//	t.Error(err)
	//	return
	//} else {
	//	standard, err := json.Marshal(dayKline)
	//	if err != nil {
	//		t.Error(err)
	//		return
	//	}
	//
	//	t.Log("dayKline standard struct: ")
	//	t.Log(string(standard))
	//
	//	t.Log("dayKline remote api response: ")
	//	t.Log(string(body))
	//
	//	for _, kline := range dayKline {
	//		if kline.Timestamp < 1000000000000 {
	//			t.Error("The timestamp must be 13 number. ")
	//			return
	//		}
	//	}
	//}

	//if index, resp, err := ok.Future.GetIndex(Pair{BTC, USDT}); err != nil {
	//	t.Error(err)
	//	return
	//} else {
	//
	//	t.Log("btc usdt index: ")
	//	t.Log(index)
	//
	//	t.Log("index remote api response: ")
	//	t.Log(string(resp))
	//}
	//
	//if index, resp, err := ok.Future.GetMark(Pair{BTC, USDT}, THIS_WEEK_CONTRACT); err != nil {
	//	t.Error(err)
	//	return
	//} else {
	//
	//	t.Log("btc usdt mark price: ")
	//	t.Log(index)
	//
	//	t.Log("mark price remote api response: ")
	//	t.Log(string(resp))
	//}

	//if Contract, err := ok.Future.GetContract(
	//	Pair{Basis: BTC, Counter: USDT},
	//	NEXT_WEEK_CONTRACT,
	//); err != nil {
	//	t.Error(err)
	//	return
	//} else {
	//	content, _ := json.Marshal(*Contract)
	//	t.Log(string(content))
	//}

	//ContractName := ok.Future.GetInstrumentId(
	//	Pair{Basis: BTC, Counter: USD},
	//	QUARTER_CONTRACT,
	//)
	//t.Log(ContractName)

	//if marketPrice, body, err := ok.Future.GetFutureMarkPrice(
	//	Pair{Basis: BTC, Counter: USD},
	//	THIS_WEEK_CONTRACT,
	//); err != nil {
	//	t.Log(marketPrice)
	//
	//	t.Log("the remote api response:")
	//	t.Log(string(body))
	//}

}

func TestFuture_ErrorAPI(t *testing.T) {

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
		raw, err := json.Marshal(account.SubAccount[BTC])
		if err != nil {
			t.Log(err)
			return
		}
		t.Log(string(raw))

		t.Log("Future account remote api struct: ")
		t.Log(string(resp))

		if account.SubAccount[BTC].BalanceAvail <= 0 {
			t.Error("You do not have enough BTC for test. ")
			return
		}
	}

	_, low, err := ok.Future.GetLimit(
		Pair{Basis: BTC, Counter: USD},
		THIS_WEEK_CONTRACT,
	)
	if err != nil {
		t.Error(err)
		return
	}

	orderShort := FutureOrder{
		Cid:          UUID(),
		Price:        low * 0.9,
		Amount:       1,
		PlaceType:    NORMAL,
		Type:         OPEN_SHORT,
		LeverRate:    20,
		Pair:         Pair{Basis: BTC, Counter: USD},
		ContractType: THIS_WEEK_CONTRACT,
		Exchange:     OKEX,
	}

	if resp, err := ok.Future.PlaceOrder(&orderShort); err != nil {
		t.Log(err)
		t.Log(string(resp))
		return
	} else {
		t.Error(errors.New("no error here, why? "))
		return
	}

}
