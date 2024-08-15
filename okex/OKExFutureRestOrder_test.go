package okex

import (
	"encoding/json"
	"net/http"
	"net/url"
	"testing"
	"time"

	. "github.com/deforceHK/goghostex"
)

/**
 *
 * The func of order unit test step is:
 * 1. Get the account, and find have the enough crypto. (removed)
 * 2. Get BTC-USD this_week ticker.
 * 2. Order the open_long without deal.
 * 3. Get the unfinished orders info, and find the order in step 1.
 * 4. Get the order info.
 * 5. Cancel the open_long Order
 *
 * unit test cmd
 * go test -v ./okex/... -count=1 -run=TestFuture_TradeAPI
 **/

func TestFuture_TradeAPI(t *testing.T) {

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
	//if account, resp, err := ok.Future.GetAccount(); err != nil {
	//	t.Error(err)
	//	return
	//} else {
	//
	//	t.Log("Future account standard struct: ")
	//	t.Log(*account)
	//
	//	t.Log("Future account remote api struct: ")
	//	t.Log(string(resp))
	//
	//	if account.SubAccount[LTC].BalanceAvail <= 1 {
	//		t.Error("You do not have enough LTC for test. ")
	//		return
	//	}
	//}

	ticker, _, err := ok.Future.GetTicker(
		Pair{Basis: BTC, Counter: USD},
		THIS_WEEK_CONTRACT,
	)
	if err != nil {
		t.Error(err)
		return
	}

	order := FutureOrder{
		Cid:          UUID(),
		Price:        ticker.Last * 1.03,
		Amount:       1,
		PlaceType:    NORMAL,
		Type:         OPEN_SHORT,
		LeverRate:    20,
		Pair:         Pair{Basis: BTC, Counter: USD},
		ContractType: THIS_WEEK_CONTRACT,
		Exchange:     OKEX,
	}
	preCid := order.Cid

	if resp, err := ok.Future.PlaceOrder(&order); err != nil {
		t.Error(err)
		return
	} else {
		if standard, err := json.Marshal(order); err != nil {
			t.Error(err)
			return
		} else {
			t.Log("Place Order standard struct: ")
			t.Log(string(standard))

			t.Log("Place Order remote api struct: ")
			t.Log(string(resp))
		}

		if preCid != order.Cid {
			t.Error("The cid is not same in the api. ")
			return
		}

		if order.OrderId == "" {
			t.Error("The order_id can not be empty string. ")
			return
		}
	}

	if resp, err := ok.Future.GetOrder(&order); err != nil {
		t.Error(err)
		return
	} else {

		if standard, err := json.Marshal(order); err != nil {
			t.Error(err)
			return
		} else {
			t.Log("Get Order standard struct: ")
			t.Log(string(standard))

			t.Log("Get Order remote api struct: ")
			t.Log(string(resp))
		}
	}

	if resp, err := ok.Future.CancelOrder(&order); err != nil {
		t.Error(err)
		return
	} else {
		if standard, err := json.Marshal(order); err != nil {
			t.Error(err)
			return
		} else {
			t.Log("Cancel Order standard struct: ")
			t.Log(string(standard))

			t.Log("Cancel Order remote api struct: ")
			t.Log(string(resp))
		}
	}

	for i := 0; i < 3; i++ {
		if resp, err := ok.Future.GetOrder(&order); err != nil {
			t.Error(err)
			return
		} else {

			if standard, err := json.Marshal(order); err != nil {
				t.Error(err)
				return
			} else {
				t.Log("Get Order after standard struct: ")
				t.Log(string(standard))

				t.Log("Get Order after remote api struct: ")
				t.Log(string(resp))
			}

			if order.Status == ORDER_CANCEL {
				break
			}
		}
	}

	if order.Status != ORDER_CANCEL {
		t.Error("The order must be canceled. ")
		return
	}

	//onlyMakerOrder := FutureOrder{
	//	Cid:          UUID(),
	//	Price:        ticker.Last * 0.99,
	//	Amount:       1,
	//	PlaceType:    ONLY_MAKER,
	//	Type:         OPEN_SHORT,
	//	LeverRate:    20,
	//	Pair:     Pair{LTC, USD},
	//	ContractType: THIS_WEEK_CONTRACT,
	//	MatchPrice:   0,
	//}
	//
	//if resp, err := ok.Future.PlaceFutureOrder(&onlyMakerOrder); err != nil {
	//	t.Error(err)
	//	return
	//} else {
	//
	//	if standard, err := json.Marshal(onlyMakerOrder); err != nil {
	//		t.Error(err)
	//		return
	//	} else {
	//		t.Log("Place only maker Order standard struct: ")
	//		t.Log(string(standard))
	//
	//		t.Log("Place only maker Order remote api struct: ")
	//		t.Log(string(resp))
	//	}
	//}
	//
	//if resp, err := ok.Future.GetOrder(&onlyMakerOrder); err != nil {
	//	t.Error(err)
	//	return
	//} else {
	//
	//	if standard, err := json.Marshal(onlyMakerOrder); err != nil {
	//		t.Error(err)
	//		return
	//	} else {
	//		t.Log("Get only maker Order after standard struct: ")
	//		t.Log(string(standard))
	//
	//		t.Log("Get only maker Order after remote api struct: ")
	//		t.Log(string(resp))
	//	}
	//
	//	if onlyMakerOrder.Status != ORDER_CANCEL {
	//		t.Error("The only maker order must bi canceled. ")
	//		return
	//	}
	//}

}

/**
 *
 * unit test cmd
 * go test -v ./okex/... -count=1 -run=TestFuture_DealAPI
 *
 **/

func TestFuture_DealAPI(t *testing.T) {

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
	if account, resp, err := ok.Future.GetAccount(); err != nil {
		t.Error(err)
		return
	} else {
		t.Log("Future account standard struct: ")
		raw, err := json.Marshal(account.SubAccount[BTC])
		if err != nil {
			t.Log(err)
			return
		}
		t.Log(string(raw))

		t.Log("Future account remote api struct: ")
		t.Log(string(resp))

		if account.SubAccount[BTC].BalanceAvail <= 0 {
			t.Error("You do not have enough BTC for test. ")
			return
		}
	}

	ticker, _, err := ok.Future.GetTicker(
		Pair{Basis: BTC, Counter: USD},
		THIS_WEEK_CONTRACT,
	)
	if err != nil {
		t.Error(err)
		return
	}

	orderShort := FutureOrder{
		Cid:          UUID(),
		Price:        ticker.Last * 0.99,
		Amount:       1,
		PlaceType:    NORMAL,
		Type:         OPEN_SHORT,
		LeverRate:    20,
		Pair:         Pair{Basis: BTC, Counter: USD},
		ContractType: THIS_WEEK_CONTRACT,
		Exchange:     OKEX,
	}

	orderLiquidate := FutureOrder{
		Cid:          UUID(),
		Price:        ticker.Last * 1.01,
		Amount:       1,
		PlaceType:    NORMAL,
		Type:         LIQUIDATE_SHORT,
		LeverRate:    20,
		Pair:         Pair{Basis: BTC, Counter: USD},
		ContractType: THIS_WEEK_CONTRACT,
		Exchange:     OKEX,
	}

	if resp, err := ok.Future.PlaceOrder(&orderShort); err != nil {
		t.Error(err)
		return
	} else {
		if standard, err := json.Marshal(orderShort); err != nil {
			t.Error(err)
			return
		} else {
			t.Log("Place Order standard struct: ")
			t.Log(string(standard))

			t.Log("Place Order remote api struct: ")
			t.Log(string(resp))
		}
	}

	for i := 0; ; i++ {
		if i > 5 {
			t.Error("too many time try. ")
			return
		}
		time.Sleep(time.Second)

		if resp, err := ok.Future.GetOrder(&orderShort); err != nil {
			t.Error(err)
			return
		} else {

			if standard, err := json.Marshal(orderShort); err != nil {
				t.Error(err)
				return
			} else {
				t.Log("Get Order standard struct: ")
				t.Log(string(standard))

				t.Log("Get Order remote api struct: ")
				t.Log(string(resp))
			}
			if orderShort.Status == ORDER_FINISH {
				orderLiquidate.Price = orderShort.AvgPrice * 1.01
				break
			}
		}
	}

	if resp, err := ok.Future.PlaceOrder(&orderLiquidate); err != nil {
		t.Error(err)
		return
	} else {
		if standard, err := json.Marshal(orderLiquidate); err != nil {
			t.Error(err)
			return
		} else {
			t.Log("Place liquidate Order standard struct: ")
			t.Log(string(standard))

			t.Log("Place liquidate Order remote api struct: ")
			t.Log(string(resp))
		}
	}

	for i := 0; ; i++ {
		if i > 5 {
			t.Error("too many time try. ")
			return
		}
		time.Sleep(time.Second)
		if resp, err := ok.Future.GetOrder(&orderLiquidate); err != nil {
			t.Error(err)
			return
		} else {
			if standard, err := json.Marshal(orderLiquidate); err != nil {
				t.Error(err)
				return
			} else {
				t.Log("Get Order after standard struct: ")
				t.Log(string(standard))
				t.Log("Get Order after remote api struct: ")
				t.Log(string(resp))
			}
			if orderShort.Status == ORDER_FINISH {
				break
			}
		}
	}

}
