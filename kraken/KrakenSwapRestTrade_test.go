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
* go test -v ./kraken/... -count=1 -run=TestSwap_GetOrder
*
**/
func TestSwap_GetOrder(t *testing.T) {

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

	var orderLong = SwapOrder{
		Cid:       UUID(),
		Price:     58888,
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
}
