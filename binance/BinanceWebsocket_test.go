package binance

import (
	"net/http"
	"testing"
	"time"

	. "github.com/deforceHK/goghostex"
)

// go test -v ./binance/... -count=1 -run=TestBinanceFutureWebsocket_Start
func TestBinanceFutureWebsocket_Start(t *testing.T) {

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

	var wsBN = &WSTradeUMBN{
		Config: config,
	}

	err := wsBN.Start()
	if err != nil {
		t.Error(err)
		return
	}

	wsBN.Subscribe("balance")

	time.Sleep(600 * time.Second)
}

// go test -v ./binance/... -count=1 -run=TestBinanceSwapMarketWebsocket
func TestBinanceSwapMarketWebsocket(t *testing.T) {
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

	var wsBN = &WSMarketUMBN{
		&WSTradeUMBN{
			Config: config,
		},
	}

	err := wsBN.Start()
	if err != nil {
		t.Error(err)
		return
	}
	time.Sleep(10 * time.Second)

	wsBN.Subscribe("solusdt@depth@500ms")
	select {}
}
