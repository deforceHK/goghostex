package okex

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	. "github.com/deforceHK/goghostex"
)

// go test -v ./okex/... -count=1 -run=TestLocalOrderBooks_Init
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

	var wsOK = &LocalOrderBooks{
		WSMarketOKEx: &WSMarketOKEx{
			Config: config,
		},
	}

	err := wsOK.Init()
	if err != nil {
		t.Error(err)
		return
	}
	time.Sleep(1 * time.Second)

	var sol = Pair{SOL, USDT}
	var eth = Pair{ETH, USDT}
	wsOK.Subscribe(sol)
	wsOK.Subscribe(eth)

	go func() {
		time.Sleep(7 * time.Second)
		wsOK.Resubscribe("SOL-USDT-SWAP")
	}()

	for i := 0; i < 10; i++ {
		time.Sleep(5 * time.Second)
		if i%2 == 0 {
			depth, depthErr := wsOK.Snapshot(sol)
			if depthErr != nil {
				t.Error(depthErr)
				return
			}
			var depthData, _ = json.Marshal(depth)
			t.Log(string(depthData))
		} else {
			depth, depthErr := wsOK.Snapshot(eth)
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
