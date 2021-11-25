package okex

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	. "github.com/strengthening/goghostex"
)

type Swap struct {
	*OKEx
}

func (swap *Swap) GetTicker(pair Pair) (*SwapTicker, []byte, error) {

	params := &url.Values{}
	params.Set("instId", pair.ToSymbol("-", true)+"-SWAP")

	var uri = "/api/v5/market/ticker?" + params.Encode()
	var response struct {
		Code string `json:"code"`
		Msg  string `json:"msg"`
		Data []*struct {
			InstType  string  `json:"instType"`
			InstId    string  `json:"instId"`
			Last      float64 `json:"last,string"`
			High24h   float64 `json:"high24h,string"`
			Low24h    float64 `json:"low24h,string"`
			BidPx     float64 `json:"bidPx,string"`
			AskPx     float64 `json:"askPx,string"`
			Volume24h float64 `json:"volCcy24h,string"`
			Timestamp int64   `json:"ts,string"`
		} `json:"data"`
	}
	resp, err := swap.DoRequestMarket(
		http.MethodGet,
		uri,
		"",
		&response,
	)

	if err != nil {
		return nil, nil, err
	}
	if response.Code != "0" {
		return nil, nil, errors.New(response.Msg)
	}
	if len(response.Data) == 0 {
		err = errors.New("lack response data. ")
	}

	date := time.Unix(response.Data[0].Timestamp/1000, 0)
	ticker := SwapTicker{
		Pair:      pair,
		Sell:      response.Data[0].AskPx,
		Buy:       response.Data[0].BidPx,
		Low:       response.Data[0].Low24h,
		High:      response.Data[0].High24h,
		Last:      response.Data[0].Last,
		Vol:       response.Data[0].Volume24h,
		Timestamp: response.Data[0].Timestamp,
		Date:      date.In(swap.config.Location).Format(GO_BIRTHDAY),
	}

	return &ticker, resp, nil
}

func (swap *Swap) GetDepth(pair Pair, size int) (*SwapDepth, []byte, error) {

	params := &url.Values{}
	params.Set("instId", pair.ToSymbol("-", true)+"-SWAP")
	params.Set("sz", fmt.Sprintf("%d", size))

	var uri = "/api/v5/market/books?" + params.Encode()
	var response struct {
		Code string `json:"code"`
		Msg  string `json:"msg"`
		Data []*struct {
			Asks      [][]string `json:"asks"`
			Bids      [][]string `json:"bids"`
			Timestamp int64      `json:"ts,string"`
		} `json:"data"`
	}

	resp, err := swap.DoRequestMarket(
		http.MethodGet,
		uri,
		"",
		&response,
	)

	if err != nil {
		return nil, nil, err
	}
	if response.Code != "0" {
		return nil, nil, errors.New(response.Msg)
	}
	if len(response.Data) == 0 {
		err = errors.New("lack response data. ")
	}

	depth := new(SwapDepth)
	depth.Pair = pair

	date := time.Unix(response.Data[0].Timestamp/1000, 0)
	depth.Timestamp = response.Data[0].Timestamp
	depth.Date = date.In(swap.config.Location).Format(GO_BIRTHDAY)
	depth.Sequence = response.Data[0].Timestamp

	for _, bid := range response.Data[0].Bids {
		price := ToFloat64(bid[0])
		amount := ToFloat64(bid[1])
		depthItem := DepthRecord{Price: price, Amount: amount}
		depth.BidList = append(depth.BidList, depthItem)
	}

	for _, ask := range response.Data[0].Asks {
		price := ToFloat64(ask[0])
		amount := ToFloat64(ask[1])
		depthItem := DepthRecord{Price: price, Amount: amount}
		depth.AskList = append(depth.AskList, depthItem)
	}

	return depth, resp, nil
}

func (swap *Swap) GetLimit(pair Pair) (float64, float64, error) {

	params := &url.Values{}
	params.Set("instId", pair.ToSymbol("-", true)+"-SWAP")

	var uri = "/api/v5/public/price-limit?" + params.Encode()
	var response struct {
		Code string `json:"code"`
		Msg  string `json:"msg"`
		Data []*struct {
			InstType string `json:"instType"`
			InstId   string `json:"instId"`

			BuyLmt    float64 `json:"buyLmt,string"`
			SellLmt   float64 `json:"sellLmt,string"`
			Timestamp int64   `json:"ts,string"`
		} `json:"data"`
	}

	_, err := swap.DoRequestMarket(
		http.MethodGet,
		uri,
		"",
		&response,
	)

	if err != nil {
		return 0, 0, err
	}
	if response.Code != "0" {
		return 0, 0, errors.New(response.Msg)
	}
	if len(response.Data) == 0 {
		return 0, 0, errors.New("lack response data. ")
	}

	return response.Data[0].SellLmt, response.Data[0].BuyLmt, nil

}

func (swap *Swap) GetKline(pair Pair, period, size, since int) ([]*SwapKline, []byte, error) {

	if size > 100 {
		size = 100
	}
	params := url.Values{}
	params.Set("instId", pair.ToSymbol("-", true)+"-SWAP")
	params.Set("bar", _INERNAL_V5_CANDLE_PERIOD_CONVERTER[period])
	params.Set("limit", strconv.Itoa(size))
	if since > 0 {
		endTime := time.Now()
		params.Set("before", strconv.Itoa(since))
		params.Set("after", strconv.Itoa(int(endTime.UnixNano()/1000000)))
	}

	var uri = "/api/v5/market/candles?" + params.Encode()
	var response struct {
		Code string     `json:"code"`
		Msg  string     `json:"msg"`
		Data [][]string `json:"data"`
	}
	resp, err := swap.DoRequestMarket(
		http.MethodGet,
		uri,
		"",
		&response,
	)
	if err != nil {
		return nil, nil, err
	}
	if response.Code != "0" {
		return nil, nil, errors.New(response.Msg)
	}

	var klines []*SwapKline
	for _, itm := range response.Data {
		timestamp := ToInt64(itm[0])
		t := time.Unix(timestamp/1000, 0)
		klines = append(klines, &SwapKline{
			Timestamp: timestamp,
			Date:      t.In(swap.config.Location).Format(GO_BIRTHDAY),
			Pair:      pair,
			Exchange:  OKEX,
			Open:      ToFloat64(itm[1]),
			High:      ToFloat64(itm[2]),
			Low:       ToFloat64(itm[3]),
			Close:     ToFloat64(itm[4]),
			Vol:       ToFloat64(itm[6]),
		})
	}

	return GetAscSwapKline(klines), resp, nil
}

func (swap *Swap) GetOpenAmount(pair Pair) (float64, int64, []byte, error) {
	panic("implement me")
}

func (swap *Swap) GetFundingFees(pair Pair) ([][]interface{}, []byte, error) {
	panic("implement me")
}

func (swap *Swap) GetFundingFee(pair Pair) (float64, error) {
	panic("implement me")
}

func (swap *Swap) GetAccount() (*SwapAccount, []byte, error) {
	panic("implement me")
}

var _INERNAL_V5_FUTURE_TYPE_CONVERTER = map[FutureType][]string{
	OPEN_LONG:       {"buy", "long"},
	OPEN_SHORT:      {"sell", "short"},
	LIQUIDATE_LONG:  {"sell", "long"},
	LIQUIDATE_SHORT: {"buy", "short"},
}

var _INERNAL_V5_FUTURE_PLACE_TYPE_CONVERTER = map[PlaceType]string{
	NORMAL:     "limit",
	ONLY_MAKER: "post_only",
	FOK:        "fok",
	IOC:        "ioc",
}

func (swap *Swap) PlaceOrder(order *SwapOrder) ([]byte, error) {
	var request = struct {
		InstId  string `json:"instId"`
		TdMode  string `json:"tdMode"`
		Side    string `json:"side"`
		PosSide string `json:"posSide,omitempty"`
		OrdType string `json:"ordType"`
		Sz      string `json:"sz"`
		Px      string `json:"px"`
		ClOrdId string `json:"clOrdId,omitempty"`
	}{}

	request.InstId = order.Pair.ToSymbol("-", true) + "-SWAP"
	request.TdMode = "cross" // todo 目前写死全仓，后续调整成可逐仓
	sideInfo, _ := _INERNAL_V5_FUTURE_TYPE_CONVERTER[order.Type]
	request.Side = sideInfo[0]
	request.PosSide = sideInfo[1]
	placeInfo, _ := _INERNAL_V5_FUTURE_PLACE_TYPE_CONVERTER[order.PlaceType]
	request.OrdType = placeInfo
	request.Sz = strconv.FormatFloat(order.Amount, 'f', -1, 64)
	request.Px = strconv.FormatFloat(order.Price, 'f', -1, 64)
	request.ClOrdId = order.Cid

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
	order.PlaceDatetime = now.In(swap.config.Location).Format(GO_BIRTHDAY)
	reqBody, _, _ := swap.BuildRequestBody(request)
	resp, err := swap.DoRequest(
		http.MethodPost,
		uri,
		reqBody,
		&response,
	)

	if err != nil {
		return resp, err
	}
	if len(response.Data) > 0 && response.Data[0].SCode != "0" {
		return resp, errors.New(response.Data[0].SMsg)
	}
	if response.Code != "0" {
		return resp, errors.New(response.Msg)
	}

	now = time.Now()
	order.DealTimestamp = now.UnixNano() / int64(time.Millisecond)
	order.DealDatetime = now.In(swap.config.Location).Format(GO_BIRTHDAY)
	order.OrderId = response.Data[0].OrdId
	return resp, nil
}

func (swap *Swap) CancelOrder(order *SwapOrder) ([]byte, error) {

	var request = struct {
		InstId string `json:"instId"`
		OrdId  string `json:"ordId"`
	}{
		order.Pair.ToSymbol("-", true) + "-SWAP",
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
	reqBody, _, _ := swap.BuildRequestBody(request)

	resp, err := swap.DoRequest(
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

func (swap *Swap) GetOrders(pair Pair) ([]*SwapOrder, []byte, error) {
	panic("implement me")
}

var _INERNAL_V5_FUTURE_ORDER_STATUE_CONVERTER = map[string]TradeStatus{
	"canceled":         ORDER_CANCEL,
	"live":             ORDER_UNFINISH,
	"partially_filled": ORDER_PART_FINISH,
	"filled":           ORDER_FINISH,
}

func (swap *Swap) GetOrder(order *SwapOrder) ([]byte, error) {

	var params = url.Values{}
	params.Set("instId", order.Pair.ToSymbol("-", true)+"-SWAP")
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

	resp, err := swap.DoRequest(
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
	order.Price = response.Data[0].Px
	order.Amount = response.Data[0].Sz

	order.AvgPrice = ToFloat64(response.Data[0].AvgPx)
	order.DealAmount = response.Data[0].AccFillSz
	order.LeverRate = ToInt64(response.Data[0].Lever)
	order.Fee = response.Data[0].Fee

	order.DealTimestamp = response.Data[0].UTime
	order.DealDatetime = time.Unix(
		order.DealTimestamp/1000, 0,
	).In(swap.config.Location).Format(GO_BIRTHDAY)

	order.PlaceTimestamp = response.Data[0].CTime
	order.PlaceDatetime = time.Unix(
		order.PlaceTimestamp/1000, 0,
	).In(swap.config.Location).Format(GO_BIRTHDAY)
	return resp, err

}

func (swap *Swap) GetUnFinishOrders(pair Pair) ([]*SwapOrder, []byte, error) {
	panic("implement me")
}

func (swap *Swap) GetPosition(pair Pair, openType FutureType) (*SwapPosition, []byte, error) {
	panic("implement me")
}

func (swap *Swap) AddMargin(pair Pair, openType FutureType, marginAmount float64) ([]byte, error) {
	panic("implement me")
}

func (swap *Swap) ReduceMargin(pair Pair, openType FutureType, marginAmount float64) ([]byte, error) {
	panic("implement me")
}

func (swap *Swap) GetAccountFlow() ([]*SwapAccountItem, []byte, error) {
	panic("implement me")
}

func (swap *Swap) GetExchangeRule(pair Pair) (*SwapRule, []byte, error) {
	panic("implement me")
	//return nil,nil,nil
}

func (swap *Swap) KeepAlive() {
	panic("implement me")
}
