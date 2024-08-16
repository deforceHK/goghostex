package binance

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	. "github.com/deforceHK/goghostex"
)

const (
	FUTURE_CM_ENDPOINT    = "https://dapi.binance.com"
	FUTURE_UM_ENDPOINT    = "https://fapi.binance.com"
	FUTURE_KEEP_ALIVE_URI = "/dapi/v1/ping"

	FUTURE_TICKER_URI           = "/dapi/v1/ticker/24hr?"
	FUTURE_EXCHANGE_INFO_URI    = "/dapi/v1/exchangeInfo"
	FUTURE_UM_EXCHANGE_INFO_URI = "/fapi/v1/exchangeInfo"

	FUTURE_DEPTH_URI     = "/dapi/v1/depth?"
	FUTURE_KLINE_URI     = "/dapi/v1/continuousKlines?"
	FUTURE_CM_CANDLE_URI = "/dapi/v1/continuousKlines?"
	FUTURE_UM_CANDLE_URI = "/fapi/v1/continuousKlines?"
	FUTURE_TRADE_URI     = "/dapi/v1/trades?"

	FUTURE_INCOME_URI       = "/dapi/v1/income?"
	FUTURE_ACCOUNT_URI      = "/dapi/v1/account?"
	FUTURE_POSITION_URI     = "/dapi/v1/positionRisk?"
	FUTURE_PLACE_ORDER_URI  = "/dapi/v1/order?"
	FUTURE_CANCEL_ORDER_URI = "/dapi/v1/order?"
	FUTURE_GET_ORDER_URI    = "/dapi/v1/order?"
	FUTURE_GET_ORDERS_URI   = "/dapi/v1/allOrders?"
)

type Future struct {
	*Binance
	Locker                 sync.Locker
	Contracts              FutureContracts
	nextUpdateContractTime time.Time

	FutureContracts []*FutureContract
	LastTimestamp   int64
}

func (future *Future) GetTicker(pair Pair, contractType string) (*FutureTicker, []byte, error) {
	if contractType == THIS_WEEK_CONTRACT || contractType == NEXT_WEEK_CONTRACT {
		return nil, nil, errors.New("binance have not this_week next_week contract. ")
	}

	var contract, errContract = future.GetContract(pair, contractType)
	if errContract != nil {
		return nil, nil, errContract
	}
	var params = url.Values{}
	params.Add("symbol", future.getBNSymbol(contract.ContractName))

	var response = make([]struct {
		Symbol     string  `json:"symbol"`
		Pair       string  `json:"pair"`
		LastPrice  float64 `json:"lastPrice,string"`
		OpenPrice  float64 `json:"openPrice,string"`
		HighPrice  float64 `json:"highPrice,string"`
		LowPrice   float64 `json:"lowPrice,string"`
		Volume     float64 `json:"volume,string"`
		BaseVolume float64 `json:"baseVolume,string"`
	}, 0)

	var resp, err = future.DoRequest(
		http.MethodGet,
		FUTURE_CM_ENDPOINT,
		FUTURE_TICKER_URI+params.Encode(),
		"",
		&response,
	)
	if err != nil {
		return nil, nil, err
	}
	if len(response) == 0 {
		return nil, nil, errors.New("Can not find the pair. ")
	}

	var nowTime = time.Now()
	return &FutureTicker{
		Ticker: Ticker{
			Pair:      pair,
			Last:      response[0].LastPrice,
			Buy:       response[0].LastPrice,
			Sell:      response[0].LastPrice,
			High:      response[0].HighPrice,
			Low:       response[0].LowPrice,
			Vol:       response[0].BaseVolume,
			Timestamp: nowTime.UnixNano() / int64(time.Millisecond),
			Date:      nowTime.Format(GO_BIRTHDAY),
		},
		ContractType: contract.ContractType,
		ContractName: contract.ContractName,
	}, resp, nil
}

func (future *Future) GetDepth(pair Pair, contractType string, size int) (*FutureDepth, []byte, error) {
	if contractType == THIS_WEEK_CONTRACT || contractType == NEXT_WEEK_CONTRACT {
		return nil, nil, errors.New("binance have not this_week next_week contract. ")
	}
	var contract, err = future.GetContract(pair, contractType)
	if err != nil {
		return nil, nil, err
	}

	var params = url.Values{}
	params.Add("symbol", future.getBNSymbol(contract.ContractName))
	params.Add("limit", fmt.Sprintf("%d", size))

	response := struct {
		LastUpdateId int64      `json:"lastUpdateId"`
		E            int64      `json:"E"`
		T            int64      `json:"T"`
		Bids         [][]string `json:"bids"`
		Asks         [][]string `json:"asks"`
	}{}

	resp, err := future.DoRequest(
		http.MethodGet,
		FUTURE_CM_ENDPOINT,
		FUTURE_DEPTH_URI+params.Encode(),
		"",
		&response,
	)
	if err != nil {
		return nil, resp, err
	} else {
		dateFmt := time.Unix(response.E/1000, response.E%1000).In(future.config.Location).Format(GO_BIRTHDAY)
		depth := FutureDepth{
			ContractType: contract.ContractType,
			ContractName: contract.ContractName,
			Pair:         pair,
			Timestamp:    response.E,
			DueTimestamp: contract.DueTimestamp,
			Sequence:     response.LastUpdateId,
			Date:         dateFmt,
			AskList:      DepthRecords{},
			BidList:      DepthRecords{},
		}

		for _, items := range response.Asks {
			depth.AskList = append(depth.AskList, DepthRecord{Price: ToFloat64(items[0]), Amount: ToFloat64(items[1])})
		}
		for _, items := range response.Bids {
			depth.BidList = append(depth.BidList, DepthRecord{Price: ToFloat64(items[0]), Amount: ToFloat64(items[1])})
		}
		return &depth, resp, nil
	}
}

func (future *Future) GetLimit(pair Pair, contractType string) (float64, float64, error) {
	if contractType == THIS_WEEK_CONTRACT || contractType == NEXT_WEEK_CONTRACT {
		return 0, 0, errors.New("binance have not this_week next_week contract. ")
	}

	var contract, err = future.GetContract(pair, contractType)
	if err != nil {
		return 0, 0, err
	}

	var bnSymbol = future.getBNSymbol(contract.ContractName)
	var response = make([]struct {
		Symbol string  `json:"symbol"`
		Price  float64 `json:"markPrice,string"` //  mark price
	}, 0)

	if _, err := future.DoRequest(
		http.MethodGet,
		FUTURE_CM_ENDPOINT,
		fmt.Sprintf("/dapi/v1/premiumIndex?symbol=%s", bnSymbol),
		"",
		&response,
	); err != nil {
		return 0, 0, err
	}
	if len(response) == 0 {
		return 0, 0, errors.New("the remote return no data. ")
	}

	var highLimit = response[0].Price * contract.MaxScalePriceLimit
	var lowLimit = response[0].Price * contract.MinScalePriceLimit
	return highLimit, lowLimit, nil
}

func (future *Future) GetIndex(pair Pair) (float64, []byte, error) {
	panic("implement me")
}

func (future *Future) GetMark(pair Pair, contractType string) (float64, []byte, error) {
	if contractType == THIS_WEEK_CONTRACT || contractType == NEXT_WEEK_CONTRACT {
		return 0, nil, errors.New("binance have not this_week next_week contract. ")
	}
	var contract, errContract = future.GetContract(pair, contractType)
	if errContract != nil {
		return 0, nil, errContract
	}

	var bnSymbol = future.getBNSymbol(contract.ContractName)
	var response = make([]struct {
		Symbol string  `json:"symbol"`
		Price  float64 `json:"markPrice,string"` //  mark price
	}, 0)
	var resp, err = future.DoRequest(
		http.MethodGet,
		FUTURE_CM_ENDPOINT,
		fmt.Sprintf("/dapi/v1/premiumIndex?symbol=%s", bnSymbol),
		"",
		&response,
	)
	if err != nil {
		return 0, resp, err
	}
	if len(response) == 0 {
		return 0, resp, errors.New("the remote return no data. ")
	}
	return response[0].Price, resp, nil
}

func (future *Future) GetKlineRecords(
	contractType string,
	pair Pair,
	period, size, since int,
) ([]*FutureKline, []byte, error) {
	if contractType == THIS_WEEK_CONTRACT || contractType == NEXT_WEEK_CONTRACT {
		return nil, nil, errors.New("binance have not the this_week next_week contract. ")
	}

	var endTimestamp = since + size*_INERNAL_KLINE_SECOND_CONVERTER[period]
	if endTimestamp > since+200*24*60*60*1000 {
		endTimestamp = since + 200*24*60*60*1000
	}
	if endTimestamp > int(time.Now().Unix()*1000) {
		endTimestamp = int(time.Now().Unix() * 1000)
	}
	var paramContractType = "CURRENT_QUARTER"
	if contractType == NEXT_QUARTER_CONTRACT {
		paramContractType = "NEXT_QUARTER"
	}

	params := url.Values{}
	params.Set("pair", pair.ToSymbol("", true))
	params.Set("contractType", paramContractType)
	params.Set("interval", _INERNAL_KLINE_PERIOD_CONVERTER[period])
	params.Set("startTime", fmt.Sprintf("%d", since))
	params.Set("endTime", fmt.Sprintf("%d", endTimestamp))
	params.Set("limit", fmt.Sprintf("%d", size))

	uri := FUTURE_KLINE_URI + params.Encode()
	klines := make([][]interface{}, 0)
	resp, err := future.DoRequest(
		http.MethodGet,
		FUTURE_CM_ENDPOINT,
		uri,
		"",
		&klines,
	)
	if err != nil {
		return nil, resp, err
	}

	var list []*FutureKline
	for _, k := range klines {
		var timestamp = ToInt64(k[0])
		var _, dueBoard = GetDueTimestamp(timestamp)
		var dueTimestamp = dueBoard[contractType]
		var dueDate = time.Unix(dueTimestamp/1000, 0).In(future.config.Location).Format(GO_BIRTHDAY)
		var r = &FutureKline{
			Kline: Kline{
				Pair:      pair,
				Exchange:  BINANCE,
				Timestamp: timestamp,
				Date:      time.Unix(timestamp/1000, 0).In(future.config.Location).Format(GO_BIRTHDAY),
				Open:      ToFloat64(k[1]),
				High:      ToFloat64(k[2]),
				Low:       ToFloat64(k[3]),
				Close:     ToFloat64(k[4]),
				Vol:       ToFloat64(k[7]),
			},
			ContractType: contractType,
			DueTimestamp: dueTimestamp,
			DueDate:      dueDate,
			Vol2:         ToFloat64(k[5]),
		}
		list = append(list, r)
	}
	return GetAscFutureKline(list), resp, nil
}

func (future *Future) GetCandles(
	dueTimestamp int64,
	symbol string,
	period,
	size int,
	since int64,
) ([]*FutureCandle, []byte, error) {
	if resp, err := future.updateFutureContracts(); err != nil {
		return nil, resp, err
	}
	if future.FutureContracts == nil {
		return nil, nil, errors.New("future contracts have not update. ")
	}

	var contract *FutureContract = nil
	for _, c := range future.FutureContracts {
		if c.Symbol == symbol && c.DueTimestamp == dueTimestamp {
			contract = c
			break
		}
	}
	if contract == nil {
		return nil, nil, errors.New("the contract not found. ")
	}

	if contract.Type == FUTURE_TYPE_LINEAR {
		return future.getUMCandles(contract, period, size, since)
	} else {
		return future.getCMCandles(contract, period, size, since)
	}
}

func (future *Future) KeepAlive() {
	nowTimestamp := time.Now().Unix() * 1000
	// last timestamp in 5s, no need to keep alive
	if (nowTimestamp - future.LastTimestamp) < 5*1000 {
		return
	}

	_, _ = future.DoRequest(http.MethodGet, FUTURE_CM_ENDPOINT, FUTURE_KEEP_ALIVE_URI, "", nil)
}

func (future *Future) DoRequest(httpMethod, endPoint, uri, reqBody string, response interface{}) ([]byte, error) {
	resp, err := NewHttpRequest(
		future.config.HttpClient,
		httpMethod,
		endPoint+uri,
		reqBody,
		map[string]string{
			"X-MBX-APIKEY": future.config.ApiKey,
		},
	)
	if err != nil {
		return nil, err
	} else {
		var nowTimestamp = time.Now().Unix() * 1000
		if future.LastTimestamp < nowTimestamp {
			future.LastTimestamp = nowTimestamp
		}
		return resp, json.Unmarshal(resp, &response)
	}
}

func (future *Future) getFutureType(side, sidePosition string) FutureType {
	if side == "BUY" && sidePosition == "LONG" {
		return OPEN_LONG
	} else if side == "SELL" && sidePosition == "SHORT" {
		return OPEN_SHORT
	} else if side == "SELL" && sidePosition == "LONG" {
		return LIQUIDATE_LONG
	} else if side == "BUY" && sidePosition == "SHORT" {
		return LIQUIDATE_SHORT
	} else {
		panic("input error, do not use BOTH. ")
	}

}

// return the binance style symbol
func (future *Future) getBNSymbol(contractName string) string {
	var infos = strings.Split(contractName, "-")
	return infos[0] + infos[1] + "_" + infos[2]
}

func (future *Future) transferSubject(income float64, remoteSubject string) string {
	if remoteSubject == "TRANSFER" {
		if income > 0 {
			return SUBJECT_TRANSFER_IN
		}
		return SUBJECT_TRANSFER_OUT
	}

	if subject, exist := subjectKV[remoteSubject]; exist {
		return subject
	} else {
		return strings.ToLower(remoteSubject)
	}

}

func (future *Future) getCMCandles(
	contract *FutureContract,
	period, size int, since int64,
) ([]*FutureCandle, []byte, error) {

	var endTimestamp = since + int64(size*_INERNAL_KLINE_SECOND_CONVERTER[period])
	if endTimestamp > since+200*24*60*60*1000 {
		endTimestamp = since + 200*24*60*60*1000
	}
	if endTimestamp > time.Now().Unix()*1000 {
		endTimestamp = time.Now().Unix() * 1000
	}

	var pairBN = strings.ToUpper(strings.Replace(contract.Symbol, "_", "", -1))
	params := url.Values{}
	params.Set("pair", pairBN)
	params.Set("contractType", contract.ContractType)
	params.Set("interval", _INERNAL_KLINE_PERIOD_CONVERTER[period])
	params.Set("startTime", fmt.Sprintf("%d", since))
	params.Set("endTime", fmt.Sprintf("%d", endTimestamp))
	params.Set("limit", fmt.Sprintf("%d", size))

	var uri = FUTURE_CM_CANDLE_URI + params.Encode()
	var results = make([][]interface{}, 0)
	resp, err := future.DoRequest(
		http.MethodGet,
		FUTURE_CM_ENDPOINT,
		uri,
		"",
		&results,
	)
	if err != nil {
		return nil, resp, err
	}

	var candles []*FutureCandle = make([]*FutureCandle, 0)
	for _, r := range results {
		var timestamp = ToInt64(r[0])
		var date = time.Unix(timestamp/1000, 0).In(future.config.Location).Format(GO_BIRTHDAY)
		var dueTimestamp = contract.DueTimestamp
		var dueDate = time.Unix(dueTimestamp/1000, 0).In(future.config.Location).Format(GO_BIRTHDAY)

		var c = &FutureCandle{
			Symbol:       contract.Symbol,
			Exchange:     BINANCE,
			Timestamp:    timestamp,
			Date:         date,
			Open:         ToFloat64(r[1]),
			High:         ToFloat64(r[2]),
			Low:          ToFloat64(r[3]),
			Close:        ToFloat64(r[4]),
			Vol:          ToFloat64(r[7]),
			Vol2:         ToFloat64(r[5]),
			Type:         contract.Type,
			DueTimestamp: dueTimestamp,
			DueDate:      dueDate,
		}

		candles = append(candles, c)
	}
	return GetAscFutureCandle(candles), resp, nil
}

func (future *Future) getUMCandles(
	contract *FutureContract,
	period, size int, since int64,
) ([]*FutureCandle, []byte, error) {

	var endTimestamp = since + int64(size*_INERNAL_KLINE_SECOND_CONVERTER[period])
	if endTimestamp > since+200*24*60*60*1000 {
		endTimestamp = since + 200*24*60*60*1000
	}
	if endTimestamp > time.Now().Unix()*1000 {
		endTimestamp = time.Now().Unix() * 1000
	}

	var pairBN = strings.ToUpper(strings.Replace(contract.Symbol, "_", "", -1))
	params := url.Values{}
	params.Set("pair", pairBN)
	params.Set("contractType", contract.ContractType)
	params.Set("interval", _INERNAL_KLINE_PERIOD_CONVERTER[period])
	params.Set("startTime", fmt.Sprintf("%d", since))
	params.Set("endTime", fmt.Sprintf("%d", endTimestamp))
	params.Set("limit", fmt.Sprintf("%d", size))

	var uri = FUTURE_UM_CANDLE_URI + params.Encode()
	var results = make([][]interface{}, 0)
	resp, err := future.DoRequest(
		http.MethodGet,
		FUTURE_UM_ENDPOINT,
		uri,
		"",
		&results,
	)
	if err != nil {
		return nil, resp, err
	}

	var candles = make([]*FutureCandle, 0)
	for _, r := range results {
		var timestamp = ToInt64(r[0])
		var date = time.Unix(timestamp/1000, 0).In(future.config.Location).Format(GO_BIRTHDAY)
		var dueTimestamp = contract.DueTimestamp
		var dueDate = time.Unix(dueTimestamp/1000, 0).In(future.config.Location).Format(GO_BIRTHDAY)

		var c = &FutureCandle{
			Symbol:       contract.Symbol,
			Exchange:     BINANCE,
			Timestamp:    timestamp,
			Date:         date,
			Open:         ToFloat64(r[1]),
			High:         ToFloat64(r[2]),
			Low:          ToFloat64(r[3]),
			Close:        ToFloat64(r[4]),
			Vol:          ToFloat64(r[5]),
			Vol2:         ToFloat64(r[7]),
			Type:         contract.Type,
			DueTimestamp: dueTimestamp,
			DueDate:      dueDate,
		}

		candles = append(candles, c)
	}
	return GetAscFutureCandle(candles), resp, nil
}
