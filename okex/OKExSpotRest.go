package okex

import (
	"fmt"
	"sort"
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
			OrderType:    ord.OrderType})
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

func (ok *Spot) PlaceOrder(ty string, ord *Order) (*Order, []byte, error) {
	urlPath := "/api/spot/v3/orders"
	param := PlaceOrderParam{
		ClientOid:    UUID(),
		InstrumentId: ord.Currency.AdaptUsdToUsdt().ToLower().ToSymbol("-"),
	}

	var response PlaceOrderResponse

	switch ord.Side {
	case BUY, SELL:
		param.Side = strings.ToLower(ord.Side.String())
		param.Price = ord.Price
		param.Size = ord.Amount
	case SELL_MARKET:
		param.Side = "sell"
		param.Size = ord.Amount
	case BUY_MARKET:
		param.Side = "buy"
		param.Notional = ord.Price
	default:
		param.Size = ord.Amount
		param.Price = ord.Price
	}

	switch ty {
	case "limit":
		param.Type = "limit"
	case "market":
		param.Type = "market"
	case "post_only":
		param.OrderType = POST_ONLY
	case "fok":
		param.OrderType = FOK
	case "ioc":
		param.OrderType = IOC
	}

	jsonStr, _, _ := ok.OKEx.BuildRequestBody(param)
	resp, err := ok.OKEx.DoRequest("POST", urlPath, jsonStr, &response)
	if err != nil {
		return nil, nil, err
	}

	if !response.Result {
		return nil, nil, NewError(ToInt(response.ErrorCode), response.ErrorMessage)
	}

	ord.Cid = response.ClientOid
	ord.OrderId = response.OrderId

	return ord, resp, nil
}

func (this *Spot) LimitBuy(amount, price string, currency CurrencyPair) (*Order, []byte, error) {
	return this.PlaceOrder("limit", &Order{
		Price:    ToFloat64(price),
		Amount:   ToFloat64(amount),
		Currency: currency,
		Side:     BUY,
	})
}

func (this *Spot) LimitSell(amount, price string, currency CurrencyPair) (*Order, []byte, error) {
	return this.PlaceOrder("limit", &Order{
		Price:    ToFloat64(price),
		Amount:   ToFloat64(amount),
		Currency: currency,
		Side:     SELL,
	})
}

func (this *Spot) MarketBuy(amount, price string, currency CurrencyPair) (*Order, []byte, error) {
	return this.PlaceOrder("market", &Order{
		Price:    ToFloat64(price),
		Amount:   ToFloat64(amount),
		Currency: currency,
		Side:     BUY_MARKET,
	})
}

func (this *Spot) MarketSell(amount, price string, currency CurrencyPair) (*Order, []byte, error) {
	return this.PlaceOrder("market", &Order{
		Price:    ToFloat64(price),
		Amount:   ToFloat64(amount),
		Currency: currency,
		Side:     SELL_MARKET,
	})
}

//orderId can set client oid or orderId
func (this *Spot) CancelOrder(orderId string, currency CurrencyPair) (bool, []byte, error) {
	urlPath := "/api/spot/v3/cancel_orders/" + orderId
	param := struct {
		InstrumentId string `json:"instrument_id"`
	}{currency.AdaptUsdToUsdt().ToLower().ToSymbol("-")}
	reqBody, _, _ := this.BuildRequestBody(param)
	var response struct {
		ClientOid string `json:"client_oid"`
		OrderId   string `json:"order_id"`
		Result    bool   `json:"result"`
	}

	resp, err := this.OKEx.DoRequest("POST", urlPath, reqBody, &response)
	if err != nil {
		return false, nil, err
	}
	if response.Result {
		return true, resp, nil
	}
	return false, resp, NewError(400, "cancel fail, unknown error")
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
		ordInfo.OrderTime = int(date.UnixNano() / int64(time.Millisecond))
	}

	return ordInfo
}

//orderId can set client oid or orderId
func (this *Spot) GetOneOrder(orderId string, currency CurrencyPair) (*Order, []byte, error) {
	urlPath := "/api/spot/v3/orders/" + orderId + "?instrument_id=" + currency.AdaptUsdToUsdt().ToSymbol("-")
	//param := struct {
	//	InstrumentId string `json:"instrument_id"`
	//}{currency.AdaptUsdToUsdt().ToLower().ToSymbol("-")}
	//reqBody, _, _ := ok.BuildRequestBody(param)
	var response OrderResponse
	resp, err := this.OKEx.DoRequest("GET", urlPath, "", &response)
	if err != nil {
		return nil, nil, err
	}

	ordInfo := this.adaptOrder(response)
	ordInfo.Currency = currency

	return ordInfo, resp, nil
}

func (this *Spot) GetUnfinishOrders(currency CurrencyPair) ([]Order, []byte, error) {
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

func (this *Spot) GetTicker(currency CurrencyPair) (*Ticker, []byte, error) {
	urlPath := fmt.Sprintf("/api/spot/v3/instruments/%s/ticker", currency.AdaptUsdToUsdt().ToSymbol("-"))
	var response struct {
		Last          float64 `json:"last,string"`
		High24h       float64 `json:"high_24h,string"`
		Low24h        float64 `json:"low_24h,string"`
		BestBid       float64 `json:"best_bid,string"`
		BestAsk       float64 `json:"best_ask,string"`
		BaseVolume24h float64 `json:"base_volume_24_h,string"`
		Timestamp     string  `json:"timestamp"`
	}
	resp, err := this.OKEx.DoRequest("GET", urlPath, "", &response)
	if err != nil {
		return nil, resp, err
	}

	date, _ := time.Parse(time.RFC3339, response.Timestamp)
	return &Ticker{
		Pair:      currency,
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
	urlPath := fmt.Sprintf("/api/spot/v3/instruments/%s/book?size=%d", currency.AdaptUsdToUsdt().ToSymbol("-"), size)

	var response struct {
		Asks      [][]interface{} `json:"asks"`
		Bids      [][]interface{} `json:"bids"`
		Timestamp string          `json:"timestamp"`
	}

	resp, err := this.OKEx.DoRequest("GET", urlPath, "", &response)
	if err != nil {
		return nil, nil, err
	}

	dep := new(Depth)
	dep.Pair = currency
	uTime, _ := time.Parse(time.RFC3339, response.Timestamp)
	dep.Timestamp = uint64(uTime.UnixNano())
	dep.Date = uTime.In(this.config.Location).Format(GO_BIRTHDAY)

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

	sort.Sort(sort.Reverse(dep.AskList))

	return dep, resp, nil
}

func (this *Spot) GetKlineRecords(currency CurrencyPair, period, size, since int) ([]Kline, []byte, error) {
	urlPath := "/api/spot/v3/instruments/%s/candles?granularity=%d"

	if since > 0 {
		sinceTime := time.Unix(int64(since), 0).UTC()
		if since/int(time.Second) != 1 { //如果不为秒，转为秒
			sinceTime = time.Unix(int64(since)/int64(time.Second), 0).UTC()
		}
		urlPath += "&start=" + sinceTime.Format(time.RFC3339)
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

	var response [][]interface{}
	resp, err := this.DoRequest("GET", fmt.Sprintf(urlPath, currency.AdaptUsdToUsdt().ToSymbol("-"), granularity), "", &response)
	if err != nil {
		return nil, nil, err
	}

	var klines []Kline
	for _, itm := range response {
		t, _ := time.Parse(time.RFC3339, fmt.Sprint(itm[0]))
		klines = append(klines, Kline{
			Timestamp: t.Unix(),
			Pair:      currency,
			Open:      ToFloat64(itm[1]),
			High:      ToFloat64(itm[2]),
			Low:       ToFloat64(itm[3]),
			Close:     ToFloat64(itm[4]),
			Vol:       ToFloat64(itm[5])})
	}

	return klines, resp, nil
}

//非个人，整个交易所的交易记录
func (ok *Spot) GetTrades(currencyPair CurrencyPair, since int64) ([]Trade, error) {
	panic("unsupported")
}
