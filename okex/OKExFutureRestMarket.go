package okex

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	. "github.com/deforceHK/goghostex"
)

type Future struct {
	*OKEx
	Contracts FutureContracts

	Locker                 sync.Locker
	nextUpdateContractTime time.Time //  下一次更新合约时间
}

func (future *Future) GetTicker(pair Pair, contractType string) (*FutureTicker, []byte, error) {
	contract, err := future.GetContract(pair, contractType)
	if err != nil {
		return nil, nil, err
	}
	nowTimestamp := time.Now().UnixNano() / int64(time.Millisecond)
	if nowTimestamp > contract.DueTimestamp {
		return nil, nil, errors.New("The new contract is generating. ")
	}

	var params = url.Values{}
	params.Set("instId", contract.ContractName)
	var uri = "/api/v5/market/ticker?" + params.Encode()
	var response struct {
		Code string `json:"code"`
		Msg  string `json:"msg"`
		Data []struct {
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

	resp, err := future.DoRequestMarket(
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
		return nil, nil, errors.New("lack response data. ")
	}

	date := time.Unix(response.Data[0].Timestamp/1000, 0)
	ticker := FutureTicker{
		Ticker: Ticker{
			Pair:      pair,
			Sell:      response.Data[0].AskPx,
			Buy:       response.Data[0].BidPx,
			Low:       response.Data[0].Low24h,
			High:      response.Data[0].High24h,
			Last:      response.Data[0].Last,
			Vol:       response.Data[0].Volume24h,
			Timestamp: response.Data[0].Timestamp,
			Date:      date.In(future.config.Location).Format(GO_BIRTHDAY),
		},
		ContractType: contractType,
		ContractName: response.Data[0].InstId,
	}

	return &ticker, resp, nil
}

func (future *Future) GetDepth(
	pair Pair,
	contractType string,
	size int,
) (*FutureDepth, []byte, error) {
	contract, err := future.GetContract(pair, contractType)
	if err != nil {
		return nil, nil, err
	}
	nowTimestamp := time.Now().UnixNano() / int64(time.Millisecond)
	if nowTimestamp > contract.DueTimestamp {
		return nil, nil, errors.New("The new contract is listing. ")
	}

	if size < 20 {
		size = 20
	}
	if size > 400 {
		size = 400
	}

	var params = url.Values{}
	params.Set("instId", contract.ContractName)
	params.Set("sz", fmt.Sprintf("%d", size))

	var response struct {
		Code string `json:"code"`
		Msg  string `json:"msg"`
		Data []struct {
			Bids      [][4]string `json:"bids"`
			Asks      [][4]string `json:"asks"`
			Timestamp int64       `json:"timestamp,string"`
		} `json:"data"`
	}
	var uri = "/api/v5/market/books?"
	resp, err := future.DoRequestMarket(
		http.MethodGet,
		uri+params.Encode(),
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
		return nil, nil, errors.New("lack response data. ")
	}

	date := time.Unix(response.Data[0].Timestamp/1000, 0)
	var dep FutureDepth
	dep.Pair = pair
	dep.ContractType = contractType
	dep.DueTimestamp = contract.DueTimestamp
	dep.Timestamp = response.Data[0].Timestamp
	dep.Sequence = dep.Timestamp
	dep.Date = date.In(future.config.Location).Format(GO_BIRTHDAY)
	for _, itm := range response.Data[0].Asks {
		dep.AskList = append(dep.AskList, DepthRecord{
			Price:  ToFloat64(itm[0]),
			Amount: ToFloat64(itm[1])})
	}
	for _, itm := range response.Data[0].Bids {
		dep.BidList = append(dep.BidList, DepthRecord{
			Price:  ToFloat64(itm[0]),
			Amount: ToFloat64(itm[1])})
	}

	return &dep, resp, nil
}

func (future *Future) GetLimit(pair Pair, contractType string) (float64, float64, error) {
	info, err := future.GetContract(pair, contractType)
	if err != nil {
		return 0, 0, err
	}

	params := url.Values{}
	params.Set("instId", info.ContractName)
	var uri = "/api/v5/public/price-limit?" + params.Encode()
	var response = struct {
		Code string `json:"code"`
		Msg  string `json:"msg"`
		Data []struct {
			BuyLmt  float64 `json:"buyLmt,string"`
			SellLmt float64 `json:"sellLmt,string"`
		} `json:"data"`
	}{}

	_, err = future.DoRequestMarket(
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

// 次季生成日，在交割时间段前后kline所属contract_type对照
var listKlineKV = map[string]string{
	THIS_WEEK_CONTRACT:    NEXT_WEEK_CONTRACT,
	NEXT_WEEK_CONTRACT:    QUARTER_CONTRACT,
	QUARTER_CONTRACT:      NEXT_QUARTER_CONTRACT,
	NEXT_QUARTER_CONTRACT: THIS_WEEK_CONTRACT,
}

// 非次季生成日，在交割时间段前后kline所属contract_type对照
var nonListKlineKV = map[string]string{
	THIS_WEEK_CONTRACT:    NEXT_WEEK_CONTRACT,
	NEXT_WEEK_CONTRACT:    THIS_WEEK_CONTRACT,
	QUARTER_CONTRACT:      QUARTER_CONTRACT,
	NEXT_QUARTER_CONTRACT: NEXT_QUARTER_CONTRACT,
}

/**
 * since : 单位毫秒,开始时间
**/
func (future *Future) GetKlineRecords(
	contractType string,
	pair Pair,
	period,
	size,
	since int,
) ([]*FutureKline, []byte, error) {
	contract, err := future.GetContract(pair, contractType)
	if err != nil {
		return nil, nil, err
	}

	if size > 300 {
		size = 300
	}

	uri := "/api/v5/market/candles?"
	params := url.Values{}
	params.Set("instId", contract.ContractName)
	params.Set("bar", _INERNAL_V5_CANDLE_PERIOD_CONVERTER[period])
	params.Set("limit", strconv.Itoa(size))

	if since > 0 {
		endTime := time.Now()
		params.Set("before", strconv.Itoa(since))
		params.Set("after", strconv.Itoa(int(endTime.UnixNano()/1000000)))
	}

	var response struct {
		Code string     `json:"code"`
		Msg  string     `json:"msg"`
		Data [][]string `json:"data"`
	}
	resp, err := future.DoRequestMarket(
		http.MethodGet,
		uri+params.Encode(),
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
		return make([]*FutureKline, 0), resp, nil
	}

	var maxKlineTS = ToInt64(response.Data[len(response.Data)-1][0])
	if ToInt64(response.Data[0][0]) > maxKlineTS {
		maxKlineTS = ToInt64(response.Data[0][0])
	}
	var flag = (maxKlineTS - okTimestampFlags[0]) / (7 * 24 * 60 * 60 * 1000)
	var swapTimestamp = okTimestampFlags[flag]
	var dueTimestamp = okDueTimestampBoard[contractType][flag]
	var dueDate = time.Unix(dueTimestamp/1000, 0).In(future.config.Location).Format(GO_BIRTHDAY)

	// 如果是次季生成日，则情况有所不同。
	var prevContractType = nonListKlineKV[contractType]
	if _, exist := okNextQuarterListKV[swapTimestamp]; exist {
		prevContractType = listKlineKV[contractType]
	}
	var prevDueTimestamp = okDueTimestampBoard[prevContractType][flag-1]
	var prevDueDate = time.Unix(prevDueTimestamp/1000, 0).In(future.config.Location).Format(GO_BIRTHDAY)

	var klines []*FutureKline
	for _, itm := range response.Data {
		timestamp := ToInt64(itm[0])
		var ct = contractType
		var dt = dueTimestamp
		var dd = dueDate
		// 如果时间间隔小的话这样使用没问题，但是时间间隔长，ok这个设计没法实现。
		if timestamp < swapTimestamp && period <= KLINE_PERIOD_1H {
			ct = prevContractType
			dt = prevDueTimestamp
			dd = prevDueDate
		}

		t := time.Unix(timestamp/1000, 0)
		klines = append(klines, &FutureKline{
			Kline: Kline{
				Timestamp: timestamp,
				Date:      t.In(future.config.Location).Format(GO_BIRTHDAY),
				Pair:      pair,
				Exchange:  OKEX,
				Open:      ToFloat64(itm[1]),
				High:      ToFloat64(itm[2]),
				Low:       ToFloat64(itm[3]),
				Close:     ToFloat64(itm[4]),
				Vol:       ToFloat64(itm[6]),
			},
			ContractType: ct,
			DueTimestamp: dt,
			DueDate:      dd,
			Vol2:         ToFloat64(itm[5]),
		})
	}

	return GetAscFutureKline(klines), resp, nil
}

func (future *Future) GetCandles(
	dueTimestamp int64,
	symbol string,
	period,
	size,
	since int,
) ([]*FutureCandle, []byte, error) {

	var ct, err = future.getContractByDueTimestamp(symbol, dueTimestamp)
	if err != nil {
		return nil, nil, err
	}

	var instId = ct.ContractName
	if size > 300 {
		size = 300
	}

	var uri = "/api/v5/market/candles?"
	var params = url.Values{}
	params.Set("instId", instId)
	params.Set("bar", _INERNAL_V5_CANDLE_PERIOD_CONVERTER[period])
	params.Set("limit", strconv.Itoa(size))

	if since > 0 {
		endTime := time.Now()
		params.Set("before", strconv.Itoa(since))
		params.Set("after", strconv.Itoa(int(endTime.UnixNano()/1000000)))
	}

	var response struct {
		Code string     `json:"code"`
		Msg  string     `json:"msg"`
		Data [][]string `json:"data"`
	}
	resp, err := future.DoRequestMarket(
		http.MethodGet,
		uri+params.Encode(),
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
		return make([]*FutureCandle, 0), resp, nil
	}

	var candles []*FutureCandle
	for _, itm := range response.Data {
		var timestamp = ToInt64(itm[0])
		if timestamp <= ct.ListTimestamp || timestamp >= ct.DueTimestamp {
			continue
		}
		// this candle haven't been confirmed
		if itm[8] == "0" {
			continue
		}

		candles = append(candles, &FutureCandle{
			Symbol:       symbol,
			Exchange:     OKEX,
			Timestamp:    timestamp,
			Date:         time.Unix(timestamp/1000, 0).In(future.config.Location).Format(GO_BIRTHDAY),
			Open:         ToFloat64(itm[1]),
			High:         ToFloat64(itm[2]),
			Low:          ToFloat64(itm[3]),
			Close:        ToFloat64(itm[4]),
			Vol:          ToFloat64(itm[6]),
			Vol2:         ToFloat64(itm[5]),
			Type:         ct.Type,
			DueTimestamp: ct.DueTimestamp,
			DueDate:      ct.DueDate,
		})
	}

	return GetAscFutureCandle(candles), resp, nil
}

func (future *Future) GetIndex(pair Pair) (float64, []byte, error) {
	var params = url.Values{}
	params.Set("instId", pair.ToSymbol("-", true))
	var uri = "/api/v5/market/index-tickers?"

	var response struct {
		Code string `json:"code"`
		Msg  string `json:"msg"`
		Data []struct {
			IdxPx float64 `json:"idxPx,string"`
		} `json:"data"`
	}
	resp, err := future.DoRequestMarket(
		http.MethodGet,
		uri+params.Encode(),
		"",
		&response,
	)
	if err != nil {
		return 0, resp, err
	}
	if response.Code != "0" {
		return 0, resp, errors.New(response.Msg)
	}

	return response.Data[0].IdxPx, resp, nil
}

func (future *Future) GetMark(pair Pair, contractType string) (float64, []byte, error) {
	var instId = future.GetInstrumentId(pair, contractType)
	var params = url.Values{}
	params.Set("instId", instId)
	params.Set("instType", "FUTURES")

	var uri = "/api/v5/public/mark-price?"
	var response = struct {
		Code string `json:"code"`
		Msg  string `json:"msg"`
		Data []struct {
			MarkPx float64 `json:"MarkPx,string"`
		} `json:"data"`
	}{}

	resp, err := future.DoRequestMarket(
		http.MethodGet,
		uri+params.Encode(),
		"",
		&response,
	)

	if err != nil {
		return 0, resp, err
	}
	if response.Code != "0" {
		return 0, resp, errors.New(response.Msg)
	}

	return response.Data[0].MarkPx, resp, nil
}

func (future *Future) KeepAlive() {
	var nowTimestamp = time.Now().Unix() * 1000
	// last in 5s, no need to keep alive.
	if (nowTimestamp - future.config.LastTimestamp) < 5*1000 {
		return
	}

	// call the rate api to update lastTimestamp
	_, _, _ = future.GetTicker(Pair{BTC, USD}, QUARTER_CONTRACT)
}
