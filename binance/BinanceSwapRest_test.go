package binance

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
	SWAP_PROXY_URL      = "socks5://127.0.0.1:1090"
)

/**
* unit test cmd
* go test -v ./binance/... -count=1 -run=TestSwap_MarketAPI_Counter
*
**/

func TestSwap_MarketAPI_Counter(t *testing.T) {

	config := &APIConfig{
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

	bn := New(config)

	var contract = bn.Swap.GetContract(Pair{Basis: BTC, Counter: USDT})
	t.Log(*contract)

	// ticker unit test
	if ticker, resp, err := bn.Swap.GetTicker(
		Pair{Basis: BTC, Counter: USDT},
	); err != nil {
		t.Error(err)
		return
	} else {
		standard, err := json.Marshal(ticker)
		if err != nil {
			t.Error(err)
			return
		}

		t.Log("Ticker standard struct: ")
		t.Log(string(standard))
		t.Log("Ticker remote api response: ")
		t.Log(string(resp))
	}

	// depth unit test
	if depth, resp, err := bn.Swap.GetDepth(
		Pair{Basis: BTC, Counter: USDT},
		100,
	); err != nil {
		t.Error(err)
		return
	} else {
		standard, err := json.Marshal(depth)
		if err != nil {
			t.Error(err)
			return
		}

		t.Log("Depth standard struct:")
		t.Log(string(standard))
		t.Log("Depth remote api response: ")
		t.Log(string(resp))

		// make sure the later request get bigger sequence
		depth1, _, _ := bn.Swap.GetDepth(
			Pair{Basis: BTC, Counter: USDT},
			100,
		)

		if depth1.Sequence < depth.Sequence {
			t.Error("later request get smaller sequence!!")
			return
		}

		if err := depth.Verify(); err != nil {
			t.Error(err)
			return
		}

		if err := depth1.Verify(); err != nil {
			t.Error(err)
			return
		}
	}

	if price, err := bn.Swap.GetMark(Pair{Basis: BTC, Counter: USDT}); err != nil {
		t.Error(err)
		return
	} else {
		t.Log("mark price: ", price)
	}

	if highest, lowest, err := bn.Swap.GetLimit(Pair{Basis: BTC, Counter: USDT}); err != nil {
		t.Error(err)
		return
	} else {
		t.Log("highest: ", highest)
		t.Log("lowest: ", lowest)
	}

	if klines, resp, err := bn.Swap.GetKline(
		Pair{Basis: BTC, Counter: USDT},
		KLINE_PERIOD_1DAY,
		20,
		1638288000000,
	); err != nil {
		t.Error(err)
		return
	} else {
		klineRaw, _ := json.Marshal(klines)
		t.Log(string(klineRaw))
		t.Log(string(resp))
		//return
	}

	if openAmount, timestamp, _, err := bn.Swap.GetOpenAmount(Pair{Basis: BTC, Counter: USDT}); err != nil {
		t.Error(err)
		return
	} else {
		t.Log(openAmount)
		t.Log(timestamp)
	}

	if fees, _, err := bn.Swap.GetFundingFees(Pair{Basis: BTC, Counter: USDT}); err != nil {
		t.Error(err)
		return
	} else {
		t.Log(fees)
		//t.Log(string(resp))
	}

	bn.Swap.KeepAlive()
}

func TestSwap_MarketAPI_Basis(t *testing.T) {

	config := &APIConfig{
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

	var bn = New(config)
	var contract = bn.Swap.GetContract(Pair{Basis: ETH, Counter: USD})
	content, _ := json.Marshal(*contract)
	t.Log(string(content))

	// ticker unit test
	//if ticker, resp, err := bn.Swap.GetTicker(
	//	Pair{Basis: BTC, Counter: USDT},
	//); err != nil {
	//	t.Error(err)
	//	return
	//} else {
	//	standard, err := json.Marshal(ticker)
	//	if err != nil {
	//		t.Error(err)
	//		return
	//	}
	//
	//	t.Log("Ticker standard struct: ")
	//	t.Log(string(standard))
	//	t.Log("Ticker remote api response: ")
	//	t.Log(string(resp))
	//}
	//
	//// depth unit test
	//if depth, resp, err := bn.Swap.GetDepth(
	//	Pair{Basis: BTC, Counter: USD},
	//	100,
	//); err != nil {
	//	t.Error(err)
	//	return
	//} else {
	//	standard, err := json.Marshal(depth)
	//	if err != nil {
	//		t.Error(err)
	//		return
	//	}
	//
	//	t.Log("Depth standard struct:")
	//	t.Log(string(standard))
	//	t.Log("Depth remote api response: ")
	//	t.Log(string(resp))
	//
	//	// make sure the later request get bigger sequence
	//	depth1, _, _ := bn.Swap.GetDepth(
	//		Pair{Basis: BTC, Counter: USD},
	//		20,
	//	)
	//
	//	if depth1.Sequence < depth.Sequence {
	//		t.Error("later request get smaller sequence!!")
	//		return
	//	}
	//
	//	if err := depth.Verify(); err != nil {
	//		t.Error(err)
	//		return
	//	}
	//
	//	if err := depth1.Verify(); err != nil {
	//		t.Error(err)
	//		return
	//	}
	//}
	//
	//if price, err := bn.Swap.GetMark(Pair{Basis: BTC, Counter: USD}); err != nil {
	//	t.Error(err)
	//	return
	//} else {
	//	t.Log("mark price: ", price)
	//}
	//
	//if highest, lowest, err := bn.Swap.GetLimit(Pair{Basis: BTC, Counter: USD}); err != nil {
	//	t.Error(err)
	//	return
	//} else {
	//	t.Log("highest:", highest)
	//	t.Log("lowest:", lowest)
	//}
	//
	//if klines, resp, err := bn.Swap.GetKline(
	//	Pair{Basis: BTC, Counter: USD},
	//	KLINE_PERIOD_1MIN,
	//	20,
	//	1638288000000,
	//); err != nil {
	//	t.Error(err)
	//	return
	//} else {
	//	klineRaw, _ := json.Marshal(klines)
	//	t.Log(string(klineRaw))
	//	t.Log(string(resp))
	//}

	//if openAmount, timestamp, _, err := bn.Swap.GetOpenAmount(Pair{Basis: BTC, Counter: USDT}); err != nil {
	//	t.Error(err)
	//	return
	//} else {
	//	t.Log(openAmount)
	//	t.Log(timestamp)
	//}
	//
	//if fees, _, err := bn.Swap.GetFundingFees(Pair{Basis: BTC, Counter: USDT}); err != nil {
	//	t.Error(err)
	//	return
	//} else {
	//	t.Log(fees)
	//	//t.Log(string(resp))
	//}
	//
	//bn.Swap.KeepAlive()
}

// must set both
// place the order ---> get the order info ---> cancel the order -> get the order info
// go test -v ./binance/... -count=1 -run=TestSwap_TradeAPI_COUNTER

func TestSwap_TradeAPI_COUNTER(t *testing.T) {

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

	bn := New(config)
	if account, raw, err := bn.Swap.GetAccount(); err != nil {
		t.Error(err)
		return
	} else {

		rawAccount, _ := json.Marshal(account)
		t.Log(string(rawAccount))
		t.Log(string(raw))

		if account.BalanceAvail < 1 {
			t.Error("There have no enough asset to trade. ")
			return
		}
	}

	pair := Pair{Basis: BTC, Counter: USDT}
	ticker, _, err := bn.Swap.GetTicker(pair)
	if err != nil {
		t.Error(err)
		return
	}

	orderShort := SwapOrder{
		Cid:       UUID(),
		Price:     ticker.Sell * 1.03,
		Amount:    0.01,
		PlaceType: NORMAL,
		Type:      OPEN_SHORT,
		LeverRate: 20,
		Pair:      pair,
		Exchange:  BINANCE,
	}

	orderLong := SwapOrder{
		Cid:       UUID(),
		Price:     ticker.Buy * 0.97,
		Amount:    0.01,
		PlaceType: NORMAL,
		Type:      OPEN_LONG,
		LeverRate: 20,
		Pair:      pair,
		Exchange:  BINANCE,
	}

	// 下空单
	if resp, err := bn.Swap.PlaceOrder(&orderShort); err != nil {
		t.Error(err)
		return
	} else {
		stdOrder, _ := json.Marshal(orderShort)
		t.Log(string(resp))
		t.Log(string(stdOrder))
	}

	// 下多单
	if resp, err := bn.Swap.PlaceOrder(&orderLong); err != nil {
		t.Error(err)
		return
	} else {
		stdOrder, _ := json.Marshal(orderLong)
		t.Log(string(resp))
		t.Log(string(stdOrder))
	}

	if resp, err := bn.Swap.GetOrder(&orderShort); err != nil {
		t.Error(err)
		return
	} else {
		stdOrder, _ := json.Marshal(orderShort)
		t.Log(string(resp))
		t.Log(string(stdOrder))
	}

	if resp, err := bn.Swap.CancelOrder(&orderShort); err != nil {
		t.Error(err)
		return
	} else {
		stdOrder, _ := json.Marshal(orderShort)
		t.Log(string(resp))
		t.Log(string(stdOrder))
	}

	if resp, err := bn.Swap.CancelOrder(&orderLong); err != nil {
		t.Error(err)
		return
	} else {
		stdOrder, _ := json.Marshal(orderLong)
		t.Log(string(resp))
		t.Log(string(stdOrder))
	}

}

// place the order ---> get the order info ---> deal
// go test -v ./binance/... -count=1 -run=TestSwap_DEALAPI_COUNTER
func TestSwap_DEALAPI_COUNTER(t *testing.T) {

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

	bn := New(config)
	if account, raw, err := bn.Swap.GetAccount(); err != nil {
		t.Error(err)
		return
	} else {

		rawAccount, _ := json.Marshal(account)
		t.Log(string(rawAccount))
		t.Log(string(raw))

		if account.BalanceAvail < 1 {
			t.Error("There have no enough asset to trade. ")
			return
		}
	}

	pair := Pair{Basis: BTC, Counter: USDT}
	ticker, _, err := bn.Swap.GetTicker(pair)
	if err != nil {
		t.Error(err)
		return
	}

	openShort := SwapOrder{
		Cid:       UUID(),
		Price:     ticker.Sell * 0.99,
		Amount:    0.012,
		PlaceType: NORMAL,
		Type:      OPEN_SHORT,
		LeverRate: 20,
		Pair:      pair,
		Exchange:  BINANCE,
	}

	// 下空单
	if resp, err := bn.Swap.PlaceOrder(&openShort); err != nil {
		t.Error(err)
		return
	} else {
		stdOrder, _ := json.Marshal(openShort)
		t.Log(string(resp))
		t.Log(string(stdOrder))
	}

	var isDeal = false
	for i := 0; i < 5; i++ {
		if resp, err := bn.Swap.GetOrder(&openShort); err != nil {
			t.Error(err)
			return
		} else {
			stdOrder, _ := json.Marshal(openShort)
			t.Log(string(resp))
			t.Log(string(stdOrder))

			if openShort.Status == ORDER_FINISH {
				isDeal = true
				break
			}
		}
	}

	if !isDeal {
		t.Error("The open order hasn't deal. ")
		return
	}

	liquidateShort := SwapOrder{
		Cid:       UUID(),
		Price:     ticker.Buy * 1.01,
		Amount:    openShort.DealAmount,
		PlaceType: NORMAL,
		Type:      LIQUIDATE_SHORT,
		LeverRate: 20,
		Pair:      pair,
		Exchange:  BINANCE,
	}

	// 平空单
	if resp, err := bn.Swap.PlaceOrder(&liquidateShort); err != nil {
		t.Error(err)
		return
	} else {
		stdOrder, _ := json.Marshal(liquidateShort)
		t.Log(string(resp))
		t.Log(string(stdOrder))
	}

	isDeal = false
	for i := 0; i < 5; i++ {
		if resp, err := bn.Swap.GetOrder(&liquidateShort); err != nil {
			t.Error(err)
			return
		} else {
			stdOrder, _ := json.Marshal(liquidateShort)
			t.Log(string(resp))
			t.Log(string(stdOrder))

			if liquidateShort.Status == ORDER_FINISH {
				isDeal = true
				break
			}
		}
	}

	if !isDeal {
		t.Error("The liquidate order hasn't deal. ")
		return
	}

}

// place the order ---> get the order info ---> cancel the order -> get the order info
// go test -v ./binance/... -count=1 -run=TestSwap_TradeAPI_BASIS
func TestSwap_TradeAPI_BASIS(t *testing.T) {

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

	var bn = New(config)
	//if account, raw, err := bn.Swap.GetAccount(); err != nil {
	//	t.Error(err)
	//	return
	//} else {
	//
	//	rawAccount, _ := json.Marshal(account)
	//	t.Log(string(rawAccount))
	//	t.Log(string(raw))
	//
	//	if account.BalanceAvail < 1 {
	//		t.Error("There have no enough asset to trade. ")
	//		return
	//	}
	//}

	var pair = Pair{Basis: BTC, Counter: USD}
	ticker, _, err := bn.Swap.GetTicker(pair)
	if err != nil {
		t.Error(err)
		return
	}

	orderShort := SwapOrder{
		Cid:       UUID(),
		Price:     ticker.Sell * 1.03,
		Amount:    1,
		PlaceType: NORMAL,
		Type:      OPEN_SHORT,
		LeverRate: 20,
		Pair:      pair,
		Exchange:  BINANCE,
	}

	orderLong := SwapOrder{
		Cid:       UUID(),
		Price:     ticker.Buy * 0.97,
		Amount:    1,
		PlaceType: NORMAL,
		Type:      OPEN_LONG,
		LeverRate: 20,
		Pair:      pair,
		Exchange:  BINANCE,
	}

	// 开空单
	if resp, err := bn.Swap.PlaceOrder(&orderShort); err != nil {
		t.Error(err)
		return
	} else {
		stdOrder, _ := json.Marshal(orderShort)
		t.Log(string(resp))
		t.Log(string(stdOrder))
	}

	// 开多单
	if resp, err := bn.Swap.PlaceOrder(&orderLong); err != nil {
		t.Error(err)
		return
	} else {
		stdOrder, _ := json.Marshal(orderLong)
		t.Log(string(resp))
		t.Log(string(stdOrder))
	}

	if resp, err := bn.Swap.GetOrder(&orderShort); err != nil {
		t.Error(err)
		return
	} else {
		stdOrder, _ := json.Marshal(orderShort)
		t.Log(string(resp))
		t.Log(string(stdOrder))
	}

	if resp, err := bn.Swap.CancelOrder(&orderShort); err != nil {
		t.Error(err)
		return
	} else {
		stdOrder, _ := json.Marshal(orderShort)
		t.Log(string(resp))
		t.Log(string(stdOrder))
	}

	if resp, err := bn.Swap.CancelOrder(&orderLong); err != nil {
		t.Error(err)
		return
	} else {
		stdOrder, _ := json.Marshal(orderLong)
		t.Log(string(resp))
		t.Log(string(stdOrder))
	}
}

// place the order ---> get the order info ---> deal
// go test -v ./binance/... -count=1 -run=TestSwap_DEALAPI_BASIS
func TestSwap_DEALAPI_BASIS(t *testing.T) {

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
		ApiPassphrase: "",
		Location:      time.Now().Location(),
	}

	bn := New(config)
	//if account, raw, err := bn.Swap.GetAccount(); err != nil {
	//	t.Error(err)
	//	return
	//} else {
	//
	//	rawAccount, _ := json.Marshal(account)
	//	t.Log(string(rawAccount))
	//	t.Log(string(raw))
	//
	//	if account.BalanceAvail < 1 {
	//		t.Error("There have no enough asset to trade. ")
	//		return
	//	}
	//}

	var pair = Pair{Basis: BTC, Counter: USD}
	var ticker, tickerResp, err = bn.Swap.GetTicker(pair)
	if err != nil {
		t.Error(err)
		return
	}

	t.Log(string(tickerResp))
	t.Log(ticker)

	openShort := SwapOrder{
		Cid:       UUID(),
		Price:     ticker.Sell * 0.99,
		Amount:    1,
		PlaceType: NORMAL,
		Type:      OPEN_SHORT,
		LeverRate: 20,
		Pair:      pair,
		Exchange:  BINANCE,
	}

	// 下空单
	if resp, err := bn.Swap.PlaceOrder(&openShort); err != nil {
		t.Error(err)
		return
	} else {
		stdOrder, _ := json.Marshal(openShort)
		t.Log(string(resp))
		t.Log(string(stdOrder))
	}

	var isDeal = false
	for i := 0; i < 5; i++ {
		if resp, err := bn.Swap.GetOrder(&openShort); err != nil {
			t.Error(err)
			return
		} else {
			stdOrder, _ := json.Marshal(openShort)
			t.Log(string(resp))
			t.Log(string(stdOrder))

			if openShort.Status == ORDER_FINISH {
				isDeal = true
				break
			}
		}
	}

	if !isDeal {
		t.Error("The open order hasn't deal. ")
		return
	}

	// 这里测试了取消已经成交的订单会报什么错误。
	//if resp, err := bn.Swap.CancelOrder(&openShort);err!=nil{
	//	t.Error(err)
	//	t.Error(string(resp))
	//	return
	//}
	// 错误信息
	//HttpStatusCode: 400, HttpMethod: DELETE, Response: {"code":-2011,"msg":"Unknown order sent."}, Request: , Url: https://dapi.binance.com/dapi/v1/order?orderId=115424246344&recvWindow=60000&signature=4f3a0e6524f93e4f21954c223f02e72f861ce40142b936a933417cd9f0c94299&symbol=BTCUSD_PERP&timestamp=1707296884739

	liquidateShort := SwapOrder{
		Cid:       UUID(),
		Price:     ticker.Buy * 1.01,
		Amount:    openShort.DealAmount,
		PlaceType: NORMAL,
		Type:      LIQUIDATE_SHORT,
		LeverRate: 20,
		Pair:      pair,
		Exchange:  BINANCE,
	}

	// 平空单
	if resp, err := bn.Swap.PlaceOrder(&liquidateShort); err != nil {
		t.Error(err)
		return
	} else {
		stdOrder, _ := json.Marshal(liquidateShort)
		t.Log(string(resp))
		t.Log(string(stdOrder))
	}

	isDeal = false
	for i := 0; i < 5; i++ {
		if resp, err := bn.Swap.GetOrder(&liquidateShort); err != nil {
			t.Error(err)
			return
		} else {
			stdOrder, _ := json.Marshal(liquidateShort)
			t.Log(string(resp))
			t.Log(string(stdOrder))

			if liquidateShort.Status == ORDER_FINISH {
				isDeal = true
				break
			}
		}
	}

	if !isDeal {
		t.Error("The liquidate order hasn't deal. ")
		return
	}

}
