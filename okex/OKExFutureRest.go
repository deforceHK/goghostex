package okex

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	. "github.com/strengthening/goghostex"
)

type Future struct {
	*OKEx
	Locker                 sync.Locker
	LastUpdateContractTime time.Time
	Contracts              FutureContracts
}

var okTimestampFlags = []int64{
	1420185600000,
}

var okDueTimestampBoard = map[string][]int64{
	THIS_WEEK_CONTRACT:    {1420790400000},
	NEXT_WEEK_CONTRACT:    {1421395200000},
	QUARTER_CONTRACT:      {1427443200000},
	NEXT_QUARTER_CONTRACT: {1435305600000},
}

var okNextQuarterListKV = make(map[int64]int64, 0)        //k list_timestamp v due_timestamp
var okNextQuarterListReverseKV = make(map[int64]int64, 0) // k due_timestamp v list_timestamp

func init() {
	fridays := getLastFridayQuarter()
	for i := 2; i < len(fridays); i++ {
		var generateTS = fridays[i-2].AddDate(0, 0, -14).Unix() * 1000
		var nextQuarterTS = fridays[i].Unix() * 1000
		okNextQuarterListKV[generateTS] = nextQuarterTS
		okNextQuarterListReverseKV[nextQuarterTS] = generateTS
	}

	// 计算2500周，大概50年
	for i := int64(1); i < 2500; i++ {
		var ts = okTimestampFlags[0] + i*7*24*60*60*1000
		okTimestampFlags = append(okTimestampFlags, ts)
		if nextQuarterTS, exist := okNextQuarterListKV[ts]; exist {

			okDueTimestampBoard[THIS_WEEK_CONTRACT] = append(
				okDueTimestampBoard[THIS_WEEK_CONTRACT],
				okDueTimestampBoard[NEXT_WEEK_CONTRACT][len(okDueTimestampBoard[NEXT_WEEK_CONTRACT])-1],
			)
			okDueTimestampBoard[NEXT_WEEK_CONTRACT] = append(
				okDueTimestampBoard[NEXT_WEEK_CONTRACT],
				okDueTimestampBoard[QUARTER_CONTRACT][len(okDueTimestampBoard[QUARTER_CONTRACT])-1],
			)
			okDueTimestampBoard[QUARTER_CONTRACT] = append(
				okDueTimestampBoard[QUARTER_CONTRACT],
				okDueTimestampBoard[NEXT_QUARTER_CONTRACT][len(okDueTimestampBoard[NEXT_QUARTER_CONTRACT])-1],
			)
			okDueTimestampBoard[NEXT_QUARTER_CONTRACT] = append(
				okDueTimestampBoard[NEXT_QUARTER_CONTRACT],
				nextQuarterTS,
			)
		} else {
			okDueTimestampBoard[THIS_WEEK_CONTRACT] = append(
				okDueTimestampBoard[THIS_WEEK_CONTRACT],
				okDueTimestampBoard[NEXT_WEEK_CONTRACT][len(okDueTimestampBoard[NEXT_WEEK_CONTRACT])-1],
			)
			okDueTimestampBoard[NEXT_WEEK_CONTRACT] = append(
				okDueTimestampBoard[NEXT_WEEK_CONTRACT],
				okDueTimestampBoard[THIS_WEEK_CONTRACT][len(okDueTimestampBoard[THIS_WEEK_CONTRACT])-1]+7*24*60*60*1000,
			)
			okDueTimestampBoard[QUARTER_CONTRACT] = append(
				okDueTimestampBoard[QUARTER_CONTRACT],
				okDueTimestampBoard[QUARTER_CONTRACT][len(okDueTimestampBoard[QUARTER_CONTRACT])-1],
			)
			okDueTimestampBoard[NEXT_QUARTER_CONTRACT] = append(
				okDueTimestampBoard[NEXT_QUARTER_CONTRACT],
				okDueTimestampBoard[NEXT_QUARTER_CONTRACT][len(okDueTimestampBoard[NEXT_QUARTER_CONTRACT])-1],
			)
		}
	}
}

func getLastFridayQuarter() []time.Time {
	var lastFridayQuarter = make([]time.Time, 0)
	var loc, _ = time.LoadLocation("Asia/Shanghai")
	var months = []time.Month{time.March, time.June, time.September, time.December}
	for year := 2015; year < 2050; year += 1 {

		for _, month := range months {
			var lastFriday = time.Date(
				year, month, 1,
				16, 0, 0, 0, loc,
			)
			// 月末最后一天
			lastFriday = lastFriday.AddDate(0, 1, 0).AddDate(0, 0, -1)
			for lastFriday.Weekday() != time.Friday {
				lastFriday = lastFriday.AddDate(0, 0, -1)
			}
			lastFridayQuarter = append(lastFridayQuarter, lastFriday)
		}
	}
	return lastFridayQuarter
}

// 获取合约信息
func (future *Future) getFutureContract(pair Pair, contractType string) (*FutureContract, error) {
	future.Locker.Lock()
	defer future.Locker.Unlock()

	loc, _ := time.LoadLocation("Asia/Shanghai")
	now := time.Now().In(loc)

	if now.After(future.LastUpdateContractTime) {
		_, err := future.updateFutureContracts()
		//重试三次
		for i := 0; err != nil && i < 3; i++ {
			time.Sleep(time.Second)
			_, err = future.updateFutureContracts()
		}

		if err != nil {
			return nil, err
		}
	}

	currencies := strings.Split(pair.ToSymbol("_", false), "_")
	contractTypeItem := fmt.Sprintf("%s,%s,%s", currencies[0], currencies[1], contractType)
	if cf, exist := future.Contracts.ContractTypeKV[contractTypeItem]; !exist {
		return nil, errors.New("Can not find the contract by contract_type. ")
	} else {
		return cf, nil
	}

}

func (future *Future) updateFutureContracts() ([]byte, error) {
	var response struct {
		Code string `json:"code"`
		Msg  string `json:"msg"`
		Data []struct {
			Alias     string  `json:"alias"`
			CtVal     float64 `json:"ctVal,string"`
			CtValCcy  string  `json:"ctValCcy"`
			ExpTime   int64   `json:"expTime,string"`
			InstId    string  `json:"instId"`
			ListTime  int64   `json:"listTime,string"`
			SettleCcy string  `json:"settleCcy"`
			TickSz    float64 `json:"tickSz,string"`
			LotSz     float64 `json:"lotSz,string"`
			Uly       string  `json:"uly"`
			State     string  `json:"state"`
		} `json:"data"`
	}
	resp, err := future.DoRequest(
		http.MethodGet,
		"/api/v5/public/instruments?instType=FUTURES",
		"",
		&response,
	)

	if err != nil {
		return nil, err
	}
	if response.Code != "0" {
		return nil, errors.New(response.Msg)
	}
	if len(response.Data) == 0 {
		return nil, errors.New("The contract api not ready. ")
	}

	var nowTime = time.Now()
	futureContracts := FutureContracts{
		ContractTypeKV: make(map[string]*FutureContract, 0),
		ContractNameKV: make(map[string]*FutureContract, 0),
		DueTimestampKV: make(map[string]*FutureContract, 0),
	}

	var flag = (nowTime.Unix()*1000 - okTimestampFlags[0]) / (7 * 24 * 60 * 60 * 1000)
	var isFreshNext10min = false

	for _, item := range response.Data {
		// 只要有合约状态不是live，那就是十分钟后更新
		if isFreshNext10min == false && item.State != "live" {
			isFreshNext10min = true
		}

		var contractType = item.Alias
		var dueTimestamp = okDueTimestampBoard[contractType][flag]
		var dueTime = time.Unix(dueTimestamp/1000, 0).In(future.config.Location)
		var openTimestamp int64
		if tmpTS, exist := okNextQuarterListReverseKV[dueTimestamp]; exist {
			openTimestamp = tmpTS
		} else {
			openTimestamp = dueTimestamp - 14*24*60*60*1000
		}
		var openTime = time.Unix(openTimestamp/1000, 0).In(future.config.Location)

		pair := NewPair(item.Uly, "-")
		settleMode := SETTLE_MODE_BASIS
		if item.SettleCcy != strings.Split(item.Uly, "-")[0] {
			settleMode = SETTLE_MODE_COUNTER
		}

		pricePrecision, amountPrecision := int64(0), int64(0)
		for i := int64(0); item.TickSz < 1.0; i++ {
			item.TickSz *= 10
			pricePrecision += 1
		}

		for i := int64(0); item.LotSz < 1.0; i++ {
			item.LotSz *= 10
			amountPrecision += 1
		}

		contract := &FutureContract{
			Pair:         pair,
			Symbol:       pair.ToSymbol("_", false),
			Exchange:     OKEX,
			ContractType: contractType,
			ContractName: item.Uly + "-" + dueTime.Format("060102"),
			SettleMode:   settleMode,

			OpenTimestamp: openTime.UnixNano() / int64(time.Millisecond),
			OpenDate:      openTime.Format(GO_BIRTHDAY),

			DueTimestamp: dueTime.UnixNano() / int64(time.Millisecond),
			DueDate:      dueTime.Format(GO_BIRTHDAY),

			UnitAmount:      item.CtVal,
			PricePrecision:  pricePrecision,
			AmountPrecision: amountPrecision,
		}

		currencies := strings.Split(contract.Symbol, "_")
		contractTypeItem := fmt.Sprintf("%s,%s,%s", currencies[0], currencies[1], contract.ContractType)
		contractNameItem := fmt.Sprintf("%s,%s,%s", currencies[0], currencies[1], contract.ContractName)
		dueTimestampItem := fmt.Sprintf("%s,%s,%d", currencies[0], currencies[1], contract.DueTimestamp)
		futureContracts.ContractTypeKV[contractTypeItem] = contract
		futureContracts.ContractNameKV[contractNameItem] = contract
		futureContracts.DueTimestampKV[dueTimestampItem] = contract
	}
	future.Contracts = futureContracts

	var nextUpdateTime = time.Unix(okTimestampFlags[flag+1]/1000, 0).In(future.config.Location)
	if isFreshNext10min || futureContracts.ContractTypeKV["btc,usd,this_week"] == nil {
		nextUpdateTime = nowTime.Add(10 * time.Minute)
	}
	future.LastUpdateContractTime = nextUpdateTime
	return resp, nil
}

func (future *Future) GetExchangeName() string {
	return OKEX
}

// 获取instrument_id
func (future *Future) GetInstrumentId(pair Pair, contractAlias string) string {
	if contractAlias != NEXT_QUARTER_CONTRACT &&
		contractAlias != QUARTER_CONTRACT &&
		contractAlias != NEXT_WEEK_CONTRACT &&
		contractAlias != THIS_WEEK_CONTRACT {
		return contractAlias
	}
	fc, _ := future.getFutureContract(pair, contractAlias)
	return fc.ContractName
}

func (future *Future) GetContract(pair Pair, contractType string) (*FutureContract, error) {
	return future.getFutureContract(pair, contractType)
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

func (future *Future) GetAccount() (*FutureAccount, []byte, error) {
	var response = struct {
		Code string `json:"code"`
		Msg  string `json:"msg"`
		Data []struct {
			UTime   int64 `json:"uTime,string"`
			Details []struct {
				Ccy       string `json:"ccy"`
				Eq        string `json:"eq"`
				CashBal   string `json:"cashBal"`
				AvailEq   string `json:"availEq"`
				FrozenBal string `json:"frozenBal"`
				OrdFrozen string `json:"ordFrozen"`
				MgnRatio  string `json:"mgnRatio"`
				Upl       string `json:"upl"`
			} `json:"details"`
		} `json:"data"`
	}{}

	var urlPath = "/api/v5/account/balance"
	resp, err := future.DoRequest(
		http.MethodGet,
		urlPath,
		"",
		&response,
	)
	if err != nil {
		return nil, nil, err
	}
	if response.Code != "0" {
		return nil, nil, errors.New(response.Msg)
	}

	acc := new(FutureAccount)
	acc.Exchange = future.GetExchangeName()
	acc.SubAccount = make(map[Currency]FutureSubAccount, 0)

	for _, detail := range response.Data[0].Details {
		currency := NewCurrency(detail.Ccy, "")
		acc.SubAccount[currency] = FutureSubAccount{
			Currency:       currency,
			Margin:         ToFloat64(detail.FrozenBal), //总体被占用的保证金，
			MarginDealed:   ToFloat64(detail.FrozenBal) - ToFloat64(detail.OrdFrozen),
			MarginUnDealed: ToFloat64(detail.OrdFrozen),
			MarginRate:     ToFloat64(detail.MgnRatio),
			BalanceTotal:   ToFloat64(detail.CashBal),
			BalanceNet:     ToFloat64(detail.Eq),
			BalanceAvail:   ToFloat64(detail.AvailEq),
			ProfitReal:     0,
			ProfitUnreal:   ToFloat64(detail.Upl),
		}
	}

	return acc, resp, nil
}

func (future *Future) PlaceOrder(order *FutureOrder) ([]byte, error) {
	if order == nil {
		return nil, errors.New("ord param is nil")
	}
	if order.ContractName == "" {
		order.ContractName = future.GetInstrumentId(order.Pair, order.ContractType)
	}

	var sideInfo, _ = _INERNAL_V5_FUTURE_TYPE_CONVERTER[order.Type]
	var placeInfo, _ = _INERNAL_V5_FUTURE_PLACE_TYPE_CONVERTER[order.PlaceType]
	var request = struct {
		InstId  string `json:"instId"`
		TdMode  string `json:"tdMode"`
		Side    string `json:"side"`
		PosSide string `json:"posSide,omitempty"`
		OrdType string `json:"ordType"`
		Sz      string `json:"sz"`
		Px      string `json:"px"`
		ClOrdId string `json:"clOrdId,omitempty"`
	}{
		order.ContractName,
		"cross",
		sideInfo[0],
		sideInfo[1],
		placeInfo,
		strconv.FormatInt(order.Amount, 10),
		strconv.FormatFloat(order.Price, 'f', -1, 64),
		order.Cid,
	}

	var response = struct {
		Code string `json:"code"`
		Msg  string `json:"msg"`
		Data []struct {
			ClOrdId string `json:"clOrdId"`
			OrdId   string `json:"ordId"`
			SCode   string `json:"sCode"`
			SMsg    string `json:"sMsg"`
		} `json:"data"`
	}{}
	var uri = "/api/v5/trade/order"

	now := time.Now()
	order.PlaceTimestamp = now.UnixNano() / int64(time.Millisecond)
	order.PlaceDatetime = now.In(future.config.Location).Format(GO_BIRTHDAY)

	reqBody, _, _ := future.BuildRequestBody(request)
	resp, err := future.DoRequest(
		http.MethodPost,
		uri,
		reqBody,
		&response,
	)

	if err != nil {
		return resp, err
	}
	if len(response.Data) > 0 && response.Data[0].SCode != "0" {
		return resp, errors.New(string(resp)) //todo 更好的获取错误码的方案
	}
	if response.Code != "0" {
		return resp, errors.New(string(resp))
	}

	now = time.Now()
	order.DealTimestamp = now.UnixNano() / int64(time.Millisecond)
	order.DealDatetime = now.In(future.config.Location).Format(GO_BIRTHDAY)
	order.OrderId = response.Data[0].OrdId
	return resp, nil
}

func (future *Future) GetOrder(order *FutureOrder) ([]byte, error) {
	if order == nil {
		return nil, errors.New("ord param is nil")
	}
	if order.ContractName == "" {
		order.ContractName = future.GetInstrumentId(order.Pair, order.ContractType)
	}

	var params = url.Values{}
	params.Set("instId", order.ContractName)
	params.Set("ordId", order.OrderId)

	var response = struct {
		Code string `json:"code"`
		Msg  string `json:"msg"`
		Data []struct {
			ClOrdId   string  `json:"clOrdId"`
			OrdId     string  `json:"ordId"`
			Px        float64 `json:"px,string"`
			Sz        float64 `json:"sz,string"`
			AvgPx     string  `json:"avgPx"`
			AccFillSz float64 `json:"accFillSz,string"`
			State     string  `json:"state"`
			Lever     float64 `json:"lever,string"`
			Fee       float64 `json:"fee,string"`
			UTime     int64   `json:"uTime,string"`
			CTime     int64   `json:"cTime,string"`
		} `json:"data"`
	}{}
	var uri = "/api/v5/trade/order?"

	resp, err := future.DoRequest(
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
	if len(response.Data) == 0 || response.Data[0].State == "live" {
		return resp, nil
	}

	if status, exist := _INERNAL_V5_FUTURE_ORDER_STATUE_CONVERTER[response.Data[0].State]; exist {
		order.Status = status
	}
	if order.Exchange == "" {
		order.Exchange = future.GetExchangeName()
	}

	order.Price = response.Data[0].Px
	order.Amount = ToInt64(response.Data[0].Sz)

	order.AvgPrice = ToFloat64(response.Data[0].AvgPx)
	order.DealAmount = ToInt64(response.Data[0].AccFillSz)
	order.LeverRate = ToInt64(response.Data[0].Lever)
	order.Fee = response.Data[0].Fee

	order.DealTimestamp = response.Data[0].UTime
	order.DealDatetime = time.Unix(
		order.DealTimestamp/1000, 0,
	).In(future.config.Location).Format(GO_BIRTHDAY)

	order.PlaceTimestamp = response.Data[0].CTime
	order.PlaceDatetime = time.Unix(
		order.PlaceTimestamp/1000, 0,
	).In(future.config.Location).Format(GO_BIRTHDAY)
	return resp, err
}

func (future *Future) CancelOrder(order *FutureOrder) ([]byte, error) {
	if order == nil || order.OrderId == "" {
		return nil, errors.New("order necessary param is nil")
	}
	if order.ContractName == "" {
		order.ContractName = future.GetInstrumentId(order.Pair, order.ContractType)
	}

	var request = struct {
		InstId string `json:"instId"`
		OrdId  string `json:"ordId"`
	}{
		order.ContractName,
		order.OrderId,
	}

	var response = struct {
		Code string `json:"code"`
		Msg  string `json:"msg"`
		Data []struct {
			ClOrdId string `json:"clOrdId"`
			OrdId   string `json:"ordId"`
			SCode   string `json:"sCode"`
			SMsg    string `json:"sMsg"`
		} `json:"data"`
	}{}

	var uri = "/api/v5/trade/cancel-order"
	reqBody, _, _ := future.BuildRequestBody(request)
	resp, err := future.DoRequest(
		http.MethodPost,
		uri,
		reqBody,
		&response,
	)
	if err != nil {
		return resp, err
	}
	if len(response.Data) == 0 {
		return resp, errors.New("request lack the data. ")
	}
	if len(response.Data) != 0 && response.Data[0].SCode != "0" {
		return resp, errors.New(response.Data[0].SMsg)
	}

	return resp, nil
}

func (future *Future) GetOrders(
	pair Pair,
	contractType string,
) ([]*FutureOrder, []byte, error) {
	panic("")
}

func (future *Future) GetTrades(pair Pair, contractType string) ([]*Trade, []byte, error) {
	panic("")
}

func (future *Future) KeepAlive() {
	nowTimestamp := time.Now().Unix() * 1000
	// last in 5s, no need to keep alive.
	if (nowTimestamp - future.config.LastTimestamp) < 5*1000 {
		return
	}

	// call the rate api to update lastTimestamp
	_, _, _ = future.GetTicker(Pair{BTC, USD}, QUARTER_CONTRACT)
}
