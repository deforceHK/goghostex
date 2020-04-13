package binance

import (
	"encoding/json"
	"net/http"
	"net/url"
	"testing"
	"time"

	. "github.com/strengthening/goghostex"
)

const (
	SWAP_API_KEY       = ""
	SWAP_API_SECRETKEY = ""
	SWAP_PROXY_URL     = "socks5://127.0.0.1:1090"
)

func TestSwap_MarketAPI(t *testing.T) {

	config := &APIConfig{
		Endpoint: ENDPOINT,
		HttpClient: &http.Client{
			Transport: &http.Transport{
				Proxy: func(req *http.Request) (*url.URL, error) {
					return url.Parse(SWAP_PROXY_URL)
				},
			},
		},
		ApiKey:        SWAP_API_KEY,
		ApiSecretKey:  SWAP_API_SECRETKEY,
		ApiPassphrase: "",
		Location:      time.Now().Location(),
	}

	bn := New(config)
	// ticker unit test
	if ticker, resp, err := bn.Swap.GetTicker(
		Pair{BTC, USDT},
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
	if depth, resp, err := bn.Swap.GetDepth(
		Pair{BTC, USDT},
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
		depth1, _, _ := bn.Swap.GetDepth(
			Pair{BTC, USDT},
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

	if highest, lowest, err := bn.Swap.GetLimit(Pair{BTC, USDT}); err != nil {
		t.Error(err)
		return
	} else {
		t.Log("highest:", highest)
		t.Log("lowest:", lowest)
	}

	if klines, resp, err := bn.Swap.GetKline(
		Pair{BTC, USDT},
		KLINE_PERIOD_1DAY,
		20,
		1271149752000,
	); err != nil {
		t.Error(err)
		return
	} else {
		klineRaw, _ := json.Marshal(klines)
		t.Log(string(klineRaw))
		t.Log(string(resp))
	}
}
