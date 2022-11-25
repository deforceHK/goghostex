package kraken

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	. "github.com/strengthening/goghostex"
)

type Spot struct {
	*Kraken
}

func (s *Spot) GetTicker(pair Pair) (*Ticker, []byte, error) {
	//TODO implement me
	panic("implement me")
}

func (s *Spot) GetDepth(pair Pair, size int) (*Depth, []byte, error) {
	//TODO implement me
	panic("implement me")
}

func (s *Spot) GetKlineRecords(pair Pair, period, size, since int) ([]*Kline, []byte, error) {

	var startTimeFmt = fmt.Sprintf("%d", since)
	var pairStd = strings.ToUpper(pair.ToSymbol("", true))
	if pairStd == "BTCUSD" {
		pairStd = "XXBTZUSD"
	} else if pairStd == "ETHUSD" {
		pairStd = "XETHZUSD"
	}

	if len(startTimeFmt) > 13 {
		startTimeFmt = startTimeFmt[0:13]
	}

	var params = url.Values{}
	params.Set("pair", pairStd)
	params.Set("interval", _INERNAL_KLINE_PERIOD_CONVERTER[period])
	params.Set("since", startTimeFmt)

	var uri = API_V1 + KLINE_URI + "?" + params.Encode()
	var result = struct {
		Error  []string                   `json:"error"`
		Result map[string]json.RawMessage `json:"result"`
	}{}

	resp, err := s.DoRequest("GET", uri, "", &result)
	if err != nil {
		return nil, nil, err
	}

	if len(result.Error) != 0 {
		return nil, nil, errors.New(strings.Join(result.Error, ","))
	}

	var records = make([][]interface{}, 0)
	err = json.Unmarshal(result.Result[pairStd], &records)
	if err != nil {
		return nil, nil, err
	}

	var klineRecords []*Kline
	for _, record := range records {
		r := Kline{Pair: pair, Exchange: KRAKEN}
		for i, e := range record {
			switch i {
			case 0:
				r.Timestamp = int64(e.(float64)*1000)
				r.Date = time.Unix(
					r.Timestamp/1000,
					0,
				).In(s.config.Location).Format(GO_BIRTHDAY)
			case 1:
				r.Open = ToFloat64(e)
			case 2:
				r.High = ToFloat64(e)
			case 3:
				r.Low = ToFloat64(e)
			case 4:
				r.Close = ToFloat64(e)
			case 6:
				r.Vol = ToFloat64(e)
			}
		}
		klineRecords = append(klineRecords, &r)
	}

	return GetAscKline(klineRecords), resp, nil
}

func (s *Spot) GetTrades(pair Pair, since int64) ([]*Trade, error) {
	//TODO implement me
	panic("implement me")
}

func (s *Spot) GetAccount() (*Account, []byte, error) {
	//TODO implement me
	panic("implement me")
}

func (s *Spot) PlaceOrder(order *Order) ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (s *Spot) CancelOrder(order *Order) ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (s *Spot) GetOrder(order *Order) ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (s *Spot) GetOrders(pair Pair) ([]*Order, error) {
	//TODO implement me
	panic("implement me")
}

func (s *Spot) GetUnFinishOrders(pair Pair) ([]*Order, []byte, error) {
	//TODO implement me
	panic("implement me")
}

func (s *Spot) KeepAlive() {
	//TODO implement me
	panic("implement me")
}

func (s *Spot) GetOHLCs(symbol string, period, size, since int) ([]*OHLC, []byte, error) {
	//TODO implement me
	panic("implement me")
}
