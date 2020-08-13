package coinbase

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"sync"
	"time"

	. "github.com/strengthening/goghostex"
)

type Spot struct {
	*Coinbase
}

func (*Spot) LimitBuy(*Order) ([]byte, error) {
	panic("implement me")
}

func (*Spot) LimitSell(*Order) ([]byte, error) {
	panic("implement me")
}

func (*Spot) MarketBuy(*Order) ([]byte, error) {
	panic("implement me")
}

func (*Spot) MarketSell(*Order) ([]byte, error) {
	panic("implement me")
}

func (*Spot) CancelOrder(*Order) ([]byte, error) {
	panic("implement me")
}

func (*Spot) GetOneOrder(*Order) ([]byte, error) {
	panic("implement me")
}

func (*Spot) GetUnFinishOrders(pair Pair) ([]Order, []byte, error) {
	panic("implement me")
}

func (*Spot) GetOrderHistorys(pair Pair, currentPage, pageSize int) ([]Order, error) {
	panic("implement me")
}

func (*Spot) GetAccount() (*Account, []byte, error) {
	panic("implement me")
}

func (spot *Spot) GetTicker(pair Pair) (*Ticker, []byte, error) {
	t := struct {
		Volume float64 `json:"volume,string"`
		Buy float64 `json:"bid,string"`
		Sell float64 `json:"ask,string"`
	}{}

	s := struct {
		Last   float64 `json:"last,string"`
		High float64 `json:"high,string"`
		Low float64 `json:"low,string"`
	}{}

	wg := sync.WaitGroup{}
	wg.Add(2)

	var tickerResp, statResp []byte
	var tickerErr, statErr error
	go func() {
		defer wg.Done()
		uri := fmt.Sprintf("/products/%s/ticker", pair.ToSymbol("-", true))
		tickerResp, tickerErr = spot.DoRequest("GET", uri, "", &t)
	}()

	go func() {
		defer wg.Done()
		uri := fmt.Sprintf("/products/%s/stats", pair.ToSymbol("-", true))
		statResp, statErr = spot.DoRequest("GET", uri, "", &s)
	}()

	wg.Wait()

	if tickerErr!=nil{
		return nil,nil, tickerErr
	}
	if statErr!=nil{
		return nil,nil, statErr
	}

	now := time.Now()
	timestamp := now.UnixNano() / int64(time.Millisecond)
	datetime := now.In(spot.config.Location).Format(GO_BIRTHDAY)
	ticker := &Ticker{
		Pair:      pair,
		High:      s.High,
		Low:       s.Low,
		Last:      s.Last,
		Vol:       t.Volume,
		Buy:       t.Buy,
		Sell:      t.Sell,
		Timestamp: timestamp,
		Date:      datetime,
	}
	return ticker, tickerResp, nil
}

func (*Spot) GetDepth(size int, pair Pair) (*Depth, []byte, error) {
	panic("implement me")
}

func (spot *Spot) GetKlineRecords(pair Pair, period, size, since int) ([]Kline, []byte, error) {
	if size > 300 {
		return nil, nil, errors.New("Can not request more than 300. ")
	}

	granularity, exist := _INERNAL_KLINE_PERIOD_CONVERTER[period]
	if !exist {
		return nil, nil, errors.New("The coinbase just support 1min 5min 15min 6h 1day. ")
	}

	uri := fmt.Sprintf(
		"/products/%s/candles?",
		pair.ToSymbol("-", true),
	)

	params := url.Values{}

	if since > 0 {
		startTimeFmt := fmt.Sprintf("%d", since)
		if len(startTimeFmt) >= 10 {
			startTimeFmt = startTimeFmt[0:10]
		}
		ts, err := strconv.ParseInt(startTimeFmt, 10, 64)
		if err != nil {
			return nil, nil, err
		}
		startTime := time.Unix(ts, 0).UTC()
		endTime := time.Unix(ts+int64(size*granularity), 0).UTC()

		params.Add("start", startTime.Format(time.RFC3339))
		params.Add("end", endTime.Format(time.RFC3339))
	}

	params.Add("granularity", fmt.Sprintf("%d", granularity))
	var response [][]interface{}
	resp, err := spot.DoRequest(
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
		t := time.Unix(ToInt64(item[0]), 0)
		klines = append(klines, Kline{
			Exchange:  COINBASE,
			Timestamp: t.UnixNano() / int64(time.Millisecond),
			Date:      t.In(spot.config.Location).Format(GO_BIRTHDAY),
			Pair:      pair,
			Open:      ToFloat64(item[3]),
			High:      ToFloat64(item[2]),
			Low:       ToFloat64(item[1]),
			Close:     ToFloat64(item[4]),
			Vol:       ToFloat64(item[5])},
		)
	}

	return klines, resp, nil
}

func (*Spot) GetTrades(pair Pair, since int64) ([]Trade, error) {
	panic("implement me")
}

func (*Spot) GetExchangeName() string {
	return COINBASE
}
