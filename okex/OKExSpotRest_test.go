package okex

import (
	"encoding/json"
	"net/http"
	"net/url"
	"testing"
	"time"

	. "github.com/deforceHK/goghostex"
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
	if depth, resp, err := ok.Spot.GetDepth(
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
		depth1, _, _ := ok.Spot.GetDepth(
			Pair{Basis: BTC, Counter: USDT},
			20,
		)

		if depth1.Sequence <= depth.Sequence {
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
	if minKlines, resp, err := ok.Spot.GetKlineRecords(
		Pair{Basis: BTC, Counter: USDT},
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

		for _, kline := range minKlines {
			if kline.Timestamp < 1000000000000 {
				t.Error("The timestamp must be 13 number. ")
				return
			}
		}
	}

	if dayKlines, resp, err := ok.Spot.GetKlineRecords(
		Pair{Basis: BTC, Counter: USDT},
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

		for _, kline := range dayKlines {
			if kline.Timestamp < 1000000000000 {
				t.Error("The timestamp must be 13 number. ")
				return
			}
		}
	}
}

/**
* unit test cmd
* go test -v ./okex/... -count=1 -run=TestSpot_InstrumentsAPI
*
**/
func TestSpot_InstrumentsAPI(t *testing.T) {

	var config = &APIConfig{
		Endpoint:   ENDPOINT,
		HttpClient: &http.Client{
			//Transport: &http.Transport{
			//	Proxy: func(req *http.Request) (*url.URL, error) {
			//		return url.Parse(PROXY_URL)
			//	},
			//},
		},
		ApiKey:        SPOT_API_KEY,
		ApiSecretKey:  SPOT_API_SECRETKEY,
		ApiPassphrase: SPOT_API_PASSPHRASE,
		Location:      time.Now().Location(),
	}

	var ok = New(config)

	var ins = ok.Spot.GetInstruments(BTC_USDT)
	var result, err = json.Marshal(ins)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(string(result))
}
