package coinbase

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	. "github.com/strengthening/goghostex"
)

const (
	SPOT_API_KEY        = ""
	SPOT_API_SECRETKEY  = ""
	SPOT_API_PASSPHRASE = ""

	SPOT_PROXY_URL = "socks5://127.0.0.1:1090"
)

func TestSpot_GetKlineRecords(t *testing.T) {

	config := &APIConfig{
		Endpoint:   ENDPOINT,
		HttpClient: &http.Client{},
		//ApiKey:        SPOT_API_KEY,
		//ApiSecretKey:  SPOT_API_SECRETKEY,
		//ApiPassphrase: SPOT_API_PASSPHRASE,
		Location: time.Now().Location(),
	}

	cb := New(config)

	klines, _, err := cb.Spot.GetKlineRecords(
		CurrencyPair{BTC, USD},
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

}
