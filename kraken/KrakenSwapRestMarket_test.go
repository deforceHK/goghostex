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
* go test -v ./kraken/... -count=1 -run=TestSwap_Kraken_Market
*
**/
func TestSwap_Kraken_Market(t *testing.T) {

	var config = &APIConfig{
		Endpoint: ENDPOINT,
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
	}

	var kraken = New(config)

	if depth, resp, err := kraken.Swap.GetDepth(NewPair("btc_usd", "_"), 10); err != nil {
		t.Error(err)
		return
	} else {
		t.Log(depth)
		t.Log(string(resp))
	}

	if contract, resp, err := kraken.Swap.GetTicker(NewPair("btc_usd", "_")); err != nil {
		t.Error(err)
		return
	} else {
		t.Log(contract)
		t.Log(string(resp))
	}

	//if klines, resp, err := kraken.Swap.GetKline(
	//	NewPair("btc_usd", "_"),
	//	KLINE_PERIOD_1MIN,
	//	1440,
	//	1654045200000,
	//); err != nil {
	//	t.Error(err)
	//	return
	//} else {
	//	//t.Log(string(resp))
	//	for _, kline := range klines {
	//		t.Log(kline)
	//	}
	//	go func() {
	//		t.Log(string(resp))
	//	}()
	//}
}
