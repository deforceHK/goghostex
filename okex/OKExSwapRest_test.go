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

const (
	SWAP_API_KEY        = ""
	SWAP_API_SECRETKEY  = ""
	SWAP_API_PASSPHRASE = ""
)

// must set both
// place the order ---> get the order info ---> cancel the order -> get the order info
func TestSwap_TradeAPI(t *testing.T) {

	config := &APIConfig{
		Endpoint:   ENDPOINT,
		HttpClient: &http.Client{
			//Transport: &http.Transport{
			//	Proxy: func(req *http.Request) (*url.URL, error) {
			//		return url.Parse(PROXY_URL)
			//	},
			//},
		},
		ApiKey:        SWAP_API_KEY,
		ApiSecretKey:  SWAP_API_SECRETKEY,
		ApiPassphrase: SWAP_API_PASSPHRASE,
		Location:      time.Now().Location(),
	}

	ok := New(config)
	pair := Pair{Basis: BTC, Counter: USDT}
	ticker, _, err := ok.Swap.GetTicker(pair)
	if err != nil {
		t.Error(err)
		return
	}

	orderShort := SwapOrder{
		Cid:       UUID(),
		Price:     ticker.Sell * 1.03, //61506.1,
		Amount:    1,
		PlaceType: NORMAL,
		Type:      OPEN_SHORT,
		LeverRate: 20,
		Pair:      pair,
		Exchange:  OKEX,
	}

	orderLong := SwapOrder{
		Cid:       UUID(),
		Price:     ticker.Buy * 0.97,
		Amount:    1,
		PlaceType: NORMAL,
		Type:      OPEN_LONG,
		LeverRate: 20,
		Pair:      pair,
		Exchange:  OKEX,
	}

	// 下空单
	if resp, err := ok.Swap.PlaceOrder(&orderShort); err != nil {
		t.Error(err)
		return
	} else {
		stdOrder, _ := json.Marshal(orderShort)
		t.Log(string(resp))
		t.Log(string(stdOrder))
	}

	// 下多单
	if resp, err := ok.Swap.PlaceOrder(&orderLong); err != nil {
		t.Error(err)
		return
	} else {
		stdOrder, _ := json.Marshal(orderLong)
		t.Log(string(resp))
		t.Log(string(stdOrder))
	}

	if resp, err := ok.Swap.GetOrder(&orderShort); err != nil {
		t.Error(err)
		return
	} else {
		stdOrder, _ := json.Marshal(orderShort)
		t.Log(string(resp))
		t.Log(string(stdOrder))
	}
}
