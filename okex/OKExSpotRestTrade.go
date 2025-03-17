package okex

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	. "github.com/deforceHK/goghostex"
)

var _INERNAL_V5_SPOT_TRADE_SIDE_CONVERTER = map[TradeSide]string{
	BUY:  "buy",
	SELL: "sell",
	//BU:        "fok",
	//IOC:        "ioc",
	//MARKET:     "market",
}

var _INERNAL_V5_SPOT_PLACE_TYPE_CONVERTER = map[PlaceType]string{
	NORMAL:     "limit",
	ONLY_MAKER: "post_only",
	FOK:        "fok",
	IOC:        "ioc",
	MARKET:     "market",
}

func (spot *Spot) PlaceOrder(order *Order) ([]byte, error) {
	var instrument = spot.getInstruments(order.Pair)
	var request = struct {
		InstId  string `json:"instId"`
		TdMode  string `json:"tdMode"`
		Side    string `json:"side"`
		PosSide string `json:"posSide,omitempty"`
		OrdType string `json:"ordType"`
		Sz      string `json:"sz"`
		Px      string `json:"px,omitempty"`
		ClOrdId string `json:"clOrdId,omitempty"`
		TgtCcy  string `json:"tgtCcy,omitempty"`
	}{}

	request.InstId = instrument.InstId
	request.TdMode = "cross"
	request.Side = _INERNAL_V5_SPOT_TRADE_SIDE_CONVERTER[order.Side]
	request.OrdType = _INERNAL_V5_SPOT_PLACE_TYPE_CONVERTER[order.OrderType]
	request.Sz = FloatToString(order.Amount, instrument.AmountPrecision)
	request.Px = FloatToString(order.Price, instrument.PricePrecision)
	request.ClOrdId = order.Cid
	request.TgtCcy = "base_ccy"

	var response = struct {
		Code string `json:"code"`
		Msg  string `json:"msg"`
		Data []struct {
			ClOrdId string `json:"clOrdId"`
			OrdId   string `json:"ordId"`
			SCode   string `json:"sCode"`
			SMsg    string `json:"sMsg"`
		} `json:"data"`
	}{}
	var uri = "/api/v5/trade/order"

	now := time.Now()
	order.PlaceTimestamp = now.UnixNano() / int64(time.Millisecond)
	order.PlaceDatetime = now.In(spot.config.Location).Format(GO_BIRTHDAY)
	reqBody, _, _ := spot.BuildRequestBody(request)
	resp, err := spot.DoRequest(
		http.MethodPost,
		uri,
		reqBody,
		&response,
	)

	if err != nil {
		return resp, err
	}
	if len(response.Data) > 0 && response.Data[0].SCode != "0" {
		return resp, errors.New(string(resp)) // very important cause it has the error code
	}
	if response.Code != "0" {
		return resp, errors.New(string(resp)) // very important cause it has the error code
	}

	now = time.Now()
	order.DealTimestamp = now.UnixNano() / int64(time.Millisecond)
	order.DealDatetime = now.In(spot.config.Location).Format(GO_BIRTHDAY)
	order.OrderId = response.Data[0].OrdId
	return resp, nil
}

// orderId can set client oid or orderId
func (spot *Spot) CancelOrder(order *Order) ([]byte, error) {
	urlPath := "/api/spot/v3/cancel_orders/" + order.OrderId
	param := struct {
		InstrumentId string `json:"instrument_id"`
	}{
		order.Pair.ToSymbol("-", true),
	}
	reqBody, _, _ := spot.BuildRequestBody(param)
	var response struct {
		ClientOid string `json:"client_oid"`
		OrderId   string `json:"order_id"`
		Result    bool   `json:"result"`
	}

	resp, err := spot.DoRequest(
		"POST",
		urlPath,
		reqBody,
		&response,
	)
	if err != nil {
		return nil, err
	}
	if response.Result {
		return resp, nil
	}
	return resp, NewError(400, "cancel fail, unknown error")
}

func (spot *Spot) adaptOrder(order *Order, response *OrderResponse) error {

	order.Cid = response.ClientOid
	order.OrderId = response.OrderId
	order.Price = response.Price
	order.Amount = response.Size
	order.AvgPrice = ToFloat64(response.PriceAvg)
	order.DealAmount = ToFloat64(response.FilledSize)
	order.Status = spot.adaptOrderState(response.State)

	switch response.Side {
	case "buy":
		if response.Type == "market" {
			order.Side = BUY_MARKET
			order.DealAmount = ToFloat64(response.Notional) //成交金额
		} else {
			order.Side = BUY
		}
	case "sell":
		if response.Type == "market" {
			order.Side = SELL_MARKET
			order.DealAmount = ToFloat64(response.Notional) //成交数量
		} else {
			order.Side = SELL
		}
	}

	switch response.OrderType {
	case 0:
		order.OrderType = NORMAL
	case 1:
		order.OrderType = ONLY_MAKER
	case 2:
		order.OrderType = FOK
	case 3:
		order.OrderType = IOC
	default:
		order.OrderType = NORMAL
	}

	if date, err := time.Parse(time.RFC3339, response.Timestamp); err != nil {
		return err
	} else {
		order.DealTimestamp = date.UnixNano() / int64(time.Millisecond)
		order.DealDatetime = date.In(spot.config.Location).Format(GO_BIRTHDAY)
		return nil
	}
}

// orderId can set client oid or orderId
func (spot *Spot) GetOrder(order *Order) ([]byte, error) {
	uri := "/api/spot/v3/orders/" + order.OrderId + "?instrument_id=" + order.Pair.ToSymbol("-", true)
	var response OrderResponse
	resp, err := spot.DoRequest(
		"GET",
		uri,
		"",
		&response,
	)

	if err != nil {
		return nil, err
	}

	if err := spot.adaptOrder(order, &response); err != nil {
		return nil, err
	}
	return resp, nil
}

func (spot *Spot) GetUnFinishOrders(pair Pair) ([]*Order, []byte, error) {
	uri := fmt.Sprintf(
		"/api/spot/v3/orders_pending?instrument_id=%s",
		pair.ToSymbol("-", true),
	)
	var response []OrderResponse
	resp, err := spot.DoRequest(
		"GET",
		uri,
		"",
		&response,
	)
	if err != nil {
		return nil, nil, err
	}

	var orders []*Order
	for _, itm := range response {
		order := Order{Pair: pair}
		if err := spot.adaptOrder(&order, &itm); err != nil {
			return nil, nil, err
		}
		orders = append(orders, &order)
	}

	return orders, resp, nil
}
