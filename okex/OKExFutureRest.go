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
	Locker        sync.Locker
	Contracts     FutureContracts
	LastTimestamp int64
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

// 获取合约信息
func (future *Future) getFutureContract(pair Pair, contractType string) (*FutureContract, error) {
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

	SyncTime := time.Now().In(future.config.Location)
	futureContracts := FutureContracts{
		ContractTypeKV: make(map[string]*FutureContract, 0),
		ContractNameKV: make(map[string]*FutureContract, 0),
		DueTimestampKV: make(map[string]*FutureContract, 0),
		SyncTime:       SyncTime,
	}

	for _, item := range response.Data {
		dueTime := time.Unix(item.ExpTime/1000, 0).In(future.config.Location)
		openTime := time.Unix(item.ListTime/1000, 0).In(future.config.Location)

		contractType := item.Alias
		pair := NewPair(item.Uly, "-")
		settleMode := SETTLE_MODE_BASIS
		if item.CtValCcy != "USD" {
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
			ContractName: item.InstId,
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
	return resp, nil
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

func (future *Future) GetExchangeName() string {
	return OKEX
}

// cny -> usd rate
func (future *Future) GetRate() (float64, []byte, error) {
	var response struct {
		Rate         float64   `json:"rate,string"`
		InstrumentId string    `json:"instrument_id"` //USD_CNY
		Timestamp    time.Time `json:"timestamp"`
	}
	resp, err := future.DoRequest("GET", "/api/futures/v3/rate", "", &response)
	if err != nil {
		return 0, nil, err
	}

	return response.Rate, resp, nil
}

func (future *Future) GetEstimatedPrice(pair Pair) (float64, []byte, error) {
	contract, err := future.getV3FutureContract(pair, QUARTER_CONTRACT)
	if err != nil {
		return 0, nil, err
	}

	urlPath := fmt.Sprintf(
		"/api/futures/v3/instruments/%s/estimated_price",
		contract.ContractName,
	)
	var response struct {
		InstrumentId    string  `json:"instrument_id"`
		SettlementPrice float64 `json:"settlement_price,string"`
		Timestamp       string  `json:"timestamp"`
	}
	resp, err := future.DoRequest("GET", urlPath, "", &response)
	if err != nil {
		return 0, nil, err
	}
	return response.SettlementPrice, resp, nil
}

// 获取instrument_id
func (future *Future) GetInstrumentId(pair Pair, contractAlias string) string {
	if contractAlias != NEXT_QUARTER_CONTRACT &&
		contractAlias != QUARTER_CONTRACT &&
		contractAlias != NEXT_WEEK_CONTRACT &&
		contractAlias != THIS_WEEK_CONTRACT {
		return contractAlias
	}

	future.Locker.Lock()
	defer future.Locker.Unlock()
	fc, _ := future.getFutureContract(pair, contractAlias)
	return fc.ContractName
}

func (future *Future) GetContract(pair Pair, contractType string) (*FutureContract, error) {
	future.Locker.Lock()
	defer future.Locker.Unlock()
	return future.getFutureContract(pair, contractType)
}

func (future *Future) GetTicker(pair Pair, contractType string) (*FutureTicker, []byte, error) {
	var params = url.Values{}
	params.Set("instId", future.GetInstrumentId(pair, contractType))

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

	resp, err := future.DoRequestV5Market(
		"GET",
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
	//return future.getV3Ticker(pair, contractType)
}

func (future *Future) GetDepth(
	pair Pair,
	contractType string,
	size int,
) (*FutureDepth, []byte, error) {
	info, err := future.GetContract(pair, contractType)
	if err != nil {
		return nil, nil, err
	}

	if size < 20 {
		size = 20
	}
	if size > 400 {
		size = 400
	}

	var params = url.Values{}
	params.Set("instId", info.ContractName)
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
	resp, err := future.DoRequestV5Market(
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
	dep.DueTimestamp = info.DueTimestamp
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
	//return future.getV3Depth(pair, contractType, size)
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

	_, err = future.DoRequestV5Market(
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
	//return future.getV3Limit(pair, contractType)
}

func (future *Future) GetIndex(pair Pair) (float64, []byte, error) {
	//统一交易对，当周，次周，季度指数一样的
	urlPath := fmt.Sprintf(
		"/api/futures/v3/instruments/%s/index",
		future.GetInstrumentId(pair, QUARTER_CONTRACT),
	)
	var response struct {
		InstrumentId string  `json:"instrument_id"`
		Index        float64 `json:"index,string"`
		Timestamp    string  `json:"timestamp"`
	}
	resp, err := future.DoRequest("GET", urlPath, "", &response)
	if err != nil {
		return 0, nil, err
	}
	return response.Index, resp, nil
}

type CrossedAccountInfo struct {
	MarginMode string  `json:"margin_mode"`
	Equity     float64 `json:"equity,string"`
}

func (future *Future) GetAccount() (*FutureAccount, []byte, error) {
	urlPath := "/api/futures/v3/accounts"
	var response struct {
		Info map[string]map[string]interface{}
	}

	resp, err := future.DoRequest("GET", urlPath, "", &response)
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

func (future *Future) normalizePrice(price float64, pair Pair) string {
	fc, _ := future.GetContract(pair, QUARTER_CONTRACT)
	return FloatToString(price, fc.PricePrecision)
}

//matchPrice:是否以对手价下单(0:不是 1:是)，默认为0;当取值为1时,price字段无效，当以对手价下单，order_type只能选择0:普通委托
func (future *Future) PlaceOrder(order *FutureOrder) ([]byte, error) {
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
	param.MatchPrice = order.MatchPrice
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
	order.OrderTimestamp = now.UnixNano() / int64(time.Millisecond)
	order.OrderDate = now.In(future.config.Location).Format(GO_BIRTHDAY)
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
	order.OrderTimestamp = response.Timestamp.UnixNano() / int64(time.Millisecond)
	order.OrderDate = response.Timestamp.In(future.config.Location).Format(GO_BIRTHDAY)
	if order.Exchange == "" {
		order.Exchange = future.GetExchangeName()
	}
	return
}

func (future *Future) GetOrder(order *FutureOrder) ([]byte, error) {
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

func (future *Future) CancelOrder(order *FutureOrder) ([]byte, error) {
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

func (future *Future) GetPosition(
	pair Pair,
	contractType string,
) ([]*FuturePosition, []byte, error) {
	urlPath := fmt.Sprintf(
		"/api/futures/v3/%s/position",
		future.GetInstrumentId(pair, contractType),
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
	resp, err := future.DoRequest("GET", urlPath, "", &response)
	if err != nil {
		return nil, nil, err
	}

	var postions []*FuturePosition

	if !response.Result {
		return nil, nil, errors.New("unknown error")
	}

	if response.MarginMode == "fixed" {
		panic("not support the fix future")
	}

	for _, pos := range response.Holding {
		postions = append(postions, &FuturePosition{
			Symbol:              pair,
			ContractType:        contractType,
			ContractId:          ToInt64(pos.InstrumentId[8:]),
			BuyAmount:           pos.LongQty,
			BuyAvailable:        pos.LongAvailQty,
			BuyPriceAvg:         pos.LongAvgCost,
			BuyPriceCost:        pos.LongAvgCost,
			BuyProfitReal:       pos.LongPnl,
			SellAmount:          pos.ShortQty,
			SellAvailable:       pos.ShortAvailQty,
			SellPriceAvg:        pos.ShortAvgCost,
			SellPriceCost:       pos.ShortAvgCost,
			SellProfitReal:      pos.ShortPnl,
			ForceLiquidatePrice: pos.LiquidationPrice,
			LeverRate:           pos.Leverage,
			CreateDate:          pos.CreatedAt.Unix(),
		})
	}

	return postions, resp, nil
}

func (future *Future) GetOrders(
	pair Pair,
	contractType string,
) ([]*FutureOrder, []byte, error) {
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

func (future *Future) GetUnFinishOrders(
	pair Pair,
	contractType string,
) ([]*FutureOrder, []byte, error) {
	urlPath := fmt.Sprintf(
		"/api/futures/v3/orders/%s?state=6&limit=100",
		future.GetInstrumentId(pair, contractType),
	)
	var response struct {
		Result    bool                  `json:"result"`
		OrderInfo []futureOrderResponse `json:"order_info"`
	}
	resp, err := future.DoRequest("GET", urlPath, "", &response)
	if err != nil {
		return nil, nil, err
	}
	if !response.Result {
		return nil, nil, errors.New("error")
	}

	var orders []*FutureOrder
	for _, itm := range response.OrderInfo {
		ord := FutureOrder{}
		future.adaptOrder(itm, &ord)
		orders = append(orders, &ord)
	}

	return orders, resp, nil
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
	info, err := future.GetContract(pair, contractType)
	if err!=nil{
		return nil, nil, err
	}

	if size > 100 {
		size = 100
	}

	uri := "/api/v5/market/candles?"
	params := url.Values{}
	params.Set("instId", info.ContractName)
	params.Set("bar",  _INERNAL_V5_CANDLE_PERIOD_CONVERTER[period])
	params.Set("limit",  strconv.Itoa(size))

	if since > 0 {
		endTime := time.Now()
		params.Set("before",  strconv.Itoa(since))
		params.Set("after",  strconv.Itoa(int(endTime.UnixNano()/1000000)))
	}

	var response struct {
		Code string     `json:"code"`
		Msg  string     `json:"msg"`
		Data [][]string `json:"data"`
	}
	resp, err := future.DoRequestV5Market(
		http.MethodGet,
		uri+ params.Encode(),
		"",
		&response,
	)
	if err != nil {
		return nil, nil, err
	}
	if response.Code != "0" {
		return nil, nil, errors.New(response.Msg)
	}

	var klines []*FutureKline
	for _, itm := range response.Data {
		timestamp := ToInt64(itm[0])
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
				Vol:       ToFloat64(itm[5]),
			},
			DueTimestamp: info.DueTimestamp,
			DueDate:      info.DueDate,
			Vol2:         ToFloat64(itm[6]),
		})
	}

	return GetAscFutureKline(klines), resp, nil
	//return future.getV3KlineRecords(contractType, pair, period, size, since)
}

func (future *Future) GetTrades(pair Pair, contractType string) ([]*Trade, []byte, error) {
	panic("")
}

func (future *Future) GetFutureMarkPrice(pair Pair, contractType string) (float64, []byte, error) {
	uri := fmt.Sprintf(
		"/api/futures/v3/instruments/%s/mark_price",
		future.GetInstrumentId(pair, contractType),
	)

	response := struct {
		InstrumentId string  `json:"instrument_id"`
		MarkPrice    float64 `json:"mark_price,string"`
		Timestamp    string  `json:"timestamp"`
	}{}

	resp, err := future.DoRequest(
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
func (future *Future) MarketCloseAllPosition(pair Pair, contract string, oType int) (bool, []byte, error) {
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

	param.InstrumentId = future.GetInstrumentId(pair, contract)
	if oType == int(LIQUIDATE_LONG) { //CLOSE_BUY {
		param.Direction = "long"
	} else {
		param.Direction = "short"
	}
	reqBody, _, _ := future.BuildRequestBody(param)
	resp, err := future.DoRequest("POST", urlPath, reqBody, &response)
	if err != nil {
		return false, nil, err
	}

	if !response.Result {
		return false, nil, errors.New(response.Message)
	}

	return true, resp, nil
}

func (future *Future) GetExchangeRule(pair Pair) (*FutureRule, []byte, error) {
	uri := "/api/futures/v3/instruments"
	rules := make([]*struct {
		BaseCurrency   string  `json:"base_currency"`
		QuotaCurrency  string  `json:"quota_currency"`
		TickSize       float64 `json:"tick_size,string"`
		TradeIncrement float64 `json:"trade_increment"`
		ContractVal    float64 `json:"contract_val,string"`
	}, 0)
	resp, err := future.DoRequest(http.MethodGet, uri, "", &rules)
	if err != nil {
		return nil, resp, err
	}

	base := pair.Basis.String()
	counter := pair.Counter.String()
	for _, r := range rules {
		if base != r.BaseCurrency || counter != r.QuotaCurrency {
			continue
		}

		rule := FutureRule{
			Rule: Rule{
				Pair:             pair,
				Base:             NewCurrency(r.BaseCurrency, ""),
				BasePrecision:    GetPrecision(r.TradeIncrement),
				BaseMinSize:      r.TradeIncrement,
				Counter:          NewCurrency(r.QuotaCurrency, ""),
				CounterPrecision: GetPrecision(r.TickSize),
			},
			ContractVal: r.ContractVal,
		}

		return &rule, resp, nil
	}
	return nil, resp, errors.New("Can not find the pair in exchange. ")
}

func (future *Future) KeepAlive() {
	nowTimestamp := time.Now().Unix() * 1000
	// last in 5s, no need to keep alive.
	if (nowTimestamp - future.config.LastTimestamp) < 5*1000 {
		return
	}

	// call the rate api to update lastTimestamp
	_, _, _ = future.GetRate()
}
