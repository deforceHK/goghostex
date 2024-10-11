package binance

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	. "github.com/deforceHK/goghostex"
)

// go test -v ./binance/... -count=1 -run=TestBinanceWebsocketBook
func TestBinanceWebsocketBook(t *testing.T) {
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

	var wsBN = LocalOrderBooks{
		WSMarketUMBN: &WSMarketUMBN{
			&WSTradeUMBN{
				Config: config,
			},
		},
	}

	err := wsBN.Init()
	if err != nil {
		t.Error(err)
		return
	}
	time.Sleep(1 * time.Second)

	wsBN.Subscribe("solusdt@depth")

	for i := 0; i < 10; i++ {
		time.Sleep(10 * time.Second)
		depth, depthErr := wsBN.Snapshot("solusdt")
		if depthErr != nil {
			t.Error(depthErr)
			return
		}
		var depthData,_ = json.Marshal(depth)
		t.Log(string(depthData))
	}

	select {}
}
