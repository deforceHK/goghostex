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

	. "github.com/strengthening/goghostex"
)

const (
	FUTURE_ENDPOINT       = "https://dapi.binance.com"
	FUTURE_KEEP_ALIVE_URI = "/dapi/v1/ping"

	FUTURE_TICKER_URI        = "/dapi/v1/ticker/24hr?"
	FUTURE_EXCHANGE_INFO_URI = "/dapi/v1/exchangeInfo"

	FUTURE_DEPTH_URI = "/dapi/v1/depth?"
	FUTURE_KLINE_URI = "/dapi/v1/klines?"
	FUTURE_TRADE_URI = "/dapi/v1/trades?"

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
	Locker        sync.Locker
	Contracts     FutureContracts
	LastTimestamp int64
}

func (future *Future) GetExchangeRule(pair Pair) (*FutureRule, []byte, error) {
	panic("implement me")
}

func (future *Future) GetEstimatedPrice(pair Pair) (float64, []byte, error) {
	panic("implement me")
}

func (future *Future) GetTicker(pair Pair, contractType string) (*FutureTicker, []byte, error) {
	if contractType == THIS_WEEK_CONTRACT || contractType == NEXT_WEEK_CONTRACT {
		contractType = QUARTER_CONTRACT
	}

	contract, err := future.GetContract(pair, contractType)
	if err != nil {
		return nil, nil, err
	}

	params := url.Values{}
	params.Add("symbol", contract.ContractName)

	response := make([]struct {
		Symbol     string  `json:"symbol"`
		Pair       string  `json:"pair"`
		LastPrice  float64 `json:"lastPrice,string"`
		OpenPrice  float64 `json:"openPrice,string"`
		HighPrice  float64 `json:"highPrice,string"`
		LowPrice   float64 `json:"lowPrice,string"`
		Volume     float64 `json:"volume,string"`
		BaseVolume float64 `json:"baseVolume,string"`
	}, 0)

	resp, err := future.DoRequest(
		http.MethodGet,
		FUTURE_TICKER_URI+params.Encode(),
		"",
		&response,
	)
	if err != nil {
		return nil, nil, err
	}

	for _, item := range response {
		if item.Symbol == contract.ContractName {
			nowTime := time.Now()
			return &FutureTicker{
				Ticker: Ticker{
					Pair:      pair,
					Last:      item.LastPrice,
					Buy:       item.LastPrice,
					Sell:      item.LastPrice,
					High:      item.HighPrice,
					Low:       item.LowPrice,
					Vol:       item.BaseVolume,
					Timestamp: nowTime.UnixNano() / int64(time.Millisecond),
					Date:      nowTime.Format(GO_BIRTHDAY),
				},
				ContractType: contract.ContractType,
				ContractName: contract.ContractName,
			}, resp, nil
		}
	}

	return nil, nil, errors.New("Can not find the contract type. " + contractType)
}
func (future *Future) GetContract(pair Pair, contractType string) (*FutureContract, error) {
	future.Locker.Lock()
	defer future.Locker.Unlock()
	return future.getFutureContract(pair, contractType)
}

func (future *Future) GetDepth(pair Pair, contractType string, size int) (*FutureDepth, []byte, error) {
	if contractType == THIS_WEEK_CONTRACT || contractType == NEXT_WEEK_CONTRACT {
		contractType = QUARTER_CONTRACT
	}
	contract, err := future.GetContract(pair, contractType)
	if err != nil {
		return nil, nil, err
	}

	params := url.Values{}
	params.Add("symbol", contract.ContractName)
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
		contractType = QUARTER_CONTRACT
	}

	contract, err := future.GetContract(pair, contractType)
	if err != nil {
		return 0, 0, err
	}

	response := make([]struct {
		Symbol string  `json:"symbol"`
		Price  float64 `json:"price,string"`
	}, 0)

	if _, err := future.DoRequest(
		http.MethodGet,
		fmt.Sprintf("/dapi/v1/ticker/price?symbol=%s", contract.ContractName),
		"",
		&response,
	); err != nil {
		return 0, 0, nil
	}

	for _, item := range response {
		if item.Symbol == contract.ContractName {
			highLimit := item.Price * contract.MaxScalePriceLimit
			lowLimit := item.Price * contract.MinScalePriceLimit
			return highLimit, lowLimit, nil
		}
	}
	return 0, 0, errors.New("Can not find the contract. ")
}

func (future *Future) GetIndex(pair Pair) (float64, []byte, error) {
	panic("implement me")
}

func (future *Future) GetKlineRecords(
	contractType string,
	pair Pair,
	period, size, since int,
) ([]*FutureKline, []byte, error) {
	if contractType == THIS_WEEK_CONTRACT || contractType == NEXT_WEEK_CONTRACT {
		contractType = QUARTER_CONTRACT
	}

	contract, err := future.GetContract(pair, contractType)
	if err != nil {
		return nil, nil, err
	}

	startTimeFmt, endTimeFmt := fmt.Sprintf("%d", since), fmt.Sprintf("%d", time.Now().UnixNano())
	if len(startTimeFmt) > 13 {
		startTimeFmt = startTimeFmt[0:13]
	}

	if len(endTimeFmt) > 13 {
		endTimeFmt = endTimeFmt[0:13]
	}

	if size > 1500 {
		size = 1500
	}

	params := url.Values{}
	params.Set("symbol", contract.ContractName)
	params.Set("interval", _INERNAL_KLINE_PERIOD_CONVERTER[period])
	params.Set("startTime", startTimeFmt)
	params.Set("endTime", endTimeFmt)
	params.Set("limit", fmt.Sprintf("%d", size))

	uri := FUTURE_KLINE_URI + params.Encode()
	klines := make([][]interface{}, 0)
	resp, err := future.DoRequest(http.MethodGet, uri, "", &klines)
	if err != nil {
		return nil, resp, err
	}

	var list []*FutureKline
	for _, k := range klines {
		timestamp := ToInt64(k[0])
		r := &FutureKline{
			Kline: Kline{
				Pair:      pair,
				Exchange:  BINANCE,
				Timestamp: timestamp,
				Date:      time.Unix(timestamp/1000, 0).In(future.config.Location).Format(GO_BIRTHDAY),
				Open:      ToFloat64(k[1]),
				High:      ToFloat64(k[2]),
				Low:       ToFloat64(k[3]),
				Close:     ToFloat64(k[4]),
				Vol:       ToFloat64(k[5]),
			},
			DueTimestamp: contract.DueTimestamp,
			DueDate:      contract.DueDate,
			Vol2:         ToFloat64(k[7]),
		}
		list = append(list, r)
	}
	return GetAscFutureKline(list), resp, nil
}

func (future *Future) GetTrades(pair Pair, contractType string) ([]*Trade, []byte, error) {
	if contractType == THIS_WEEK_CONTRACT || contractType == NEXT_WEEK_CONTRACT {
		contractType = QUARTER_CONTRACT
	}

	contract, err := future.GetContract(pair, contractType)
	if err != nil {
		return nil, nil, err
	}

	params := url.Values{}
	params.Set("symbol", contract.ContractName)

	uri := FUTURE_TRADE_URI + params.Encode()
	response := make([]struct {
		Id           int64   `json:"id"`
		Price        float64 `json:"price,string"`
		Qty          int64   `json:"qty,string"`
		BaseQty      float64 `json:"baseQty,string"`
		Time         int64   `json:"time"`
		IsBuyerMaker bool    `json:"isBuyerMaker"`
	}, 0)
	resp, err := future.DoRequest(http.MethodGet, uri, "", &response)
	if err != nil {
		return nil, resp, err
	}

	trades := make([]*Trade, 0)
	for _, item := range response {
		tradeType := BUY
		if !item.IsBuyerMaker {
			tradeType = SELL
		}
		trade := Trade{
			Tid:       item.Id,
			Type:      tradeType,
			Amount:    item.BaseQty,
			Price:     item.Price,
			Timestamp: item.Time,
			Pair:      pair,
		}
		trades = append(trades, &trade)
	}

	return trades, resp, nil
}

func (future *Future) GetAccount() (*FutureAccount, []byte, error) {
	params := url.Values{}
	if err := future.buildParamsSigned(&params); err != nil {
		return nil, nil, err
	}

	response := struct {
		Asset []struct {
			Asset                  string  `json:"asset"`
			WalletBalance          float64 `json:"walletBalance,string"`
			UnrealizedProfit       float64 `json:"unrealizedProfit,string"`
			MarginBalance          float64 `json:"marginBalance,string"`
			MaintMargin            float64 `json:"maintMargin,string"`
			InitialMargin          float64 `json:"initialMargin,string"`
			PositionInitialMargin  float64 `json:"positionInitialMargin,string"`
			OpenOrderInitialMargin float64 `json:"openOrderInitialMargin,string"`
			MaxWithdrawAmount      float64 `json:"maxWithdrawAmount,string"`
			CrossWalletBalance     float64 `json:"crossWalletBalance,string"`
			CrossUnPnl             float64 `json:"crossUnPnl,string"`
			AvailableBalance       float64 `json:"availableBalance,string"`
		} `json:"assets"`
	}{}

	resp, err := future.DoRequest(
		http.MethodGet,
		"/dapi/v1/account?"+params.Encode(),
		"", &response,
	)
	if err != nil {
		return nil, resp, err
	}

	futureAccount := FutureAccount{
		SubAccount: make(map[Currency]FutureSubAccount, 0),
		Exchange:   BINANCE,
	}

	for _, item := range response.Asset {
		currency := NewCurrency(item.Asset, "")
		marginRate := float64(0.0)
		if item.MarginBalance > 0 {
			marginRate = item.MaintMargin / item.MarginBalance
		}

		futureAccount.SubAccount[currency] = FutureSubAccount{
			Currency: currency,

			Margin:         item.MarginBalance,
			MarginDealed:   item.PositionInitialMargin,
			MarginUnDealed: item.OpenOrderInitialMargin,
			MarginRate:     marginRate,

			BalanceTotal: item.WalletBalance,
			BalanceNet:   item.WalletBalance + item.UnrealizedProfit,
			BalanceAvail: item.MaxWithdrawAmount,

			ProfitReal:   0,
			ProfitUnreal: item.UnrealizedProfit,
		}
	}

	return &futureAccount, resp, nil
}

func (future *Future) PlaceOrder(order *FutureOrder) ([]byte, error) {
	if order == nil {
		return nil, errors.New("ord param is nil")
	}

	contract, err := future.GetContract(order.Pair, order.ContractType)
	if err != nil {
		return nil, err
	}

	side, positionSide, placeType := "", "", ""
	exist := false
	if side, exist = sideRelation[order.Type]; !exist {
		return nil, errors.New("future type not found. ")
	}
	if positionSide, exist = positionSideRelation[order.Type]; !exist {
		return nil, errors.New("future type not found. ")
	}
	if placeType, exist = placeTypeRelation[order.PlaceType]; !exist {
		return nil, errors.New("place type not found. ")
	}

	param := url.Values{}
	param.Set("symbol", contract.ContractName)
	param.Set("side", side)
	param.Set("positionSide", positionSide)
	param.Set("type", "LIMIT")
	param.Set("price", FloatToString(order.Price, contract.PricePrecision))
	param.Set("quantity", fmt.Sprintf("%d", order.Amount))
	// "GTC": 成交为止, 一直有效。
	// "IOC": 无法立即成交(吃单)的部分就撤销。
	// "FOK": 无法全部立即成交就撤销。
	// "GTX": 无法成为挂单方就撤销。
	param.Set("timeInForce", placeType)
	if order.Cid != "" {
		param.Set("newClientOrderId", order.Cid)
	}
	if err := future.buildParamsSigned(&param); err != nil {
		return nil, err
	}

	var response struct {
		Cid        string  `json:"clientOrderId"`
		Status     string  `json:"status"`
		OrderId    int64   `json:"orderId"`
		UpdateTime int64   `json:"updateTime"`
		Price      float64 `json:"price,string"`
		AvgPrice   float64 `json:"avgPrice,string"`
		Amount     int64   `json:"origQty,string"`
		DealAmount int64   `json:"executedQty,string"`
	}

	now := time.Now()
	resp, err := future.DoRequest(
		http.MethodPost,
		FUTURE_PLACE_ORDER_URI+param.Encode(),
		"",
		&response,
	)
	if err != nil {
		return nil, err
	}
	orderTime := time.Unix(response.UpdateTime/1000, 0)

	order.OrderId = fmt.Sprintf("%d", response.OrderId)
	order.PlaceTimestamp = now.UnixNano() / int64(time.Millisecond)
	order.PlaceDatetime = now.In(future.config.Location).Format(GO_BIRTHDAY)
	order.OrderTimestamp = response.UpdateTime
	order.OrderDate = orderTime.In(future.config.Location).Format(GO_BIRTHDAY)
	order.Status = statusRelation[response.Status]
	order.Price = response.Price
	order.Amount = response.Amount
	order.ContractName = contract.ContractName

	if response.Cid != "" {
		order.Cid = response.Cid
	}
	if response.DealAmount > 0 {
		order.AvgPrice = response.AvgPrice
		order.DealAmount = response.DealAmount
	}

	return resp, nil
}

func (future *Future) CancelOrder(order *FutureOrder) ([]byte, error) {
	contract, err := future.GetContract(order.Pair, order.ContractType)
	if err != nil {
		return nil, err
	}

	if order.OrderId == "" && order.Cid == "" {
		return nil, errors.New("The order_id and cid is empty. ")
	}

	param := url.Values{}
	param.Set("symbol", contract.ContractName)
	if order.OrderId != "" {
		param.Set("orderId", order.OrderId)
	} else {
		param.Set("origClientOrderId", order.Cid)
	}
	if err := future.buildParamsSigned(&param); err != nil {
		return nil, err
	}

	var response struct {
		Cid        string  `json:"clientOrderId"`
		Status     string  `json:"status"`
		DealAmount int64   `json:"executedQty,string"`
		AvgPrice   float64 `json:"avgPrice,string"`
		OrderId    int64   `json:"orderId"`
		UpdateTime int64   `json:"updateTime"`
	}

	now := time.Now()
	resp, err := future.DoRequest(
		http.MethodDelete,
		FUTURE_CANCEL_ORDER_URI+param.Encode(),
		"",
		&response,
	)
	if err != nil {
		return nil, err
	}

	orderTime := time.Unix(response.UpdateTime/1000, 0)
	order.PlaceTimestamp = now.UnixNano() / int64(time.Millisecond)
	order.PlaceDatetime = now.In(future.config.Location).Format(GO_BIRTHDAY)
	order.OrderTimestamp = response.UpdateTime
	order.OrderDate = orderTime.In(future.config.Location).Format(GO_BIRTHDAY)
	order.Status = statusRelation[response.Status]
	if response.DealAmount > 0 {
		order.AvgPrice = response.AvgPrice
		order.DealAmount = response.DealAmount
	}
	return resp, nil
}

func (future *Future) GetPosition(pair Pair, contractType string) ([]*FuturePosition, []byte, error) {
	contract, err := future.GetContract(pair, contractType)
	if err != nil {
		return nil, nil, err
	}

	param := url.Values{}
	param.Set("pair", pair.ToSymbol("", true))

	if err := future.buildParamsSigned(&param); err != nil {
		return nil, nil, err
	}

	var response []struct {
		IsolatedMargin   float64 `json:"isolatedMargin,string"`
		EntryPrice       float64 `json:"entryPrice,string"`
		Leverage         float64 `json:"leverage,string"`
		LiquidationPrice float64 `json:"liquidationPrice,string"`
		PositionAmt      float64 `json:"positionAmt,string"`
		UnRealizedProfit float64 `json:"unRealizedProfit,string"`
		PositionSide     string  `json:"positionSide"`
		Symbol           string  `json:"symbol"`
		MarginType       string  `json:"marginType"`
	}

	resp, err := future.DoRequest(
		http.MethodGet,
		FUTURE_POSITION_URI+param.Encode(),
		"",
		&response,
	)
	if err != nil {
		return nil, nil, err
	}

	fmt.Println(string(resp))
	for _, item := range response {
		if item.Symbol == contract.ContractName {
			//todo to be define
			//FuturePosition{
			//	Symbol:
			//}
		}
	}
	return nil, resp, errors.New("Can not find the contract position. ")

}

func (future *Future) GetOrders(pair Pair, contractType string) ([]*FutureOrder, []byte, error) {
	contract, err := future.GetContract(pair, contractType)
	if err != nil {
		return nil, nil, err
	}

	param := url.Values{}
	param.Set("symbol", contract.ContractName)

	if err := future.buildParamsSigned(&param); err != nil {
		return nil, nil, err
	}

	response := make([]struct {
		AvgPrice      float64 `json:"avgPrice,string"`
		ClientOrderId string  `json:"clientOrderId"`
		ExecutedQty   int64   `json:"executedQty"`
		OrderId       int64   `json:"orderId"`
		OrigQty       float64 `json:"origQty,string"`
		Price         float64 `json:"price,string"`
		Side          string  `json:"side"`
		PositionSide  string  `json:"positionSide"`
		Status        string  `json:"status"`
		Symbol        string  `json:"symbol"`
		Pair          string  `json:"pair"`
		Time          int64   `json:"time"`
		TimeInForce   string  `json:"timeInForce"`
		UpdateTime    int64   `json:"updateTime"`
	}, 0)

	resp, err := future.DoRequest(
		http.MethodGet,
		FUTURE_GET_ORDERS_URI+param.Encode(),
		"",
		&response,
	)
	if err != nil {
		return nil, resp, err
	}

	orders := make([]*FutureOrder, 0)
	for _, item := range response {
		placeTime := time.Unix(item.Time/1000, item.Time%1000).In(future.config.Location)
		updateTime := time.Unix(item.UpdateTime/1000, item.UpdateTime%1000).In(future.config.Location)

		order := FutureOrder{
			Cid:            item.ClientOrderId,
			OrderId:        fmt.Sprintf("%d", item.OrderId),
			Price:          item.Price,
			AvgPrice:       item.AvgPrice,
			Amount:         ToInt64(item.OrigQty),
			DealAmount:     item.ExecutedQty,
			PlaceTimestamp: placeTime.UnixNano() / int64(time.Millisecond),
			PlaceDatetime:  placeTime.Format(GO_BIRTHDAY),
			OrderTimestamp: updateTime.UnixNano() / int64(time.Millisecond),
			OrderDate:      updateTime.Format(GO_BIRTHDAY),
			Status:         _INTERNAL_ORDER_STATUS_REVERSE_CONVERTER[item.Status],
			PlaceType:      _INTERNAL_PLACE_TYPE_REVERSE_CONVERTER[item.TimeInForce],
			Type:           future.getFutureType(item.Side, item.PositionSide),
			//LeverRate: item.,
			//Fee:item.,
			Pair:         pair,
			ContractType: contractType,
			ContractName: contract.ContractName,
			Exchange:     BINANCE,
		}
		orders = append(orders, &order)
	}
	return orders, resp, nil
}

func (future *Future) GetOrder(order *FutureOrder) ([]byte, error) {
	if order.OrderId == "" && order.Cid == "" {
		return nil, errors.New("The order id and cid is empty. ")
	}

	contract, err := future.GetContract(order.Pair, order.ContractType)
	if err != nil {
		return nil, err
	}

	param := url.Values{}
	param.Set("symbol", contract.ContractName)
	if order.OrderId != "" {
		param.Set("orderId", order.OrderId)
	} else {
		param.Set("origClientOrderId", order.Cid)
	}
	if err := future.buildParamsSigned(&param); err != nil {
		return nil, err
	}

	var response struct {
		Cid    string `json:"clientOrderId"`
		Status string `json:"status"`

		Price    float64 `json:"price,string"`
		AvgPrice float64 `json:"avgPrice,string"`

		Amount     int64 `json:"origQty,string"`
		DealAmount int64 `json:"executedQty,string"`

		OrderId    int64 `json:"orderId"`
		UpdateTime int64 `json:"updateTime"`
	}

	now := time.Now()
	resp, err := future.DoRequest(http.MethodGet, FUTURE_GET_ORDER_URI+param.Encode(), "", &response)
	if err != nil {
		return nil, err
	}

	orderTime := time.Unix(response.UpdateTime/1000, 0)
	order.PlaceTimestamp = now.UnixNano() / int64(time.Millisecond)
	order.PlaceDatetime = now.In(future.config.Location).Format(GO_BIRTHDAY)
	order.OrderTimestamp = response.UpdateTime
	order.OrderDate = orderTime.In(future.config.Location).Format(GO_BIRTHDAY)
	order.Status = statusRelation[response.Status]
	order.Price = response.Price
	order.Amount = response.Amount
	if response.DealAmount > 0 {
		order.AvgPrice = response.AvgPrice
		order.DealAmount = response.DealAmount
	}
	return resp, nil
}

func (future *Future) GetUnFinishOrders(pair Pair, contractType string) ([]*FutureOrder, []byte, error) {
	panic("implement me")
}

func (future *Future) KeepAlive() {
	nowTimestamp := time.Now().Unix() * 1000
	// last timestamp in 5s, no need to keep alive
	if (nowTimestamp - future.LastTimestamp) < 5*1000 {
		return
	}

	_, _ = future.DoRequest(http.MethodGet, FUTURE_KEEP_ALIVE_URI, "", nil)
}

func (future *Future) DoRequest(httpMethod, uri, reqBody string, response interface{}) ([]byte, error) {
	resp, err := NewHttpRequest(
		future.config.HttpClient,
		httpMethod,
		FUTURE_ENDPOINT+uri,
		reqBody,
		map[string]string{
			"X-MBX-APIKEY": future.config.ApiKey,
		},
	)
	if err != nil {
		return nil, err
	} else {
		nowTimestamp := time.Now().Unix() * 1000
		if future.LastTimestamp < nowTimestamp {
			future.LastTimestamp = nowTimestamp
		}
		return resp, json.Unmarshal(resp, &response)
	}
}

// get the future contract info.
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
	contractTypeItem := fmt.Sprintf(
		"%s,%s,%s",
		currencies[0],
		currencies[1],
		contractType,
	)

	cf, exist := future.Contracts.ContractTypeKV[contractTypeItem]
	if !exist {
		return nil, errors.New("Can not find the contract by contract_type. ")
	}
	return cf, nil
}

type binanceContractInfo struct {
	Symbol      string `json:"symbol"`
	Pair        string `json:"pair"`
	BaseAsset   string `json:"baseAsset"`
	QuoteAsset  string `json:"quoteAsset"`
	MarginAsset string `json:"marginAsset"`

	ContractType      string `json:"contractType"`
	DeliveryDate      int64  `json:"deliveryDate"`
	OnboardDate       int64  `json:"onboardDate"`
	ContractStatus    string `json:"contractStatus"`
	ContractSize      int64  `json:"contractSize"`
	PricePrecision    int64  `json:"pricePrecision"`
	QuantityPrecision int64  `json:"quantityPrecision"`

	Filters []map[string]interface{} `json:"filters"`
}

// update the future contracts info.
func (future *Future) updateFutureContracts() ([]byte, error) {
	response := struct {
		Symbols    []*binanceContractInfo `json:"symbols"`
		ServerTime int64                  `json:"serverTime"`
	}{}

	resp, err := future.DoRequest(
		http.MethodGet, FUTURE_EXCHANGE_INFO_URI, "", &response,
	)
	if err != nil {
		return nil, err
	}

	syncTime := time.Unix(response.ServerTime/1000, response.ServerTime%1000).In(future.config.Location)
	contracts := FutureContracts{
		ContractTypeKV: make(map[string]*FutureContract, 0),
		ContractNameKV: make(map[string]*FutureContract, 0),
		DueTimestampKV: make(map[string]*FutureContract, 0),

		SyncTime: syncTime, // sync from remote time.
	}

	for _, item := range response.Symbols {

		// it is not future , it's swap in this project.
		if item.ContractType == "PERPETUAL" {
			continue
		}

		if item.ContractType != "CURRENT_QUARTER" && item.ContractType != "NEXT_QUARTER" {
			continue
		}

		contractType := ""
		if item.ContractType == "CURRENT_QUARTER" {
			contractType = QUARTER_CONTRACT
		} else if item.ContractType == "NEXT_QUARTER" {
			contractType = NEXT_QUARTER_CONTRACT
		} else {
			continue
		}

		settleMode := SETTLE_MODE_BASIS
		if item.MarginAsset == item.QuoteAsset {
			settleMode = SETTLE_MODE_COUNTER
		}

		priceMaxScale, priceMinScale := float64(1.2), float64(0.8)
		for _, filter := range item.Filters {
			if value, ok := filter["filterType"].(string); ok && value == "PERCENT_PRICE" {
				priceMaxScale = ToFloat64(filter["multiplierUp"])
				priceMinScale = ToFloat64(filter["multiplierDown"])
			}
		}

		dueTime := time.Unix(item.DeliveryDate/1000, 0).In(future.config.Location)
		openTime := time.Unix(item.OnboardDate/1000, 0).In(future.config.Location)

		pair := Pair{
			Basis:   NewCurrency(item.BaseAsset, ""),
			Counter: NewCurrency(item.QuoteAsset, ""),
		}

		contract := &FutureContract{
			Pair:         pair,
			Symbol:       pair.ToSymbol("_", false),
			Exchange:     BINANCE,
			ContractType: contractType,
			ContractName: item.Symbol,

			SettleMode:    settleMode,
			OpenTimestamp: openTime.UnixNano() / int64(time.Millisecond),
			OpenDate:      openTime.Format(GO_BIRTHDAY),
			DueTimestamp:  dueTime.UnixNano() / int64(time.Millisecond),
			DueDate:       dueTime.Format(GO_BIRTHDAY),

			UnitAmount:      item.ContractSize,
			PricePrecision:  item.PricePrecision,
			AmountPrecision: item.QuantityPrecision,

			MaxScalePriceLimit: priceMaxScale,
			MinScalePriceLimit: priceMinScale,
		}

		currencies := strings.Split(contract.Symbol, "_")
		contractTypeItem := fmt.Sprintf("%s,%s,%s", currencies[0], currencies[1], contractType)
		contractNameItem := fmt.Sprintf("%s,%s,%s", currencies[0], currencies[1], contract.ContractName)
		dueTimestampItem := fmt.Sprintf("%s,%s,%d", currencies[0], currencies[1], contract.DueTimestamp)

		contracts.ContractTypeKV[contractTypeItem] = contract
		contracts.ContractNameKV[contractNameItem] = contract
		contracts.DueTimestampKV[dueTimestampItem] = contract
	}

	future.Contracts = contracts
	return resp, nil
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
