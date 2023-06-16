package gate

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"testing"
	"time"

	. "github.com/deforceHK/goghostex"
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
		ApiKey:       "",
		ApiSecretKey: "",
		Location:     time.Now().Location(),
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
	//	fmt.Print(string(resp))
	//}

	//if klines, resp, err := gateCli.Swap.GetKline(
	//	Pair{Currency{"AMPL", ""}, USDT},
	//	KLINE_PERIOD_1MIN,
	//	50,
	//	0,
	//); err != nil {
	//	t.Error(err)
	//	return
	//} else {
	//	body, _ := json.Marshal(klines)
	//	fmt.Print(string(body))
	//	fmt.Print(string(resp))
	//}

	//if fees, resp, err := gateCli.Swap.GetFundingFees(
	//	Pair{Currency{"AMPL", ""}, USDT},
	//); err != nil {
	//	t.Error(err)
	//	return
	//} else {
	//	body, _ := json.Marshal(fees)
	//	fmt.Print(string(body))
	//	fmt.Print(string(resp))
	//}

	//if fees, err := gateCli.Swap.GetFundingFee(
	//	Pair{Currency{"AMPL", ""}, USDT},
	//); err != nil {
	//	t.Error(err)
	//	return
	//} else {
	//	body, _ := json.Marshal(fees)
	//	fmt.Print(string(body))
	//	//fmt.Print(string(resp))
	//}

	if _, resp, err := gateCli.Swap.GetAccount(); err != nil {
		t.Error(err)
		return
	} else {
		//body, _ := json.Marshal(fees)
		fmt.Print(string(resp))
		//fmt.Print(string(resp))
	}
}

func TestSwap_TradeAPI(t *testing.T) {
	config := &APIConfig{
		Endpoint: "",
		HttpClient: &http.Client{
			Transport: &http.Transport{
				Proxy: func(req *http.Request) (*url.URL, error) {
					return url.Parse("socks5://127.0.0.1:1090")
				},
			},
		},
		ApiKey:       "",
		ApiSecretKey: "",
		Location:     time.Now().Location(),
	}

	gateCli := New(config)
	pair := BTC_USDT
	rule, _, err := gateCli.Swap.GetTicker(pair)
	if err != nil {
		t.Error(err)
		return
	}

	order := &SwapOrder{
		Price:     rule.Last * 1.1,
		Amount:    0.001,
		PlaceType: NORMAL,
		Type:      OPEN_LONG,
		Pair:      pair,
		Exchange:  GATE,
	}

	if resp, err := gateCli.Swap.PlaceOrder(order); err != nil {
		t.Error(err)
		return
	} else {
		fmt.Println("~~~~~~~~~~~~")
		fmt.Println(string(resp))
		fmt.Println(order)
	}

	if resp, err := gateCli.Swap.GetOrder(order); err != nil {
		t.Error(err)
		return
	} else {
		fmt.Println("~~~~~~~~~~~~")
		fmt.Println(string(resp))
		fmt.Println(order)

		orderRaw, _ := json.Marshal(order)
		fmt.Println(string(orderRaw))
	}

	//time.Sleep(5*time.Second)
	//if resp, err := gateCli.Swap.GetOrder(order); err != nil {
	//	t.Error(err)
	//	return
	//} else {
	//	fmt.Println("~~~~~~~~~~~~")
	//	fmt.Println(string(resp))
	//	fmt.Println(err)
	//}
}
