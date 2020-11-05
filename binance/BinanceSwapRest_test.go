package binance

import (
	"encoding/json"
	"net/http"
	"net/url"
	"testing"
	"time"

	. "github.com/strengthening/goghostex"
)

const (
	SWAP_API_KEY       = ""
	SWAP_API_SECRETKEY = ""
	SWAP_PROXY_URL     = "socks5://127.0.0.1:1090"
)

func TestSwap_MarketAPI(t *testing.T) {

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
		ApiPassphrase: "",
		Location:      time.Now().Location(),
	}

	bn := New(config)
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
		20,
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
			20,
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

	if highest, lowest, err := bn.Swap.GetLimit(Pair{Basis: BTC, Counter: USDT}); err != nil {
		t.Error(err)
		return
	} else {
		t.Log("highest:", highest)
		t.Log("lowest:", lowest)
	}

	if klines, resp, err := bn.Swap.GetKline(
		Pair{Basis: BTC, Counter: USDT},
		KLINE_PERIOD_1DAY,
		20,
		1271149752000,
	); err != nil {
		t.Error(err)
		return
	} else {
		klineRaw, _ := json.Marshal(klines)
		t.Log(string(klineRaw))
		t.Log(string(resp))
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
}

func TestSwap_Account(t *testing.T) {

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
	if items, raw, err := bn.Swap.GetAccountFlow(); err != nil {
		t.Error(err)
		return
	} else {
		t.Log(string(raw))

		for _, i := range items {
			t.Log(*i)
		}

		//rawAccount, _ := json.Marshal(account)
		//t.Log(string(rawAccount))
		//t.Log(string(raw))
		//
		//if account.BalanceAvail < 1 {
		//	t.Error("There have no enough asset to trade. ")
		//	return
		//}
	}
}

// must set both
// place the order ---> get the order info ---> cancel the order -> get the order info
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
		ApiPassphrase: "",
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
		Amount:    0.012,
		PlaceType: NORMAL,
		Type:      OPEN_SHORT,
		LeverRate: 20,
		Pair:      pair,
		Exchange:  BINANCE,
	}

	orderLong := SwapOrder{
		Cid:       UUID(),
		Price:     ticker.Buy * 0.97,
		Amount:    0.012,
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

	//if orders, resp, err := bn.Swap.GetOrders(Pair{LTC, USDT}); err != nil {
	//	t.Error(err)
	//	return
	//} else {
	//	raw, _ := json.Marshal(orders)
	//	t.Log(string(raw))
	//	t.Log(string(resp))
	//}

	//if orders, resp, err := bn.Swap.GetUnFinishOrders(Pair{LTC, USDT}); err != nil {
	//	t.Error(err)
	//	return
	//} else {
	//
	//	raw, _ := json.Marshal(orders)
	//	t.Log(string(raw))
	//	t.Log(string(resp))
	//}
	//
	//if s, resp, err := bn.Swap.GetPosition(Pair{LTC, USDT}, OPEN_LONG); err != nil {
	//	t.Error(err)
	//	return
	//} else {
	//	t.Log(string(resp))
	//	t.Log(s)
	//}
}
