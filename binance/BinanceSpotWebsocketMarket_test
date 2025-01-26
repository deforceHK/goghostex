
package binance

import (
	"net/http"
	"testing"
	"time"

	. "github.com/deforceHK/goghostex"
)

// go test -v ./binance/... -count=1 -run=TestBinancSpoteWebsocketMarket
func TestBinancSpoteWebsocketMarket(t *testing.T) {
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

	var wsBN = &WSMarketSpot{
		Config: config,
	}


	err := wsBN.Start()
	if err != nil {
		t.Error(err)
		return
	}
	time.Sleep(10 * time.Second)

	wsBN.Subscribe("ethusdt_250627@depth")
	select {}
}