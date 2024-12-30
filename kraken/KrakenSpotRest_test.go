package kraken

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	. "github.com/deforceHK/goghostex"
)

const (
	SPOT_API_KEY        = ""
	SPOT_API_SECRETKEY  = ""
	SPOT_API_PASSPHRASE = ""
)

/**
* unit test cmd
* go test -v ./kraken/... -count=1 -run=TestSpot_GetKlineRecords
*
**/
func TestSpot_GetKlineRecords(t *testing.T) {

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
	klines, _, err := kraken.Spot.GetKlineRecords(
		Pair{Basis: MATIC, Counter: USD},
		KLINE_PERIOD_1MIN,
		300,
		1546898760000,
	)

	if err != nil {
		t.Error(err)
		return
	}

	raw, err := json.Marshal(klines)
	if err != nil {
		t.Error(err)
		return
	}

	fmt.Println(string(raw))
	//fmt.Println(string(resp))
	var resp, token, tokenErr = kraken.GetToken()
	if tokenErr != nil {
		t.Error(tokenErr)
		return
	} else {
		fmt.Println(string(resp))
		fmt.Println(token)
	}
}
