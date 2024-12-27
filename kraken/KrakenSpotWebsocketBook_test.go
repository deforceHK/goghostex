package kraken

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	. "github.com/deforceHK/goghostex"
)

/**
* unit test cmd
* go test -v ./kraken/... -count=1 -run=TestWSSpotWebsocketBook_Start
*
**/
func TestWSSpotWebsocketBook_Start(t *testing.T) {
	//var receivedNum = 0
	var book = SpotOrderBooks{
		WSSpotMarketKK: &WSSpotMarketKK{

			Config: &APIConfig{
				Endpoint:   SPOT_KRAKEN_ENDPOINT,
				HttpClient: &http.Client{
					//Transport: &http.Transport{
					//	Proxy: func(req *http.Request) (*url.URL, error) {
					//		return url.Parse(SWAP_PROXY_URL)
					//	},
					//},
				},
				ApiKey:        SPOT_API_KEY,
				ApiSecretKey:  SPOT_API_SECRETKEY,
				ApiPassphrase: SPOT_API_PASSPHRASE,
				Location:      time.Now().Location(),
			},
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
