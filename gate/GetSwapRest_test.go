package gate

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"testing"
	"time"

	. "github.com/strengthening/goghostex"
)

func TestSwap_GetExchangeRule(t *testing.T) {

	config := &APIConfig{
		Endpoint: "",
		HttpClient: &http.Client{
			Transport: &http.Transport{
				Proxy: func(req *http.Request) (*url.URL, error) {
					return url.Parse("socks5://127.0.0.1:1090")
				},
			},
		},
		//ApiKey:        SPOT_API_KEY,
		//ApiSecretKey:  SPOT_API_SECRETKEY,
		//ApiPassphrase: SPOT_API_PASSPHRASE,
		Location: time.Now().Location(),
	}

	gateCli := New(config)

	//if rule, resp, err := gateCli.Swap.GetExchangeRule(Pair{Currency{"AMPL", ""}, USDT}); err != nil {
	//	t.Error(err)
	//	return
	//} else {
	//	fmt.Print(rule)
	//	fmt.Print(string(resp))
	//}

	//if rule, resp, err := gateCli.Swap.GetTicker(
	//	Pair{Currency{"AMPL", ""}, USDT},
	//); err != nil {
	//	t.Error(err)
	//	return
	//} else {
	//	fmt.Print(rule)
	//	fmt.Print(string(resp))
	//}

	//if depth, resp, err := gateCli.Swap.GetDepth(
	//	Pair{Currency{"AMPL", ""}, USDT},
	//	50,
	//); err != nil {
	//	t.Error(err)
	//	return
	//} else {
	//	fmt.Print(depth)
	//	//fmt.Print(string(resp))
	//}

	if klines, _, err := gateCli.Swap.GetKline(
		Pair{Currency{"AMPL", ""}, USDT},
		KLINE_PERIOD_1MIN,
		10,
		0,
	); err != nil {
		t.Error(err)
		return
	} else {
		body, _ := json.Marshal(klines)
		fmt.Print(string(body))
	}
}
