package okex

import (
	"encoding/json"
	"net/http"
	"net/url"
	"testing"
	"time"

	. "github.com/deforceHK/goghostex"
)

// go test -v ./okex/... -count=1 -run=TestSwap_Account_Counter
func TestSwap_Account_Counter(t *testing.T) {

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

	var ok = New(config)
	if items, resp, err := ok.Swap.GetPairFlow(Pair{BTC, USDT}); err != nil {
		t.Error(err)
		t.Log(string(resp))
		return
	} else {
		var _, _ = json.Marshal(items)
		t.Log(string(resp))
		for _, i := range items {
			t.Log(*i)
		}
	}

	//if items, resp, err := ok.Swap.GetAccountFlow(); err != nil {
	//	t.Error(err)
	//	t.Log(string(resp))
	//	return
	//} else {
	//	var _, _ = json.Marshal(items)
	//	t.Log(string(resp))
	//	for _, i := range items {
	//		t.Log(*i)
	//	}
	//}
}

// go test -v ./okex/... -count=1 -run=TestSwap_Account_Basis
func TestSwap_Account_Basis(t *testing.T) {

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

	var ok = New(config)
	if items, resp, err := ok.Swap.GetPairFlow(Pair{BTC, USD}); err != nil {
		t.Error(err)
		t.Log(string(resp))
		return
	} else {
		var _, _ = json.Marshal(items)
		t.Log(string(resp))
		for _, i := range items {
			t.Log(*i)
		}
	}

	//if items, resp, err := ok.Swap.GetAccountFlow(); err != nil {
	//	t.Error(err)
	//	t.Log(string(resp))
	//	return
	//} else {
	//	var _, _ = json.Marshal(items)
	//	t.Log(string(resp))
	//	for _, i := range items {
	//		t.Log(*i)
	//	}
	//}
}
