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

	var wsBN = &LocalOrderBooks{
		WSMarketUMBN: &WSMarketUMBN{
			Config: config,
		},
	}

	err := wsBN.Init()
	if err != nil {
		t.Error(err)
		return
	}

	var ethSwap = "ethusdt"
	var ethfuture = "ethusdt_250627"
	//wsBN.Subscribe(sol)
	wsBN.SubscribeByProductId(ethSwap)
	wsBN.SubscribeByProductId(ethfuture)

	time.Sleep(10 * time.Second)

	for i := 0; i < 10; i++ {
		time.Sleep(5 * time.Second)
		if i%2 == 0 {
			depth, depthErr := wsBN.SnapshotById(ethfuture)
			if depthErr != nil {
				t.Error(depthErr)
				return
			}
			var depthData, _ = json.Marshal(depth)
			t.Log(string(depthData))
			break
		} else {
			depth, depthErr := wsBN.SnapshotById(ethSwap)
			if depthErr != nil {
				t.Error(depthErr)
				return
			}
			var _, _ = json.Marshal(depth)
			//t.Log(string(depthData))
		}
	}

	select {}
}
