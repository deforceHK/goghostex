package binance

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	. "github.com/strengthening/goghostex"
)

const (
	SWAP_ENDPOINT = "https://fapi.binance.com"

	SWAP_TICKER_URI = "/fapi/v1/ticker/24hr"
	SWAP_DEPTH_URI  = "/fapi/v1/depth?symbol=%s&limit=%d"
	SWAP_KLINE_URI  = "/fapi/v1/klines"

	SWAP_INCOME_URI       = "/fapi/v1/income?"
	SWAP_ACCOUNT_URI      = "/fapi/v1/account?"
	SWAP_PLACE_ORDER_URI  = "/fapi/v1/order?"
	SWAP_CANCEL_ORDER_URI = "/fapi/v1/order?"
	SWAP_GET_ORDER_URI    = "/fapi/v1/order?"
	SWAP_GET_ORDERS_URI   = "/fapi/v1/allOrders?"
)

type Swap struct {
	*Binance
	sync.Locker
	swapContracts SwapContracts
	uTime         time.Time
	bnbAvgPrice   float64 // 抵扣交易费用的 bnb 平均持仓成本
}

type SwapContract struct {
	Symbol string `json:"symbol"`

	PricePrecision int64   `json:"pricePrecision"` // 下单价格精度
	PriceMaxScale  float64 `json:"priceMaxScale"`  // 下单价格最大值
	PriceMinScale  float64 `json:"priceMinScale"`  // 下单价格最小值

	AmountPrecision int64   `json:"quantityPrecision"` // 下单数量精度
	AmountMax       float64 `json:"amountMax"`         // 下单数量最大值
	AmountMin       float64 `json:"amountMin"`         // 下单数量最小值
}

type SwapContracts map[string]SwapContract

func (swap *Swap) GetExchangeRule(pair Pair) (*SwapRule, []byte, error) {
	uri := "/fapi/v1/exchangeInfo"
	r := struct {
		Symbols []struct {
			Symbol            string `json:"symbol"`
			BaseAsset         string `json:"baseAsset"`
			QuotaAsset        string `json:"quotaAsset"`
			PricePrecision    int    `json:"pricePrecision"`
			QuantityPrecision int    `json:"quantityPrecision"`
		} `json:"symbols"`
	}{}
	resp, err := swap.DoRequest(http.MethodGet, uri, "", r)
	if err != nil {
		return nil, resp, err
	}
	symbol := pair.ToSymbol("", true)
	for _, s := range r.Symbols {
		if s.Symbol != symbol {
			continue
		}
		rule := SwapRule{
			Rule: Rule{
				Pair:             pair,
				Base:             NewCurrency(s.BaseAsset, ""),
				BasePrecision:    s.QuantityPrecision,
				BaseMinSize:      1 / math.Pow10(s.QuantityPrecision),
				Counter:          NewCurrency(s.QuotaAsset, ""),
				CounterPrecision: s.PricePrecision,
			},
			ContractVal: 1,
		}
		return &rule, resp, nil
	}

	return nil, resp, errors.New("Can not find the pair in exchange. ")
}

func (swap *Swap) GetTicker(pair Pair) (*SwapTicker, []byte, error) {

	wg := sync.WaitGroup{}
	wg.Add(2)

	var tickerRaw []byte
	var tickerErr error
	tickerResp := make(map[string]interface{}, 0)

	var swapDepth *SwapDepth
	var depthErr error

	go func() {
		defer wg.Done()
		params := url.Values{}
		params.Set("symbol", pair.ToSymbol("", true))
		tickerRaw, tickerErr = swap.DoRequest(
			http.MethodGet,
			SWAP_TICKER_URI+"?"+params.Encode(),
			"",
			&tickerResp,
		)
	}()

	go func() {
		defer wg.Done()
		swapDepth, _, depthErr = swap.GetDepth(pair, 5)
	}()

	wg.Wait()

	if tickerErr != nil {
		return nil, nil, tickerErr
	}

	if depthErr != nil {
		return nil, nil, depthErr
	}

	now := time.Now()
	var ticker SwapTicker
	ticker.Pair = pair
	ticker.Timestamp = now.UnixNano() / int64(time.Millisecond)
	ticker.Date = now.In(swap.config.Location).Format(GO_BIRTHDAY)

	ticker.Last = ToFloat64(tickerResp["lastPrice"])
	ticker.Vol = ToFloat64(tickerResp["volume"])
	ticker.High = ToFloat64(tickerResp["highPrice"])
	ticker.Low = ToFloat64(tickerResp["lowPrice"])

	ticker.Buy = ToFloat64(swapDepth.BidList[0].Price)
	ticker.Sell = ToFloat64(swapDepth.AskList[0].Price)
	return &ticker, tickerRaw, nil
}

func (swap *Swap) GetDepth(pair Pair, size int) (*SwapDepth, []byte, error) {

	sizes := map[int]int{5: 2, 10: 2, 20: 2, 50: 2, 100: 5, 500: 10, 1000: 20}
	_, exist := sizes[size]
	if !exist {
		size = 100
	}

	response := struct {
		Code         int64           `json:"code,-"`
		Message      string          `json:"message,-"`
		Asks         [][]interface{} `json:"asks"` // The binance return asks Ascending ordered
		Bids         [][]interface{} `json:"bids"` // The binance return bids Descending ordered
		LastUpdateId int64           `json:"lastUpdateId"`
	}{}

	resp, err := swap.DoRequest(
		"GET",
		fmt.Sprintf(SWAP_DEPTH_URI, pair.ToSymbol("", true), size),
		"",
		&response,
	)

	depth := new(SwapDepth)
	depth.Pair = pair
	now := time.Now()
	depth.Timestamp = now.UnixNano() / int64(time.Millisecond)
	depth.Date = now.In(swap.config.Location).Format(GO_BIRTHDAY)
	depth.Sequence = response.LastUpdateId

	for _, bid := range response.Bids {
		price := ToFloat64(bid[0])
		amount := ToFloat64(bid[1])
		depthItem := DepthItem{Price: price, Amount: amount}
		depth.BidList = append(depth.BidList, depthItem)
	}

	for _, ask := range response.Asks {
		price := ToFloat64(ask[0])
		amount := ToFloat64(ask[1])
		depthItem := DepthItem{Price: price, Amount: amount}
		depth.AskList = append(depth.AskList, depthItem)
	}

	return depth, resp, err
}

func (swap *Swap) GetStdDepth(pair Pair, size int) (*SwapStdDepth, []byte, error) {

	if size > 1000 {
		size = 1000
	} else if size < 5 {
		size = 5
	}

	response := struct {
		Code         int64           `json:"code,-"`
		Message      string          `json:"message,-"`
		Asks         [][]interface{} `json:"asks"` // The binance return asks Ascending ordered
		Bids         [][]interface{} `json:"bids"` // The binance return bids Descending ordered
		LastUpdateId int64           `json:"lastUpdateId"`
	}{}

	resp, err := swap.DoRequest(
		"GET",
		fmt.Sprintf(SWAP_DEPTH_URI, pair.ToSymbol("", true), size),
		"",
		&response,
	)

	depth := &SwapStdDepth{}
	depth.Pair = pair
	now := time.Now()
	depth.Timestamp = now.UnixNano() / int64(time.Millisecond)
	depth.Date = now.In(swap.config.Location).Format(GO_BIRTHDAY)
	depth.Sequence = response.LastUpdateId

	for _, bid := range response.Bids {
		price := int64(math.Floor(ToFloat64(bid[0])*100000000 + 0.5))
		amount := ToFloat64(bid[1])
		dsi := DepthStdItem{Price: price, Amount: amount}
		depth.BidList = append(depth.BidList, dsi)
	}

	for _, ask := range response.Asks {
		price := int64(math.Floor(ToFloat64(ask[0])*100000000 + 0.5))
		amount := ToFloat64(ask[1])
		dsi := DepthStdItem{Price: price, Amount: amount}
		depth.AskList = append(depth.AskList, dsi)
	}

	return depth, resp, err

}

func (swap *Swap) GetLimit(pair Pair) (float64, float64, error) {
	response := struct {
		MarkPrice float64 `json:"markPrice,string"`
	}{}

	_, err := swap.DoRequest(
		"GET",
		fmt.Sprintf("/fapi/v1/premiumIndex?symbol=%s", pair.ToSymbol("", true)),
		"",
		&response,
	)

	if err != nil {
		return 0, 0, err
	}

	contract := swap.getContract(pair)
	floatTemplate := "%." + fmt.Sprintf("%d", contract.PricePrecision) + "f"

	highest := response.MarkPrice * contract.PriceMaxScale
	highest, _ = strconv.ParseFloat(fmt.Sprintf(floatTemplate, highest), 64)

	lowest := response.MarkPrice * contract.PriceMinScale
	lowest, _ = strconv.ParseFloat(fmt.Sprintf(floatTemplate, lowest), 64)

	return highest, lowest, nil
}

func (swap *Swap) GetKline(pair Pair, period, size, since int) ([]*SwapKline, []byte, error) {
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
	params.Set("symbol", pair.ToSymbol("", true))
	params.Set("interval", _INERNAL_KLINE_PERIOD_CONVERTER[period])
	params.Set("startTime", startTimeFmt)
	params.Set("endTime", endTimeFmt)
	params.Set("limit", fmt.Sprintf("%d", size))

	uri := SWAP_KLINE_URI + "?" + params.Encode()
	klines := make([][]interface{}, 0)
	resp, err := swap.DoRequest(http.MethodGet, uri, "", &klines)
	if err != nil {
		return nil, nil, err
	}

	var swapKlines []*SwapKline
	for _, k := range klines {
		timestamp := ToInt64(k[0])
		r := &SwapKline{
			Pair:      pair,
			Exchange:  BINANCE,
			Timestamp: timestamp,
			Date:      time.Unix(timestamp/1000, 0).In(swap.config.Location).Format(GO_BIRTHDAY),
			Open:      ToFloat64(k[1]),
			High:      ToFloat64(k[2]),
			Low:       ToFloat64(k[3]),
			Close:     ToFloat64(k[4]),
			Vol:       ToFloat64(k[5]),
		}

		swapKlines = append(swapKlines, r)
	}

	return swapKlines, resp, nil
}

func (swap *Swap) GetOpenAmount(pair Pair) (float64, int64, []byte, error) {
	params := url.Values{}
	params.Set("symbol", pair.ToSymbol("", true))
	params.Set("period", "5m")
	params.Set("limit", "500")

	responses := make([]struct {
		Amount    float64 `json:"sumOpenInterest,string"`
		Quota     float64 `json:"sumOpenInterestValue,string"`
		Timestamp int64   `json:"timestamp"`
	}, 0)
	resp, err := swap.DoRequest(
		http.MethodGet,
		"/futures/data/openInterestHist?"+params.Encode(),
		"",
		&responses,
	)
	if err != nil {
		return 0, 0, nil, err
	}
	return responses[len(responses)-1].Amount, responses[len(responses)-1].Timestamp, resp, nil
}

func (swap *Swap) GetFee() (float64, error) {
	panic("implement me")
}

func (swap *Swap) GetFundingFees(pair Pair) ([][]interface{}, []byte, error) {
	param := url.Values{}
	param.Set("symbol", pair.ToSymbol("", true))
	param.Set("limit", "500")

	rawFees := make([]struct {
		FundingRate float64 `json:"fundingRate,string"`
		FundingTime int64   `json:"fundingTime"`
	}, 0)

	if resp, err := swap.DoRequest(
		http.MethodGet,
		"/fapi/v1/fundingRate?"+param.Encode(),
		"",
		&rawFees,
	); err != nil {
		return nil, nil, err
	} else {
		fees := make([][]interface{}, 0)
		for _, f := range rawFees {
			fee := []interface{}{f.FundingRate, f.FundingTime}
			fees = append(fees, fee)
		}
		return fees, resp, nil
	}
}

var placeTypeRelation = map[PlaceType]string{
	NORMAL:     "GTC",
	ONLY_MAKER: "GTX",
	IOC:        "IOC",
	FOK:        "FOK",
}

var sideRelation = map[FutureType]string{
	OPEN_LONG:       "BUY",
	OPEN_SHORT:      "SELL",
	LIQUIDATE_LONG:  "SELL",
	LIQUIDATE_SHORT: "BUY",
}

var positionSideRelation = map[FutureType]string{
	OPEN_LONG:       "LONG",
	OPEN_SHORT:      "SHORT",
	LIQUIDATE_LONG:  "LONG",
	LIQUIDATE_SHORT: "SHORT",
}

var statusRelation = map[string]TradeStatus{
	"NEW":              ORDER_UNFINISH,
	"PARTIALLY_FILLED": ORDER_PART_FINISH,
	"FILLED":           ORDER_FINISH,
	"CANCELED":         ORDER_CANCEL,
	"REJECTED":         ORDER_FAIL,
	"EXPIRED":          ORDER_CANCEL,
}

func (swap *Swap) PlaceOrder(order *SwapOrder) ([]byte, error) {
	if order == nil {
		return nil, errors.New("order param is nil")
	}

	side, positionSide, placeType := "", "", ""
	exist := false
	if side, exist = sideRelation[order.Type]; !exist {
		panic("future type not found. ")
	}
	if positionSide, exist = positionSideRelation[order.Type]; !exist {
		panic("future type not found. ")
	}
	if placeType, exist = placeTypeRelation[order.PlaceType]; !exist {
		panic("place type not found. ")
	}

	param := url.Values{}
	param.Set("symbol", order.Pair.ToSymbol("", true))
	param.Set("side", side)
	param.Set("positionSide", positionSide)
	param.Set("type", "LIMIT")
	param.Set("price", swap.normalPrice(order.Price, order.Pair))
	param.Set("quantity", swap.normalAmount(order.Amount, order.Pair))
	// "GTC": 成交为止, 一直有效。
	// "IOC": 无法立即成交(吃单)的部分就撤销。
	// "FOK": 无法全部立即成交就撤销。
	// "GTX": 无法成为挂单方就撤销。
	param.Set("timeInForce", placeType)
	if order.Cid != "" {
		param.Set("newClientOrderId", order.Cid)
	}

	var response struct {
		Cid        string  `json:"clientOrderId"`
		Status     string  `json:"status"`
		CumQuote   float64 `json:"cumQuote,string"`
		DealAmount float64 `json:"executedQty,string"`
		OrderId    int64   `json:"orderId"`
		UpdateTime int64   `json:"updateTime"`
		Price      float64 `json:"price,string"`
		Amount     float64 `json:"origQty,string"`
	}

	now := time.Now()
	if err := swap.buildParamsSigned(&param); err != nil {
		return nil, err
	}
	resp, err := swap.DoRequest(http.MethodPost, SWAP_PLACE_ORDER_URI+param.Encode(), "", &response)
	if err != nil {
		return nil, err
	}
	orderTime := time.Unix(response.UpdateTime/1000, 0)
	order.OrderId = fmt.Sprintf("%d", response.OrderId)
	order.PlaceTimestamp = now.UnixNano() / int64(time.Millisecond)
	order.PlaceDatetime = now.In(swap.config.Location).Format(GO_BIRTHDAY)
	order.DealTimestamp = response.UpdateTime
	order.DealDatetime = orderTime.In(swap.config.Location).Format(GO_BIRTHDAY)
	order.Status = statusRelation[response.Status]
	order.Price = response.Price
	order.Amount = response.Amount
	if response.DealAmount > 0 {
		order.AvgPrice = response.CumQuote / response.DealAmount
		order.DealAmount = response.DealAmount
	}
	return resp, nil
}

func (swap *Swap) CancelOrder(order *SwapOrder) ([]byte, error) {
	if order.OrderId == "" && order.Cid == "" {
		return nil, errors.New("The orderid and cid is empty. ")
	}

	param := url.Values{}
	param.Set("symbol", order.Pair.ToSymbol("", true))
	if order.OrderId != "" {
		param.Set("orderId", order.OrderId)
	} else {
		param.Set("origClientOrderId", order.Cid)
	}

	var response struct {
		Cid        string  `json:"clientOrderId"`
		Status     string  `json:"status"`
		CumQuote   float64 `json:"cumQuote,string"`
		DealAmount float64 `json:"executedQty,string"`
		OrderId    int64   `json:"orderId"`
		UpdateTime int64   `json:"updateTime"`
	}

	now := time.Now()
	if err := swap.buildParamsSigned(&param); err != nil {
		return nil, err
	}
	resp, err := swap.DoRequest(http.MethodDelete, SWAP_CANCEL_ORDER_URI+param.Encode(), "", &response)
	if err != nil {
		return nil, err
	}

	orderTime := time.Unix(response.UpdateTime/1000, 0)
	order.PlaceTimestamp = now.UnixNano() / int64(time.Millisecond)
	order.PlaceDatetime = now.In(swap.config.Location).Format(GO_BIRTHDAY)
	order.DealTimestamp = response.UpdateTime
	order.DealDatetime = orderTime.In(swap.config.Location).Format(GO_BIRTHDAY)
	order.Status = statusRelation[response.Status]
	if response.DealAmount > 0 {
		order.AvgPrice = response.CumQuote / response.DealAmount
		order.DealAmount = response.DealAmount
	}
	return resp, nil
}

func (swap *Swap) GetOrder(order *SwapOrder) ([]byte, error) {
	if order.OrderId == "" && order.Cid == "" {
		return nil, errors.New("The orderid and cid is empty. ")
	}

	param := url.Values{}
	param.Set("symbol", order.Pair.ToSymbol("", true))
	if order.OrderId != "" {
		param.Set("orderId", order.OrderId)
	} else {
		param.Set("origClientOrderId", order.Cid)
	}

	var response struct {
		Cid        string  `json:"clientOrderId"`
		Status     string  `json:"status"`
		CumQuote   float64 `json:"cumQuote,string"`
		DealAmount float64 `json:"executedQty,string"`
		Price      float64 `json:"price,string"`
		Amount     float64 `json:"origQty,string"`
		OrderId    int64   `json:"orderId"`
		UpdateTime int64   `json:"updateTime"`
	}

	now := time.Now()
	if err := swap.buildParamsSigned(&param); err != nil {
		return nil, err
	}
	resp, err := swap.DoRequest(http.MethodGet, SWAP_GET_ORDER_URI+param.Encode(), "", &response)
	if err != nil {
		return nil, err
	}

	orderTime := time.Unix(response.UpdateTime/1000, 0)
	order.PlaceTimestamp = now.UnixNano() / int64(time.Millisecond)
	order.PlaceDatetime = now.In(swap.config.Location).Format(GO_BIRTHDAY)
	order.DealTimestamp = response.UpdateTime
	order.DealDatetime = orderTime.In(swap.config.Location).Format(GO_BIRTHDAY)
	order.Status = statusRelation[response.Status]
	order.Price = response.Price
	order.Amount = response.Amount
	if response.DealAmount > 0 {
		order.AvgPrice = response.CumQuote / response.DealAmount
		order.DealAmount = response.DealAmount
	}
	return resp, nil
}

func (swap *Swap) GetOrders(pair Pair) ([]*SwapOrder, []byte, error) {
	var rawOrders = make([]struct {
		Price          float64 `json:"price,string"`
		Amount         float64 `json:"origQty,string"`
		AvgPrice       float64 `json:"avgPrice,string"`
		DealAmount     float64 `json:"executedQty,string"`
		Cid            string  `json:"clientOrderId"`
		OrderId        int64   `json:"orderId"`
		Status         string  `json:"status"`
		OrderTimestamp int64   `json:"time"`
		DealTimestamp  int64   `json:"updateTime"`
		Side           string  `json:"side"`
		PositionSide   string  `json:"positionSide"`
	}, 0)

	params := url.Values{}
	params.Set("symbol", pair.ToSymbol("", true))
	if err := swap.buildParamsSigned(&params); err != nil {
		return nil, nil, err
	}

	resp, err := swap.DoRequest(http.MethodGet, SWAP_GET_ORDERS_URI+params.Encode(), "", &rawOrders)
	if err != nil {
		return nil, nil, err
	}
	if len(rawOrders) == 0 {
		return make([]*SwapOrder, 0), resp, nil
	}

	swapOrders := make([]*SwapOrder, 0)
	for _, rawOrder := range rawOrders {
		orderTime := time.Unix(rawOrder.OrderTimestamp/1000, 0)
		dealTime := time.Unix(rawOrder.DealTimestamp/1000, 0)
		// do not support the both position in binance
		if rawOrder.PositionSide == "BOTH" {
			continue
		}

		s := &SwapOrder{
			Cid:        rawOrder.Cid,
			OrderId:    fmt.Sprintf("%d", rawOrder.OrderId),
			Type:       swap.getFutureType(rawOrder.Side, rawOrder.PositionSide),
			Price:      rawOrder.Price,
			Amount:     rawOrder.Amount,
			AvgPrice:   rawOrder.AvgPrice,
			DealAmount: rawOrder.DealAmount,
			Status:     statusRelation[rawOrder.Status],

			Pair:           pair,
			Exchange:       BINANCE,
			PlaceTimestamp: rawOrder.OrderTimestamp,
			PlaceDatetime:  orderTime.In(swap.config.Location).Format(GO_BIRTHDAY),
			DealTimestamp:  rawOrder.DealTimestamp,
			DealDatetime:   dealTime.In(swap.config.Location).Format(GO_BIRTHDAY),
		}
		swapOrders = append(swapOrders, s)
	}

	return swapOrders, resp, nil
}

func (swap *Swap) GetOrders1(orderIds []string, pair Pair) ([]*SwapOrder, []byte, error) {
	param := url.Values{}
	param.Set("symbol", pair.ToSymbol("", true))
	if err := swap.buildParamsSigned(&param); err != nil {
		return nil, nil, err
	}

	response := make([]struct {
		Fee           float64 `json:"commission,string"`
		FeeAsset      string  `json:"commissionAsset"`
		OrderId       int64   `json:"orderId"`
		AvgPrice      float64 `json:"price,string"`
		DealAmount    float64 `json:"qty,string"`
		DealTimestamp int64   `json:"time"`
	}, 0)
	resp, err := swap.DoRequest(http.MethodGet, SWAP_GET_ORDERS_URI+param.Encode(), "", &response)
	if err != nil {
		return nil, nil, err
	}
	if len(response) == 0 {
		return make([]*SwapOrder, 0), resp, nil
	}

	idKV := map[string]int{}
	for i, rawOrder := range response {
		idKV[fmt.Sprintf("%d", rawOrder.OrderId)] = i
	}
	swapOrders := make([]*SwapOrder, 0)
	if len(orderIds) == 0 {
		for _, rawOrder := range response {
			fee := rawOrder.Fee
			if rawOrder.FeeAsset == "BNB" {
				fee *= swap.bnbAvgPrice
			}
			if fee > 0 {
				fee = 0 - fee
			}
			s := &SwapOrder{
				OrderId:    fmt.Sprintf("%d", rawOrder.OrderId),
				AvgPrice:   rawOrder.AvgPrice,
				DealAmount: rawOrder.DealAmount,
				Fee:        fee,
				Pair:       pair,
				Exchange:   BINANCE,
			}
			swapOrders = append(swapOrders, s)
		}
	} else {
		for _, orderId := range orderIds {
			i, exist := idKV[orderId]
			if !exist {
				continue
			}
			fee := response[i].Fee
			if response[i].FeeAsset == "BNB" {
				fee *= swap.bnbAvgPrice
			}
			if fee > 0 {
				fee = 0 - fee
			}
			s := &SwapOrder{
				OrderId:    fmt.Sprintf("%d", response[i].OrderId),
				AvgPrice:   response[i].AvgPrice,
				DealAmount: response[i].DealAmount,
				Fee:        fee,
				Pair:       pair,
				Exchange:   BINANCE,
			}
			swapOrders = append(swapOrders, s)
		}
	}
	return swapOrders, resp, nil
}

func (swap *Swap) GetUnFinishOrders(pair Pair) ([]*SwapOrder, []byte, error) {
	param := url.Values{}
	param.Set("symbol", pair.ToSymbol("", true))
	if err := swap.buildParamsSigned(&param); err != nil {
		return nil, nil, err
	}

	response := make([]struct {
		Price          float64 `json:"price,string"`
		Amount         float64 `json:"origQty,string"`
		AvgPrice       float64 `json:"avgPrice,string"`
		DealAmount     float64 `json:"executedQty,string"`
		Cid            string  `json:"clientOrderId"`
		OrderId        int64   `json:"orderId"`
		Status         string  `json:"status"`
		OrderTimestamp int64   `json:"time"`
		DealTimestamp  int64   `json:"updateTime"`
		Side           string  `json:"side"`
		PositionSide   string  `json:"positionSide"`
	}, 0)

	resp, err := swap.DoRequest(http.MethodGet, "/fapi/v1/openOrders?"+param.Encode(), "", &response)
	if err != nil {
		return nil, nil, err
	}

	swapOrders := make([]*SwapOrder, 0)
	for _, rawOrder := range response {
		if rawOrder.PositionSide == "BOTH" {
			continue
		}
		orderTime := time.Unix(rawOrder.OrderTimestamp/1000, 0)
		dealTime := time.Unix(rawOrder.DealTimestamp/1000, 0)
		s := &SwapOrder{
			Cid:            rawOrder.Cid,
			OrderId:        fmt.Sprintf("%d", rawOrder.OrderId),
			Price:          rawOrder.Price,
			Amount:         rawOrder.Amount,
			AvgPrice:       rawOrder.AvgPrice,
			DealAmount:     rawOrder.DealAmount,
			Status:         statusRelation[rawOrder.Status],
			Pair:           pair,
			Exchange:       BINANCE,
			PlaceTimestamp: rawOrder.OrderTimestamp,
			PlaceDatetime:  orderTime.In(swap.config.Location).Format(GO_BIRTHDAY),
			DealTimestamp:  rawOrder.DealTimestamp,
			DealDatetime:   dealTime.In(swap.config.Location).Format(GO_BIRTHDAY),
		}
		swapOrders = append(swapOrders, s)
	}

	return swapOrders, resp, nil
}

func (swap *Swap) GetPosition(pair Pair, openType FutureType) (*SwapPosition, []byte, error) {
	param := url.Values{}
	if err := swap.buildParamsSigned(&param); err != nil {
		return nil, nil, err
	}

	response := make([]struct {
		EntryPrice       float64 `json:"entryPrice,string"`
		MarginType       string  `json:"marginType"`
		Leverage         int64   `json:"leverage,string"`
		IsolatedMargin   float64 `json:"isolatedMargin,string"`
		LiquidatePrice   float64 `json:"liquidationPrice,string"`
		PositionAmt      float64 `json:"positionAmt,string"`
		Symbol           string  `json:"symbol"`
		UnRealizedProfit float64 `json:"unRealizedProfit,string"`
		PositionSide     string  `json:"positionSide"`
		MarkPrice        float64 `json:"markPrice,string"`
	}, 0)

	resp, err := swap.DoRequest(http.MethodGet, "/fapi/v1/positionRisk?"+param.Encode(), "", &response)
	if err != nil {
		return nil, nil, err
	}

	for _, p := range response {
		//do not support the binance both mode.
		if p.PositionSide == "BOTH" {
			continue
		}
		if p.Symbol != pair.ToSymbol("", true) {
			continue
		}
		positionType := OPEN_LONG
		if p.PositionSide == "SHORT" {
			positionType = OPEN_SHORT
		}
		if openType != positionType {
			continue
		}

		s := &SwapPosition{
			Pair:           pair,
			Type:           positionType,
			Amount:         p.PositionAmt,
			Price:          p.EntryPrice,
			MarkPrice:      p.MarkPrice,
			LiquidatePrice: p.LiquidatePrice,
			MarginType:     p.MarginType,
			MarginAmount:   p.IsolatedMargin,
			Leverage:       p.Leverage,
		}
		return s, resp, nil
	}

	return nil, resp, errors.New("Can not find the position. ")
}

func (swap *Swap) GetAccount() (*SwapAccount, []byte, error) {
	params := url.Values{}
	if err := swap.buildParamsSigned(&params); err != nil {
		return nil, nil, err
	}

	response := struct {
		// The future margin 期货保证金 == marginFilled+ marginUnFilled
		Margin float64 `json:"totalInitialMargin,string"`
		// The future is filled  持有头寸占用的保证金
		MarginPosition float64 `json:"totalPositionInitialMargin,string"`
		// The future is unfilled 未成交的挂单占用的保证金
		MarginOpen float64 `json:"totalOpenOrderInitialMargin,string"`
		// 保证金率
		MarginRate float64
		// 总值
		BalanceTotal float64 `json:"totalWalletBalance,string"`
		// 净值
		// BalanceNet = BalanceTotal + ProfitUnreal + ProfitReal
		BalanceNet float64
		// 可提取
		// BalanceAvail = BalanceNet - Margin
		BalanceAvail float64 `json:"maxWithdrawAmount,string"`
		// 未实现盈亏
		ProfitUnreal float64 `json:"totalUnrealizedProfit,string"`

		Positions []struct {
			Symbol           string  `json:"symbol"`
			EntryPrice       float64 `json:"entryPrice,string"`
			Leverage         int64   `json:"leverage,string"`
			Isolated         bool    `json:"isolated"`
			Margin           float64 `json:"initialMargin,string"`
			MarginPosition   float64 `json:"positionInitialMargin,string"`
			MarginOpen       float64 `json:"openOrderInitialMargin,string"`
			UnRealizedProfit float64 `json:"unRealizedProfit,string"`
			PositionSide     string  `json:"positionSide"`
		} `json:"positions"`
	}{}

	resp, err := swap.DoRequest(http.MethodGet, SWAP_ACCOUNT_URI+params.Encode(), "", &response)
	if err != nil {
		return nil, nil, err
	}

	account := &SwapAccount{
		Exchange:       BINANCE,
		Currency:       USDT,
		Margin:         response.Margin,
		MarginPosition: response.MarginPosition,
		MarginOpen:     response.MarginOpen,
		BalanceTotal:   response.BalanceTotal,
		BalanceNet:     response.BalanceTotal + response.ProfitUnreal + 0, //总资产+未实现盈利+已实现盈利（binance实时结算为0）
		BalanceAvail:   response.BalanceAvail,
		ProfitReal:     0,
		ProfitUnreal:   response.ProfitUnreal,
		Positions:      make([]*SwapPosition, 0),
	}

	for _, p := range response.Positions {
		// do not support binance, the one side trade.
		if p.PositionSide == "BOTH" {
			continue
		}
		// There don't have position.
		if p.Margin == 0.0 {
			continue
		}

		pair := Pair{
			Basis:   NewCurrency(p.Symbol[0:len(p.Symbol)-4], ""),
			Counter: NewCurrency(p.Symbol[len(p.Symbol)-4:len(p.Symbol)], ""),
		}
		futureType := OPEN_LONG
		if p.PositionSide == "SHORT" {
			futureType = OPEN_SHORT
		}

		marginType := CROSS
		if p.Isolated {
			marginType = ISOLATED
		}

		sp := &SwapPosition{
			Pair:  pair,
			Type:  futureType,
			Price: p.EntryPrice,
			//Amount:       p.MarginPosition * float64(p.Leverage) / p.EntryPrice,
			MarginType:   marginType,
			MarginAmount: p.Margin,
			Leverage:     p.Leverage,
		}

		account.Positions = append(account.Positions, sp)
	}

	return account, resp, nil

}

func (swap *Swap) GetAccountFlow() ([]*SwapAccountItem, []byte, error) {

	params := url.Values{}
	if err := swap.buildParamsSigned(&params); err != nil {
		return nil, nil, err
	}

	responses := make([]*struct {
		Symbol     string  `json:"symbol"`
		IncomeType string  `json:"incomeType"`
		Income     float64 `json:"income,string"`
		Asset      string  `json:"asset"`
		Info       string  `json:"info"`
		Time       int64   `json:"time"`
	}, 0)

	resp, err := swap.DoRequest(http.MethodGet, SWAP_INCOME_URI+params.Encode(), "", &responses)
	if err != nil {
		return nil, resp, err
	}

	pairRecord := make(map[string]Pair, 0)
	items := make([]*SwapAccountItem, 0)

	for i := len(responses) - 1; i >= 0; i-- {
		r := responses[i]
		if r.Symbol == "" {
			r.Symbol = "BTCUSDT" //默认btcusdt为主要操作账户。
		}

		p, exist := pairRecord[r.Symbol]
		if !exist {
			lenSymbol := len(r.Symbol)
			p = NewPair(
				r.Symbol[:lenSymbol-4]+"_"+r.Symbol[lenSymbol-4:],
				"_",
			)
			pairRecord[r.Symbol] = p
		}

		dateTime := time.Unix(r.Time/1000, 0).In(swap.config.Location).Format(GO_BIRTHDAY)
		sai := &SwapAccountItem{
			Pair: p, Exchange: BINANCE, Subject: swap.transferSubject(r.Income, r.IncomeType),
			SettleMode: 2, SettleCurrency: NewCurrency(r.Asset, ""), Amount: r.Income,
			Timestamp: r.Time, DateTime: dateTime, Info: r.Info,
		}

		items = append(items, sai)
	}

	return items, resp, nil
}

var subjectKV = map[string]string{
	"COMMISSION":   SUBJECT_COMMISSION,
	"REALIZED_PNL": SUBJECT_SETTLE,
	"FUNDING_FEE":  SUBJECT_FUNDING_FEE,
}

func (swap *Swap) transferSubject(income float64, remoteSubject string) string {
	if remoteSubject == "TRANSFER" {
		if income > 0 {
			return SUBJECT_TRANSFER_IN
		}
		return SUBJECT_TRANSFER_OUT
	}

	if remoteSubject == "CROSS_COLLATERAL_TRANSFER" {
		if income > 0 {
			return SUBJECT_COLLATERAL_TRANSFER_IN
		}
		return SUBJECT_COLLATERAL_TRANSFER_OUT
	}

	if subject, exist := subjectKV[remoteSubject]; exist {
		return subject
	} else {
		return strings.ToLower(remoteSubject)
	}

}

func (swap *Swap) AddMargin(pair Pair, openType FutureType, marginAmount float64) ([]byte, error) {
	return swap.modifyMargin(pair, openType, marginAmount, 1)
}

func (swap *Swap) ReduceMargin(pair Pair, openType FutureType, marginAmount float64) ([]byte, error) {
	return swap.modifyMargin(pair, openType, marginAmount, 2)
}

func (swap *Swap) modifyMargin(pair Pair, openType FutureType, marginAmount float64, opType int) ([]byte, error) {
	sidePosition := "LONG"
	if openType == OPEN_SHORT || openType == LIQUIDATE_SHORT {
		sidePosition = "SHORT"
	}

	param := url.Values{}
	param.Set("symbol", pair.ToSymbol("", true))
	param.Set("positionSide", sidePosition)
	param.Set("amount", fmt.Sprintf("%f", marginAmount))
	param.Set("type", fmt.Sprintf("%d", opType))
	if err := swap.buildParamsSigned(&param); err != nil {
		return nil, err
	}

	if resp, err := swap.DoRequest(
		http.MethodPost,
		"/fapi/v1/positionMargin?"+param.Encode(),
		"",
		nil,
	); err != nil {
		return nil, err
	} else {
		return resp, nil
	}
}

func (swap *Swap) DoRequest(httpMethod, uri, reqBody string, response interface{}) ([]byte, error) {
	header := map[string]string{
		"X-MBX-APIKEY": swap.config.ApiKey,
	}
	if httpMethod == http.MethodPost {
		header["Content-Type"] = "application/x-www-form-urlencoded"
	}

	resp, err := NewHttpRequest(
		swap.config.HttpClient,
		httpMethod,
		SWAP_ENDPOINT+uri,
		reqBody,
		header,
	)

	if err != nil {
		return nil, err
	} else {
		return resp, json.Unmarshal(resp, &response)
	}
}

func (swap *Swap) getContract(pair Pair) SwapContract {

	now := time.Now().In(swap.config.Location)
	//第一次调用或者
	if swap.uTime.IsZero() || now.After(swap.uTime.AddDate(0, 0, 1)) {
		swap.Lock()
		_, err := swap.updateContracts()
		//重试三次
		for i := 0; err != nil && i < 3; i++ {
			time.Sleep(time.Second)
			_, err = swap.updateContracts()
		}

		// 初次启动必须可以吧。
		if swap.uTime.IsZero() && err != nil {
			panic(err)
		}
		swap.Unlock()
	}
	return swap.swapContracts[pair.ToSymbol("_", false)]
}

func (swap *Swap) updateContracts() ([]byte, error) {

	var rawExchangeInfo struct {
		ServerTime int64 `json:"serverTime"`
		Symbols    []struct {
			SwapContract `json:",-"`
			BaseAsset    string                   `json:"baseAsset"`
			CounterAsset string                   `json:"quoteAsset"`
			Filters      []map[string]interface{} `json:"filters"`
		} `json:"symbols"`
	}

	resp, err := swap.DoRequest(http.MethodGet, "/fapi/v1/exchangeInfo", "", &rawExchangeInfo)
	if err != nil {
		return nil, err
	}

	uTime := time.Unix(rawExchangeInfo.ServerTime/1000, 0).In(swap.config.Location)
	for _, c := range rawExchangeInfo.Symbols {
		pair := Pair{Basis: NewCurrency(c.BaseAsset, ""), Counter: NewCurrency(c.CounterAsset, "")}
		var stdSymbol = pair.ToSymbol("_", false)
		var priceMaxScale float64
		var priceMinScale float64

		var amountMax float64
		var amountMin float64

		for _, f := range c.Filters {
			if f["filterType"] == "PERCENT_PRICE" {
				minPercent := 1 / math.Pow10(ToInt(f["multiplierDecimal"]))
				priceMaxScale = ToFloat64(f["multiplierUp"]) - minPercent
				priceMinScale = ToFloat64(f["multiplierDown"]) + minPercent
			} else if f["filterType"] == "LOT_SIZE" {
				amountMax = ToFloat64(f["maxQty"])
				amountMin = ToFloat64(f["minQty"])
			}
		}

		contract := SwapContract{
			stdSymbol,
			c.PricePrecision,
			priceMaxScale,
			priceMinScale,
			c.AmountPrecision,
			amountMax,
			amountMin,
		}

		swap.swapContracts[stdSymbol] = contract
	}
	swap.uTime = uTime
	return resp, nil
}

// not support the binance both mode.
func (swap *Swap) getFutureType(side, sidePosition string) FutureType {
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

// standard the price
func (swap *Swap) normalPrice(price float64, pair Pair) string {
	contract := swap.getContract(pair)
	return FloatToString(price, int(contract.PricePrecision))
}

// standard the amount
func (swap *Swap) normalAmount(amount float64, pair Pair) string {
	contract := swap.getContract(pair)
	return FloatToString(amount, int(contract.AmountPrecision))
}
