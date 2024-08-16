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
	FUTURE_API_KEY        = ""
	FUTURE_API_SECRETKEY  = ""
	FUTURE_API_PASSPHRASE = ""
	FUTURE_PROXY_URL      = "socks5://127.0.0.1:1090"
)

/**
* unit test cmd
* go test -v ./binance/... -count=1 -run=TestFuture_MarketAPI
*
**/

func TestFuture_MarketAPI(t *testing.T) {

	var client = &http.Client{}
	if FUTURE_PROXY_URL != "" {
		client = &http.Client{
			Transport: &http.Transport{
				Proxy: func(req *http.Request) (*url.URL, error) {
					return url.Parse(FUTURE_PROXY_URL)
				},
			},
		}
	}

	config := &APIConfig{
		Endpoint:      ENDPOINT,
		HttpClient:    client,
		ApiKey:        FUTURE_API_KEY,
		ApiSecretKey:  FUTURE_API_SECRETKEY,
		ApiPassphrase: FUTURE_API_PASSPHRASE,
		Location:      time.Now().Location(),
	}

	bn := New(config)
	// ticker unit test
	if ticker, resp, err := bn.Future.GetTicker(Pair{Basis: BTC, Counter: USD}, NEXT_QUARTER_CONTRACT); err != nil {
		t.Error(err)
		return
	} else {
		t.Log(ticker)
		t.Log(string(resp))
	}

	if depth, resp, err := bn.Future.GetDepth(Pair{Basis: BTC, Counter: USD}, NEXT_QUARTER_CONTRACT, 5); err != nil {
		t.Error(err)
		return
	} else {
		depthRaw, _ := json.Marshal(depth)
		t.Log(string(depthRaw))
		t.Log(string(resp))
	}

	if max, min, err := bn.Future.GetLimit(Pair{Basis: BTC, Counter: USD}, NEXT_QUARTER_CONTRACT); err != nil {
		t.Error(err)
		return
	} else {
		t.Log(max, min)
	}

	if klineList, resp, err := bn.Future.GetKlineRecords(
		NEXT_QUARTER_CONTRACT,
		Pair{Basis: ETH, Counter: USD},
		KLINE_PERIOD_1MIN,
		200,
		1609430400000,
	); err != nil {
		t.Error(err)
		return
	} else {
		raw, _ := json.Marshal(klineList)
		t.Log(string(raw))
		t.Log(string(resp))
	}

	if trades, _, err := bn.Future.GetTrades(
		Pair{Basis: BTC, Counter: USD},
		NEXT_QUARTER_CONTRACT,
	); err != nil {
		t.Error(err)
		return
	} else {
		stdTrades, _ := json.Marshal(trades)
		t.Log(string(stdTrades))
	}

	if contract, err := bn.Future.GetContract(Pair{ETH, USD}, NEXT_QUARTER_CONTRACT); err != nil {
		t.Error(err)
		return
	} else {
		t.Log(contract)
	}

	if candles, _, err := bn.Future.GetCandles(
		1735286400000, "btc_usd",
		KLINE_PERIOD_1MIN, 300, 0,
	); err != nil {
		t.Error(err)
		return
	} else {
		for _, candle := range candles {
			t.Log(candle)
		}
	}

}
