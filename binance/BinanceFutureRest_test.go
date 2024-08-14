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
	FUTURE_API_KEY        = ""
	FUTURE_API_SECRETKEY  = ""
	FUTURE_API_PASSPHRASE = ""
	FUTURE_PROXY_URL      = "socks5://127.0.0.1:1090"
)

/**
* unit test cmd
* go test -v ./binance/... -count=1 -run=TestFuture_MarketAPI
*
**/

func TestFuture_MarketAPI(t *testing.T) {

	var client = &http.Client{}
	if FUTURE_PROXY_URL != "" {
		client = &http.Client{
			Transport: &http.Transport{
				Proxy: func(req *http.Request) (*url.URL, error) {
					return url.Parse(FUTURE_PROXY_URL)
				},
			},
		}
	}

	config := &APIConfig{
		Endpoint:      ENDPOINT,
		HttpClient:    client,
		ApiKey:        FUTURE_API_KEY,
		ApiSecretKey:  FUTURE_API_SECRETKEY,
		ApiPassphrase: FUTURE_API_PASSPHRASE,
		Location:      time.Now().Location(),
	}

	bn := New(config)
	// ticker unit test
	if ticker, resp, err := bn.Future.GetTicker(Pair{Basis: BTC, Counter: USD}, NEXT_QUARTER_CONTRACT); err != nil {
		t.Error(err)
		return
	} else {
		t.Log(ticker)
		t.Log(string(resp))
	}

	if depth, resp, err := bn.Future.GetDepth(Pair{Basis: BTC, Counter: USD}, NEXT_QUARTER_CONTRACT, 5); err != nil {
		t.Error(err)
		return
	} else {
		depthRaw, _ := json.Marshal(depth)
		t.Log(string(depthRaw))
		t.Log(string(resp))
	}

	if max, min, err := bn.Future.GetLimit(Pair{Basis: BTC, Counter: USD}, NEXT_QUARTER_CONTRACT); err != nil {
		t.Error(err)
		return
	} else {
		t.Log(max, min)
	}

	if klineList, resp, err := bn.Future.GetKlineRecords(
		NEXT_QUARTER_CONTRACT,
		Pair{Basis: ETH, Counter: USD},
		KLINE_PERIOD_1MIN,
		200,
		1609430400000,
	); err != nil {
		t.Error(err)
		return
	} else {
		raw, _ := json.Marshal(klineList)
		t.Log(string(raw))
		t.Log(string(resp))
	}

	if trades, _, err := bn.Future.GetTrades(
		Pair{Basis: BTC, Counter: USD},
		NEXT_QUARTER_CONTRACT,
	); err != nil {
		t.Error(err)
		return
	} else {
		stdTrades, _ := json.Marshal(trades)
		t.Log(string(stdTrades))
	}

	if contract, err := bn.Future.GetContract(Pair{ETH, USD}, NEXT_QUARTER_CONTRACT); err != nil {
		t.Error(err)
		return
	} else {
		t.Log(contract)
	}

}

/**
*
* unit test cmd
* make sure you have enough bnb in account, then go to the shell.
* go test -v ./binance/... -count=1 -run=TestFuture_TradeAPI
*
**/

func TestFuture_TradeAPI(t *testing.T) {

	config := &APIConfig{
		Endpoint: ENDPOINT,
		HttpClient: &http.Client{
			Transport: &http.Transport{
				Proxy: func(req *http.Request) (*url.URL, error) {
					return url.Parse(FUTURE_PROXY_URL)
				},
			},
		},
		ApiKey:        FUTURE_API_KEY,
		ApiSecretKey:  FUTURE_API_SECRETKEY,
		ApiPassphrase: FUTURE_API_PASSPHRASE,
		Location:      time.Now().Location(),
	}

	bn := New(config)
	if account, _, err := bn.Future.GetAccount(); err != nil {
		t.Error(err)
		return
	} else {
		t.Log(account)
		//t.Log(string(resp))

		r, _ := json.Marshal(account.SubAccount[BNB])
		t.Log(string(r))
		if account.SubAccount[BTC].Margin <= 0.01 {
			t.Error("There is no enough bnb to continue. ")
		}
	}

	pair := Pair{BTC, USD}
	if ticker, _, err := bn.Future.GetTicker(pair, QUARTER_CONTRACT); err != nil {
		t.Error(err)
		return
	} else {

		undealOrder := &FutureOrder{
			Price:        ticker.Last * 0.96,
			Amount:       1,
			PlaceType:    NORMAL,
			Type:         OPEN_LONG,
			LeverRate:    20,
			Pair:         pair,
			ContractType: QUARTER_CONTRACT,
			Exchange:     BINANCE,
		}

		if resp, err := bn.Future.PlaceOrder(undealOrder); err != nil {
			t.Error(err)
			return
		} else {
			t.Log(string(resp))

			o, _ := json.Marshal(undealOrder)
			t.Log(string(o))
		}

		if resp, err := bn.Future.CancelOrder(undealOrder); err != nil {
			t.Error(err)
			return
		} else {
			t.Log(string(resp))
			o, _ := json.Marshal(undealOrder)
			t.Log(string(o))
		}

		time.Sleep(5 * time.Second)

		longOrder := &FutureOrder{
			Price:        ticker.Last * 1.04,
			Amount:       1,
			PlaceType:    NORMAL,
			Type:         OPEN_LONG,
			LeverRate:    20,
			Pair:         pair,
			ContractType: QUARTER_CONTRACT,
			Exchange:     BINANCE,
		}
		if _, err := bn.Future.PlaceOrder(longOrder); err != nil {
			t.Error(err)
			return
		} else {
			if _, err := bn.Future.GetOrder(longOrder); err != nil {
				t.Error(err)
				return
			}

			o, _ := json.Marshal(longOrder)
			t.Log(string(o))
		}

		liquidateOrder := &FutureOrder{
			Price:        ticker.Last * 0.96,
			Amount:       1,
			PlaceType:    NORMAL,
			Type:         LIQUIDATE_LONG,
			LeverRate:    20,
			Pair:         pair,
			ContractType: QUARTER_CONTRACT,
			Exchange:     BINANCE,
		}
		if _, err := bn.Future.PlaceOrder(liquidateOrder); err != nil {
			t.Error(err)
			return
		} else {
			if _, err := bn.Future.GetOrder(liquidateOrder); err != nil {
				t.Error(err)
				return
			}
			o, _ := json.Marshal(liquidateOrder)
			t.Log(string(o))
		}

	}
}

/**
* unit test cmd
* go test -v ./binance/... -count=1 -run=TestFuture_AccountAPI
*
**/

func TestFuture_AccountAPI(t *testing.T) {

	var config = &APIConfig{
		Endpoint: ENDPOINT,
		HttpClient: &http.Client{
			Transport: &http.Transport{
				Proxy: func(req *http.Request) (*url.URL, error) {
					return url.Parse(FUTURE_PROXY_URL)
				},
			},
		},
		ApiKey:        FUTURE_API_KEY,
		ApiSecretKey:  FUTURE_API_SECRETKEY,
		ApiPassphrase: FUTURE_API_PASSPHRASE,
		Location:      time.Now().Location(),
	}

	var bn = New(config)
	if flow, resp, err := bn.Future.GetPairFlow(Pair{BTC, USD}); err != nil {
		t.Error(err)
		return
	} else {
		t.Log(string(resp))
		sss, _ := json.Marshal(flow)
		t.Log(string(sss))
	}

}
