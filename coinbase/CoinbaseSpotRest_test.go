package coinbase

import (
	"encoding/json"
	"net/http"
	"net/url"
	"testing"
	"time"

	. "github.com/deforceHK/goghostex"
)

const (
	SPOT_API_KEY        = ""
	SPOT_API_SECRETKEY  = ""
	SPOT_API_PASSPHRASE = ""

	SPOT_PROXY_URL = "socks5://127.0.0.1:1090"
)

/**
go test -v ./coinbase/... -count=1 -run=TestSpot_GetKlineRecords

**/
func TestSpot_GetKlineRecords(t *testing.T) {

	config := &APIConfig{
		Endpoint: ENDPOINT,
		HttpClient: &http.Client{
			Transport: &http.Transport{
				Proxy: func(req *http.Request) (*url.URL, error) {
					return url.Parse("socks5://127.0.0.1:1090")
				},
			},
		},
		ApiKey:        SPOT_API_KEY,
		ApiSecretKey:  SPOT_API_SECRETKEY,
		ApiPassphrase: SPOT_API_PASSPHRASE,
		Location:      time.Now().Location(),
	}

	var cb = New(config)
	var klines, _, err = cb.Spot.GetKlineRecords(
		Pair{Basis: BTC, Counter: USD},
		KLINE_PERIOD_1MIN,
		300,
		1645571700000,
	)

	if err != nil {
		t.Error(err)
		return
	}

	raw, err := json.Marshal(klines[0])
	if err != nil {
		t.Error(err)
		return
	}

	t.Log(string(raw))

	if rule, raw, err := cb.Spot.GetExchangeRule(Pair{BTC, USD}); err != nil {
		t.Error(err)
		return
	} else {
		t.Log(string(raw))
		t.Log(rule)
	}

}
