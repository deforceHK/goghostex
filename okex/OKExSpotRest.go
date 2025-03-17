package okex

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	. "github.com/deforceHK/goghostex"
)

type Spot struct {
	*OKEx

	sync.Locker
	nextUpdateTime time.Time
	Instruments    map[string]*Instrument
}

type PlaceOrderParam struct {
	ClientOid     string  `json:"client_oid"`
	Type          string  `json:"type"`
	Side          string  `json:"side"`
	InstrumentId  string  `json:"instrument_id"`
	OrderType     int     `json:"order_type"`
	Price         float64 `json:"price"`
	Size          float64 `json:"size"`
	Notional      float64 `json:"notional"`
	MarginTrading string  `json:"margin_trading,omitempty"`
}

type remoteOrder struct {
	Result       bool   `json:"result"`
	OrderId      string `json:"order_id"`
	ClientOid    string `json:"client_oid"`
	ErrorCode    string `json:"error_code"`
	ErrorMessage string `json:"error_message"`
}

func (this *remoteOrder) Merge(order *Order) error {
	if this.ErrorCode != "" {
		return errors.New(this.ErrorMessage)
	}
	order.OrderId = this.OrderId
	order.Cid = this.ClientOid
	return nil
}

type OrderResponse struct {
	ClientOid      string  `json:"client_oid"`
	OrderId        string  `json:"order_id"`
	Price          float64 `json:"price,string"`
	Size           float64 `json:"size,string"`
	Notional       string  `json:"notional"`
	Side           string  `json:"side"`
	Type           string  `json:"type"`
	FilledSize     string  `json:"filled_size"`
	FilledNotional string  `json:"filled_notional"`
	PriceAvg       string  `json:"price_avg"`
	State          int     `json:"state,string"`
	OrderType      int     `json:"order_type,string"`
	Timestamp      string  `json:"timestamp"`
}

// public api
func (spot *Spot) GetTicker(pair Pair) (*Ticker, []byte, error) {

	var params = url.Values{}
	params.Set("instId", pair.ToSymbol("-", true))

	var uri = "/api/v5/market/ticker?"

	var response struct {
		Code int64  `json:"code,string"`
		Msg  string `json:"msg"`
		Data []*struct {
			Last      float64 `json:"last,string"`
			Ask       float64 `json:"askPx,string"`
			Bid       float64 `json:"bidPx,string"`
			High      float64 `json:"high24h,string"`
			Low       float64 `json:"low24h,string"`
			Vol       float64 `json:"vol24h,string"`
			Timestamp int64   `json:"ts,string"`
		} `json:"data"`
	}

	resp, err := spot.DoRequest(http.MethodGet, uri+params.Encode(), "", &response)
	if err != nil {
		return nil, resp, err
	}
	if response.Code != 0 {
		return nil, resp, errors.New(response.Msg)
	}
	if len(response.Data) == 0 {
		return nil, resp, errors.New("The api data not ready. ")
	}

	return &Ticker{
		Pair:      pair,
		Last:      response.Data[0].Last,
		High:      response.Data[0].High,
		Low:       response.Data[0].Low,
		Sell:      response.Data[0].Ask,
		Buy:       response.Data[0].Bid,
		Vol:       response.Data[0].Vol,
		Timestamp: response.Data[0].Timestamp,
		Date:      time.UnixMilli(response.Data[0].Timestamp).In(spot.config.Location).Format(GO_BIRTHDAY),
	}, resp, nil
}

func (spot *Spot) GetDepth(pair Pair, size int) (*Depth, []byte, error) {
	uri := fmt.Sprintf(
		"/api/spot/v3/instruments/%s/book?size=%d",
		pair.ToSymbol("-", true),
		size,
	)

	var response struct {
		Asks      [][]interface{} `json:"asks"`
		Bids      [][]interface{} `json:"bids"`
		Timestamp string          `json:"timestamp"`
	}

	resp, err := spot.DoRequest("GET", uri, "", &response)
	if err != nil {
		return nil, nil, err
	}

	dep := new(Depth)
	dep.Pair = pair
	date, _ := time.Parse(time.RFC3339, response.Timestamp)
	dep.Timestamp = date.UnixNano() / int64(time.Millisecond)
	dep.Date = date.In(spot.config.Location).Format(GO_BIRTHDAY)
	dep.Sequence = dep.Timestamp

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

	return dep, resp, nil
}

func (spot *Spot) GetKlineRecords(pair Pair, period, size, since int) ([]*Kline, []byte, error) {
	uri := fmt.Sprintf(
		"/api/spot/v3/instruments/%s/candles?",
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
		sinceTime := time.Unix(ts, 0).UTC()
		endTime := time.Now().UTC()
		params.Add("start", sinceTime.Format(time.RFC3339))
		params.Add("end", endTime.Format(time.RFC3339))
	}
	granularity, isExist := _INERNAL_KLINE_PERIOD_CONVERTER[period]
	if !isExist {
		return nil, nil, errors.New("The period is not supported. ")
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

	var klines []*Kline
	for _, item := range response {
		t, _ := time.Parse(time.RFC3339, fmt.Sprint(item[0]))
		klines = append(klines, &Kline{
			Timestamp: t.UnixNano() / int64(time.Millisecond),
			Date:      t.In(spot.config.Location).Format(GO_BIRTHDAY),
			Exchange:  OKEX,
			Pair:      pair,
			Open:      ToFloat64(item[1]),
			High:      ToFloat64(item[2]),
			Low:       ToFloat64(item[3]),
			Close:     ToFloat64(item[4]),
			Vol:       ToFloat64(item[5])},
		)
	}

	return GetAscKline(klines), resp, nil
}

// 非个人，整个交易所的交易记录
func (spot *Spot) GetTrades(pair Pair, since int64) ([]*Trade, error) {
	panic("unsupported")
}

func (spot *Spot) GetExchangeRule(pair Pair) (*Rule, []byte, error) {
	uri := "/api/spot/v3/instruments"
	r := make([]struct {
		InstrumentId  string  `json:"instrument_id"`
		BaseCurrency  string  `json:"base_currency"`
		QuoteCurrency string  `json:"quote_currency"`
		MinSize       float64 `json:"min_size,string"`
		TickSize      float64 `json:"tick_size,string"`
		SizeIncrement float64 `json:"size_increment,string"`
	}, 0)

	resp, err := spot.DoRequest("GET", uri, "", &r)
	if err != nil {
		return nil, resp, err
	}

	symbol := pair.ToSymbol("-", true)
	for _, p := range r {
		if p.InstrumentId != symbol {
			continue
		}

		if raw, err := json.Marshal(p); err != nil {
			return nil, resp, err
		} else {
			rule := Rule{
				Pair:             pair,
				Base:             NewCurrency(p.BaseCurrency, ""),
				Counter:          NewCurrency(p.QuoteCurrency, ""),
				BaseMinSize:      p.MinSize,
				BasePrecision:    GetPrecision(p.SizeIncrement),
				CounterPrecision: GetPrecision(p.TickSize),
			}
			return &rule, raw, nil
		}
	}

	return nil, resp, errors.New("Can not find the pair in remote. ")
}

// private api
func (spot *Spot) GetAccount() (*Account, []byte, error) {
	uri := "/api/spot/v3/accounts"
	var response []struct {
		Currency  string
		Frozen    float64 `json:"frozen,string"`
		Hold      float64 `json:"hold,string"`
		Balance   float64 `json:"balance,string"`
		Available float64 `json:"available,string"`
		Holds     float64 `json:"holds,string"`
	}

	resp, err := spot.DoRequest(
		http.MethodGet,
		uri,
		"",
		&response,
	)
	if err != nil {
		return nil, nil, err
	}

	account := &Account{
		SubAccounts: make(map[string]SubAccount, 0),
	}
	for _, item := range response {
		currency := NewCurrency(item.Currency, "")
		account.SubAccounts[strings.ToUpper(item.Currency)] = SubAccount{
			Currency:     currency,
			Amount:       item.Available,
			AmountFrozen: item.Hold,
		}
	}
	return account, resp, nil
}

func (spot *Spot) GetOrders(pair Pair) ([]*Order, error) {
	panic("unsupported")
}

// util api
func (spot *Spot) KeepAlive() {
	nowTimestamp := time.Now().Unix() * 1000
	if (nowTimestamp - spot.config.LastTimestamp) < 5*1000 {
		return
	}
	_, _, _ = spot.GetTicker(Pair{Basis: BTC, Counter: USDT})
}

func (spot *Spot) GetOHLCs(symbol string, period, size, since int) ([]*OHLC, []byte, error) {
	panic("implement me")
}

func (spot *Spot) getInstruments(pair Pair) *Instrument {
	defer spot.Unlock()
	spot.Lock()

	var now = time.Now().In(spot.config.Location)
	if now.After(spot.nextUpdateTime) {
		_, err := spot.updateInstruments()
		//重试三次
		for i := 0; err != nil && i < 3; i++ {
			time.Sleep(time.Second)
			_, err = spot.updateInstruments()
		}
		// 初次启动必须可以吧。
		if spot.nextUpdateTime.IsZero() && err != nil {
			panic(err)
		}

	}
	//return spot.swapContracts.ContractNameKV[pair.ToSwapContractName()]
	return spot.Spot.Instruments[pair.ToSymbol("-", true)]
}

func (spot *Spot) GetInstruments(pair Pair) *Instrument {
	return spot.getInstruments(pair)
}

func (spot *Spot) updateInstruments() ([]byte, error) {

	var params = url.Values{}
	params.Set("instType", "SPOT")

	var uri = "/api/v5/public/instruments?"
	var response = struct {
		Code string        `json:"code"`
		Msg  string        `json:"msg"`
		Data []*Instrument `json:"data"`
	}{}
	resp, err := spot.DoRequestMarket(
		http.MethodGet,
		uri+params.Encode(),
		"",
		&response,
	)
	if err != nil {
		return resp, err
	}
	if response.Code != "0" {
		return resp, errors.New(response.Msg)
	}
	if len(response.Data) == 0 {
		return resp, errors.New("The api data not ready. ")
	}

	for _, item := range response.Data {
		if item.State != "live" {
			continue
		}
		item.PricePrecision = GetPrecisionInt64(ToFloat64(item.TickSz))
		item.AmountPrecision = GetPrecisionInt64(ToFloat64(item.LotSz))

		spot.Instruments[item.InstId] = item
	}

	// setting next update time.
	var nowTime = time.Now().In(spot.config.Location)
	var nextUpdateTime = time.Date(
		nowTime.Year(), nowTime.Month(), nowTime.Day(),
		16, 0, 0, 0, spot.config.Location,
	)
	if nowTime.Hour() >= 16 {
		nextUpdateTime = nextUpdateTime.AddDate(0, 0, 1)
	}

	spot.nextUpdateTime = nextUpdateTime
	return resp, nil
}
