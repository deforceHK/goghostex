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
* go test -v ./kraken/... -count=1 -run=TestWSSwapKK_Start
*
**/
func TestWSSwapKK_Start(t *testing.T) {

	var ws = WSSwapKK{
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

	var ticker = struct {
		Event string `json:"event"`
		Feed  string `json:"feed"`
		//ProductIDs []string `json:"product_ids"`
	}{
		Event: "subscribe",
		Feed:  "heartbeat", // "ticker",
		//ApiKey: SWAP_API_KEY,
		//ProductIDs: []string{"PF_XBTUSD", "PF_ETHUSD"},
	}
	ws.Subscribe(ticker)
	select {}

}
