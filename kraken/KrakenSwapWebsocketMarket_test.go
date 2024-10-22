package kraken

import (
	"net/http"
	"net/url"
	"testing"
	"time"

	. "github.com/deforceHK/goghostex"
)

/**
* unit test cmd
* go test -v ./kraken/... -count=1 -run=TestWSSwapMarketKK_Start
*
**/
func TestWSSwapMarketKK_Start(t *testing.T) {
	//var receivedNum = 0
	var ws = WSSwapMarketKK{
		Config: &APIConfig{
			Endpoint: SWAP_KRAKEN_ENDPOINT,
			HttpClient: &http.Client{
				Transport: &http.Transport{
					Proxy: func(req *http.Request) (*url.URL, error) {
						return url.Parse(SWAP_PROXY_URL)
					},
				},
			},
			ApiKey:        SWAP_API_KEY,
			ApiSecretKey:  SWAP_API_SECRETKEY,
			ApiPassphrase: SWAP_API_PASSPHRASE,
			Location:      time.Now().Location(),
		},
	}
	var err = ws.Start()
	if err != nil {
		t.Error(err)
		return
	}

	var subXBT = struct {
		Event      string   `json:"event"`
		Feed       string   `json:"feed"`
		ProductIds []string `json:"product_ids"`
	}{
		"subscribe", "book", []string{"PF_XBTUSD"},
	}
	var subETH = struct {
		Event      string   `json:"event"`
		Feed       string   `json:"feed"`
		ProductIds []string `json:"product_ids"`
	}{
		"subscribe", "book", []string{"PF_ETHUSD"},
	}

	ws.Subscribe(subXBT)
	ws.Subscribe(subETH)

	time.Sleep(20 * time.Second)
	ws.Restart()
	select {}
}
