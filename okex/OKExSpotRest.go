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

// [{
//        "frozen":"0",
//        "hold":"0",
//        "id":"9150707",
//        "currency":"BTC",
//        "balance":"0.0049925",
//        "available":"0.0049925",
//        "holds":"0"
//    },
//    ...]

func (this *Spot) GetAccount() (*Account, []byte, error) {
	urlPath := "/api/spot/v3/accounts"
	var response []struct {
		Frozen    float64 `json:"frozen,string"`
		Hold      float64 `json:"hold,string"`
		Currency  string
		Balance   float64 `json:"balance,string"`
		Available float64 `json:"available,string"`
		Holds     float64 `json:"holds,string"`
	}

	resp, err := this.OKEx.DoRequest("GET", urlPath, "", &response)
	if err != nil {
		return nil, nil, err
	}

	account := &Account{
		SubAccounts: make(map[Currency]SubAccount, 2)}

	for _, itm := range response {
		currency := NewCurrency(itm.Currency, "")
		account.SubAccounts[currency] = SubAccount{
			Currency:     currency,
			ForzenAmount: itm.Hold,
			Amount:       itm.Available,
		}
	}

	return account, resp, nil
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

type PlaceOrderResponse struct {
	OrderId      string `json:"order_id"`
	ClientOid    string `json:"client_oid"`
	Result       bool   `json:"result"`
	ErrorCode    string `json:"error_code"`
	ErrorMessage string `json:"error_message"`
}

/**
Must Set Client Oid
*/
func (this *Spot) BatchPlaceOrders(orders []Order) ([]PlaceOrderResponse, []byte, error) {
	var param []PlaceOrderParam
	var response map[string][]PlaceOrderResponse

	for _, ord := range orders {
		param = append(param, PlaceOrderParam{
			InstrumentId: ord.Currency.AdaptUsdToUsdt().ToSymbol("-"),
			ClientOid:    ord.Cid,
			Side:         strings.ToLower(ord.Side.String()),
			Size:         ord.Amount,
			Price:        ord.Price,
			Type:         "limit",
			OrderType:    0})
	}
	reqBody, _, _ := this.BuildRequestBody(param)
	resp, err := this.DoRequest("POST", "/api/spot/v3/batch_orders", reqBody, &response)
	if err != nil {
		return nil, nil, err
	}

	var ret []PlaceOrderResponse

	for _, v := range response {
		ret = append(ret, v...)
	}

	return ret, resp, nil
}

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
	}{order.Currency.AdaptUsdToUsdt().ToLower().ToSymbol("-")}
	reqBody, _, _ := this.BuildRequestBody(param)
	var response struct {
		ClientOid string `json:"client_oid"`
		OrderId   string `json:"order_id"`
		Result    bool   `json:"result"`
	}

	resp, err := this.OKEx.DoRequest("POST", urlPath, reqBody, &response)
	if err != nil {
		return nil, err
	}
	if response.Result {
		return resp, nil
	}
	return resp, NewError(400, "cancel fail, unknown error")
}

type OrderResponse struct {
	InstrumentId   string  `json:"instrument_id"`
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

func (this *Spot) adaptOrder(response OrderResponse) *Order {
	ordInfo := &Order{
		Cid:        response.ClientOid,
		OrderId:    response.OrderId,
		Price:      response.Price,
		Amount:     response.Size,
		AvgPrice:   ToFloat64(response.PriceAvg),
		DealAmount: ToFloat64(response.FilledSize),
		Status:     this.adaptOrderState(response.State)}

	switch response.Side {
	case "buy":
		if response.Type == "market" {
			ordInfo.Side = BUY_MARKET
			ordInfo.DealAmount = ToFloat64(response.Notional) //成交金额
		} else {
			ordInfo.Side = BUY
		}
	case "sell":
		if response.Type == "market" {
			ordInfo.Side = SELL_MARKET
			ordInfo.DealAmount = ToFloat64(response.Notional) //成交数量
		} else {
			ordInfo.Side = SELL
		}
	}

	date, err := time.Parse(time.RFC3339, response.Timestamp)
	if err != nil {
		println(err)
	} else {
		ordInfo.OrderTimestamp = uint64(date.UnixNano() / 1000000)
		ordInfo.OrderDate = date.In(this.config.Location).Format(GO_BIRTHDAY)
	}

	return ordInfo
}

//orderId can set client oid or orderId
func (this *Spot) GetOneOrder(order *Order) ([]byte, error) {
	urlPath := "/api/spot/v3/orders/" + order.OrderId + "?instrument_id=" + order.Currency.AdaptUsdToUsdt().ToSymbol("-")
	//param := struct {
	//	InstrumentId string `json:"instrument_id"`
	//}{currency.AdaptUsdToUsdt().ToLower().ToSymbol("-")}
	//reqBody, _, _ := ok.BuildRequestBody(param)
	var response OrderResponse
	resp, err := this.OKEx.DoRequest("GET", urlPath, "", &response)
	if err != nil {
		return nil, err
	}

	ordInfo := this.adaptOrder(response)
	ordInfo.Currency = order.Currency

	return resp, nil
}

func (this *Spot) GetUnFinishOrders(currency CurrencyPair) ([]Order, []byte, error) {
	urlPath := fmt.Sprintf("/api/spot/v3/orders_pending?instrument_id=%s", currency.AdaptUsdToUsdt().ToSymbol("-"))
	var response []OrderResponse
	resp, err := this.OKEx.DoRequest("GET", urlPath, "", &response)
	if err != nil {
		return nil, nil, err
	}

	var ords []Order
	for _, itm := range response {
		ord := this.adaptOrder(itm)
		ord.Currency = currency
		ords = append(ords, *ord)
	}

	return ords, resp, nil
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
	resp, err := this.OKEx.DoRequest("GET", uri, "", &response)
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

	resp, err := this.OKEx.DoRequest("GET", uri, "", &response)
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
	uri := fmt.Sprintf("/api/spot/v3/instruments/%s/candles?", currency.AdaptUsdToUsdt().ToSymbol("-"))

	params := url.Values{}
	if since > 0 {
		ts, err := strconv.ParseInt(strconv.Itoa(since)[0:10], 10, 64)
		if err != nil {
			return nil, nil, err
		}
		sinceTime := time.Unix(ts, 0).UTC()
		params.Add("start", sinceTime.Format(time.RFC3339))
	}

	granularity := 60
	switch period {
	case KLINE_PERIOD_1MIN:
		granularity = 60
	case KLINE_PERIOD_3MIN:
		granularity = 180
	case KLINE_PERIOD_5MIN:
		granularity = 300
	case KLINE_PERIOD_15MIN:
		granularity = 900
	case KLINE_PERIOD_30MIN:
		granularity = 1800
	case KLINE_PERIOD_1H, KLINE_PERIOD_60MIN:
		granularity = 3600
	case KLINE_PERIOD_2H:
		granularity = 7200
	case KLINE_PERIOD_4H:
		granularity = 14400
	case KLINE_PERIOD_6H:
		granularity = 21600
	case KLINE_PERIOD_1DAY:
		granularity = 86400
	case KLINE_PERIOD_1WEEK:
		granularity = 604800
	default:
		granularity = 1800
	}
	params.Add("granularity", fmt.Sprintf("%d", granularity))

	fmt.Println(uri+params.Encode())
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
	for _, itm := range response {
		t, _ := time.Parse(time.RFC3339, fmt.Sprint(itm[0]))
		klines = append(klines, Kline{
			Timestamp: t.Unix(),
			Date:      t.In(this.config.Location).Format(GO_BIRTHDAY),
			Pair:      currency,
			Open:      ToFloat64(itm[1]),
			High:      ToFloat64(itm[2]),
			Low:       ToFloat64(itm[3]),
			Close:     ToFloat64(itm[4]),
			Vol:       ToFloat64(itm[5])},
		)
	}

	return klines, resp, nil
}

//非个人，整个交易所的交易记录
func (this *Spot) GetTrades(currencyPair CurrencyPair, since int64) ([]Trade, error) {
	panic("unsupported")
}

func (this *Spot) placeOrder(order *Order) ([]byte, error) {
	urlPath := "/api/spot/v3/orders"
	param := PlaceOrderParam{
		ClientOid:    UUID(),
		InstrumentId: order.Currency.AdaptUsdToUsdt().ToLower().ToSymbol("-"),
	}

	var response PlaceOrderResponse

	switch order.Side {
	case BUY, SELL:
		param.Side = strings.ToLower(order.Side.String())
		param.Price = order.Price
		param.Size = order.Amount
	case SELL_MARKET:
		param.Side = "sell"
		param.Size = order.Amount
	case BUY_MARKET:
		param.Side = "buy"
		param.Notional = order.Price
	default:
		param.Size = order.Amount
		param.Price = order.Price
	}

	switch order.OrderType {
	case NORMAL:
		param.OrderType = 0
	case ONLY_MAKER:
		param.OrderType = 1
	case FOK:
		param.OrderType = 2
	case IOC:
		param.OrderType = 3
	default:
		param.OrderType = 0
	}

	jsonStr, _, _ := this.OKEx.BuildRequestBody(param)
	resp, err := this.OKEx.DoRequest("POST", urlPath, jsonStr, &response)
	if err != nil {
		return nil, err
	}

	if !response.Result {
		return nil, NewError(ToInt(response.ErrorCode), response.ErrorMessage)
	}

	order.Cid = response.ClientOid
	order.OrderId = response.OrderId

	return resp, nil
}
