package kraken

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	. "github.com/deforceHK/goghostex"
)

// go test -v ./kraken/... -count=1 -run=TestLocalOrderBooks_Init
func TestLocalOrderBooks_Init(t *testing.T) {
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

	var book = &LocalOrderBooks{
		WSSwapMarketKK: &WSSwapMarketKK{
			//WSSwapTradeKK: &WSSwapTradeKK{
			Config: config,
			//},
		},
	}
	var err = book.Init()
	if err != nil {
		t.Error(err)
		return
	}

	var pair = NewPair("btc_usd", "_")
	book.Subscribe(pair)

	for i := 0; i < 100; i++ {
		time.Sleep(5 * time.Second)
		depth, depthErr := book.Snapshot(pair)
		if depthErr != nil {
			t.Error(depthErr)
			return
		}
		var depthData, _ = json.Marshal(depth)
		t.Log(string(depthData))
	}

	select {}
}
