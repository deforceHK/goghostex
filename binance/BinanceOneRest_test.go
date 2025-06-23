package binance

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	. "github.com/deforceHK/goghostex"
)

const (
	ONE_API_KEY        = ""
	ONE_API_SECRETKEY  = ""
	ONE_API_PASSPHRASE = ""
	ONE_PROXY_URL      = ""
)

// go test -v ./binance/... -count=1 -run=TestOne_GetInfos
func TestOne_GetInfos(t *testing.T) {

	config := &APIConfig{
		Endpoint:   ENDPOINT,
		HttpClient: &http.Client{
			//Transport: &http.Transport{
			//	Proxy: func(req *http.Request) (*url.URL, error) {
			//		return url.Parse(ONE_PROXY_URL)
			//	},
			//},
		},
		ApiKey:        ONE_API_KEY,
		ApiSecretKey:  ONE_API_SECRETKEY,
		ApiPassphrase: ONE_API_PASSPHRASE,
		Location:      time.Now().Location(),
	}

	bn := New(config)

	var infos, _, err = bn.One.GetCMInfos()
	if err != nil {
		t.Error(err)
		return
	}

	for _, info := range infos {
		t.Log(*info)
	}

	infos, _, err = bn.One.GetUMInfos()
	if err != nil {
		t.Error(err)
		return
	}

	for _, info := range infos {
		t.Log(*info)
	}
}

// go test -v ./binance/... -count=1 -run=TestOne_PlaceOrder
func TestOne_PlaceOrder(t *testing.T) {
	config := &APIConfig{
		Endpoint:   "https://papi.binance.com",
		HttpClient: &http.Client{
			//Transport: &http.Transport{
			//	Proxy: func(req *http.Request) (*url.URL, error) {
			//		return url.Parse(ONE_PROXY_URL)
			//	},
			//},
		},
		ApiKey:        ONE_API_KEY,
		ApiSecretKey:  ONE_API_SECRETKEY,
		ApiPassphrase: ONE_API_PASSPHRASE,
		Location:      time.Now().Location(),
	}

	bn := New(config)

	// We can use a coin-margined pair for test, e.g. BTC/USD perpetual
	var pair = Pair{
		Basis:   BTC,
		Counter: USDT,
	}

	// 1. Get Ticker to get the latest price for placing an order.
	// Assumes bn.One has GetTicker method.
	//ticker, _, err := bn.One.GetTicker(pair)
	//if err != nil {
	//	t.Errorf("Failed to get ticker: %v", err)
	//	return
	//}
	//t.Logf("Got ticker for %s, last price: %f", pair.String(), ticker.Last)

	// 2. Create a limit sell order.
	// The price is 10% higher than the latest price to ensure it won't be filled immediately.
	var order = OneOrder{
		Cid:        UUID(),
		Pair:       pair,
		ProductId:  "BTCUSDT",
		Amount:     0.001, // Assuming trading 1 contract
		PlaceType:  MARKET,
		Type:       OPEN_LONG,
		LeverRate:  20, // 20x leverage
		SettleMode: 2,
	}

	// 3. Place the order
	if resp, err := bn.One.PlaceOrder(&order); err != nil {
		t.Errorf("PlaceOrder failed: %v", err)
		return
	} else {
		standard, _ := json.Marshal(order)
		t.Log("Order standard struct:")
		t.Log(string(standard))
		t.Log("PlaceOrder remote api response: ")
		t.Log(string(resp))
	}
	t.Logf("Successfully placed order with ID: %s", order.Cid)

	//// 4. Get the order to check status
	//if resp, err := bn.One.GetOrder(&order); err != nil {
	//	t.Errorf("GetOrder failed: %v", err)
	//	return
	//} else {
	//	standard, _ := json.Marshal(order)
	//	t.Log("GetOrder standard struct:")
	//	t.Log(string(standard))
	//	t.Log("GetOrder remote api response: ")
	//	t.Log(string(resp))
	//}
	//t.Logf("Successfully get order %s, status is %s", order.OrderID, order.Status)
	//
	//// 5. Cancel the order
	//if resp, err := bn.One.CancelOrder(&order); err != nil {
	//	t.Errorf("CancelOrder failed: %v", err)
	//	return
	//} else {
	//	standard, _ := json.Marshal(order)
	//	t.Log("CancelOrder standard struct:")
	//	t.Log(string(standard))
	//	t.Log("CancelOrder remote api response: ")
	//	t.Log(string(resp))
	//}
	//t.Logf("Successfully canceled order %s", order.OrderID)
}
