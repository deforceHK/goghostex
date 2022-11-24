package binance

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
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

// public api
func (spot *Spot) GetTicker(pair Pair) (*Ticker, []byte, error) {
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

	if resp, err := spot.DoRequest(
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
		).In(spot.config.Location).Format(GO_BIRTHDAY)
		ticker.Last = ToFloat64(response.Last)
		ticker.Buy = ToFloat64(response.Buy)
		ticker.Sell = ToFloat64(response.Sell)
		ticker.Low = ToFloat64(response.Low)
		ticker.High = ToFloat64(response.High)
		ticker.Vol = ToFloat64(response.Volume)
		return &ticker, resp, nil
	}
}

func (spot *Spot) GetDepth(pair Pair, size int) (*Depth, []byte, error) {
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
	resp, err := spot.DoRequest(
		"GET",
		apiUri,
		"",
		&response,
	)

	depth := new(Depth)
	depth.Pair = pair
	now := time.Now()
	depth.Timestamp = now.UnixNano() / int64(time.Millisecond)
	depth.Date = now.In(spot.config.Location).Format(GO_BIRTHDAY)
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

func (spot *Spot) GetKlineRecords(pair Pair, period, size, since int) ([]*Kline, []byte, error) {
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
	resp, err := spot.DoRequest("GET", uri, "", &klines)
	if err != nil {
		return nil, nil, err
	}

	var klineRecords []*Kline
	for _, record := range klines {
		r := Kline{Pair: pair, Exchange: BINANCE}
		for i, e := range record {
			switch i {
			case 0:
				r.Timestamp = int64(e.(float64))
				r.Date = time.Unix(
					int64(r.Timestamp)/1000,
					0,
				).In(spot.config.Location).Format(GO_BIRTHDAY)
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
		klineRecords = append(klineRecords, &r)
	}

	return GetAscKline(klineRecords), resp, nil
}

func (spot *Spot) GetTrades(pair Pair, since int64) ([]*Trade, error) {
	panic("implement me")
}

// private api
func (spot *Spot) GetAccount() (*Account, []byte, error) {
	params := url.Values{}
	if err := spot.buildParamsSigned(&params); err != nil {
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

	if resp, err := spot.DoRequest("GET", uri, "", &response); err != nil {
		return nil, nil, err
	} else {
		account := &Account{
			Exchange:    BINANCE,
			SubAccounts: make(map[string]SubAccount, 0),
		}

		for _, itm := range response.Balances {
			currency := NewCurrency(itm.Asset, "")
			account.SubAccounts[strings.ToUpper(itm.Asset)] = SubAccount{
				Currency:     currency,
				AmountFrozen: itm.Locked,
				Amount:       itm.Free,
			}
		}
		return account, resp, nil
	}
}

func (spot *Spot) PlaceOrder(order *Order) ([]byte, error) {
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

	if err := spot.buildParamsSigned(&params); err != nil {
		return nil, err
	}

	response := remoteOrder{}
	resp, err := spot.DoRequest(
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
	response.merge(order, spot.config.Location)
	return resp, nil
}
func (spot *Spot) CancelOrder(order *Order) ([]byte, error) {
	if order.OrderId == "" {
		return nil, errors.New("You must get the order_id. ")
	}

	uri := API_V3 + ORDER_URI
	params := url.Values{}
	params.Set("symbol", order.Pair.ToSymbol("", true))
	params.Set("orderId", order.OrderId)
	if err := spot.buildParamsSigned(&params); err != nil {
		return nil, err
	}

	response := remoteOrder{}
	resp, err := spot.DoRequest(
		"DELETE",
		uri,
		params.Encode(),
		&response,
	)

	if err != nil {
		return nil, err
	}
	response.merge(order, spot.config.Location)
	return resp, nil
}

func (spot *Spot) GetOrder(order *Order) ([]byte, error) {
	if order.OrderId == "" {
		return nil, errors.New("You must get the order_id. ")
	}

	params := url.Values{}
	params.Set("symbol", order.Pair.ToSymbol("", true))
	params.Set("orderId", order.OrderId)
	if err := spot.buildParamsSigned(&params); err != nil {
		return nil, err
	}

	uri := API_V3 + ORDER_URI + params.Encode()
	response := remoteOrder{}
	resp, err := spot.DoRequest(
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
	response.merge(order, spot.config.Location)
	return resp, nil
}

func (spot *Spot) GetOrders(pair Pair) ([]*Order, error) {
	panic("implement me")
}

func (spot *Spot) GetUnFinishOrders(pair Pair) ([]*Order, []byte, error) {
	params := url.Values{}
	params.Set("symbol", pair.ToSymbol("", true))
	if err := spot.buildParamsSigned(&params); err != nil {
		return nil, nil, err
	}

	uri := API_V3 + UNFINISHED_ORDERS_INFO
	remoteOrders := make([]*remoteOrder, 0)
	resp, err := spot.DoRequest(
		http.MethodGet,
		uri,
		params.Encode(),
		&remoteOrders,
	)
	if err != nil {
		return nil, nil, err
	}

	orders := make([]*Order, 0)
	for _, remoteOrder := range remoteOrders {
		order := Order{}
		remoteOrder.merge(&order, spot.config.Location)
		orders = append(orders, &order)
	}

	return orders, resp, nil
}

func (spot *remoteOrder) merge(order *Order, location *time.Location) {
	if spot.TransactTime != 0 || spot.Time != 0 {
		ts := spot.Time
		if spot.TransactTime > spot.Time {
			ts = spot.TransactTime
		}
		transactTime := time.Unix(int64(ts)/1000, int64(ts)%1000)
		order.OrderDate = transactTime.In(location).Format(GO_BIRTHDAY)
		order.OrderTimestamp = spot.TransactTime
	}

	status, exist := _INTERNAL_ORDER_STATUS_REVERSE_CONVERTER[spot.Status]
	if !exist {
		status = ORDER_FAIL
	}

	if spot.Type == "LIMIT" && spot.Side == "SELL" {
		order.Side = BUY
	} else if spot.Type == "LIMIT" && spot.Side == "BUY" {
		order.Side = SELL
	} else if spot.Type == "MARKET" && spot.Side == "SELL" {
		order.Side = SELL_MARKET
	} else {
		order.Side = BUY_MARKET
	}

	order.Status = status
	order.OrderId = fmt.Sprintf("%d", spot.OrderId)
	order.Cid = spot.ClientOrderId
	order.Price = spot.Price
	order.Amount = spot.OrigQty
	order.AvgPrice = spot.CummulativeQuoteQty
	order.DealAmount = spot.ExecutedQty
}

func (spot *Spot) GetExchangeRule(pair Pair) (*Rule, []byte, error) {
	uri := "/api/v3/exchangeInfo"

	pairsInfo := struct {
		Symbols []map[string]json.RawMessage `json:"symbols"`
	}{}
	resp, err := spot.DoRequest(http.MethodGet, uri, "", &pairsInfo)
	if err != nil {
		return nil, resp, err
	}

	symbol := pair.ToSymbol("", true)
	for _, s := range pairsInfo.Symbols {
		input, _ := json.Marshal(s)
		r := struct {
			Symbol     string                   `json:"symbol"`
			BaseAsset  string                   `json:"baseAsset"`
			QuotaAsset string                   `json:"quotaAsset"`
			Filters    []map[string]interface{} `json:"filters"`
		}{}

		if err := json.Unmarshal(input, &r); err != nil {
			return nil, input, err
		} else {
			if r.Symbol != symbol {
				continue
			}

			basePrecision, counterPrecision, baseMinSize := int(0), int(0), float64(0)
			for _, f := range r.Filters {
				if f["filterType"] == "PRICE_FILTER" {
					counterPrecision = GetPrecision(ToFloat64(f["tickSize"]))
				}
				if f["filterType"] == "LOT_SIZE" {
					basePrecision = GetPrecision(ToFloat64(f["stepSize"]))
					baseMinSize = ToFloat64(f["minQty"])
				}
			}

			rule := Rule{
				Pair:          pair,
				Base:          NewCurrency(r.BaseAsset, ""),
				BasePrecision: basePrecision,
				BaseMinSize:   baseMinSize,

				Counter:          NewCurrency(r.QuotaAsset, ""),
				CounterPrecision: counterPrecision,
			}
			return &rule, input, nil
		}
	}

	return nil, resp, errors.New("Can not find the pair in exchange. ")
}

//util api
func (spot *Spot) KeepAlive() {
	nowTimestamp := time.Now().Unix() * 1000
	if (nowTimestamp - spot.config.LastTimestamp) < 5*1000 {
		return
	}
	_, _, _ = spot.GetTicker(Pair{Basis: BNB, Counter: BTC})
}

func (spot *Spot) GetOHLCs(symbol string, period, size, since int) ([]*OHLC, []byte, error) {
	panic("implement me")
}
