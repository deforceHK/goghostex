package kraken

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
* go test -v ./kraken/... -count=1 -run=TestSwap_Trade_And_Cancel
*
**/
func TestSwap_Trade_And_Cancel(t *testing.T) {

	config := &APIConfig{
		Endpoint: SWAP_KRAKEN_ENDPOINT,
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

	var kr = New(config)
	var pair = Pair{BTC, USD}

	var ticker, _, err = kr.Swap.GetTicker(pair)
	if err != nil {
		t.Error(err)
		return
	}

	var orderLong = SwapOrder{
		Cid:       UUID(),
		Price:     ticker.Buy * 0.9,
		Amount:    0.0001,
		PlaceType: NORMAL,
		Type:      OPEN_LONG,
		LeverRate: 20,
		Pair:      pair,
		Exchange:  KRAKEN,
	}

	// 下多单
	if resp, err := kr.Swap.PlaceOrder(&orderLong); err != nil {
		t.Error(err)
		return
	} else {
		stdOrder, _ := json.Marshal(orderLong)
		t.Log(string(resp))
		t.Log(string(stdOrder))
	}

	if resp, err := kr.Swap.GetOrder(&orderLong); err != nil {
		t.Error(err)
		return
	} else {
		stdOrder, _ := json.Marshal(orderLong)
		t.Log(string(resp))
		t.Log(string(stdOrder))
	}
	time.Sleep(5 * time.Second)

	if resp, err := kr.Swap.CancelOrder(&orderLong); err != nil {
		t.Error(err)
		return
	} else {
		stdOrder, _ := json.Marshal(orderLong)
		t.Log(string(resp))
		t.Log(string(stdOrder))
	}

	var orderShort = SwapOrder{
		Cid:       UUID(),
		Price:     ticker.Sell * 1.1,
		Amount:    0.0001,
		PlaceType: NORMAL,
		Type:      OPEN_SHORT,
		LeverRate: 20,
		Pair:      pair,
		Exchange:  KRAKEN,
	}

	// 下空单
	if resp, err := kr.Swap.PlaceOrder(&orderShort); err != nil {
		t.Error(err)
		return
	} else {
		stdOrder, _ := json.Marshal(orderShort)
		t.Log(string(resp))
		t.Log(string(stdOrder))
	}

	if resp, err := kr.Swap.GetOrder(&orderShort); err != nil {
		t.Error(err)
		return
	} else {
		stdOrder, _ := json.Marshal(orderShort)
		t.Log(string(resp))
		t.Log(string(stdOrder))
	}
	time.Sleep(5 * time.Second)
	if resp, err := kr.Swap.CancelOrder(&orderShort); err != nil {
		t.Error(err)
		return
	} else {
		stdOrder, _ := json.Marshal(orderShort)
		t.Log(string(resp))
		t.Log(string(stdOrder))
	}

}

/**
* unit test cmd
* go test -v ./kraken/... -count=1 -run=TestSwap_Trade_Deal
*
**/
func TestSwap_Trade_Deal(t *testing.T) {

	config := &APIConfig{
		Endpoint:   SWAP_KRAKEN_ENDPOINT,
		HttpClient: &http.Client{
			//Transport: &http.Transport{
			//	Proxy: func(req *http.Request) (*url.URL, error) {
			//		return url.Parse(SWAP_PROXY_URL)
			//	},
			//},
		},
		ApiKey:        SWAP_API_KEY,
		ApiSecretKey:  SWAP_API_SECRETKEY,
		ApiPassphrase: SWAP_API_PASSPHRASE,
		Location:      time.Now().Location(),
	}

	var kr = New(config)
	var pair = Pair{
		BTC, USD,
	}
	var ticker, _, err = kr.Swap.GetTicker(pair)
	if err != nil {
		t.Error(err)
		return
	}
	//var orderShort = SwapOrder{
	//	Cid:       UUID(),
	//	Price:     59000,
	//	Amount:    1,
	//	PlaceType: NORMAL,
	//	Type:      OPEN_SHORT,
	//	LeverRate: 20,
	//	Pair:      pair,
	//	Exchange:  KRAKEN,
	//}

	//var orderLong = SwapOrder{
	//	Cid:       UUID(),
	//	Price:     61999,
	//	Amount:    0.0001,
	//	PlaceType: NORMAL,
	//	Type:      OPEN_LONG,
	//	LeverRate: 20,
	//	Pair:      pair,
	//	Exchange:  KRAKEN,
	//}
	var orderLong = SwapOrder{
		Cid:       UUID(),
		Price:     ticker.Sell * 1.05,
		Amount:    0.0001,
		PlaceType: NORMAL,
		Type:      OPEN_LONG,
		LeverRate: 20,
		Pair:      pair,
		Exchange:  KRAKEN,
	}

	//// 下空单
	//if resp, err := kr.Swap.PlaceOrder(&orderShort); err != nil {
	//	t.Error(err)
	//	return
	//} else {
	//	stdOrder, _ := json.Marshal(orderShort)
	//	t.Log(string(resp))
	//	t.Log(string(stdOrder))
	//}

	// 下多单
	if resp, err := kr.Swap.PlaceOrder(&orderLong); err != nil {
		t.Error(err)
		return
	} else {
		stdOrder, _ := json.Marshal(orderLong)
		t.Log(string(resp))
		t.Log(string(stdOrder))
	}

	if resp, err := kr.Swap.GetOrder(&orderLong); err != nil {
		t.Error(err)
		return
	} else {
		stdOrder, _ := json.Marshal(orderLong)
		t.Log(string(resp))
		t.Log(string(stdOrder))
	}

	time.Sleep(5 * time.Second)

	// 平多单
	var liquidateLong = SwapOrder{
		Cid:       UUID(),
		Price:     ticker.Buy * 0.95,
		Amount:    0.0001,
		PlaceType: NORMAL,
		Type:      LIQUIDATE_LONG,
		LeverRate: 20,
		Pair:      pair,
		Exchange:  KRAKEN,
	}
	if resp, err := kr.Swap.PlaceOrder(&liquidateLong); err != nil {
		t.Error(err)
		return
	} else {
		stdOrder, _ := json.Marshal(liquidateLong)
		t.Log(string(resp))
		t.Log(string(stdOrder))
	}

	if resp, err := kr.Swap.GetOrder(&liquidateLong); err != nil {
		t.Error(err)
		return
	} else {
		stdOrder, _ := json.Marshal(liquidateLong)
		t.Log(string(resp))
		t.Log(string(stdOrder))
	}

}

/**
* unit test cmd
* go test -v ./kraken/... -count=1 -run=TestSwap_Trade_Short_Deal
*
**/
func TestSwap_Trade_Short_Deal(t *testing.T) {

	config := &APIConfig{
		Endpoint:   SWAP_KRAKEN_ENDPOINT,
		HttpClient: &http.Client{
			//Transport: &http.Transport{
			//	Proxy: func(req *http.Request) (*url.URL, error) {
			//		return url.Parse(SWAP_PROXY_URL)
			//	},
			//},
		},
		ApiKey:        SWAP_API_KEY,
		ApiSecretKey:  SWAP_API_SECRETKEY,
		ApiPassphrase: SWAP_API_PASSPHRASE,
		Location:      time.Now().Location(),
	}

	var kr = New(config)
	var pair = Pair{
		ETH, USD,
	}
	var ticker, _, err = kr.Swap.GetTicker(pair)
	if err != nil {
		t.Error(err)
		return
	}
	var orderShort = SwapOrder{
		Cid:       UUID(),
		Price:     ticker.Buy * 0.95,
		Amount:    0.013,
		PlaceType: NORMAL,
		Type:      OPEN_SHORT,
		LeverRate: 20,
		Pair:      pair,
		Exchange:  KRAKEN,
	}

	//var orderLong = SwapOrder{
	//	Cid:       UUID(),
	//	Price:     61999,
	//	Amount:    0.0001,
	//	PlaceType: NORMAL,
	//	Type:      OPEN_LONG,
	//	LeverRate: 20,
	//	Pair:      pair,
	//	Exchange:  KRAKEN,
	//}
	//var orderLong = SwapOrder{
	//	Cid:       UUID(),
	//	Price:     ticker.Sell * 1.05,
	//	Amount:    0.0001,
	//	PlaceType: NORMAL,
	//	Type:      OPEN_LONG,
	//	LeverRate: 20,
	//	Pair:      pair,
	//	Exchange:  KRAKEN,
	//}

	//// 下空单
	if resp, err := kr.Swap.PlaceOrder(&orderShort); err != nil {
		t.Error(err)
		return
	} else {
		stdOrder, _ := json.Marshal(orderShort)
		t.Log(string(resp))
		t.Log(string(stdOrder))
	}

	if resp, err := kr.Swap.GetOrder(&orderShort); err != nil {
		t.Error(err)
		return
	} else {
		stdOrder, _ := json.Marshal(orderShort)
		t.Log(string(resp))
		t.Log(string(stdOrder))
	}
	time.Sleep(5 * time.Second)

	if resp, err := kr.Swap.GetOrder(&orderShort); err != nil {
		t.Error(err)
		return
	} else {
		stdOrder, _ := json.Marshal(orderShort)
		t.Log(string(resp))
		t.Log(string(stdOrder))
	}

}

/**
* unit test cmd
* go test -v ./kraken/... -count=1 -run=TestSwap_GetORders
*
**/
func TestSwap_GetORders(t *testing.T) {

	config := &APIConfig{
		Endpoint:   SWAP_KRAKEN_ENDPOINT,
		HttpClient: &http.Client{
			//Transport: &http.Transport{
			//	Proxy: func(req *http.Request) (*url.URL, error) {
			//		return url.Parse(SWAP_PROXY_URL)
			//	},
			//},
		},
		ApiKey:        SWAP_API_KEY,
		ApiSecretKey:  SWAP_API_SECRETKEY,
		ApiPassphrase: SWAP_API_PASSPHRASE,
		Location:      time.Now().Location(),
	}

	var kr = New(config)
	var pair = Pair{
		BTC, USD,
	}
	if orders, resp, err := kr.Swap.GetOrders(pair); err != nil {
		t.Error(err)
		return
	} else {
		t.Log(string(resp))
		t.Log(orders)
	}
}

/**
* unit test cmd
* go test -v ./kraken/... -count=1 -run=TestSwap_GetUnFinishOrders
*
**/

func TestSwap_GetUnFinishOrders(t *testing.T) {
	config := &APIConfig{
		Endpoint: SWAP_KRAKEN_ENDPOINT,
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

	var kr = New(config)
	var pair = Pair{
		BTC, USD,
	}
	var _, resp, err = kr.Swap.GetUnFinishOrders(pair)
	if err != nil {
		t.Error(err)
		return
	} else {
		t.Log(string(resp))
	}
}

/**
* unit test cmd
* go test -v ./kraken/... -count=1 -run=TestSwap_GetOrders
*
**/

func TestSwap_GetOrders(t *testing.T) {
	config := &APIConfig{
		Endpoint: SWAP_KRAKEN_ENDPOINT,
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

	var kr = New(config)
	var pair = Pair{
		BTC, USD,
	}
	var _, resp, err = kr.Swap.GetOrders(pair)
	if err != nil {
		t.Error(err)
		return
	} else {
		t.Log(string(resp))
	}
}
