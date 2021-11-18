package okex

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"testing"
	"time"

	. "github.com/strengthening/goghostex"
)

func TestSwap_MarketAPI(t *testing.T) {
	config := &APIConfig{
		Endpoint: ENDPOINT,
		HttpClient: &http.Client{
			Transport: &http.Transport{
				Proxy: func(req *http.Request) (*url.URL, error) {
					return url.Parse(PROXY_URL)
				},
			},
		},
		ApiKey:        SPOT_API_KEY,
		ApiSecretKey:  SPOT_API_SECRETKEY,
		ApiPassphrase: SPOT_API_PASSPHRASE,
		Location:      time.Now().Location(),
	}

	ok := New(config)
	// ticker unit test
	//if ticker, _, err := ok.Swap.GetTicker(
	//	Pair{Basis: BTC, Counter: USDT},
	//); err != nil {
	//	t.Error(err)
	//	return
	//}else{
	//	fmt.Println(ticker)
	//	raw,_ :=json.Marshal(ticker)
	//	fmt.Println(string(raw))
	//}

	//if depth, resp, err := ok.Swap.GetDepth(
	//	Pair{Basis: BTC, Counter: USDT},
	//	200,
	//); err != nil {
	//	t.Error(err)
	//	return
	//} else {
	//	fmt.Println(depth)
	//	fmt.Println(string(resp))
	//}
	//
	//if high, low, err := ok.Swap.GetLimit(Pair{Basis: BTC, Counter: USD}); err != nil {
	//	t.Error(err)
	//	return
	//} else {
	//	fmt.Println(high, low)
	//}

	if klines, resp, err := ok.Swap.GetKline(
		Pair{Basis: BTC, Counter: USD},
		KLINE_PERIOD_1DAY,
		100,
		0,
	); err != nil {
		t.Error(err)
		return
	} else {
		raw, _ := json.Marshal(klines)

		fmt.Println(string(raw))
		fmt.Println(string(resp))
	}
}
