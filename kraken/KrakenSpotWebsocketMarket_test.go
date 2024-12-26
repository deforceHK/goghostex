package kraken

import (
	"net/http"
	"testing"
	"time"

	. "github.com/deforceHK/goghostex"
)

/**
* unit test cmd
* go test -v ./kraken/... -count=1 -run=TestWSSpotMarketKK_Start
*
**/
func TestWSSpotMarketKK_Start(t *testing.T) {
	//var receivedNum = 0
	var ws = WSSpotMarketKK{
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

	var xbt = struct {
		Method string `json:"method"`
		Params struct {
			Channel string   `json:"channel"`
			Symbol  []string `json:"symbol"`
			Depth   int64    `json:"depth"`
		} `json:"params"`
	}{
		Method: "subscribe",
		Params: struct {
			Channel string `json:"channel"`
			Symbol []string `json:"symbol"`
			Depth  int64    `json:"depth"`
		}{
			"book",
			[]string{"BTC/USD"},
			500,
		},
	}

	//var subXBT = struct {
	//	Event      string   `json:"event"`
	//	Feed       string   `json:"feed"`
	//	ProductIds []string `json:"product_ids"`
	//}{
	//	"subscribe", "book", []string{"PF_XBTUSD"},
	//}
	//var subETH = struct {
	//	Event      string   `json:"event"`
	//	Feed       string   `json:"feed"`
	//	ProductIds []string `json:"product_ids"`
	//}{
	//	"subscribe", "book", []string{"PF_ETHUSD"},
	//}

	ws.Subscribe(xbt)
	//ws.Subscribe(subETH)

	time.Sleep(20 * time.Second)
	ws.Restart()
	select {}
}
