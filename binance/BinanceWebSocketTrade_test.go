package binance

import (
	"net/http"
	"testing"
	"time"

	. "github.com/deforceHK/goghostex"
)

// go test -v ./binance/... -count=1 -run=TestBinanceWebsocketTrade_Start
func TestBinanceWebsocketTrade_Start(t *testing.T) {

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

	time.Sleep(5 * time.Second)

	var params = WSParamsBN{
		Id:     UUID(),
		Method: "order.place",
		Params: map[string]interface{}{
			"positionSide": "LONG",
			"price":        98999,
			"quantity":     0.002,
			"side":         "BUY",
			"symbol":       "BTCUSDT",
			"timeInForce":  "GTC",
			"type":         "LIMIT",
		},
	}
	if err = wsBN.Write(params); err != nil {
		t.Error(err)
		return
	}

	time.Sleep(600 * time.Second)
}
