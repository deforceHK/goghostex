package binance

import (
	"encoding/json"
	"net/http"
	"net/url"
	"testing"
	"time"

	. "github.com/deforceHK/goghostex"
)

/**
* unit test cmd
* go test -v ./binance/... -count=1 -run=TestFuture_AccountAPI
*
**/

func TestFuture_AccountAPI(t *testing.T) {

	var config = &APIConfig{
		Endpoint: ENDPOINT,
		HttpClient: &http.Client{
			Transport: &http.Transport{
				Proxy: func(req *http.Request) (*url.URL, error) {
					return url.Parse(FUTURE_PROXY_URL)
				},
			},
		},
		ApiKey:        FUTURE_API_KEY,
		ApiSecretKey:  FUTURE_API_SECRETKEY,
		ApiPassphrase: FUTURE_API_PASSPHRASE,
		Location:      time.Now().Location(),
	}

	var bn = New(config)
	if flow, resp, err := bn.Future.GetPairFlow(Pair{BTC, USD}); err != nil {
		t.Error(err)
		return
	} else {
		t.Log(string(resp))
		sss, _ := json.Marshal(flow)
		t.Log(string(sss))
	}

}
