package binance

import (
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

func (this *Spot) LimitBuy(amount, price string, currency CurrencyPair) (*Order, []byte, error) {
	panic("implement me")
}

func (this *Spot) LimitSell(amount, price string, currency CurrencyPair) (*Order, []byte, error) {
	panic("implement me")
}

func (this *Spot) MarketBuy(amount, price string, currency CurrencyPair) (*Order, []byte, error) {
	panic("implement me")
}

func (this *Spot) MarketSell(amount, price string, currency CurrencyPair) (*Order, []byte, error) {
	panic("implement me")
}

func (this *Spot) CancelOrder(orderId string, currency CurrencyPair) (bool, []byte, error) {
	panic("implement me")
}

func (this *Spot) GetOneOrder(orderId string, currency CurrencyPair) (*Order, []byte, error) {
	panic("implement me")
}

func (this *Spot) GetUnfinishOrders(currency CurrencyPair) ([]Order, []byte, error) {
	panic("implement me")
}

func (this *Spot) GetOrderHistorys(currency CurrencyPair, currentPage, pageSize int) ([]Order, error) {
	panic("implement me")
}

func (this *Spot) GetAccount() (*Account, []byte, error) {
	panic("implement me")
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

//func (this *Spot) GetExchangeName() string {
//	return BINANCE
//}
