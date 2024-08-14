package okex

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
* go test -v ./okex/... -count=1 -run=TestFuture_AccountAPI
*
**/

func TestFuture_AccountAPI(t *testing.T) {

	config := &APIConfig{
		Endpoint: ENDPOINT,
		HttpClient: &http.Client{
			Transport: &http.Transport{
				Proxy: func(req *http.Request) (*url.URL, error) {
					return url.Parse(PROXY_URL)
				},
			},
		},
		ApiKey:        FUTURE_API_KEY,
		ApiSecretKey:  FUTURE_API_SECRETKEY,
		ApiPassphrase: FUTURE_API_PASSPHRASE,
		Location:      time.Now().Location(),
	}

	var ok = New(config)

	if item, resp, err := ok.Future.GetPairFlow(Pair{BTC, USD}); err != nil {
		t.Error(err)
		return
	} else {
		t.Log(string(resp))
		var sss, _ = json.Marshal(item)
		t.Log(string(sss))
	}

}
