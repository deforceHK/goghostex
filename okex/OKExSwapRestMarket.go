package okex

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	. "github.com/deforceHK/goghostex"
)

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
	var contract = swap.getContract(pair)
	var params = &url.Values{}
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
		var price = ToFloat64(bid[0])
		var amountContract = ToFloat64(bid[1])
		var amount = swap.getAmount(price, amountContract, contract)
		var depthItem = DepthRecord{Price: price, Amount: amount}
		depth.BidList = append(depth.BidList, depthItem)
	}

	for _, ask := range response.Data[0].Asks {
		var price = ToFloat64(ask[0])
		var amountContract = ToFloat64(ask[1])
		var amount = swap.getAmount(price, amountContract, contract)
		var depthItem = DepthRecord{Price: price, Amount: amount}
		depth.AskList = append(depth.AskList, depthItem)
	}

	return depth, resp, nil
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

func (swap *Swap) GetContract(pair Pair) *SwapContract {
	return swap.getContract(pair)
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

	return response.Data[0].BuyLmt, response.Data[0].SellLmt, nil

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
