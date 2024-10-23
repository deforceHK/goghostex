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
	time.Sleep(1 * time.Second)

	var sol = Pair{SOL, USDT}
	var eth = Pair{ETH, USDT}
	wsBN.Subscribe(sol)
	wsBN.Subscribe(eth)

	for i := 0; i < 10; i++ {
		time.Sleep(5 * time.Second)
		if i%2 == 0 {
			depth, depthErr := wsBN.Snapshot(sol)
			if depthErr != nil {
				t.Error(depthErr)
				return
			}
			var depthData, _ = json.Marshal(depth)
			t.Log(string(depthData))
		} else {
			depth, depthErr := wsBN.Snapshot(eth)
			if depthErr != nil {
				t.Error(depthErr)
				return
			}
			var depthData, _ = json.Marshal(depth)
			t.Log(string(depthData))
		}
	}

	select {}
}
