package bitstamp

import (
	"sort"
	"strings"
	"time"

	. "github.com/strengthening/goghostex"
)

type Spot struct {
	*Bitstamp
}

func (this *Spot) GetOrderHistorys(pair CurrencyPair, currentPage, pageSize int) ([]Order, error) {
	panic("implement me")
}

func (this *Spot) GetTrades(pair CurrencyPair, since int64) ([]Trade, error) {
	panic("implement me")
}

func (this *Spot) GetExchangeName() string {
	panic("implement me")
}

func (this *Spot) LimitBuy(amount, price string, pair CurrencyPair) (*Order, []byte, error) {
	panic("implement me")
}

func (this *Spot) LimitSell(amount, price string, pair CurrencyPair) (*Order, []byte, error) {
	panic("implement me")
}

func (this *Spot) MarketBuy(amount, price string, pair CurrencyPair) (*Order, []byte, error) {
	panic("implement me")
}

func (this *Spot) MarketSell(amount, price string, pair CurrencyPair) (*Order, []byte, error) {
	panic("implement me")
}

func (this *Spot) CancelOrder(orderId string, pair CurrencyPair) (bool, []byte, error) {
	panic("implement me")
}

func (this *Spot) GetOneOrder(orderId string, pair CurrencyPair) (*Order, []byte, error) {
	panic("implement me")
}

func (this *Spot) GetUnfinishOrders(pair CurrencyPair) ([]Order, []byte, error) {
	panic("implement me")
}

func (this *Spot) GetAccount() (*Account, []byte, error) {
	panic("implement me")
}

func (this *Spot) GetTicker(pair CurrencyPair) (*Ticker, []byte, error) {

	uri := "/api/v2/ticker/" + strings.ToLower(pair.ToSymbol(""))
	response := struct {
		High      float64 `json:"high,string"`
		Low       float64 `json:"low,string"`
		Last      float64 `json:"last,string"`
		Buy       float64 `json:"bid,string"`
		Sell      float64 `json:"ask,string"`
		Volume    float64 `json:"volume,string"`
		Timestamp float64 `json:"timestamp,string"`
	}{}

	resp, err := this.DoRequest("GET", uri, "", &response)
	if err != nil {
		return nil, nil, err
	}

	return &Ticker{
		Pair:      pair,
		Last:      ToFloat64(response.Last),
		High:      ToFloat64(response.High),
		Low:       ToFloat64(response.Low),
		Vol:       ToFloat64(response.Volume),
		Sell:      ToFloat64(response.Sell),
		Buy:       ToFloat64(response.Buy),
		Timestamp: int64(response.Timestamp) * 1000,
		Date: time.Unix(
			int64(response.Timestamp),
			0,
		).In(this.config.Location).Format(GO_BIRTHDAY),
	}, resp, nil
}

func (this *Spot) GetDepth(size int, pair CurrencyPair) (*Depth, []byte, error) {
	uri := "/api/v2/order_book/" + strings.ToLower(pair.ToSymbol(""))
	response := struct {
		Bids      [][]interface{} `json:"bids"`
		Asks      [][]interface{} `json:"asks"`
		Status    string          `json:"status"`
		Reason    string          `json:"reason"`
		Timestamp int64          `json:"timestamp,string"`
	}{}

	resp, err := this.DoRequest("GET", uri, "", &response) //&response)
	if err != nil {
		return nil, nil, err
	}

	dep := new(Depth)
	dep.Pair = pair
	dep.Timestamp = response.Timestamp * 1000
	dep.Date = time.Unix(
		int64(response.Timestamp)/1000,
		0,
	).In(this.config.Location).Format(GO_BIRTHDAY)

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

	sort.Sort(sort.Reverse(dep.AskList)) //reverse
	return dep, resp, nil
}

func (this *Spot) GetKlineRecords(pair CurrencyPair, period, size, since int) ([]Kline, []byte, error) {
	panic("implement me")
}
