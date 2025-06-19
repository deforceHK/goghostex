package binance

import (
	"net/http"
	"testing"
	"time"

	. "github.com/deforceHK/goghostex"
)

const (
	ONE_API_KEY        = ""
	ONE_API_SECRETKEY  = ""
	ONE_API_PASSPHRASE = ""
	ONE_PROXY_URL      = ""
)

// go test -v ./binance/... -count=1 -run=TestOne_GetInfos
func TestOne_GetInfos(t *testing.T) {

	config := &APIConfig{
		Endpoint:   ENDPOINT,
		HttpClient: &http.Client{
			//Transport: &http.Transport{
			//	Proxy: func(req *http.Request) (*url.URL, error) {
			//		return url.Parse(ONE_PROXY_URL)
			//	},
			//},
		},
		ApiKey:        ONE_API_KEY,
		ApiSecretKey:  ONE_API_SECRETKEY,
		ApiPassphrase: ONE_API_PASSPHRASE,
		Location:      time.Now().Location(),
	}

	bn := New(config)

	var infos, _, err = bn.One.GetCMInfos()
	if err != nil {
		t.Error(err)
		return
	}

	for _, info := range infos {
		t.Log(*info)
	}

	infos, _, err = bn.One.GetUMInfos()
	if err != nil {
		t.Error(err)
		return
	}

	for _, info := range infos {
		t.Log(*info)
	}
}
