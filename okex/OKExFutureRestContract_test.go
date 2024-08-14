package okex

import (
	"net/http"
	"net/url"
	"testing"
	"time"

	. "github.com/deforceHK/goghostex"
)

// go test -v ./okex/... -count=1 -run=TestFuture_GetContracts
func TestFuture_GetContracts(t *testing.T) {

	var config = &APIConfig{
		Endpoint: ENDPOINT,
		HttpClient: &http.Client{
			Transport: &http.Transport{
				Proxy: func(req *http.Request) (*url.URL, error) {
					return url.Parse("socks5://127.0.0.1:1090")
				},
			},
		},
		ApiKey:        FUTURE_API_KEY,
		ApiSecretKey:  FUTURE_API_SECRETKEY,
		ApiPassphrase: FUTURE_API_PASSPHRASE,
		Location:      time.Now().Location(),
	}

	var okex = New(config)

	if contracts, _, err := okex.Future.GetContracts(); err != nil {
		t.Error(err)
		return
	} else {
		for _, contract := range contracts {
			if contract.Symbol == "btc_usd" || contract.Symbol == "btc_usdt" ||
				contract.Symbol == "eth_usd" || contract.Symbol == "eth_usdt" {
				t.Log(contract)
			}
		}

		//t.Log(string(resp))
	}
}
