package okex

import (
	"encoding/json"
	"net/http"
	"net/url"
	"testing"
	"time"

	. "github.com/deforceHK/goghostex"
)

const (
	SWAP_API_KEY        = ""
	SWAP_API_SECRETKEY  = ""
	SWAP_API_PASSPHRASE = ""
)

/**
* place the order ---> get the order info ---> cancel the order -> get the order info
*
* unit test cmd
* go test -v ./okex/... -count=1 -run=TestSwap_TradeAPI
*
**/
func TestSwap_TradeAPI(t *testing.T) {

	config := &APIConfig{
		Endpoint: ENDPOINT,
		HttpClient: &http.Client{
			Transport: &http.Transport{
				Proxy: func(req *http.Request) (*url.URL, error) {
					return url.Parse(PROXY_URL)
				},
			},
		},
		ApiKey:        SWAP_API_KEY,
		ApiSecretKey:  SWAP_API_SECRETKEY,
		ApiPassphrase: SWAP_API_PASSPHRASE,
		Location:      time.Now().Location(),
	}

	ok := New(config)
	//if items, resp, err := ok.Swap.GetAccountFlow(); err != nil {
	//	t.Error(err)
	//	return
	//} else {
	//	t.Log(string(resp))
	//	output, _ := json.Marshal(items)
	//	t.Log(string(output))
	//}

	pair := Pair{Basis: BTC, Counter: USDT}
	ticker, _, err := ok.Swap.GetTicker(pair)
	if err != nil {
		t.Error(err)
		return
	}

	orderShort := SwapOrder{
		Cid:       UUID(),
		Price:     ticker.Sell * 1.03,
		Amount:    1.1,
		PlaceType: NORMAL,
		Type:      OPEN_SHORT,
		LeverRate: 20,
		Pair:      pair,
		Exchange:  OKEX,
	}

	orderLong := SwapOrder{
		Cid:       UUID(),
		Price:     ticker.Buy * 0.97,
		Amount:    1.1,
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
		t.Log(string(resp))
		t.Error(err)
		return
	} else {
		stdOrder, _ := json.Marshal(orderShort)
		t.Log(string(resp))
		t.Log(string(stdOrder))
	}

	if resp, err := ok.Swap.CancelOrder(&orderShort); err != nil {
		t.Error(err)
		return
	} else {
		stdOrder, _ := json.Marshal(orderShort)
		t.Log(string(resp))
		t.Log(string(stdOrder))
	}

	if resp, err := ok.Swap.CancelOrder(&orderLong); err != nil {
		t.Error(err)
		return
	} else {
		stdOrder, _ := json.Marshal(orderLong)
		t.Log(string(resp))
		t.Log(string(stdOrder))
	}

	if resp, err := ok.Swap.GetOrder(&orderShort); err != nil {
		t.Log(string(resp))
		t.Error(err)
		return
	} else {
		stdOrder, _ := json.Marshal(orderShort)
		t.Log(string(resp))
		t.Log(string(stdOrder))
	}

}

/**
* place the order ---> get the order info
*
* unit test cmd
* go test -v ./okex/... -count=1 -run=TestSwap_DealAPI
*
**/
func TestSwap_DealAPI(t *testing.T) {

	config := &APIConfig{
		Endpoint: ENDPOINT,
		HttpClient: &http.Client{
			Transport: &http.Transport{
				Proxy: func(req *http.Request) (*url.URL, error) {
					return url.Parse(PROXY_URL)
				},
			},
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
		Price:     ticker.Sell * 0.99,
		Amount:    1,
		PlaceType: NORMAL,
		Type:      OPEN_SHORT,
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

	for {
		resp1, err1 := ok.Swap.GetOrder(&orderShort)

		if err1 != nil {
			t.Error(err1)
			return
		}

		if orderShort.Status == ORDER_FINISH || orderShort.Status == ORDER_CANCEL {
			t.Log(string(resp1))
			t.Log(orderShort)
			break
		}
		time.Sleep(2 * time.Second)
	}

	orderLiquidate := SwapOrder{
		Cid:       UUID(),
		Price:     orderShort.AvgPrice * 1.01,
		Amount:    1,
		PlaceType: NORMAL,
		Type:      LIQUIDATE_SHORT,
		LeverRate: 20,
		Pair:      pair,
		Exchange:  OKEX,
	}

	// 平空单
	if resp, err := ok.Swap.PlaceOrder(&orderLiquidate); err != nil {
		t.Error(err)
		return
	} else {
		stdOrder, _ := json.Marshal(orderLiquidate)
		t.Log(string(resp))
		t.Log(string(stdOrder))
	}

	for {
		resp1, err1 := ok.Swap.GetOrder(&orderLiquidate)

		if err1 != nil {
			t.Error(err1)
			return
		}

		if orderLiquidate.Status == ORDER_FINISH || orderLiquidate.Status == ORDER_CANCEL {
			t.Log(string(resp1))
			t.Log(orderLiquidate)
			break
		}
		time.Sleep(2 * time.Second)
	}

}
