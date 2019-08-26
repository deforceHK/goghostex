package binance

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
	*Binance
}

func (this *Spot) LimitBuy(order *Order) ([]byte, error) {
	if order.Side != BUY {
		return nil, errors.New("the order side is not BUY")
	}
	return this.placeOrder(order)
}

func (this *Spot) LimitSell(order *Order) ([]byte, error) {
	if order.Side != SELL {
		return nil, errors.New("the order side is not SELL")
	}

	return this.placeOrder(order)
}
func (this *Spot) MarketBuy(order *Order) ([]byte, error) {
	if order.Side != BUY_MARKET {
		return nil, errors.New("the order side is not BUY_MARKET")
	}
	return this.placeOrder(order)
}

func (this *Spot) MarketSell(order *Order) ([]byte, error) {
	if order.Side != SELL_MARKET {
		return nil, errors.New("the order side is not SELL_MARKET")
	}
	return this.placeOrder(order)
}

func (this *Spot) CancelOrder(order *Order) ([]byte, error) {
	pair := this.adaptCurrencyPair(order.Currency)
	uri := API_V3 + ORDER_URI
	params := url.Values{}
	params.Set("symbol", pair.ToSymbol(""))
	params.Set("orderId", order.OrderId)

	if err := this.buildParamsSigned(&params); err != nil {
		return nil, err
	}

	response := struct {
		Symbol              string  `json:"symbol"`
		OrderId             int64   `json:"orderId,sting"`
		ClientOrderId       string  `json:"clientOrderId,sting"`
		TransactTime        uint64  `json:"transactTime,sting"`
		Price               float64 `json:"price,string"`
		OrigQty             float64 `json:"origQty,string"`
		Status              string  `json:"status"`                     // NEW/FILLED/CANCELED
		ExecutedQty         float64 `json:"executedQty,string"`         // deal amount
		CummulativeQuoteQty float64 `json:"cummulativeQuoteQty,string"` //deal avg price
	}{}

	resp, err := this.DoRequest(
		"DELETE",
		uri,
		params.Encode(),
		&response,
	)

	if response.Status == "CANCELED" {
		order.Status = ORDER_CANCEL
		order.OrderId = fmt.Sprintf("%d", response.OrderId)
		order.Cid = response.ClientOrderId
		order.AvgPrice = response.CummulativeQuoteQty
		order.DealAmount = response.ExecutedQty
		transactTime := time.Unix(
			int64(response.TransactTime)/1000,
			int64(response.TransactTime)%1000,
		)
		order.OrderTimestamp = response.TransactTime
		order.OrderDate = transactTime.In(this.config.Location).Format(GO_BIRTHDAY)
	}

	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (this *Spot) GetOneOrder(order *Order) ([]byte, error) {
	pair := this.adaptCurrencyPair(order.Currency)

	params := url.Values{}
	params.Set("symbol", pair.ToSymbol(""))
	if order.OrderId == "" {
		return nil, errors.New("You must get the order_id. ")
	}
	params.Set("orderId", order.OrderId)

	if err := this.buildParamsSigned(&params); err != nil {
		return nil, err
	}

	uri := API_V3 + ORDER_URI + params.Encode()
	response := struct {
		Symbol              string  `json:"symbol"`
		OrderId             int64   `json:"orderId,sting"`
		ClientOrderId       string  `json:"clientOrderId,sting"`
		TransactTime        uint64  `json:"transactTime,sting"`
		Price               float64 `json:"price,string"`
		OrigQty             float64 `json:"origQty,string"`
		Status              string  `json:"status"`                     // NEW/FILLED/CANCELED
		ExecutedQty         float64 `json:"executedQty,string"`         // deal amount
		CummulativeQuoteQty float64 `json:"cummulativeQuoteQty,string"` //deal avg price
	}{}

	resp, err := this.DoRequest("GET", uri, "", &response)
	if err != nil {
		return nil, err
	}

	if response.OrderId <= 0 {
		return nil, errors.New(string(resp))
	}

	order.Cid = response.ClientOrderId
	order.AvgPrice = response.CummulativeQuoteQty
	order.DealAmount = response.ExecutedQty
	transactTime := time.Unix(
		int64(response.TransactTime)/1000,
		int64(response.TransactTime)%1000,
	)
	fmt.Println(response.TransactTime)
	order.OrderTimestamp = response.TransactTime
	order.OrderDate = transactTime.In(this.config.Location).Format(GO_BIRTHDAY)

	if response.Status == "FILLED" {
		order.Status = ORDER_FINISH
	} else {
		now := time.Now()
		order.OrderTimestamp = uint64(now.UnixNano())
		order.OrderDate = now.Format(GO_BIRTHDAY)
		order.Status = ORDER_UNFINISH
	}

	return resp, nil
}

func (this *Spot) GetUnfinishOrders(currency CurrencyPair) ([]Order, []byte, error) {
	panic("implement me")
}

func (this *Spot) GetOrderHistorys(currency CurrencyPair, currentPage, pageSize int) ([]Order, error) {
	panic("implement me")
}

func (this *Spot) GetAccount() (*Account, []byte, error) {

	params := url.Values{}
	if err := this.buildParamsSigned(&params); err != nil {
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

	if resp, err := this.DoRequest("GET", uri, "", &response); err != nil {
		return nil, nil, err
	} else {
		fmt.Println(string(resp))
		fmt.Println(response)

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

func (this *Spot) GetTicker(currency CurrencyPair) (*Ticker, []byte, error) {
	currency2 := this.adaptCurrencyPair(currency)
	tickerUri := API_V1 + fmt.Sprintf(TICKER_URI, strings.ToUpper(currency2.ToSymbol("")))
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

	if resp, err := this.DoRequest(
		"GET",
		tickerUri,
		"",
		&response,
	); err != nil {
		return nil, nil, err
	} else {
		var ticker Ticker
		ticker.Pair = currency
		ticker.Timestamp = uint64(response.Timestamp / 1000)
		ticker.Date = time.Unix(
			response.Timestamp/1000,
			0,
		).In(this.config.Location).Format(GO_BIRTHDAY)
		ticker.Last = ToFloat64(response.Last)
		ticker.Buy = ToFloat64(response.Buy)
		ticker.Sell = ToFloat64(response.Sell)
		ticker.Low = ToFloat64(response.Low)
		ticker.High = ToFloat64(response.High)
		ticker.Vol = ToFloat64(response.Volume)
		return &ticker, resp, nil
	}
}

func (this *Spot) GetDepth(size int, pair CurrencyPair) (*Depth, []byte, error) {
	if size > 1000 {
		size = 1000
	} else if size < 5 {
		size = 5
	}
	currencyPair2 := this.adaptCurrencyPair(pair)
	response := struct {
		Code    int64           `json:"code,-"`
		Message string          `json:"message,-"`
		Asks    [][]interface{} `json:"asks"`
		Bids    [][]interface{} `json:"bids"`
	}{}

	apiUri := fmt.Sprintf(API_V1+DEPTH_URI, currencyPair2.ToSymbol(""), size)
	resp, err := this.DoRequest(
		"GET",
		apiUri,
		"",
		&response,
	)

	depth := new(Depth)
	depth.Pair = pair
	for _, bid := range response.Bids {
		price := ToFloat64(bid[0])
		amount := ToFloat64(bid[1])
		dr := DepthRecord{price, amount}
		depth.BidList = append(depth.BidList, dr)
	}

	for _, ask := range response.Asks {
		price := ToFloat64(ask[0])
		amount := ToFloat64(ask[1])
		dr := DepthRecord{price, amount}
		depth.AskList = append(depth.AskList, dr)
	}

	return depth, resp, err
}

func (this *Spot) GetKlineRecords(pair CurrencyPair, period, size, since int) ([]Kline, []byte, error) {

	currency := this.adaptCurrencyPair(pair)
	params := url.Values{}
	params.Set("symbol", strings.ToUpper(currency.ToSymbol("")))
	params.Set("interval", _INERNAL_KLINE_PERIOD_CONVERTER[period])
	params.Set("startTime", strconv.Itoa(since/1000000))
	params.Set("endTime", strconv.Itoa(int(time.Now().UnixNano()/1000000)))
	params.Set("limit", fmt.Sprintf("%d", size))

	uri := API_V1 + KLINE_URI + "?" + params.Encode()
	klines := make([][]interface{}, 0)
	resp, err := this.DoRequest("GET", uri, "", &klines)
	if err != nil {
		return nil, nil, err
	}

	var klineRecords []Kline
	for _, record := range klines {
		r := Kline{Pair: currency}
		for i, e := range record {
			switch i {
			case 0:
				r.Timestamp = int64(e.(float64))
				r.Date = time.Unix(
					r.Timestamp/1000,
					0,
				).In(this.config.Location).Format(GO_BIRTHDAY)
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

func (this *Spot) GetTrades(currencyPair CurrencyPair, since int64) ([]Trade, error) {
	panic("implement me")
}

func (this *Spot) placeOrder(order *Order) ([]byte, error) {
	pair := this.adaptCurrencyPair(order.Currency)
	uri := API_V3 + ORDER_URI
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
	params.Set("symbol", pair.ToSymbol(""))
	params.Set("side", orderSide)
	params.Set("type", orderType)
	params.Set("quantity", fmt.Sprintf("%f", order.Amount))
	params.Set("timeInForce", "GTC")

	switch orderType {
	case "LIMIT":
		params.Set("price", fmt.Sprintf("%f", order.Price))
	}

	if err := this.buildParamsSigned(&params); err != nil {
		return nil, err
	}

	response := struct {
		Symbol              string  `json:"symbol"`
		OrderId             int64   `json:"orderId,sting"`
		ClientOrderId       string  `json:"clientOrderId,sting"`
		TransactTime        uint64  `json:"transactTime,sting"`
		Price               float64 `json:"price,string"`
		OrigQty             float64 `json:"origQty,string"`
		Status              string  `json:"status"`                     // NEW/FILLED/CANCELED
		ExecutedQty         float64 `json:"executedQty,string"`         // deal amount
		CummulativeQuoteQty float64 `json:"cummulativeQuoteQty,string"` //deal avg price
	}{}
	resp, err := this.DoRequest("POST", uri, params.Encode(), &response)
	if err != nil {
		return nil, err
	}
	fmt.Println(string(resp))
	if response.OrderId <= 0 {
		return nil, errors.New(string(resp))
	}

	order.OrderId = fmt.Sprintf("%d", response.OrderId)
	order.Cid = response.ClientOrderId
	order.AvgPrice = response.CummulativeQuoteQty
	order.DealAmount = response.ExecutedQty

	if response.Status == "FILLED" {
		transactTime := time.Unix(
			int64(response.TransactTime)/1000,
			int64(response.TransactTime)%1000,
		)
		order.OrderTimestamp = response.TransactTime
		order.OrderDate = transactTime.In(this.config.Location).Format(GO_BIRTHDAY)
		order.Status = ORDER_FINISH
	} else {
		now := time.Now()
		order.OrderTimestamp = uint64(now.UnixNano() / 1000)
		order.OrderDate = now.Format(GO_BIRTHDAY)
		order.Status = ORDER_UNFINISH
	}
	return resp, nil
}
