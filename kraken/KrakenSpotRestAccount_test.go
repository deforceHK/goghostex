package kraken

import (
	"net/http"
	"testing"
	"time"

	. "github.com/deforceHK/goghostex"
)

/**
* unit test cmd
* go test -v ./kraken/... -count=1 -run=TestSpot_GetAccount
*
**/
func TestSpot_GetAccount(t *testing.T) {

	var config = &APIConfig{
		Endpoint:   ENDPOINT,
		HttpClient: &http.Client{
			//Transport: &http.Transport{
			//	Proxy: func(req *http.Request) (*url.URL, error) {
			//		return url.Parse("socks5://127.0.0.1:1090")
			//	},
			//},
		},
		ApiKey:        SPOT_API_KEY,
		ApiSecretKey:  SPOT_API_SECRETKEY,
		ApiPassphrase: SPOT_API_PASSPHRASE,
		Location:      time.Now().Location(),
	}

	var kraken = New(config)
	if _, resp, err := kraken.Spot.GetAccount(); err != nil {
		t.Error(err)
		return
	} else {
		t.Logf(string(resp))
	}
}
