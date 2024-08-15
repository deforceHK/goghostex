package okex

import (
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"time"

	. "github.com/deforceHK/goghostex"
)

func (future *Future) PlaceOrder(order *FutureOrder) ([]byte, error) {
	contract, err := future.GetContract(order.Pair, order.ContractType)
	if err != nil {
		return nil, err
	}

	if order == nil {
		return nil, errors.New("ord param is nil")
	}
	if order.ContractName == "" {
		order.ContractName = future.GetInstrumentId(order.Pair, order.ContractType)
	}

	var sideInfo, _ = _INERNAL_V5_FUTURE_TYPE_CONVERTER[order.Type]
	var placeInfo, _ = _INERNAL_V5_FUTURE_PLACE_TYPE_CONVERTER[order.PlaceType]
	var request = struct {
		InstId  string `json:"instId"`
		TdMode  string `json:"tdMode"`
		Side    string `json:"side"`
		PosSide string `json:"posSide,omitempty"`
		OrdType string `json:"ordType"`
		Sz      string `json:"sz"`
		Px      string `json:"px"`
		ClOrdId string `json:"clOrdId,omitempty"`
	}{
		order.ContractName,
		"cross",
		sideInfo[0],
		sideInfo[1],
		placeInfo,
		strconv.FormatInt(order.Amount, 10),
		FloatToPrice(order.Price, contract.PricePrecision, contract.TickSize),
		order.Cid,
	}

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
	order.PlaceDatetime = now.In(future.config.Location).Format(GO_BIRTHDAY)

	reqBody, _, _ := future.BuildRequestBody(request)
	resp, err := future.DoRequest(
		http.MethodPost,
		uri,
		reqBody,
		&response,
	)

	if err != nil {
		return resp, err
	}
	if len(response.Data) > 0 && response.Data[0].SCode != "0" {
		return resp, errors.New(string(resp)) //todo 更好的获取错误码的方案
	}
	if response.Code != "0" {
		return resp, errors.New(string(resp))
	}

	now = time.Now()
	order.DealTimestamp = now.UnixNano() / int64(time.Millisecond)
	order.DealDatetime = now.In(future.config.Location).Format(GO_BIRTHDAY)
	order.OrderId = response.Data[0].OrdId
	return resp, nil
}

func (future *Future) GetOrder(order *FutureOrder) ([]byte, error) {
	if order == nil {
		return nil, errors.New("ord param is nil")
	}
	if order.ContractName == "" {
		order.ContractName = future.GetInstrumentId(order.Pair, order.ContractType)
	}

	var params = url.Values{}
	params.Set("instId", order.ContractName)
	params.Set("ordId", order.OrderId)

	var response = struct {
		Code string `json:"code"`
		Msg  string `json:"msg"`
		Data []struct {
			ClOrdId   string  `json:"clOrdId"`
			OrdId     string  `json:"ordId"`
			Px        float64 `json:"px,string"`
			Sz        float64 `json:"sz,string"`
			AvgPx     string  `json:"avgPx"`
			AccFillSz float64 `json:"accFillSz,string"`
			State     string  `json:"state"`
			Lever     float64 `json:"lever,string"`
			Fee       float64 `json:"fee,string"`
			UTime     int64   `json:"uTime,string"`
			CTime     int64   `json:"cTime,string"`
		} `json:"data"`
	}{}
	var uri = "/api/v5/trade/order?"

	resp, err := future.DoRequest(
		http.MethodGet,
		uri+params.Encode(),
		"",
		&response,
	)

	if err != nil {
		return resp, err
	}
	if response.Code != "0" {
		return resp, errors.New(response.Msg)
	}
	if len(response.Data) == 0 || response.Data[0].State == "live" {
		return resp, nil
	}

	if status, exist := _INERNAL_V5_FUTURE_ORDER_STATUE_CONVERTER[response.Data[0].State]; exist {
		order.Status = status
	}
	if order.Exchange == "" {
		order.Exchange = future.GetExchangeName()
	}

	order.Price = response.Data[0].Px
	order.Amount = ToInt64(response.Data[0].Sz)

	order.AvgPrice = ToFloat64(response.Data[0].AvgPx)
	order.DealAmount = ToInt64(response.Data[0].AccFillSz)
	order.LeverRate = ToInt64(response.Data[0].Lever)
	order.Fee = response.Data[0].Fee

	order.DealTimestamp = response.Data[0].UTime
	order.DealDatetime = time.Unix(
		order.DealTimestamp/1000, 0,
	).In(future.config.Location).Format(GO_BIRTHDAY)

	order.PlaceTimestamp = response.Data[0].CTime
	order.PlaceDatetime = time.Unix(
		order.PlaceTimestamp/1000, 0,
	).In(future.config.Location).Format(GO_BIRTHDAY)
	return resp, err
}

func (future *Future) CancelOrder(order *FutureOrder) ([]byte, error) {
	if order == nil || order.OrderId == "" {
		return nil, errors.New("order necessary param is nil")
	}
	if order.ContractName == "" {
		order.ContractName = future.GetInstrumentId(order.Pair, order.ContractType)
	}

	var request = struct {
		InstId string `json:"instId"`
		OrdId  string `json:"ordId"`
	}{
		order.ContractName,
		order.OrderId,
	}

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

	var uri = "/api/v5/trade/cancel-order"
	reqBody, _, _ := future.BuildRequestBody(request)
	resp, err := future.DoRequest(
		http.MethodPost,
		uri,
		reqBody,
		&response,
	)
	if err != nil {
		return resp, err
	}
	if len(response.Data) == 0 {
		return resp, errors.New("request lack the data. ")
	}
	if len(response.Data) != 0 && response.Data[0].SCode != "0" {
		return resp, errors.New(response.Data[0].SMsg)
	}

	return resp, nil
}

func (future *Future) GetOrders(
	pair Pair,
	contractType string,
) ([]*FutureOrder, []byte, error) {
	panic("")
}

func (future *Future) GetTrades(pair Pair, contractType string) ([]*Trade, []byte, error) {
	panic("")
}
