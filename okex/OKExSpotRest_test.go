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
	API_KEY       = ""
	API_SECRETKEY = ""
	PROXY_URL     = "socks5://127.0.0.1:1090"
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
		ApiKey:        API_KEY,
		ApiSecretKey:  API_SECRETKEY,
		ApiPassphrase: "",
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
		int(time.Now().Add(-2*24*time.Hour).UnixNano()),
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
}
