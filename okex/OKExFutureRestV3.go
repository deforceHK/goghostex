package okex

import (
	"fmt"
	"net/url"
	"strconv"
	"time"

	. "github.com/strengthening/goghostex"
)

func (future *Future) getV3Ticker(pair Pair, contractType string) (*FutureTicker, []byte, error) {

	var uri = fmt.Sprintf(
		"/api/futures/v3/instruments/%s/ticker",
		future.GetInstrumentId(pair, contractType),
	)

	var response struct {
		InstrumentId string  `json:"instrument_id"`
		Last         float64 `json:"last,string"`
		High24h      float64 `json:"high_24h,string"`
		Low24h       float64 `json:"low_24h,string"`
		BestBid      float64 `json:"best_bid,string"`
		BestAsk      float64 `json:"best_ask,string"`
		Volume24h    float64 `json:"volume_24h,string"`
		Timestamp    string  `json:"timestamp"`
	}
	resp, err := future.DoRequest(
		"GET",
		uri,
		"",
		&response,
	)
	if err != nil {
		return nil, nil, err
	}

	date, _ := time.Parse(time.RFC3339, response.Timestamp)
	ticker := FutureTicker{
		Ticker: Ticker{
			Pair:      pair,
			Sell:      response.BestAsk,
			Buy:       response.BestBid,
			Low:       response.Low24h,
			High:      response.High24h,
			Last:      response.Last,
			Vol:       response.Volume24h,
			Timestamp: date.UnixNano() / int64(time.Millisecond),
			Date:      date.In(future.config.Location).Format(GO_BIRTHDAY),
		},
		ContractType: contractType,
		ContractName: response.InstrumentId,
	}

	return &ticker, resp, nil
}

func (future *Future) getV3Depth(
	pair Pair,
	contractType string,
	size int,
) (*FutureDepth, []byte, error) {

	fc, err := future.GetContract(pair, contractType)
	if err != nil {
		return nil, nil, err
	}

	urlPath := fmt.Sprintf("/api/futures/v3/instruments/%s/book?size=%d", fc.ContractName, size)
	var response struct {
		Bids      [][4]interface{} `json:"bids"`
		Asks      [][4]interface{} `json:"asks"`
		Timestamp string           `json:"timestamp"`
	}
	resp, err := future.DoRequest("GET", urlPath, "", &response)
	if err != nil {
		return nil, nil, err
	}
	date, _ := time.Parse(time.RFC3339, response.Timestamp)
	if future.config.Location != nil {
		date = date.In(future.config.Location)
	}

	var dep FutureDepth
	dep.Pair = pair
	dep.ContractType = contractType
	dep.DueTimestamp = fc.DueTimestamp
	dep.Timestamp = date.UnixNano() / int64(time.Millisecond)
	dep.Sequence = dep.Timestamp
	dep.Date = date.Format(GO_BIRTHDAY)
	for _, itm := range response.Asks {
		dep.AskList = append(dep.AskList, DepthRecord{
			Price:  ToFloat64(itm[0]),
			Amount: ToFloat64(itm[1])})
	}
	for _, itm := range response.Bids {
		dep.BidList = append(dep.BidList, DepthRecord{
			Price:  ToFloat64(itm[0]),
			Amount: ToFloat64(itm[1])})
	}

	return &dep, resp, nil
}

func (future *Future) getV3Limit(pair Pair, contractType string) (float64, float64, error) {

	fc, err := future.GetContract(pair, contractType)
	if err != nil {
		return 0, 0, err
	}

	urlPath := fmt.Sprintf("/api/futures/v3/instruments/%s/price_limit", fc.ContractName)
	response := struct {
		Highest float64 `json:"highest,string"`
		Lowest  float64 `json:"lowest,string"`
	}{}

	_, err = future.DoRequest("GET", urlPath, "", &response)
	if err != nil {
		return 0, 0, err
	}

	return response.Highest, response.Lowest, nil
}


func (future *Future) getV3KlineRecords(
	contractType string,
	pair Pair,
	period,
	size,
	since int,
) ([]*FutureKline, []byte, error) {
	uri := fmt.Sprintf(
		"/api/futures/v3/instruments/%s/candles?",
		future.GetInstrumentId(pair, contractType),
	)

	params := url.Values{}
	params.Add(
		"granularity",
		fmt.Sprintf("%d", _INERNAL_KLINE_PERIOD_CONVERTER[period]),
	)

	if since > 0 {
		ts, _ := strconv.ParseInt(strconv.Itoa(since)[0:10], 10, 64)
		startTime := time.Unix(ts, 0).UTC()
		endTime := time.Now().UTC()
		params.Add("start", startTime.Format(time.RFC3339))
		params.Add("end", endTime.Format(time.RFC3339))
	}

	var response [][]interface{}
	resp, err := future.DoRequest(
		"GET",
		uri+params.Encode(),
		"",
		&response,
	)
	if err != nil {
		return nil, nil, err
	}

	contract, err := future.getV3FutureContract(pair, contractType)
	if err != nil {
		return nil, nil, err
	}
	var klines []*FutureKline
	for _, itm := range response {
		t, _ := time.Parse(time.RFC3339, fmt.Sprint(itm[0]))
		klines = append(klines, &FutureKline{
			Kline: Kline{
				Timestamp: t.UnixNano() / int64(time.Millisecond),
				Date:      t.In(future.config.Location).Format(GO_BIRTHDAY),
				Pair:      pair,
				Exchange:  OKEX,
				Open:      ToFloat64(itm[1]),
				High:      ToFloat64(itm[2]),
				Low:       ToFloat64(itm[3]),
				Close:     ToFloat64(itm[4]),
				Vol:       ToFloat64(itm[5]),
			},
			DueTimestamp: contract.DueTimestamp,
			DueDate:      contract.DueDate,
			Vol2:         ToFloat64(itm[6]),
		})
	}

	return GetAscFutureKline(klines), resp, nil
}