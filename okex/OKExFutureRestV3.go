package okex

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	. "github.com/strengthening/goghostex"
)

//future contract info
type okexFutureContract struct {
	InstrumentID        string  `json:"instrument_id"`          //instrument_id for example：BTC-USD-180213
	UnderlyingIndex     string  `json:"underlying_index"`       // 标的指数，如：BTC-USD
	BaseCurrency        string  `json:"base_currency"`          // 交易货币，如：BTC-USD中的BTC ,BTC-USDT中的BTC
	QuoteCurrency       string  `json:"quote_currency"`         // 计价货币币种，如：BTC-USD中的USD ,BTC-USDT中的USDT
	SettlementCurrency  string  `json:"settlement_currency"`    // 盈亏结算和保证金币种，如：BTC
	TickSize            float64 `json:"tick_size,string"`       //下单价格精度
	TradeIncrement      float64 `json:"trade_increment,string"` //数量精度
	ContractVal         float64 `json:"contract_val,string"`    //the contract vol in usd
	ContractValCurrency string  `json:"contract_val_currency"`  //合约面值计价币种 如 usd，btc，ltc，etc xrp eos
	Listing             string  `json:"listing"`                // 上线日期
	Delivery            string  `json:"delivery"`               // delivery date 交割日期
	DueTimestamp        int64   `json:"due_timestamp"`
	DueDate             string  `json:"due_date"`
	Alias               string  `json:"alias"` // this_week next_week quarter next_quarter
	IsInverse           bool    `json:"is_inverse,string"`
}

// 获取合约信息
func (future *Future) getV3FutureContract(pair Pair, contractType string) (*FutureContract, error) {
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

	if future.Contracts.SyncTime.IsZero() || (future.Contracts.SyncTime.Before(lastUpdateTime)) {
		_, err := future.updateV3FutureContracts()
		//重试三次
		for i := 0; err != nil && i < 3; i++ {
			time.Sleep(time.Second)
			_, err = future.updateV3FutureContracts()
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

func (future *Future) updateV3FutureContracts() ([]byte, error) {
	var response []okexFutureContract
	resp, err := future.DoRequest(
		http.MethodGet, "/api/futures/v3/instruments", "", &response,
	)
	if err != nil {
		return nil, err
	}
	SyncTime := time.Now().In(future.config.Location)
	asia, _ := time.LoadLocation("Asia/Shanghai")
	futureContracts := FutureContracts{
		ContractTypeKV: make(map[string]*FutureContract, 0),
		ContractNameKV: make(map[string]*FutureContract, 0),
		DueTimestampKV: make(map[string]*FutureContract, 0),
		SyncTime:       SyncTime,
	}

	for _, item := range response {
		dueTime, err := time.ParseInLocation("2006-01-02", item.Delivery, asia)
		if err != nil {
			return nil, err
		}

		openTime, err := time.ParseInLocation("2006-01-02", item.Listing, asia)
		if err != nil {
			return nil, err
		}

		dueTime = dueTime.Add(16 * time.Hour).In(future.config.Location)
		openTime = openTime.Add(16 * time.Hour).In(future.config.Location)

		contractType := item.Alias
		if contractType == "bi_quarter" {
			contractType = NEXT_QUARTER_CONTRACT
			item.Alias = NEXT_QUARTER_CONTRACT
		}

		pair := Pair{Basis: NewCurrency(item.BaseCurrency, ""), Counter: NewCurrency(item.QuoteCurrency, "")}
		settleMode := SETTLE_MODE_BASIS
		if item.QuoteCurrency == item.SettlementCurrency {
			settleMode = SETTLE_MODE_COUNTER
		}

		pricePrecision, amountPrecision := int64(0), int64(0)
		for i := int64(0); item.TickSize < 1.0; i++ {
			item.TickSize *= 10
			pricePrecision += 1
		}

		for i := int64(0); item.TradeIncrement < 1.0; i++ {
			item.TradeIncrement *= 10
			amountPrecision += 1
		}

		contract := &FutureContract{
			Pair:         pair,
			Symbol:       pair.ToSymbol("_", false),
			Exchange:     OKEX,
			ContractType: contractType,
			ContractName: item.InstrumentID,
			SettleMode:   settleMode,

			OpenTimestamp: openTime.UnixNano() / int64(time.Millisecond),
			OpenDate:      openTime.Format(GO_BIRTHDAY),

			DueTimestamp: dueTime.UnixNano() / int64(time.Millisecond),
			DueDate:      dueTime.Format(GO_BIRTHDAY),

			UnitAmount:      float64(item.ContractVal),
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
	return resp, nil
}

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

//matchPrice:是否以对手价下单(0:不是 1:是)，默认为0;当取值为1时,price字段无效，当以对手价下单，order_type只能选择0:普通委托
func (future *Future) placeV3Order(order *FutureOrder) ([]byte, error) {
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

	param.InstrumentId = future.GetInstrumentId(order.Pair, order.ContractType)
	param.ClientOid = order.Cid
	param.Type = int64(order.Type)
	param.Price = future.normalizePrice(order.Price, order.Pair)
	param.Size = fmt.Sprint(order.Amount)
	param.Leverage = order.LeverRate
	//param.MatchPrice = order.MatchPrice
	param.OrderType = int64(order.PlaceType)

	//当matchPrice=1以对手价下单，order_type只能选择0:普通委托
	if param.MatchPrice == 1 && param.OrderType != 0 {
		println("注意:当matchPrice=1以对手价下单时，order_type只能选择0:普通委托")
		param.OrderType = 0
	}

	now := time.Now()
	order.PlaceTimestamp = now.UnixNano() / int64(time.Millisecond)
	order.PlaceDatetime = now.In(future.config.Location).Format(GO_BIRTHDAY)

	reqBody, _, _ := future.BuildRequestBody(param)
	resp, err := future.DoRequest("POST", urlPath, reqBody, &response)
	if err != nil {
		return nil, err
	}

	now = time.Now()
	order.Cid = response.ClientOid
	order.OrderId = response.OrderId
	order.DealTimestamp = now.UnixNano() / int64(time.Millisecond)
	order.DealDatetime = now.In(future.config.Location).Format(GO_BIRTHDAY)
	return resp, nil
}

func (future *Future) adaptOrder(response futureOrderResponse, order *FutureOrder) {
	order.ContractName = response.InstrumentId
	if response.OrderId != "" {
		order.OrderId = response.OrderId
	}
	if response.ClientOid != "" {
		order.Cid = response.ClientOid
	}
	order.DealAmount = int64(response.FilledQty)
	order.AvgPrice = response.PriceAvg
	order.Status = future.adaptOrderState(response.State)
	order.Fee = response.Fee
	order.DealTimestamp = response.Timestamp.UnixNano() / int64(time.Millisecond)
	order.DealDatetime = response.Timestamp.In(future.config.Location).Format(GO_BIRTHDAY)
	if order.Exchange == "" {
		order.Exchange = future.GetExchangeName()
	}
	return
}

func (future *Future) getV3Order(order *FutureOrder) ([]byte, error) {
	urlPath := fmt.Sprintf(
		"/api/futures/v3/orders/%s/%s",
		future.GetInstrumentId(order.Pair, order.ContractType),
		order.OrderId,
	)

	var response futureOrderResponse
	resp, err := future.DoRequest("GET", urlPath, "", &response)
	if err != nil {
		return nil, err
	}

	future.adaptOrder(response, order)
	return resp, nil
}

func (future *Future) cancelV3Order(order *FutureOrder) ([]byte, error) {
	urlPath := fmt.Sprintf(
		"/api/futures/v3/cancel_order/%s/%s",
		future.GetInstrumentId(order.Pair, order.ContractType),
		order.OrderId,
	)
	var response struct {
		Result       bool   `json:"result"`
		OrderId      string `json:"order_id"`
		ClientOid    string `json:"client_oid"`
		InstrumentId string `json:"instrument_id"`
	}
	resp, err := future.DoRequest("POST", urlPath, "", &response)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
