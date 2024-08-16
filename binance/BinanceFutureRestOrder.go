package binance

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	. "github.com/deforceHK/goghostex"
)

func (future *Future) GetTrades(pair Pair, contractType string) ([]*Trade, []byte, error) {
	if contractType == THIS_WEEK_CONTRACT || contractType == NEXT_WEEK_CONTRACT {
		return nil, nil, errors.New("binance have not this_week next_week contract. ")
	}

	contract, err := future.GetContract(pair, contractType)
	if err != nil {
		return nil, nil, err
	}

	params := url.Values{}
	params.Set("symbol", future.getBNSymbol(contract.ContractName))

	uri := FUTURE_TRADE_URI + params.Encode()
	response := make([]struct {
		Id           int64   `json:"id"`
		Price        float64 `json:"price,string"`
		Qty          int64   `json:"qty,string"`
		BaseQty      float64 `json:"baseQty,string"`
		Time         int64   `json:"time"`
		IsBuyerMaker bool    `json:"isBuyerMaker"`
	}, 0)
	resp, err := future.DoRequest(
		http.MethodGet,
		FUTURE_CM_ENDPOINT,
		uri,
		"",
		&response,
	)
	if err != nil {
		return nil, resp, err
	}

	trades := make([]*Trade, 0)
	for _, item := range response {
		tradeType := BUY
		if !item.IsBuyerMaker {
			tradeType = SELL
		}
		trade := Trade{
			Tid:       item.Id,
			Type:      tradeType,
			Amount:    item.BaseQty,
			Price:     item.Price,
			Timestamp: item.Time,
			Pair:      pair,
		}
		trades = append(trades, &trade)
	}

	return trades, resp, nil
}

func (future *Future) PlaceOrder(order *FutureOrder) ([]byte, error) {
	if order == nil {
		return nil, errors.New("ord param is nil")
	}

	contract, err := future.GetContract(order.Pair, order.ContractType)
	if err != nil {
		return nil, err
	}

	side, positionSide, placeType := "", "", ""
	exist := false
	if side, exist = sideRelation[order.Type]; !exist {
		return nil, errors.New("future type not found. ")
	}
	if positionSide, exist = positionSideRelation[order.Type]; !exist {
		return nil, errors.New("future type not found. ")
	}
	if placeType, exist = placeTypeRelation[order.PlaceType]; !exist {
		return nil, errors.New("place type not found. ")
	}

	param := url.Values{}
	param.Set("symbol", future.getBNSymbol(contract.ContractName))
	param.Set("side", side)
	param.Set("positionSide", positionSide)
	param.Set("type", "LIMIT")
	param.Set("price", FloatToPrice(order.Price, contract.PricePrecision, contract.TickSize))
	param.Set("quantity", fmt.Sprintf("%d", order.Amount))
	// "GTC": 成交为止, 一直有效。
	// "IOC": 无法立即成交(吃单)的部分就撤销。
	// "FOK": 无法全部立即成交就撤销。
	// "GTX": 无法成为挂单方就撤销。
	param.Set("timeInForce", placeType)
	if order.Cid != "" {
		param.Set("newClientOrderId", order.Cid)
	}
	if err := future.buildParamsSigned(&param); err != nil {
		return nil, err
	}

	var response struct {
		Cid        string  `json:"clientOrderId"`
		Status     string  `json:"status"`
		OrderId    int64   `json:"orderId"`
		UpdateTime int64   `json:"updateTime"`
		Price      float64 `json:"price,string"`
		AvgPrice   float64 `json:"avgPrice,string"`
		Amount     int64   `json:"origQty,string"`
		DealAmount int64   `json:"executedQty,string"`
	}

	now := time.Now()
	resp, err := future.DoRequest(
		http.MethodPost,
		FUTURE_CM_ENDPOINT,
		FUTURE_PLACE_ORDER_URI+param.Encode(),
		"",
		&response,
	)
	if err != nil {
		return nil, err
	}
	orderTime := time.Unix(response.UpdateTime/1000, 0)

	order.OrderId = fmt.Sprintf("%d", response.OrderId)
	order.PlaceTimestamp = now.UnixNano() / int64(time.Millisecond)
	order.PlaceDatetime = now.In(future.config.Location).Format(GO_BIRTHDAY)
	order.DealTimestamp = response.UpdateTime
	order.DealDatetime = orderTime.In(future.config.Location).Format(GO_BIRTHDAY)
	order.Status = statusRelation[response.Status]
	order.Price = response.Price
	order.Amount = response.Amount
	order.ContractName = contract.ContractName

	if response.Cid != "" {
		order.Cid = response.Cid
	}
	if response.DealAmount > 0 {
		order.AvgPrice = response.AvgPrice
		order.DealAmount = response.DealAmount
	}

	return resp, nil
}

func (future *Future) CancelOrder(order *FutureOrder) ([]byte, error) {
	contract, err := future.GetContract(order.Pair, order.ContractType)
	if err != nil {
		return nil, err
	}

	if order.OrderId == "" && order.Cid == "" {
		return nil, errors.New("The order_id and cid is empty. ")
	}

	param := url.Values{}
	param.Set("symbol", future.getBNSymbol(contract.ContractName))
	if order.OrderId != "" {
		param.Set("orderId", order.OrderId)
	} else {
		param.Set("origClientOrderId", order.Cid)
	}
	if err := future.buildParamsSigned(&param); err != nil {
		return nil, err
	}

	var response struct {
		Cid        string  `json:"clientOrderId"`
		Status     string  `json:"status"`
		DealAmount int64   `json:"executedQty,string"`
		AvgPrice   float64 `json:"avgPrice,string"`
		OrderId    int64   `json:"orderId"`
		UpdateTime int64   `json:"updateTime"`
	}

	now := time.Now()
	resp, err := future.DoRequest(
		http.MethodDelete,
		FUTURE_CM_ENDPOINT,
		FUTURE_CANCEL_ORDER_URI+param.Encode(),
		"",
		&response,
	)
	if err != nil {
		return nil, err
	}

	orderTime := time.Unix(response.UpdateTime/1000, 0)
	order.PlaceTimestamp = now.UnixNano() / int64(time.Millisecond)
	order.PlaceDatetime = now.In(future.config.Location).Format(GO_BIRTHDAY)
	order.DealTimestamp = response.UpdateTime
	order.DealDatetime = orderTime.In(future.config.Location).Format(GO_BIRTHDAY)
	order.Status = statusRelation[response.Status]
	if response.DealAmount > 0 {
		order.AvgPrice = response.AvgPrice
		order.DealAmount = response.DealAmount
	}
	return resp, nil
}

func (future *Future) GetOrders(pair Pair, contractType string) ([]*FutureOrder, []byte, error) {
	contract, err := future.GetContract(pair, contractType)
	if err != nil {
		return nil, nil, err
	}

	var param = url.Values{}
	param.Set("symbol", future.getBNSymbol(contract.ContractName))

	if err := future.buildParamsSigned(&param); err != nil {
		return nil, nil, err
	}

	response := make([]struct {
		AvgPrice      float64 `json:"avgPrice,string"`
		ClientOrderId string  `json:"clientOrderId"`
		ExecutedQty   int64   `json:"executedQty"`
		OrderId       int64   `json:"orderId"`
		OrigQty       float64 `json:"origQty,string"`
		Price         float64 `json:"price,string"`
		Side          string  `json:"side"`
		PositionSide  string  `json:"positionSide"`
		Status        string  `json:"status"`
		Symbol        string  `json:"symbol"`
		Pair          string  `json:"pair"`
		Time          int64   `json:"time"`
		TimeInForce   string  `json:"timeInForce"`
		UpdateTime    int64   `json:"updateTime"`
	}, 0)

	resp, err := future.DoRequest(
		http.MethodGet,
		FUTURE_CM_ENDPOINT,
		FUTURE_GET_ORDERS_URI+param.Encode(),
		"",
		&response,
	)
	if err != nil {
		return nil, resp, err
	}

	orders := make([]*FutureOrder, 0)
	for _, item := range response {
		placeTime := time.Unix(item.Time/1000, item.Time%1000).In(future.config.Location)
		updateTime := time.Unix(item.UpdateTime/1000, item.UpdateTime%1000).In(future.config.Location)

		order := FutureOrder{
			Cid:            item.ClientOrderId,
			OrderId:        fmt.Sprintf("%d", item.OrderId),
			Price:          item.Price,
			AvgPrice:       item.AvgPrice,
			Amount:         ToInt64(item.OrigQty),
			DealAmount:     item.ExecutedQty,
			PlaceTimestamp: placeTime.UnixNano() / int64(time.Millisecond),
			PlaceDatetime:  placeTime.Format(GO_BIRTHDAY),
			DealTimestamp:  updateTime.UnixNano() / int64(time.Millisecond),
			DealDatetime:   updateTime.Format(GO_BIRTHDAY),
			Status:         _INTERNAL_ORDER_STATUS_REVERSE_CONVERTER[item.Status],
			PlaceType:      _INTERNAL_PLACE_TYPE_REVERSE_CONVERTER[item.TimeInForce],
			Type:           future.getFutureType(item.Side, item.PositionSide),
			//LeverRate: item.,
			//Fee:item.,
			Pair:         pair,
			ContractType: contractType,
			ContractName: contract.ContractName,
			Exchange:     BINANCE,
		}
		orders = append(orders, &order)
	}
	return orders, resp, nil
}

func (future *Future) GetOrder(order *FutureOrder) ([]byte, error) {
	if order.OrderId == "" && order.Cid == "" {
		return nil, errors.New("The order id and cid is empty. ")
	}

	contract, err := future.GetContract(order.Pair, order.ContractType)
	if err != nil {
		return nil, err
	}

	var params = url.Values{}
	params.Add("symbol", future.getBNSymbol(contract.ContractName))

	if order.OrderId != "" {
		params.Set("orderId", order.OrderId)
	} else {
		params.Set("origClientOrderId", order.Cid)
	}
	if err := future.buildParamsSigned(&params); err != nil {
		return nil, err
	}

	var response struct {
		Cid    string `json:"clientOrderId"`
		Status string `json:"status"`

		Price    float64 `json:"price,string"`
		AvgPrice float64 `json:"avgPrice,string"`

		Amount     int64 `json:"origQty,string"`
		DealAmount int64 `json:"executedQty,string"`

		OrderId    int64 `json:"orderId"`
		UpdateTime int64 `json:"updateTime"`
	}

	now := time.Now()
	resp, err := future.DoRequest(
		http.MethodGet,
		FUTURE_CM_ENDPOINT,
		FUTURE_GET_ORDER_URI+params.Encode(),
		"",
		&response,
	)
	if err != nil {
		return nil, err
	}

	orderTime := time.Unix(response.UpdateTime/1000, 0)
	order.PlaceTimestamp = now.UnixNano() / int64(time.Millisecond)
	order.PlaceDatetime = now.In(future.config.Location).Format(GO_BIRTHDAY)
	order.DealTimestamp = response.UpdateTime
	order.DealDatetime = orderTime.In(future.config.Location).Format(GO_BIRTHDAY)
	order.Status = statusRelation[response.Status]
	order.Price = response.Price
	order.Amount = response.Amount
	if response.DealAmount > 0 {
		order.AvgPrice = response.AvgPrice
		order.DealAmount = response.DealAmount
	}
	return resp, nil
}
