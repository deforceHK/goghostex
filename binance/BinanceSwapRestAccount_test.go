package binance

import (
	"net/http"
	"net/url"
	"testing"
	"time"

	. "github.com/deforceHK/goghostex"
)

// go test -v ./binance/... -count=1 -run=TestSwap_Account_Flow
func TestSwap_Account_Flow(t *testing.T) {

	var config = &APIConfig{
		Endpoint: ENDPOINT,
		HttpClient: &http.Client{
			Transport: &http.Transport{
				Proxy: func(req *http.Request) (*url.URL, error) {
					return url.Parse(PROXY_URL)
				},
			},
		},
		ApiKey:        SWAP_API_KEY,
		ApiSecretKey:  SWAP_API_SECRETKEY,
		ApiPassphrase: SWAP_API_PASSPHRASE,
		Location:      time.Now().Location(),
	}

	var bn = New(config)
	if items, raw, err := bn.Swap.GetAccountFlow(); err != nil {
		t.Error(err)
		return
	} else {
		t.Log(string(raw))
		for _, i := range items {
			t.Log(*i)
		}
	}
}

// go test -v ./binance/... -count=1 -run=TestSwap_Pair_Flow_Counter
func TestSwap_Pair_Flow_Counter(t *testing.T) {

	var config = &APIConfig{
		Endpoint:   ENDPOINT,
		HttpClient: &http.Client{
			//Transport: &http.Transport{
			//	Proxy: func(req *http.Request) (*url.URL, error) {
			//		return url.Parse(PROXY_URL)
			//	},
			//},
		},
		ApiKey:        SWAP_API_KEY,
		ApiSecretKey:  SWAP_API_SECRETKEY,
		ApiPassphrase: SWAP_API_PASSPHRASE,
		Location:      time.Now().Location(),
	}

	var bn = New(config)
	if items, raw, err := bn.Swap.GetPairFlow(Pair{SOL, USDT}); err != nil {
		t.Error(err)
		return
	} else {
		t.Log(string(raw))
		for _, i := range items {
			t.Log(*i)
		}
	}
}

// go test -v ./binance/... -count=1 -run=TestSwap_Pair_Flow_Basis
func TestSwap_Pair_Flow_Basis(t *testing.T) {

	var config = &APIConfig{
		Endpoint: ENDPOINT,
		HttpClient: &http.Client{
			Transport: &http.Transport{
				Proxy: func(req *http.Request) (*url.URL, error) {
					return url.Parse(PROXY_URL)
				},
			},
		},
		ApiKey:        SWAP_API_KEY,
		ApiSecretKey:  SWAP_API_SECRETKEY,
		ApiPassphrase: SWAP_API_PASSPHRASE,
		Location:      time.Now().Location(),
	}

	var bn = New(config)
	if items, raw, err := bn.Swap.GetPairFlow(Pair{ETH, USD}); err != nil {
		t.Error(err)
		return
	} else {
		t.Log(string(raw))
		for _, i := range items {
			t.Log(*i)
		}
	}
}
