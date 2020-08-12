package binance

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	. "github.com/strengthening/goghostex"
)

type Spot struct {
	*Binance
}

// the common resp struct of order/info/cancel
type remoteOrder struct {
	Symbol              string  `json:"symbol"`
	OrderId             int64   `json:"orderId"`
	ClientOrderId       string  `json:"clientOrderId"`
	TransactTime        int64   `json:"transactTime"` // exist when order
	Time                int64   `json:"time"`         // exist when get order info
	Price               float64 `json:"price,string"`
	OrigQty             float64 `json:"origQty,string"`
	ExecutedQty         float64 `json:"executedQty,string"`         // order deal amount
	CummulativeQuoteQty float64 `json:"cummulativeQuoteQty,string"` // order avg price

	// OrderStatus: NEW PARTIALLY_FILLED FILLED CANCELED REJECTED EXPIRED
	Status string `json:"status"`
	// OrderType: LIMIT MARKET STOP_LOSS STOP_LOSS_LIMIT TAKE_PROFIT TAKE_PROFIT_LIMIT LIMIT_MAKER
	Type string `json:"type"`
	// OrderSide: BUY SELL
	Side string `json:"side"`
}

func (binance *remoteOrder) Merge(order *Order, location *time.Location) {
	if binance.TransactTime != 0 || binance.Time != 0 {
		ts := binance.Time
		if binance.TransactTime > binance.Time {
			ts = binance.TransactTime
		}
		transactTime := time.Unix(int64(ts)/1000, int64(ts)%1000)
		order.OrderDate = transactTime.In(location).Format(GO_BIRTHDAY)
		order.OrderTimestamp = binance.TransactTime
	}

	status, exist := _INTERNAL_ORDER_STATUS_REVERSE_CONVERTER[binance.Status]
	if !exist {
		status = ORDER_FAIL
	}

	if binance.Type == "LIMIT" && binance.Side == "SELL" {
		order.Side = BUY
	} else if binance.Type == "LIMIT" && binance.Side == "BUY" {
		order.Side = SELL
	} else if binance.Type == "MARKET" && binance.Side == "SELL" {
		order.Side = SELL_MARKET
	} else {
		order.Side = BUY_MARKET
	}

	order.Status = status
	order.OrderId = fmt.Sprintf("%d", binance.OrderId)
	order.Cid = binance.ClientOrderId
	order.Price = binance.Price
	order.Amount = binance.OrigQty
	order.AvgPrice = binance.CummulativeQuoteQty
	order.DealAmount = binance.ExecutedQty
}

func (binance *Spot) LimitBuy(order *Order) ([]byte, error) {
	if order.Side != BUY {
		return nil, errors.New("The order side is not BUY or order type is not LIMIT. ")
	}
	return binance.placeOrder(order)
}

func (binance *Spot) LimitSell(order *Order) ([]byte, error) {
	if order.Side != SELL {
		return nil, errors.New("The order side is not SELL or order type is not LIMIT. ")
	}
	return binance.placeOrder(order)
}

func (binance *Spot) MarketBuy(order *Order) ([]byte, error) {
	if order.Side != BUY_MARKET {
		return nil, errors.New("the order side is not BUY_MARKET")
	}
	return binance.placeOrder(order)
}

func (binance *Spot) MarketSell(order *Order) ([]byte, error) {
	if order.Side != SELL_MARKET {
		return nil, errors.New("the order side is not SELL_MARKET")
	}
	return binance.placeOrder(order)
}

func (binance *Spot) CancelOrder(order *Order) ([]byte, error) {
	if order.OrderId == "" {
		return nil, errors.New("You must get the order_id. ")
	}

	uri := API_V3 + ORDER_URI
	params := url.Values{}
	params.Set("symbol", order.Pair.ToSymbol("", true))
	params.Set("orderId", order.OrderId)
	if err := binance.buildParamsSigned(&params); err != nil {
		return nil, err
	}

	response := remoteOrder{}
	resp, err := binance.DoRequest(
		"DELETE",
		uri,
		params.Encode(),
		&response,
	)

	if err != nil {
		return nil, err
	}
	response.Merge(order, binance.config.Location)
	return resp, nil
}

func (binance *Spot) GetOneOrder(order *Order) ([]byte, error) {
	//pair := binance.adaptCurrencyPair(order.Currency)
	if order.OrderId == "" {
		return nil, errors.New("You must get the order_id. ")
	}

	params := url.Values{}
	params.Set("symbol", order.Pair.ToSymbol("", true))
	params.Set("orderId", order.OrderId)
	if err := binance.buildParamsSigned(&params); err != nil {
		return nil, err
	}

	uri := API_V3 + ORDER_URI + params.Encode()
	response := remoteOrder{}
	resp, err := binance.DoRequest(
		"GET",
		uri,
		"",
		&response,
	)
	if err != nil {
		return nil, err
	}
	if response.OrderId <= 0 {
		return nil, errors.New(string(resp))
	}
	response.Merge(order, binance.config.Location)
	return resp, nil
}

func (binance *Spot) GetUnFinishOrders(pair Pair) ([]Order, []byte, error) {

	params := url.Values{}
	params.Set("symbol", pair.ToSymbol("", true))
	if err := binance.buildParamsSigned(&params); err != nil {
		return nil, nil, err
	}

	uri := API_V3 + UNFINISHED_ORDERS_INFO + params.Encode()
	remoteOrders := make([]*remoteOrder, 0)
	resp, err := binance.DoRequest("GET", uri, "", &remoteOrders)
	if err != nil {
		return nil, nil, err
	}

	orders := make([]Order, 0)
	for _, remoteOrder := range remoteOrders {
		order := Order{}
		remoteOrder.Merge(&order, binance.config.Location)
		orders = append(orders, order)
	}

	return orders, resp, nil
}

func (binance *Spot) GetOrderHistorys(pair Pair, currentPage, pageSize int) ([]Order, error) {
	panic("implement me")
}

func (binance *Spot) GetAccount() (*Account, []byte, error) {

	params := url.Values{}
	if err := binance.buildParamsSigned(&params); err != nil {
		return nil, nil, err
	}

	uri := API_V3 + ACCOUNT_URI + params.Encode()
	response := struct {
		Balances []*struct {
			Asset  string  `json:"asset"`
			Free   float64 `json:"free,string"`
			Locked float64 `json:"locked,string"`
		}
	}{}

	if resp, err := binance.DoRequest("GET", uri, "", &response); err != nil {
		return nil, nil, err
	} else {
		account := &Account{
			Exchange:    BINANCE,
			SubAccounts: make(map[Currency]SubAccount, 0),
		}

		for _, itm := range response.Balances {
			currency := NewCurrency(itm.Asset, "")
			account.SubAccounts[currency] = SubAccount{
				Currency:     currency,
				ForzenAmount: itm.Locked,
				Amount:       itm.Free,
			}
		}
		return account, resp, nil
	}
}

func (binance *Spot) GetTicker(pair Pair) (*Ticker, []byte, error) {
	tickerUri := API_V1 + fmt.Sprintf(TICKER_URI, pair.ToSymbol("", true))
	response := struct {
		Last   string `json:"lastPrice"`
		Buy    string `json:"bidPrice"`
		Sell   string `json:"askPrice"`
		Volume string `json:"volume"`

		Low       string `json:"lowPrice"`
		High      string `json:"highPrice"`
		Timestamp int64  `json:"closeTime"`
		Code      int64  `json:"code,-"`
		Message   string `json:"message,-"`
	}{}

	if resp, err := binance.DoRequest(
		"GET",
		tickerUri,
		"",
		&response,
	); err != nil {
		return nil, nil, err
	} else {
		var ticker Ticker
		ticker.Pair = pair
		ticker.Timestamp = response.Timestamp
		ticker.Date = time.Unix(
			response.Timestamp/1000,
			0,
		).In(binance.config.Location).Format(GO_BIRTHDAY)
		ticker.Last = ToFloat64(response.Last)
		ticker.Buy = ToFloat64(response.Buy)
		ticker.Sell = ToFloat64(response.Sell)
		ticker.Low = ToFloat64(response.Low)
		ticker.High = ToFloat64(response.High)
		ticker.Vol = ToFloat64(response.Volume)
		return &ticker, resp, nil
	}
}

func (binance *Spot) GetDepth(size int, pair Pair) (*Depth, []byte, error) {
	if size > 1000 {
		size = 1000
	} else if size < 5 {
		size = 5
	}

	response := struct {
		Code         int64           `json:"code,-"`
		Message      string          `json:"message,-"`
		Asks         [][]interface{} `json:"asks"` // The binance return asks Ascending ordered
		Bids         [][]interface{} `json:"bids"` // The binance return bids Descending ordered
		LastUpdateId int64           `json:"lastUpdateId"`
	}{}

	apiUri := fmt.Sprintf(API_V1+DEPTH_URI, pair.ToSymbol("", true), size)
	resp, err := binance.DoRequest(
		"GET",
		apiUri,
		"",
		&response,
	)

	depth := new(Depth)
	depth.Pair = pair
	now := time.Now()
	depth.Timestamp = now.UnixNano() / int64(time.Millisecond)
	depth.Date = now.In(binance.config.Location).Format(GO_BIRTHDAY)
	depth.Sequence = response.LastUpdateId

	for _, bid := range response.Bids {
		price := ToFloat64(bid[0])
		amount := ToFloat64(bid[1])
		dr := DepthRecord{Price: price, Amount: amount}
		depth.BidList = append(depth.BidList, dr)
	}

	for _, ask := range response.Asks {
		price := ToFloat64(ask[0])
		amount := ToFloat64(ask[1])
		dr := DepthRecord{Price: price, Amount: amount}
		depth.AskList = append(depth.AskList, dr)
	}

	return depth, resp, err
}

func (binance *Spot) GetKlineRecords(pair Pair, period, size, since int) ([]Kline, []byte, error) {
	startTimeFmt, endTimeFmt := fmt.Sprintf("%d", since), fmt.Sprintf("%d", time.Now().UnixNano())
	if len(startTimeFmt) > 13 {
		startTimeFmt = startTimeFmt[0:13]
	}

	if len(endTimeFmt) > 13 {
		endTimeFmt = endTimeFmt[0:13]
	}

	params := url.Values{}
	params.Set("symbol", strings.ToUpper(pair.ToSymbol("", true)))
	params.Set("interval", _INERNAL_KLINE_PERIOD_CONVERTER[period])
	params.Set("startTime", startTimeFmt)
	params.Set("endTime", endTimeFmt)
	params.Set("limit", fmt.Sprintf("%d", size))

	uri := API_V1 + KLINE_URI + "?" + params.Encode()
	klines := make([][]interface{}, 0)
	resp, err := binance.DoRequest("GET", uri, "", &klines)
	if err != nil {
		return nil, nil, err
	}

	var klineRecords []Kline
	for _, record := range klines {
		r := Kline{Pair: pair}
		for i, e := range record {
			switch i {
			case 0:
				r.Timestamp = int64(e.(float64))
				r.Date = time.Unix(
					int64(r.Timestamp)/1000,
					0,
				).In(binance.config.Location).Format(GO_BIRTHDAY)
			case 1:
				r.Open = ToFloat64(e)
			case 2:
				r.High = ToFloat64(e)
			case 3:
				r.Low = ToFloat64(e)
			case 4:
				r.Close = ToFloat64(e)
			case 5:
				r.Vol = ToFloat64(e)
			}
		}
		klineRecords = append(klineRecords, r)
	}

	return klineRecords, resp, nil
}

func (binance *Spot) GetTrades(pair Pair, since int64) ([]Trade, error) {
	panic("implement me")
}

func (binance *Spot) placeOrder(order *Order) ([]byte, error) {
	uri := API_V3 + ORDER_URI
	if order.Cid == "" {
		order.Cid = UUID()
	}

	orderSide := ""
	orderType := ""
	switch order.Side {
	case BUY:
		orderSide = "BUY"
		orderType = "LIMIT"
	case SELL:
		orderSide = "SELL"
		orderType = "LIMIT"
	case BUY_MARKET:
		orderSide = "BUY"
		orderType = "MARKET"
	case SELL_MARKET:
		orderSide = "SELL"
		orderType = "MARKET"
	default:
		return nil, errors.New("Can not deal the order side. ")
	}

	params := url.Values{}
	params.Set("symbol", order.Pair.ToSymbol("", true))
	params.Set("side", orderSide)
	params.Set("type", orderType)
	params.Set("quantity", fmt.Sprintf("%f", order.Amount))
	params.Set("newClientOrderId", order.Cid)

	switch order.OrderType {
	case NORMAL, ONLY_MAKER:
		params.Set("timeInForce", "GTC")
	case FOK:
		params.Set("timeInForce", "FOK")
	case IOC:
		params.Set("timeInForce", "IOC")
	default:
		params.Set("timeInForce", "GTC")
	}

	switch orderType {
	case "LIMIT":
		params.Set("price", fmt.Sprintf("%f", order.Price))
	}

	if err := binance.buildParamsSigned(&params); err != nil {
		return nil, err
	}

	response := remoteOrder{}
	resp, err := binance.DoRequest(
		"POST",
		uri,
		params.Encode(),
		&response,
	)
	if err != nil {
		return nil, err
	}

	if response.OrderId <= 0 {
		return nil, errors.New(string(resp))
	}
	response.Merge(order, binance.config.Location)
	return resp, nil
}
