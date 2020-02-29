package okex

import (
	"errors"
	"fmt"
	"math"
	"net/url"
	"strconv"
	"sync"
	"time"

	. "github.com/strengthening/goghostex"
)

//future contract info
type FutureContract struct {
	InstrumentID        string  `json:"instrument_id"`         //instrument_id for example：BTC-USD-180213
	UnderlyingIndex     string  `json:"underlying_index"`      // 标的指数，如：BTC-USD
	BaseCurrency        string  `json:"base_currency"`         // 交易货币，如：BTC-USD中的BTC ,BTC-USDT中的BTC
	QuoteCurrency       string  `json:"quote_currency"`        // 计价货币币种，如：BTC-USD中的USD ,BTC-USDT中的USDT
	SettlementCurrency  string  `json:"settlement_currency"`   // 盈亏结算和保证金币种，如：BTC
	TickSize            float64 `json:"tick_size,string"`      //下单价格精度
	TradeIncrement      string  `json:"trade_increment"`       //数量精度
	ContractVal         string  `json:"contract_val"`          //the contract vol in usd
	ContractValCurrency string  `json:"contract_val_currency"` //合约面值计价币种 如 usd，btc，ltc，etc xrp eos
	Listing             string  `json:"listing"`               // 上线日期
	Delivery            string  `json:"delivery"`              // delivery date 交割日期
	DueTimestamp        int64   `json:"due_timestamp"`
	DueDate             string  `json:"due_date"`
	Alias               string  `json:"alias"` // this_week next_week quarter
	IsInverse           bool    `json:"is_inverse,string"`
}

type FutureContracts struct {
	contractTypeKV map[string]FutureContract
	dueTimestampKV map[string]FutureContract
	uTime          time.Time
}

type Future struct {
	*OKEx
	sync.Locker
	Contracts FutureContracts
}

// 获取合约信息
func (ok *Future) getFutureContract(pair CurrencyPair, contractType string) FutureContract {
	loc, _ := time.LoadLocation("Asia/Shanghai")
	now := time.Now().In(loc)

	weekNum := int(now.Weekday())
	minusDay := 5 - weekNum
	if weekNum < 5 || (weekNum == 5 && now.Hour() <= 16) {
		minusDay = -7 + 5 - weekNum
	}
	//最晚更新时限。
	lastUpdateTime := time.Date(
		now.Year(), now.Month(), now.Day(),
		16, 0, 0, 0, now.Location(),
	).AddDate(0, 0, minusDay)

	if ok.Contracts.uTime.IsZero() || (ok.Contracts.uTime.Before(lastUpdateTime)) {
		_, err := ok.updateFutureContracts()
		//重试三次
		for i := 0; err != nil && i < 3; i++ {
			time.Sleep(time.Second)
			_, err = ok.updateFutureContracts()
		}
		if err != nil {
			panic(err)
		}
	}

	contractTypeItem := fmt.Sprintf(
		"%s,%s,%s",
		pair.CurrencyBasis.Symbol,
		pair.CurrencyCounter.Symbol,
		contractType,
	)
	cf, exist := ok.Contracts.contractTypeKV[contractTypeItem]
	if !exist {
		panic("Can not find the contract by contract_type. ")
	}
	return cf
}

func (ok *Future) updateFutureContracts() ([]byte, error) {
	uTime := time.Now().In(ok.config.Location)
	var response []FutureContract
	resp, err := ok.DoRequest("GET", "/api/futures/v3/instruments", "", &response)
	if err != nil {
		return nil, err
	}

	loc, _ := time.LoadLocation("Asia/Shanghai")

	futureContracts := FutureContracts{
		contractTypeKV: make(map[string]FutureContract, 0),
		dueTimestampKV: make(map[string]FutureContract, 0),
		uTime:          uTime,
	}

	for _, item := range response {
		dueTime, err := time.ParseInLocation("2006-01-02", item.Delivery, loc)
		if err != nil {
			return nil, err
		}

		dueTime = dueTime.Add(16 * time.Hour).In(ok.config.Location)
		item.DueDate = dueTime.Format(GO_BIRTHDAY)
		item.DueTimestamp = dueTime.UnixNano() / int64(time.Millisecond)

		contract1, contract2 := item, item
		contractTypeItem := fmt.Sprintf("%s,%s,%s", item.BaseCurrency, item.QuoteCurrency, item.Alias)
		dueTimestampItem := fmt.Sprintf("%s,%s,%d", item.BaseCurrency, item.QuoteCurrency, item.DueTimestamp)
		futureContracts.contractTypeKV[contractTypeItem] = contract1
		futureContracts.dueTimestampKV[dueTimestampItem] = contract2
	}

	ok.Contracts = futureContracts
	return resp, nil
}

func (ok *Future) GetExchangeName() string {
	return OKEX
}

// cny -> usd rate
func (ok *Future) GetRate() (float64, []byte, error) {
	var response struct {
		Rate         float64   `json:"rate,string"`
		InstrumentId string    `json:"instrument_id"` //USD_CNY
		Timestamp    time.Time `json:"timestamp"`
	}
	resp, err := ok.DoRequest("GET", "/api/futures/v3/rate", "", &response)
	if err != nil {
		return 0, nil, err
	}

	return response.Rate, resp, nil
}

func (ok *Future) GetFutureEstimatedPrice(currencyPair CurrencyPair) (float64, []byte, error) {
	urlPath := fmt.Sprintf(
		"/api/futures/v3/instruments/%s/estimated_price",
		ok.getFutureContract(currencyPair, QUARTER_CONTRACT).InstrumentID,
	)
	var response struct {
		InstrumentId    string  `json:"instrument_id"`
		SettlementPrice float64 `json:"settlement_price,string"`
		Timestamp       string  `json:"timestamp"`
	}
	resp, err := ok.DoRequest("GET", urlPath, "", &response)
	if err != nil {
		return 0, nil, err
	}
	return response.SettlementPrice, resp, nil
}

// 获取instrument_id
func (ok *Future) GetInstrumentId(pair CurrencyPair, contractAlias string) string {
	if contractAlias != QUARTER_CONTRACT && contractAlias != NEXT_WEEK_CONTRACT && contractAlias != THIS_WEEK_CONTRACT {
		return contractAlias
	}

	ok.Lock()
	defer ok.Unlock()
	fc := ok.getFutureContract(pair, contractAlias)
	return fc.InstrumentID
}

// 获取ok合约信息
func (ok *Future) GetFutureContract(pair CurrencyPair, contractType string) FutureContract {
	ok.Lock()
	defer ok.Unlock()
	return ok.getFutureContract(pair, contractType)
}

func (ok *Future) GetFutureTicker(currencyPair CurrencyPair, contractType string) (*FutureTicker, []byte, error) {
	var uri = fmt.Sprintf(
		"/api/futures/v3/instruments/%s/ticker",
		ok.GetInstrumentId(currencyPair, contractType),
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
	resp, err := ok.DoRequest(
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
			Pair:      currencyPair,
			Sell:      response.BestAsk,
			Buy:       response.BestBid,
			Low:       response.Low24h,
			High:      response.High24h,
			Last:      response.Last,
			Vol:       response.Volume24h,
			Timestamp: date.UnixNano() / int64(time.Millisecond),
			Date:      date.In(ok.config.Location).Format(GO_BIRTHDAY),
		},
		ContractType: contractType,
		ContractName: response.InstrumentId,
	}

	return &ticker, resp, nil
}

func (ok *Future) GetFutureDepth(
	currencyPair CurrencyPair,
	contractType string,
	size int,
) (*FutureDepth, []byte, error) {
	fc := ok.GetFutureContract(currencyPair, contractType)
	urlPath := fmt.Sprintf(
		"/api/futures/v3/instruments/%s/book?size=%d",
		fc.InstrumentID,
		size,
	)
	var response struct {
		Bids      [][4]interface{} `json:"bids"`
		Asks      [][4]interface{} `json:"asks"`
		Timestamp string           `json:"timestamp"`
	}
	resp, err := ok.DoRequest("GET", urlPath, "", &response)
	if err != nil {
		return nil, nil, err
	}
	date, _ := time.Parse(time.RFC3339, response.Timestamp)
	if ok.config.Location != nil {
		date = date.In(ok.config.Location)
	}

	var dep FutureDepth
	dep.Pair = currencyPair
	dep.ContractType = contractType
	dep.DueTimestamp = fc.DueTimestamp
	dep.Timestamp = date.UnixNano() / int64(time.Millisecond)
	dep.Sequence = dep.Timestamp
	dep.Date = date.Format(GO_BIRTHDAY)
	for _, itm := range response.Asks {
		dep.AskList = append(dep.AskList, FutureDepthRecord{
			Price:  ToFloat64(itm[0]),
			Amount: ToInt64(itm[1])})
	}
	for _, itm := range response.Bids {
		dep.BidList = append(dep.BidList, FutureDepthRecord{
			Price:  ToFloat64(itm[0]),
			Amount: ToInt64(itm[1])})
	}

	return &dep, resp, nil
}

func (ok *Future) GetFutureStdDepth(
	currencyPair CurrencyPair,
	contractType string,
	size int,
) (*FutureStdDepth, []byte, error) {
	fc := ok.GetFutureContract(currencyPair, contractType)
	urlPath := fmt.Sprintf("/api/futures/v3/instruments/%s/book?size=%d", fc.InstrumentID, size)
	var response struct {
		Bids      [][4]interface{} `json:"bids"`
		Asks      [][4]interface{} `json:"asks"`
		Timestamp string           `json:"timestamp"`
	}
	resp, err := ok.DoRequest("GET", urlPath, "", &response)
	if err != nil {
		return nil, nil, err
	}
	date, _ := time.Parse(time.RFC3339, response.Timestamp)
	if ok.config.Location != nil {
		date = date.In(ok.config.Location)
	}

	var dep FutureStdDepth
	dep.Pair = currencyPair
	dep.ContractType = contractType
	dep.Timestamp = date.UnixNano() / int64(time.Millisecond)
	dep.DueTimestamp = fc.DueTimestamp
	dep.Sequence = dep.Timestamp
	dep.Date = date.Format(GO_BIRTHDAY)
	for _, itm := range response.Asks {
		stdPrice := int64(math.Floor(ToFloat64(itm[0])*100000000 + 0.5))
		dep.AskList = append(dep.AskList, FutureStdDepthRecord{
			Price:  stdPrice,
			Amount: ToInt64(itm[1])})
	}
	for _, itm := range response.Bids {
		stdPrice := int64(math.Floor(ToFloat64(itm[0])*100000000 + 0.5))
		dep.BidList = append(dep.BidList, FutureStdDepthRecord{
			Price:  stdPrice,
			Amount: ToInt64(itm[1])})
	}

	return &dep, resp, nil
}

func (ok *Future) GetFutureLimit(pair CurrencyPair, contractType string) (float64, float64, error) {

	fc := ok.GetFutureContract(pair, contractType)
	urlPath := fmt.Sprintf("/api/futures/v3/instruments/%s/price_limit", fc.InstrumentID)

	response := struct {
		Highest float64 `json:"highest,string"`
		Lowest  float64 `json:"lowest,string"`
	}{}

	_, err := ok.DoRequest("GET", urlPath, "", &response)
	if err != nil {
		return 0, 0, err
	}

	return response.Highest, response.Lowest, nil
}

func (ok *Future) GetFutureIndex(currencyPair CurrencyPair) (float64, []byte, error) {
	//统一交易对，当周，次周，季度指数一样的
	urlPath := fmt.Sprintf(
		"/api/futures/v3/instruments/%s/index",
		ok.GetInstrumentId(currencyPair, QUARTER_CONTRACT),
	)
	var response struct {
		InstrumentId string  `json:"instrument_id"`
		Index        float64 `json:"index,string"`
		Timestamp    string  `json:"timestamp"`
	}
	resp, err := ok.DoRequest("GET", urlPath, "", &response)
	if err != nil {
		return 0, nil, err
	}
	return response.Index, resp, nil
}

type CrossedAccountInfo struct {
	MarginMode string  `json:"margin_mode"`
	Equity     float64 `json:"equity,string"`
}

func (ok *Future) GetFutureAccount() (*FutureAccount, []byte, error) {
	urlPath := "/api/futures/v3/accounts"
	var response struct {
		Info map[string]map[string]interface{}
	}

	resp, err := ok.DoRequest("GET", urlPath, "", &response)
	if err != nil {
		return nil, nil, err
	}

	acc := new(FutureAccount)
	acc.SubAccount = make(map[Currency]FutureSubAccount, 0)
	for c, info := range response.Info {
		if info["margin_mode"] == "crossed" {
			currency := NewCurrency(c, "")
			acc.SubAccount[currency] = FutureSubAccount{
				Currency:       currency,
				Margin:         ToFloat64(info["margin"]),
				MarginDealed:   ToFloat64(info["margin_frozen"]),
				MarginUnDealed: ToFloat64(info["margin_for_unfilled"]),
				MarginRate:     ToFloat64(info["margin_ratio"]),
				BalanceTotal:   ToFloat64(info["total_avail_balance"]),
				BalanceNet:     ToFloat64(info["equity"]),
				BalanceAvail:   ToFloat64(info["can_withdraw"]),
				ProfitReal:     ToFloat64(info["realized_pnl"]),
				ProfitUnreal:   ToFloat64(info["unrealized_pnl"]),
			}
		} else {
			//todo 逐仓模式
		}
	}

	return acc, resp, nil
}

func (ok *Future) normalizePrice(price float64, pair CurrencyPair) string {
	fc := ok.GetFutureContract(pair, QUARTER_CONTRACT)
	var bit = 0
	for fc.TickSize < 1 {
		bit++
		fc.TickSize *= 10
	}
	return FloatToString(price, bit)
}

//matchPrice:是否以对手价下单(0:不是 1:是)，默认为0;当取值为1时,price字段无效，当以对手价下单，order_type只能选择0:普通委托
func (ok *Future) PlaceFutureOrder(order *FutureOrder) ([]byte, error) {
	if order == nil {
		return nil, errors.New("ord param is nil")
	}

	urlPath := "/api/futures/v3/order"
	var param struct {
		ClientOid    string `json:"client_oid"`
		InstrumentId string `json:"instrument_id"`
		Type         int64  `json:"type"`
		OrderType    int64  `json:"order_type"`
		Price        string `json:"price"`
		Size         string `json:"size"`
		MatchPrice   int64  `json:"match_price"`
		Leverage     int64  `json:"leverage"`
	}

	var response struct {
		Result       bool   `json:"result"`
		ErrorMessage string `json:"error_message"`
		ErrorCode    string `json:"error_code"`
		ClientOid    string `json:"client_oid"`
		OrderId      string `json:"order_id"`
	}

	param.InstrumentId = ok.GetInstrumentId(order.Currency, order.ContractType)
	param.ClientOid = order.Cid
	param.Type = int64(order.Type)
	param.Price = ok.normalizePrice(order.Price, order.Currency)
	param.Size = fmt.Sprint(order.Amount)
	param.Leverage = order.LeverRate
	param.MatchPrice = order.MatchPrice
	param.OrderType = int64(order.PlaceType)

	//当matchPrice=1以对手价下单，order_type只能选择0:普通委托
	if param.MatchPrice == 1 && param.OrderType != 0 {
		println("注意:当matchPrice=1以对手价下单时，order_type只能选择0:普通委托")
		param.OrderType = 0
	}

	now := time.Now()
	order.PlaceTimestamp = now.UnixNano() / int64(time.Millisecond)
	order.PlaceDatetime = now.In(ok.config.Location).Format(GO_BIRTHDAY)

	reqBody, _, _ := ok.BuildRequestBody(param)
	resp, err := ok.DoRequest("POST", urlPath, reqBody, &response)
	if err != nil {
		return nil, err
	}

	now = time.Now()
	order.Cid = response.ClientOid
	order.OrderId = response.OrderId
	order.OrderTimestamp = now.UnixNano() / int64(time.Millisecond)
	order.OrderDate = now.In(ok.config.Location).Format(GO_BIRTHDAY)
	return resp, nil
}

func (ok *Future) adaptOrder(response futureOrderResponse, order *FutureOrder) {
	order.ContractName = response.InstrumentId
	if response.OrderId != "" {
		order.OrderId = response.OrderId
	}
	if response.ClientOid != "" {
		order.Cid = response.ClientOid
	}
	order.DealAmount = int64(response.FilledQty)
	order.AvgPrice = response.PriceAvg
	order.Status = ok.adaptOrderState(response.State)
	order.Fee = response.Fee
	order.OrderTimestamp = response.Timestamp.UnixNano() / int64(time.Millisecond)
	order.OrderDate = response.Timestamp.In(ok.config.Location).Format(GO_BIRTHDAY)
	if order.Exchange == "" {
		order.Exchange = ok.GetExchangeName()
	}
	return
}

func (ok *Future) GetFutureOrder(order *FutureOrder) ([]byte, error) {
	urlPath := fmt.Sprintf(
		"/api/futures/v3/orders/%s/%s",
		ok.GetInstrumentId(order.Currency, order.ContractType),
		order.OrderId,
	)

	var response futureOrderResponse
	resp, err := ok.DoRequest("GET", urlPath, "", &response)
	if err != nil {
		return nil, err
	}

	ok.adaptOrder(response, order)
	return resp, nil
}

func (ok *Future) CancelFutureOrder(order *FutureOrder) ([]byte, error) {
	urlPath := fmt.Sprintf(
		"/api/futures/v3/cancel_order/%s/%s",
		ok.GetInstrumentId(order.Currency, order.ContractType),
		order.OrderId,
	)
	var response struct {
		Result       bool   `json:"result"`
		OrderId      string `json:"order_id"`
		ClientOid    string `json:"client_oid"`
		InstrumentId string `json:"instrument_id"`
	}
	resp, err := ok.DoRequest("POST", urlPath, "", &response)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (ok *Future) GetFuturePosition(
	currencyPair CurrencyPair,
	contractType string,
) ([]FuturePosition, []byte, error) {
	urlPath := fmt.Sprintf(
		"/api/futures/v3/%s/position",
		ok.GetInstrumentId(currencyPair, contractType),
	)
	var response struct {
		Result     bool   `json:"result"`
		MarginMode string `json:"margin_mode"`
		Holding    []struct {
			InstrumentId         string    `json:"instrument_id"`
			LongQty              float64   `json:"long_qty,string"` //多
			LongAvailQty         float64   `json:"long_avail_qty,string"`
			LongAvgCost          float64   `json:"long_avg_cost,string"`
			LongSettlementPrice  float64   `json:"long_settlement_price,string"`
			LongMargin           float64   `json:"long_margin,string"`
			LongPnl              float64   `json:"long_pnl,string"`
			LongPnlRatio         float64   `json:"long_pnl_ratio,string"`
			LongUnrealisedPnl    float64   `json:"long_unrealised_pnl,string"`
			RealisedPnl          float64   `json:"realised_pnl,string"`
			Leverage             int       `json:"leverage,string"`
			ShortQty             float64   `json:"short_qty,string"`
			ShortAvailQty        float64   `json:"short_avail_qty,string"`
			ShortAvgCost         float64   `json:"short_avg_cost,string"`
			ShortSettlementPrice float64   `json:"short_settlement_price,string"`
			ShortMargin          float64   `json:"short_margin,string"`
			ShortPnl             float64   `json:"short_pnl,string"`
			ShortPnlRatio        float64   `json:"short_pnl_ratio,string"`
			ShortUnrealisedPnl   float64   `json:"short_unrealised_pnl,string"`
			LiquidationPrice     float64   `json:"liquidation_price,string"`
			CreatedAt            time.Time `json:"created_at,string"`
			UpdatedAt            time.Time `json:"updated_at"`
		}
	}
	resp, err := ok.DoRequest("GET", urlPath, "", &response)
	if err != nil {
		return nil, nil, err
	}

	var postions []FuturePosition

	if !response.Result {
		return nil, nil, errors.New("unknown error")
	}

	if response.MarginMode == "fixed" {
		panic("not support the fix future")
	}

	for _, pos := range response.Holding {
		postions = append(postions, FuturePosition{
			Symbol:         currencyPair,
			ContractType:   contractType,
			ContractId:     ToInt64(pos.InstrumentId[8:]),
			BuyAmount:      pos.LongQty,
			BuyAvailable:   pos.LongAvailQty,
			BuyPriceAvg:    pos.LongAvgCost,
			BuyPriceCost:   pos.LongAvgCost,
			BuyProfitReal:  pos.LongPnl,
			SellAmount:     pos.ShortQty,
			SellAvailable:  pos.ShortAvailQty,
			SellPriceAvg:   pos.ShortAvgCost,
			SellPriceCost:  pos.ShortAvgCost,
			SellProfitReal: pos.ShortPnl,
			ForceLiquPrice: pos.LiquidationPrice,
			LeverRate:      pos.Leverage,
			CreateDate:     pos.CreatedAt.Unix(),
		})
	}

	return postions, resp, nil
}

func (ok *Future) GetFutureOrders(
	orderIds []string,
	currencyPair CurrencyPair,
	contractType string,
) ([]FutureOrder, []byte, error) {
	panic("")
}

type futureOrderResponse struct {
	InstrumentId string    `json:"instrument_id"`
	ClientOid    string    `json:"client_oid"`
	OrderId      string    `json:"order_id"`
	Size         float64   `json:"size,string"`
	Price        float64   `json:"price,string"`
	FilledQty    float64   `json:"filled_qty,string"`
	PriceAvg     float64   `json:"price_avg,string"`
	Fee          float64   `json:"fee,string"`
	Type         int       `json:"type,string"`
	OrderType    int       `json:"order_type,string"`
	Pnl          float64   `json:"pnl,string"`
	Leverage     int       `json:"leverage,string"`
	ContractVal  float64   `json:"contract_val,string"`
	State        int       `json:"state,string"`
	Timestamp    time.Time `json:"timestamp,string"`
}

func (ok *Future) GetUnFinishFutureOrders(
	currencyPair CurrencyPair,
	contractType string,
) ([]FutureOrder, []byte, error) {
	urlPath := fmt.Sprintf(
		"/api/futures/v3/orders/%s?state=6&limit=100",
		ok.GetInstrumentId(currencyPair, contractType),
	)
	var response struct {
		Result    bool                  `json:"result"`
		OrderInfo []futureOrderResponse `json:"order_info"`
	}
	resp, err := ok.DoRequest("GET", urlPath, "", &response)
	if err != nil {
		return nil, nil, err
	}
	if !response.Result {
		return nil, nil, errors.New("error")
	}

	var ords []FutureOrder
	for _, itm := range response.OrderInfo {
		ord := FutureOrder{}
		ok.adaptOrder(itm, &ord)
		ords = append(ords, ord)
	}

	return ords, resp, nil
}

func (ok *Future) GetFee() (float64, error) { panic("") }

func (ok *Future) GetContractValue(currencyPair CurrencyPair) (float64, error) {
	fc := ok.GetFutureContract(currencyPair, QUARTER_CONTRACT)
	return ToFloat64(fc.ContractVal), nil
}

func (ok *Future) GetDeliveryTime() (int, int, int, int) {
	return 4, 16, 0, 0 //星期五，下午4点交割
}

/**
 * since : 单位秒,开始时间
**/
func (ok *Future) GetFutureKlineRecords(
	contractType string,
	pair CurrencyPair,
	period,
	size,
	since int,
) ([]FutureKline, []byte, error) {
	uri := fmt.Sprintf(
		"/api/futures/v3/instruments/%s/candles?",
		ok.GetInstrumentId(pair, contractType),
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
	resp, err := ok.DoRequest(
		"GET",
		uri+params.Encode(),
		"",
		&response,
	)
	if err != nil {
		return nil, nil, err
	}

	contract := ok.getFutureContract(pair, contractType)
	var klines []FutureKline
	for _, itm := range response {
		t, _ := time.Parse(time.RFC3339, fmt.Sprint(itm[0]))
		klines = append(klines, FutureKline{
			Kline: Kline{
				Timestamp: t.UnixNano() / int64(time.Millisecond),
				Date:      t.In(ok.config.Location).Format(GO_BIRTHDAY),
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

	return klines, resp, nil
}

func (ok *Future) GetTrades(contract_type string, currencyPair CurrencyPair, since int64) ([]Trade, error) {
	panic("")
}

func (ok *Future) GetFutureMarkPrice(pair CurrencyPair, contractType string) (float64, []byte, error) {
	uri := fmt.Sprintf(
		"/api/futures/v3/instruments/%s/mark_price",
		ok.GetInstrumentId(pair, contractType),
	)

	response := struct {
		InstrumentId string  `json:"instrument_id"`
		MarkPrice    float64 `json:"mark_price,string"`
		Timestamp    string  `json:"timestamp"`
	}{}

	resp, err := ok.DoRequest(
		"GET",
		uri,
		"",
		&response,
	)

	if err != nil {
		return 0, resp, err
	}

	return response.MarkPrice, resp, nil
}

//特殊接口
/*
 市价全平仓
 contract:合约ID
 oType：平仓方向：CLOSE_SELL平空，CLOSE_BUY平多
*/
func (ok *Future) MarketCloseAllPosition(currency CurrencyPair, contract string, oType int) (bool, []byte, error) {
	urlPath := "/api/futures/v3/close_position"
	var response struct {
		InstrumentId string `json:"instrument_id"`
		Result       bool   `json:"result"`
		Message      string `json:"message"`
		Code         int    `json:"code"`
	}

	var param struct {
		InstrumentId string `json:"instrument_id"`
		Direction    string `json:"direction"`
	}

	param.InstrumentId = ok.GetInstrumentId(currency, contract)
	if oType == int(LIQUIDATE_LONG) { //CLOSE_BUY {
		param.Direction = "long"
	} else {
		param.Direction = "short"
	}
	reqBody, _, _ := ok.BuildRequestBody(param)
	resp, err := ok.DoRequest("POST", urlPath, reqBody, &response)
	if err != nil {
		return false, nil, err
	}

	if !response.Result {
		return false, nil, errors.New(response.Message)
	}

	return true, resp, nil
}
