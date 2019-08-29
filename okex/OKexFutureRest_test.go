package okex

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"
	"time"

	. "github.com/strengthening/goghostex"
)

//var api APIConfig
//
//func init() {
//
//	dialer, err := proxy.SOCKS5(
//		"tcp",
//		"127.0.0.1:1090",
//		nil,
//		&net.Dialer{
//			Timeout:   10 * time.Second,
//			KeepAlive: 10 * time.Second,
//		},
//	)
//	if err != nil {
//		log.Fatalln("get dialer error", dialer)
//	}
//
//	httpTransport := &http.Transport{Dial: dialer.Dial}
//	api = APIConfig{
//		HttpClient:    &http.Client{Transport: httpTransport},
//		Endpoint:      ENDPOINT,
//		ApiKey:        "",
//		ApiSecretKey:  "",
//		ApiPassphrase: "",
//		Location:      time.Now().Location(),
//	}
//}

const (
	FUTURE_API_KEY        = ""
	FUTURE_API_SECRETKEY  = ""
	FUTURE_API_PASSPHRASE = ""
)

func TestFuture_MarketAPI(t *testing.T) {

	config := &APIConfig{
		Endpoint: ENDPOINT,
		HttpClient: &http.Client{
			Transport: &http.Transport{
				Proxy: func(req *http.Request) (*url.URL, error) {
					return url.Parse(PROXY_URL)
				},
			},
		},
		ApiKey:        FUTURE_API_KEY,
		ApiSecretKey:  FUTURE_API_SECRETKEY,
		ApiPassphrase: FUTURE_API_PASSPHRASE,
		Location:      time.Now().Location(),
	}

	ok := New(config)

	if ticker, body, err := ok.Future.GetFutureTicker(
		CurrencyPair{BTC, USD},
		THIS_WEEK_CONTRACT,
	); err != nil {
		t.Error(err)
		return
	} else {
		fmt.Println(*ticker)
		fmt.Println(string(body))
	}

}

//func TestOKExFuture_GetRate(t *testing.T) {
//	ok := New(&api)
//
//	if result, _, err := ok.Future.GetRate(); err != nil {
//		t.Error(err)
//		return
//	} else {
//		fmt.Println(result)
//	}
//}
//
//func TestOKExFuture_GetFutureEstimatedPrice(t *testing.T) {
//	ok := New(&api)
//	p := CurrencyPair{BTC, USD}
//	if result, resp, err := ok.Future.GetFutureEstimatedPrice(p); err != nil {
//		t.Error(err)
//		return
//	} else {
//		fmt.Println(result)
//		fmt.Println(string(resp))
//	}
//}
//
//func TestOKExFuture_GetFutureContractInfo(t *testing.T) {
//	ok := New(&api)
//	if result, resp, err := ok.Future.GetFutureContractInfo(); err != nil {
//		t.Error(err)
//		return
//	} else {
//		fmt.Println(result)
//		fmt.Println(string(resp))
//	}
//}
//
//func TestOKExFuture_GetFutureTicker(t *testing.T) {
//	ok := New(&api)
//
//	p := CurrencyPair{BTC, USD}
//	if result, resp, err := ok.Future.GetFutureTicker(p, "quarter"); err != nil {
//		t.Error(err)
//		return
//	} else {
//		fmt.Println(string(resp))
//		fmt.Println(*result)
//	}
//}
//
//func TestOKExFuture_GetFutureDepth(t *testing.T) {
//
//	ok := New(&api)
//	if result, resp, err := ok.Future.GetFutureDepth(
//		BTC_USD,
//		"next_week",
//		200,
//	); err != nil {
//		t.Error(err)
//		return
//	} else {
//		fmt.Println(string(resp))
//		fmt.Println(*result)
//	}
//
//}
//
//func TestOKExFuture_GetFutureIndex(t *testing.T) {
//
//	ok := New(&api)
//	if result, resp, err := ok.Future.GetFutureIndex(
//		BTC_USD,
//	); err != nil {
//		t.Error(err)
//		return
//	} else {
//		fmt.Println(string(resp))
//		fmt.Println(result)
//	}
//}
//
//func TestOKexFuture_GetFutureUserinfo(t *testing.T) {
//	ok := New(&api)
//	if result, resp, err := ok.Future.GetFutureUserinfo(); err != nil {
//		t.Error(err)
//		return
//	} else {
//		fmt.Println(string(resp))
//		fmt.Println(*result)
//	}
//}
//
//func TestOKexFuture_GetUnfinishFutureOrders(t *testing.T) {
//
//	ok := New(&api)
//	p := CurrencyPair{BTC, USD}
//	if result, resp, err := ok.Future.GetUnFinishFutureOrders(
//		p,
//		"quarter",
//	); err != nil {
//		t.Error(err)
//		return
//	} else {
//		fmt.Println(string(resp))
//		fmt.Println(result)
//	}
//}
//
//func TestOKExFuture_PlaceFutureOrder(t *testing.T) {
//	ok := New(&api)
//
//	p := CurrencyPair{ETC, USD}
//	ticker, _, err := ok.Future.GetFutureTicker(p, "quarter")
//	//.GetFutureTicker(p)
//	if err != nil {
//		t.Error(err)
//		return
//	}
//
//	fOrder := &FutureOrder{
//		ClientOid:    "12a31231",
//		Price:        ticker.Last * 0.95,
//		Amount:       1,
//		Currency:     p,
//		OrderType:    0,
//		OType:        1,
//		LeverRate:    20,
//		ContractName: "quarter",
//	}
//
//	if result, resp, err := ok.Future.PlaceFutureOrder(0, fOrder); err != nil || !result {
//		t.Error(err)
//		return
//	} else {
//		fmt.Println(string(resp))
//	}
//
//	rst, resp, err := ok.Future.CancelFutureOrder(fOrder)
//	if err != nil || !rst {
//		t.Error(err)
//		return
//	} else {
//		fmt.Println(string(resp))
//		fmt.Println(*fOrder)
//	}
//}
//
//// test the maker only order
//func TestOKExFuture_PlaceFutureOrder_MakerOnly(t *testing.T) {
//
//	ok := New(&api)
//	ticker, _, err := ok.Future.GetFutureTicker(CurrencyPair{ETC, USD}, "quarter")
//	if err != nil {
//		t.Error(err)
//		return
//	}
//
//	fOrder := &FutureOrder{
//		ClientOid:    UUID(),
//		Price:        ticker.Last * 1.03,
//		Amount:       1,
//		Currency:     CurrencyPair{ETC, USD},
//		OrderType:    1,
//		OType:        1,
//		LeverRate:    20,
//		ContractName: "quarter",
//	}
//
//	if result, resp, err := ok.Future.PlaceFutureOrder(0, fOrder); err != nil {
//		t.Error(err)
//		return
//	} else {
//		fmt.Println(result)
//		fmt.Println(string(resp))
//		order, _ := json.Marshal(*fOrder)
//		fmt.Println(string(order))
//	}
//
//	time.Sleep(5 * time.Second)
//
//	if resp, err := ok.Future.GetFutureOrder(fOrder); err != nil {
//		t.Error(err)
//		return
//	} else {
//		fmt.Println(string(resp))
//	}
//
//	if fOrder.Status != ORDER_CANCEL {
//		rst, resp, err := ok.Future.CancelFutureOrder(fOrder)
//		if err != nil {
//			t.Error(err)
//			return
//		} else {
//			fmt.Println(rst)
//			fmt.Println(string(resp))
//			fmt.Println(*fOrder)
//		}
//	}
//
//}
//
//// test the fok order
//func TestOKExFuture_PlaceFutureOrder_FOK(t *testing.T) {
//
//	ok := New(&api)
//	ticker, _, err := ok.Future.GetFutureTicker(
//		CurrencyPair{ETC, USD},
//		"quarter",
//	)
//	if err != nil {
//		t.Error(err)
//		return
//	}
//
//	openOrder := &FutureOrder{
//		ClientOid:    UUID(),
//		Price:        ticker.Last * 1.02,
//		Amount:       1,
//		Currency:     CurrencyPair{ETC, USD},
//		OrderType:    2,
//		OType:        1,
//		LeverRate:    20,
//		ContractName: "quarter",
//	}
//
//	if result, resp, err := ok.Future.PlaceFutureOrder(0, openOrder); err != nil {
//		t.Error(err)
//		return
//	} else {
//		fmt.Println(result)
//		fmt.Println(string(resp))
//		order, _ := json.Marshal(*openOrder)
//		fmt.Println(string(order))
//	}
//
//	time.Sleep(5 * time.Second)
//
//	if resp, err := ok.Future.GetFutureOrder(openOrder); err != nil {
//		t.Error(err)
//		return
//	} else {
//		fmt.Println(string(resp))
//	}
//
//	if openOrder.Status != ORDER_CANCEL && openOrder.Status != ORDER_FINISH {
//		rst, resp, err := ok.Future.CancelFutureOrder(openOrder)
//		if err != nil {
//			t.Error(err)
//			return
//		} else {
//			fmt.Println(rst)
//			fmt.Println(string(resp))
//			fmt.Println(*openOrder)
//		}
//	}
//
//	if openOrder.DealAmount > 0 {
//		liquidateOrder := &FutureOrder{
//			ClientOid:    UUID(),
//			Price:        openOrder.AvgPrice * 0.98,
//			Amount:       openOrder.DealAmount,
//			Currency:     openOrder.Currency,
//			OrderType:    0,
//			OType:        3,
//			LeverRate:    openOrder.LeverRate,
//			ContractName: openOrder.ContractName,
//		}
//
//		if _, resp, err := ok.Future.PlaceFutureOrder(0, liquidateOrder); err != nil {
//			t.Error(err)
//			return
//		} else {
//			fmt.Println(string(resp))
//		}
//	}
//
//}
