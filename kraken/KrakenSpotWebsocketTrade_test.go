package kraken

import (
	"net/http"
	"testing"
	"time"

	. "github.com/deforceHK/goghostex"
)

/**
* unit test cmd
* go test -v ./kraken/... -count=1 -run=TestWSSpotTradeKK_Start
*
**/
func TestWSSpotTradeKK_Start(t *testing.T) {
	//var receivedNum = 0
	var ws = WSSpotTradeKK{
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
	}
	var err = ws.Start()
	if err != nil {
		t.Error(err)
		return
	}

	time.Sleep(20 * time.Second)
	select {}
}
