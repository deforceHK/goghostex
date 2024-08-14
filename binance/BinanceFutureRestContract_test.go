package binance

import (
	"net/http"
	"net/url"
	"testing"
	"time"

	. "github.com/deforceHK/goghostex"
)

// go test -v ./binance/... -count=1 -run=TestFuture_GetContracts
func TestFuture_GetContracts(t *testing.T) {
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

	var config = &APIConfig{
		Endpoint:      ENDPOINT,
		HttpClient:    client,
		ApiKey:        FUTURE_API_KEY,
		ApiSecretKey:  FUTURE_API_SECRETKEY,
		ApiPassphrase: FUTURE_API_PASSPHRASE,
		Location:      time.Now().Location(),
	}

	var bn = New(config)

	if contracts, _, err := bn.Future.GetContracts(); err != nil {
		t.Error(err)
		return
	} else {
		for _, contract := range contracts {
			if contract.Symbol == "btc_usd" || contract.Symbol == "btc_usdt" ||
				contract.Symbol == "eth_usd" || contract.Symbol == "eth_usdt" {
				t.Log(contract)
				t.Log(contract.RawData)
			}
		}

		//t.Log(string(resp))
	}
}
