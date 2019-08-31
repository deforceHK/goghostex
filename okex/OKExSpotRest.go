package okex

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	. "github.com/strengthening/goghostex"
)

type Spot struct {
	*OKEx
}

type PlaceOrderParam struct {
	ClientOid     string  `json:"client_oid"`
	Type          string  `json:"type"`
	Side          string  `json:"side"`
	InstrumentId  string  `json:"instrument_id"`
	OrderType     int     `json:"order_type"`
	Price         float64 `json:"price"`
	Size          float64 `json:"size"`
	Notional      float64 `json:"notional"`
	MarginTrading string  `json:"margin_trading,omitempty"`
}

type remoteOrder struct {
	Result       bool   `json:"result"`
	OrderId      string `json:"order_id"`
	ClientOid    string `json:"client_oid"`
	ErrorCode    string `json:"error_code"`
	ErrorMessage string `json:"error_message"`
}

func (this *remoteOrder) Merge(order *Order) error {
	if this.ErrorCode != "" {
		return errors.New(this.ErrorMessage)
	}

	order.OrderId = this.OrderId
	order.Cid = this.ClientOid

	return nil
}

type OrderResponse struct {
	ClientOid      string  `json:"client_oid"`
	OrderId        string  `json:"order_id"`
	Price          float64 `json:"price,string"`
	Size           float64 `json:"size,string"`
	Notional       string  `json:"notional"`
	Side           string  `json:"side"`
	Type           string  `json:"type"`
	FilledSize     string  `json:"filled_size"`
	FilledNotional string  `json:"filled_notional"`
	PriceAvg       string  `json:"price_avg"`
	State          int     `json:"state,string"`
	OrderType      int     `json:"order_type,string"`
	Timestamp      string  `json:"timestamp"`
}

func (this *Spot) GetAccount() (*Account, []byte, error) {
	uri := "/api/spot/v3/accounts"
	var response []struct {
		Currency  string
		Frozen    float64 `json:"frozen,string"`
		Hold      float64 `json:"hold,string"`
		Balance   float64 `json:"balance,string"`
		Available float64 `json:"available,string"`
		Holds     float64 `json:"holds,string"`
	}

	resp, err := this.DoRequest(
		"GET",
		uri,
		"",
		&response,
	)
	if err != nil {
		return nil, nil, err
	}

	account := &Account{
		SubAccounts: make(map[Currency]SubAccount, 0),
	}
	for _, item := range response {
		currency := NewCurrency(item.Currency, "")
		account.SubAccounts[currency] = SubAccount{
			Currency:     currency,
			ForzenAmount: item.Hold,
			Amount:       item.Available,
		}
	}
	return account, resp, nil
}

/**
 *
 * Must Set Client Oid
 *
 **/
//func (this *Spot) BatchPlaceOrders(orders []Order) ([]PlaceOrderResponse, []byte, error) {
//	var param []PlaceOrderParam
//	var response map[string][]PlaceOrderResponse
//
//	for _, ord := range orders {
//		param = append(param, PlaceOrderParam{
//			InstrumentId: ord.Currency.AdaptUsdToUsdt().ToSymbol("-"),
//			ClientOid:    ord.Cid,
//			Side:         strings.ToLower(ord.Side.String()),
//			Size:         ord.Amount,
//			Price:        ord.Price,
//			Type:         "limit",
//			OrderType:    0})
//	}
//	reqBody, _, _ := this.BuildRequestBody(param)
//	resp, err := this.DoRequest("POST", "/api/spot/v3/batch_orders", reqBody, &response)
//	if err != nil {
//		return nil, nil, err
//	}
//
//	var ret []PlaceOrderResponse
//
//	for _, v := range response {
//		ret = append(ret, v...)
//	}
//
//	return ret, resp, nil
//}

func (this *Spot) LimitBuy(order *Order) ([]byte, error) {
	if order.Side != BUY {
		return nil, errors.New("The order side is not buy. ")
	}
	return this.placeOrder(order)
}

func (this *Spot) LimitSell(order *Order) ([]byte, error) {
	if order.Side != SELL {
		return nil, errors.New("The order side is not sell. ")
	}
	return this.placeOrder(order)
}

func (this *Spot) MarketBuy(order *Order) ([]byte, error) {
	if order.Side != BUY_MARKET {
		return nil, errors.New("The order side is not buy market. ")
	}
	return this.placeOrder(order)
}

func (this *Spot) MarketSell(order *Order) ([]byte, error) {
	if order.Side != SELL_MARKET {
		return nil, errors.New("The order side is not sell market. ")
	}
	return this.placeOrder(order)
}

//orderId can set client oid or orderId
func (this *Spot) CancelOrder(order *Order) ([]byte, error) {
	urlPath := "/api/spot/v3/cancel_orders/" + order.OrderId
	param := struct {
		InstrumentId string `json:"instrument_id"`
	}{
		order.Currency.AdaptUsdToUsdt().ToLower().ToSymbol("-"),
	}
	reqBody, _, _ := this.BuildRequestBody(param)
	var response struct {
		ClientOid string `json:"client_oid"`
		OrderId   string `json:"order_id"`
		Result    bool   `json:"result"`
	}

	resp, err := this.DoRequest(
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

func (this *Spot) adaptOrder(order *Order, response *OrderResponse) error {

	order.Cid = response.ClientOid
	order.OrderId = response.OrderId
	order.Price = response.Price
	order.Amount = response.Size
	order.AvgPrice = ToFloat64(response.PriceAvg)
	order.DealAmount = ToFloat64(response.FilledSize)
	order.Status = this.adaptOrderState(response.State)

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
		order.OrderTimestamp = uint64(date.UnixNano() / 1000000)
		order.OrderDate = date.In(this.config.Location).Format(GO_BIRTHDAY)
		return nil
	}
}

//orderId can set client oid or orderId
func (this *Spot) GetOneOrder(order *Order) ([]byte, error) {
	uri := "/api/spot/v3/orders/" + order.OrderId + "?instrument_id=" + order.Currency.AdaptUsdToUsdt().ToSymbol("-")
	var response OrderResponse
	resp, err := this.DoRequest(
		"GET",
		uri,
		"",
		&response,
	)

	if err != nil {
		return nil, err
	}

	if err := this.adaptOrder(order, &response); err != nil {
		return nil, err
	}
	return resp, nil
}

func (this *Spot) GetUnFinishOrders(currency CurrencyPair) ([]Order, []byte, error) {
	uri := fmt.Sprintf(
		"/api/spot/v3/orders_pending?instrument_id=%s",
		currency.AdaptUsdToUsdt().ToSymbol("-"),
	)
	var response []OrderResponse
	resp, err := this.DoRequest(
		"GET",
		uri,
		"",
		&response,
	)
	if err != nil {
		return nil, nil, err
	}

	var orders []Order
	for _, itm := range response {
		order := Order{Currency: currency}
		if err := this.adaptOrder(&order, &itm); err != nil {
			return nil, nil, err
		}
		orders = append(orders, order)
	}

	return orders, resp, nil
}

func (this *Spot) GetOrderHistorys(currency CurrencyPair, currentPage, pageSize int) ([]Order, error) {
	panic("unsupported")
}

func (this *Spot) GetTicker(pair CurrencyPair) (*Ticker, []byte, error) {
	uri := fmt.Sprintf(
		"/api/spot/v3/instruments/%s/ticker",
		pair.AdaptUsdToUsdt().ToSymbol("-"),
	)

	var response struct {
		Last          float64 `json:"last,string"`
		High24h       float64 `json:"high_24h,string"`
		Low24h        float64 `json:"low_24h,string"`
		BestBid       float64 `json:"best_bid,string"`
		BestAsk       float64 `json:"best_ask,string"`
		BaseVolume24h float64 `json:"base_volume_24h,string"`
		Timestamp     string  `json:"timestamp"`
	}
	resp, err := this.DoRequest("GET", uri, "", &response)
	if err != nil {
		return nil, resp, err
	}

	date, _ := time.Parse(time.RFC3339, response.Timestamp)
	return &Ticker{
		Pair:      pair,
		Last:      response.Last,
		High:      response.High24h,
		Low:       response.Low24h,
		Sell:      response.BestAsk,
		Buy:       response.BestBid,
		Vol:       response.BaseVolume24h,
		Timestamp: uint64(time.Duration(date.UnixNano() / int64(time.Millisecond))),
		Date:      date.In(this.config.Location).Format(GO_BIRTHDAY),
	}, resp, nil
}

func (this *Spot) GetDepth(size int, currency CurrencyPair) (*Depth, []byte, error) {
	uri := fmt.Sprintf(
		"/api/spot/v3/instruments/%s/book?size=%d",
		currency.AdaptUsdToUsdt().ToSymbol("-"),
		size,
	)

	var response struct {
		Asks      [][]interface{} `json:"asks"`
		Bids      [][]interface{} `json:"bids"`
		Timestamp string          `json:"timestamp"`
	}

	resp, err := this.DoRequest("GET", uri, "", &response)
	if err != nil {
		return nil, nil, err
	}

	dep := new(Depth)
	dep.Pair = currency
	date, _ := time.Parse(time.RFC3339, response.Timestamp)
	dep.Timestamp = uint64(date.UnixNano() / 1000000)
	dep.Date = date.In(this.config.Location).Format(GO_BIRTHDAY)
	dep.Sequence = dep.Timestamp

	for _, itm := range response.Asks {
		dep.AskList = append(dep.AskList, DepthRecord{
			Price:  ToFloat64(itm[0]),
			Amount: ToFloat64(itm[1]),
		})
	}

	for _, itm := range response.Bids {
		dep.BidList = append(dep.BidList, DepthRecord{
			Price:  ToFloat64(itm[0]),
			Amount: ToFloat64(itm[1]),
		})
	}

	return dep, resp, nil
}

func (this *Spot) GetKlineRecords(currency CurrencyPair, period, size, since int) ([]Kline, []byte, error) {
	uri := fmt.Sprintf(
		"/api/spot/v3/instruments/%s/candles?",
		currency.AdaptUsdToUsdt().ToSymbol("-"),
	)

	params := url.Values{}
	if since > 0 {
		ts, err := strconv.ParseInt(strconv.Itoa(since)[0:10], 10, 64)
		if err != nil {
			return nil, nil, err
		}
		sinceTime := time.Unix(ts, 0).UTC()
		endTime := time.Now().UTC()

		params.Add("start", sinceTime.Format(time.RFC3339))
		params.Add("end", endTime.Format(time.RFC3339))
	}
	granularity, isExist := _INERNAL_KLINE_PERIOD_CONVERTER[period]
	if !isExist {
		return nil, nil, errors.New("The period is not supported. ")
	}
	params.Add("granularity", fmt.Sprintf("%d", granularity))

	var response [][]interface{}
	resp, err := this.DoRequest(
		"GET",
		uri+params.Encode(),
		"",
		&response,
	)
	if err != nil {
		return nil, nil, err
	}

	var klines []Kline
	for _, item := range response {
		t, _ := time.Parse(time.RFC3339, fmt.Sprint(item[0]))
		klines = append(klines, Kline{
			Timestamp: uint64(t.UnixNano() / int64(time.Millisecond)),
			Date:      t.In(this.config.Location).Format(GO_BIRTHDAY),
			Pair:      currency,
			Open:      ToFloat64(item[1]),
			High:      ToFloat64(item[2]),
			Low:       ToFloat64(item[3]),
			Close:     ToFloat64(item[4]),
			Vol:       ToFloat64(item[5])},
		)
	}

	return klines, resp, nil
}

//非个人，整个交易所的交易记录
func (this *Spot) GetTrades(currencyPair CurrencyPair, since int64) ([]Trade, error) {
	panic("unsupported")
}

func (this *Spot) placeOrder(order *Order) ([]byte, error) {
	uri := "/api/spot/v3/orders"
	param := PlaceOrderParam{
		InstrumentId: order.Currency.AdaptUsdToUsdt().ToLower().ToSymbol("-"),
	}

	if order.Cid == "" {
		order.Cid = UUID()
	}
	param.ClientOid = order.Cid

	var response remoteOrder
	switch order.Side {
	case BUY, SELL:
		param.Side = strings.ToLower(order.Side.String())
		param.Price = order.Price
		param.Size = order.Amount
		param.Type = "limit"
		param.OrderType = _INTERNAL_ORDER_TYPE_CONVERTER[order.OrderType]
	case SELL_MARKET:
		param.Side = "sell"
		param.Type = "market"
		param.Size = order.Amount
	case BUY_MARKET:
		param.Side = "buy"
		param.Type = "market"
		param.Notional = order.Price
	default:
		param.Size = order.Amount
		param.Price = order.Price
	}

	jsonStr, _, _ := this.BuildRequestBody(param)
	resp, err := this.DoRequest(
		"POST",
		uri,
		jsonStr,
		&response,
	)
	if err != nil {
		return resp, err
	}
	if err := response.Merge(order); err != nil {
		return resp, err
	}
	return resp, nil
}
