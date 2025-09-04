package binance

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"sync"
	"time"

	. "github.com/deforceHK/goghostex"
)

const (
	SWAP_COUNTER_ENDPOINT   = "https://fapi.binance.com"
	SWAP_COUNTER_DEPTH_URI  = "/fapi/v1/depth?"
	SWAP_COUNTER_KLINE_URI  = "/fapi/v1/klines?"
	SWAP_COUNTER_TICKER_URI = "/fapi/v1/ticker/24hr?"

	SWAP_COUNTER_PLACE_ORDER_URI  = "/fapi/v1/order?"
	SWAP_COUNTER_GET_ORDER_URI    = "/fapi/v1/order?"
	SWAP_COUNTER_CANCEL_ORDER_URI = "/fapi/v1/order?"
	SWAP_COUNTER_INCOME_URI       = "/fapi/v1/income?"

	SWAP_BASIS_ENDPOINT   = "https://dapi.binance.com"
	SWAP_BASIS_DEPTH_URI  = "/dapi/v1/depth?"
	SWAP_BASIS_KLINE_URI  = "/dapi/v1/klines?"
	SWAP_BASIS_TICKER_URI = "/dapi/v1/ticker/24hr?"

	SWAP_BASIS_PLACE_ORDER_URI  = "/dapi/v1/order?"
	SWAP_BASIS_GET_ORDER_URI    = "/dapi/v1/order?"
	SWAP_BASIS_CANCEL_ORDER_URI = "/dapi/v1/order?"
	SWAP_BASIS_INCOME_URI       = "/dapi/v1/income?"

	SWAP_ACCOUNT_URI    = "/fapi/v1/account?"
	SWAP_GET_ORDERS_URI = "/fapi/v1/allOrders?"
)

type Swap struct {
	*Binance
	sync.Locker
	swapContracts SwapContracts
	bnbAvgPrice   float64 // 抵扣交易费用的 bnb 平均持仓成本

	nextUpdateContractTime time.Time // 下一次更新交易所contract信息
	LastKeepLiveTime       time.Time // 上一次keep live时间。
}

func (swap *Swap) GetTicker(pair Pair) (*SwapTicker, []byte, error) {
	var contract = swap.GetContract(pair)
	var wg = sync.WaitGroup{}
	wg.Add(2)

	var tickerRaw []byte
	var tickerErr error
	var tickerCounter = make(map[string]interface{}, 0)
	var tickerBasis = make([]map[string]interface{}, 0)
	var swapDepth *SwapDepth
	var depthErr error

	go func() {
		defer wg.Done()
		params := url.Values{}
		if contract.SettleMode == SETTLE_MODE_COUNTER {
			params.Set("symbol", pair.ToSymbol("", true))
			tickerRaw, tickerErr = swap.DoRequest(
				http.MethodGet,
				SWAP_COUNTER_TICKER_URI+params.Encode(),
				"",
				&tickerCounter,
				contract.SettleMode,
			)
		} else {
			params.Set("symbol", pair.ToSymbol("", true)+"_PERP")
			tickerRaw, tickerErr = swap.DoRequest(
				http.MethodGet,
				SWAP_BASIS_TICKER_URI+params.Encode(),
				"",
				&tickerBasis,
				contract.SettleMode,
			)

		}
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

	var now = time.Now()
	var last, vol, high, low float64 = 0, 0, 0, 0
	if contract.SettleMode == SETTLE_MODE_COUNTER {
		last = ToFloat64(tickerCounter["lastPrice"])
		vol = ToFloat64(tickerCounter["volume"])
		high = ToFloat64(tickerCounter["highPrice"])
		low = ToFloat64(tickerCounter["lowPrice"])
	} else {
		last = ToFloat64(tickerBasis[0]["lastPrice"])
		vol = ToFloat64(tickerBasis[0]["baseVolume"])
		high = ToFloat64(tickerBasis[0]["highPrice"])
		low = ToFloat64(tickerBasis[0]["lowPrice"])
	}

	var ticker SwapTicker
	ticker.Pair = pair
	ticker.Timestamp = now.UnixNano() / int64(time.Millisecond)
	ticker.Date = now.In(swap.config.Location).Format(GO_BIRTHDAY)
	ticker.Last = last
	ticker.Vol = vol
	ticker.High = high
	ticker.Low = low
	ticker.Buy = ToFloat64(swapDepth.BidList[0].Price)
	ticker.Sell = ToFloat64(swapDepth.AskList[0].Price)
	return &ticker, tickerRaw, nil
}

func (swap *Swap) GetMark(pair Pair) (float64, error) {
	var contract = swap.GetContract(pair)

	var settleMode = contract.SettleMode

	var price float64 = 0
	var priceErr error

	if settleMode == SETTLE_MODE_COUNTER {
		var response = struct {
			Price float64 `json:"markPrice,string"`
		}{}

		_, priceErr = swap.DoRequest(
			http.MethodGet,
			fmt.Sprintf("/fapi/v1/premiumIndex?symbol=%s", pair.ToSymbol("", true)),
			"",
			&response,
			SETTLE_MODE_COUNTER,
		)

		if priceErr != nil {
			return 0, priceErr
		}

		price = response.Price
	} else {
		var response = make([]struct {
			Price float64 `json:"markPrice,string"`
		}, 0)

		_, priceErr = swap.DoRequest(
			http.MethodGet,
			fmt.Sprintf(
				"/dapi/v1/premiumIndex?symbol=%s",
				pair.ToSymbol("", true)+"_PERP",
			),
			"",
			&response,
			SETTLE_MODE_BASIS,
		)

		if priceErr != nil {
			return 0, priceErr
		}
		if len(response) == 0 {
			return 0, errors.New("no data from remote. ")
		}
		price = response[0].Price
	}

	return price, nil
}

func (swap *Swap) GetDepth(pair Pair, size int) (*SwapDepth, []byte, error) {
	var contract = swap.GetContract(pair)

	var uri = SWAP_COUNTER_DEPTH_URI
	var symbol = pair.ToSymbol("", true)
	if contract.SettleMode == SETTLE_MODE_BASIS {
		uri = SWAP_BASIS_DEPTH_URI
		symbol = symbol + "_PERP"
	}

	var sizes = []int{5, 10, 20, 50, 100, 500, 1000}
	for _, s := range sizes {
		if size <= s {
			size = s
			break
		}
	}

	var response = struct {
		Asks         [][]interface{} `json:"asks"` // The binance return asks Ascending ordered
		Bids         [][]interface{} `json:"bids"` // The binance return bids Descending ordered
		LastUpdateId int64           `json:"lastUpdateId"`
	}{}

	var params = url.Values{}
	params.Set("symbol", symbol)
	params.Set("limit", fmt.Sprintf("%d", size))

	var resp, err = swap.DoRequest(
		http.MethodGet,
		uri+params.Encode(),
		"",
		&response,
		contract.SettleMode,
	)

	var now = time.Now()
	var depth = new(SwapDepth)
	depth.Pair = pair
	depth.Timestamp = now.UnixNano() / int64(time.Millisecond)
	depth.Date = now.In(swap.config.Location).Format(GO_BIRTHDAY)
	depth.Sequence = response.LastUpdateId

	for _, bid := range response.Bids {
		var price = ToFloat64(bid[0])
		var amount = ToFloat64(bid[1])
		depthItem := DepthRecord{Price: price, Amount: amount}
		depth.BidList = append(depth.BidList, depthItem)
	}

	for _, ask := range response.Asks {
		var price = ToFloat64(ask[0])
		var amount = ToFloat64(ask[1])
		depthItem := DepthRecord{Price: price, Amount: amount}
		depth.AskList = append(depth.AskList, depthItem)
	}

	return depth, resp, err
}

func (swap *Swap) GetContract(pair Pair) *SwapContract {
	return swap.getContract(pair)
}

func (swap *Swap) GetLimit(pair Pair) (float64, float64, error) {
	wg := sync.WaitGroup{}
	wg.Add(2)

	var price float64 = 0
	var priceErr error
	go func() {
		defer wg.Done()
		price, priceErr = swap.GetMark(pair)
		if priceErr != nil {
			return
		}
	}()

	var highestRatio float64 = 0
	var lowestRatio float64 = 0
	var pricePrecision = 0
	var limitErr error

	go func() {
		defer wg.Done()

		var contract = swap.GetContract(pair)
		var response struct {
			ServerTime int64 `json:"serverTime"`
			Symbols    []struct {
				ContractType   string                   `json:"contractType"`
				Pair           string                   `json:"pair"`
				BaseAsset      string                   `json:"baseAsset"`
				CounterAsset   string                   `json:"quoteAsset"`
				PricePrecision int                      `json:"pricePrecision"`
				Filters        []map[string]interface{} `json:"filters"`
			} `json:"symbols"`
		}

		var uri = "/fapi/v1/exchangeInfo"
		if contract.SettleMode == SETTLE_MODE_BASIS {
			uri = "/dapi/v1/exchangeInfo"
		}

		_, limitErr = swap.DoRequest(
			http.MethodGet,
			uri,
			"",
			&response,
			contract.SettleMode,
		)
		if limitErr != nil {
			return
		}

		for _, c := range response.Symbols {
			if c.Pair != pair.ToSymbol("", true) || c.ContractType != "PERPETUAL" {
				continue
			}

			pricePrecision = c.PricePrecision
			for _, f := range c.Filters {
				if f["filterType"] != "PERCENT_PRICE" {
					continue
				}
				minPercent := 1 / math.Pow10(ToInt(f["multiplierDecimal"]))
				highestRatio = ToFloat64(f["multiplierUp"]) - minPercent
				lowestRatio = ToFloat64(f["multiplierDown"]) + minPercent
				return
			}

			limitErr = errors.New(fmt.Sprintf("Can not find the symbol %s", pair.ToSymbol("_", false)))
			return
		}

		limitErr = errors.New(fmt.Sprintf("Can not find the symbol %s", pair.ToSymbol("_", false)))
		return
	}()
	wg.Wait()

	if limitErr != nil {
		return 0, 0, limitErr
	}
	if priceErr != nil {
		return 0, 0, priceErr
	}
	var tmp = math.Pow10(pricePrecision)
	return math.Floor(price*highestRatio*tmp) / tmp, math.Floor(price*lowestRatio*tmp) / tmp, nil
}

func (swap *Swap) GetKline(pair Pair, period, size, since int) ([]*SwapKline, []byte, error) {
	var contract = swap.GetContract(pair)

	if size > 1500 {
		size = 1500
	}

	var endTimestamp = since + size*_INERNAL_KLINE_SECOND_CONVERTER[period]
	if endTimestamp > since+200*24*60*60*1000 {
		endTimestamp = since + 200*24*60*60*1000
	}
	if endTimestamp > int(time.Now().Unix()*1000) {
		endTimestamp = int(time.Now().Unix() * 1000)
	}

	var symbol = pair.ToSymbol("", true)
	if contract.SettleMode == SETTLE_MODE_BASIS {
		symbol += "_PERP"
	}
	var params = url.Values{}
	params.Set("symbol", symbol)
	params.Set("interval", _INERNAL_KLINE_PERIOD_CONVERTER[period])
	params.Set("startTime", fmt.Sprintf("%d", since))
	params.Set("endTime", fmt.Sprintf("%d", endTimestamp))
	params.Set("limit", fmt.Sprintf("%d", size))

	var uri = SWAP_COUNTER_KLINE_URI + params.Encode()
	if contract.SettleMode == SETTLE_MODE_BASIS {
		uri = SWAP_BASIS_KLINE_URI + params.Encode()
	}

	klines := make([][]interface{}, 0)
	resp, err := swap.DoRequest(http.MethodGet, uri, "", &klines, contract.SettleMode)
	if err != nil {
		return nil, nil, err
	}

	var swapKlines []*SwapKline
	for _, k := range klines {
		timestamp := ToInt64(k[0])
		var vol = ToFloat64(k[5])
		if contract.SettleMode == SETTLE_MODE_BASIS {
			vol = ToFloat64(k[7])
		}
		r := &SwapKline{
			Pair:      pair,
			Exchange:  BINANCE,
			Timestamp: timestamp,
			Date:      time.Unix(timestamp/1000, 0).In(swap.config.Location).Format(GO_BIRTHDAY),
			Open:      ToFloat64(k[1]),
			High:      ToFloat64(k[2]),
			Low:       ToFloat64(k[3]),
			Close:     ToFloat64(k[4]),
			Vol:       vol,
		}

		swapKlines = append(swapKlines, r)
	}

	return GetAscSwapKline(swapKlines), resp, nil
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
		SETTLE_MODE_COUNTER,
	)
	if err != nil {
		return 0, 0, nil, err
	}
	return responses[len(responses)-1].Amount, responses[len(responses)-1].Timestamp, resp, nil
}

func (swap *Swap) GetFundingFee(pair Pair) (float64, error) {
	param := url.Values{}
	param.Set("symbol", pair.ToSymbol("", true))

	info := struct {
		Symbol          string  `json:"symbol"`
		MarkPrice       float64 `json:"markPrice,string"`
		IndexPrice      float64 `json:"indexPrice,string"`
		LastFundingRate float64 `json:"lastFundingRate,string"`
		NextFundingTime int64   `json:"nextFundingTime"`
	}{}

	if _, err := swap.DoRequest(
		http.MethodGet,
		"/fapi/v1/premiumIndex?"+param.Encode(),
		"",
		&info,
		SETTLE_MODE_COUNTER,
	); err != nil {
		return 0, err
	} else {
		return info.LastFundingRate, nil
	}
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
		SETTLE_MODE_COUNTER,
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
	MARKET:     "MARKET",
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

	var side, positionSide, placeType = "", "", ""
	var exist = false

	if side, exist = sideRelation[order.Type]; !exist {
		return nil, errors.New("swap type not found. ")
	}
	if positionSide, exist = positionSideRelation[order.Type]; !exist {
		return nil, errors.New("swap type not found. ")
	}
	if placeType, exist = placeTypeRelation[order.PlaceType]; !exist {
		return nil, errors.New("place type not found. ")
	}

	var contract = swap.GetContract(order.Pair)
	var paramSymbol = order.Pair.ToSymbol("", true)
	var uri = SWAP_COUNTER_PLACE_ORDER_URI
	if contract.SettleMode == SETTLE_MODE_BASIS {
		paramSymbol += "_PERP"
		uri = SWAP_BASIS_PLACE_ORDER_URI
	}

	var param = url.Values{}
	param.Set("symbol", paramSymbol)
	param.Set("side", side)
	param.Set("positionSide", positionSide)
	param.Set("type", "LIMIT")
	param.Set("price", FloatToPrice(order.Price, contract.PricePrecision, contract.TickSize))
	param.Set("quantity", FloatToString(order.Amount, contract.AmountPrecision))

	if placeType == "MARKET" {
		param.Set("type", "MARKET")
		param.Del("price")
	} else {
		// "GTC": 成交为止, 一直有效。
		// "IOC": 无法立即成交(吃单)的部分就撤销。
		// "FOK": 无法全部立即成交就撤销。
		// "GTX": 无法成为挂单方就撤销。
		param.Set("timeInForce", placeType)
	}
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
	resp, err := swap.DoRequest(
		http.MethodPost,
		uri+param.Encode(),
		"",
		&response,
		contract.SettleMode,
	)
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

	var contract = swap.GetContract(order.Pair)
	var paramSymbol = order.Pair.ToSymbol("", true)
	if contract.SettleMode == SETTLE_MODE_BASIS {
		paramSymbol += "_PERP"
	}

	var param = url.Values{}
	param.Set("symbol", paramSymbol)
	if order.OrderId != "" {
		param.Set("orderId", order.OrderId)
	} else {
		param.Set("origClientOrderId", order.Cid)
	}
	if err := swap.buildParamsSigned(&param); err != nil {
		return nil, err
	}

	var now = time.Now()
	var resp []byte
	var err error
	if contract.SettleMode == SETTLE_MODE_COUNTER {
		var response struct {
			Cid        string  `json:"clientOrderId"`
			Status     string  `json:"status"`
			CumQuote   float64 `json:"cumQuote,string"`
			DealAmount float64 `json:"executedQty,string"`
			OrderId    int64   `json:"orderId"`
			UpdateTime int64   `json:"updateTime"`
		}
		resp, err = swap.DoRequest(
			http.MethodDelete,
			SWAP_COUNTER_CANCEL_ORDER_URI+param.Encode(),
			"",
			&response,
			contract.SettleMode,
		)
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
	} else {
		var response struct {
			Cid        string  `json:"clientOrderId"`
			Status     string  `json:"status"`
			CumBase    float64 `json:"cumBase,string"`
			AvgPrice   float64 `json:"avgPrice,string"`
			DealAmount float64 `json:"executedQty,string"`
			OrderId    int64   `json:"orderId"`
			UpdateTime int64   `json:"updateTime"`
		}
		resp, err = swap.DoRequest(
			http.MethodDelete,
			SWAP_BASIS_CANCEL_ORDER_URI+param.Encode(),
			"",
			&response,
			contract.SettleMode,
		)
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
			order.AvgPrice = response.AvgPrice
			order.DealAmount = response.DealAmount
		}
	}
	return resp, nil
}

func (swap *Swap) GetOrder(order *SwapOrder) ([]byte, error) {
	if order.OrderId == "" && order.Cid == "" {
		return nil, errors.New("The orderid and cid is empty. ")
	}

	var contract = swap.GetContract(order.Pair)
	var paramSymbol = order.Pair.ToSymbol("", true)
	if contract.SettleMode == SETTLE_MODE_BASIS {
		paramSymbol += "_PERP"
	}
	var param = url.Values{}
	param.Set("symbol", paramSymbol)
	if order.OrderId != "" {
		param.Set("orderId", order.OrderId)
	} else {
		param.Set("origClientOrderId", order.Cid)
	}
	if err := swap.buildParamsSigned(&param); err != nil {
		return nil, err
	}

	var now = time.Now()
	var resp []byte
	var err error
	if contract.SettleMode == SETTLE_MODE_COUNTER {
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
		resp, err = swap.DoRequest(
			http.MethodGet,
			SWAP_COUNTER_GET_ORDER_URI+param.Encode(),
			"",
			&response,
			contract.SettleMode,
		)
		if err != nil {
			return resp, err
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
	} else {
		var response struct {
			Cid        string  `json:"clientOrderId"`
			Status     string  `json:"status"`
			CumBase    float64 `json:"cumBase,string"`
			DealAmount float64 `json:"executedQty,string"`
			Price      float64 `json:"price,string"`
			AvgPrice   float64 `json:"avgPrice,string"`
			Amount     float64 `json:"origQty,string"`
			OrderId    int64   `json:"orderId"`
			UpdateTime int64   `json:"updateTime"`
		}
		resp, err = swap.DoRequest(
			http.MethodGet,
			SWAP_BASIS_GET_ORDER_URI+param.Encode(),
			"",
			&response,
			contract.SettleMode,
		)
		if err != nil {
			return resp, err
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
			order.AvgPrice = response.AvgPrice
			order.DealAmount = response.DealAmount
		}
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

	resp, err := swap.DoRequest(
		http.MethodGet,
		SWAP_GET_ORDERS_URI+params.Encode(),
		"",
		&rawOrders,
		SETTLE_MODE_COUNTER,
	)
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

	resp, err := swap.DoRequest(
		http.MethodGet,
		"/fapi/v1/openOrders?"+param.Encode(),
		"",
		&response,
		SETTLE_MODE_COUNTER,
	)
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

	resp, err := swap.DoRequest(
		http.MethodGet,
		"/fapi/v1/positionRisk?"+param.Encode(),
		"",
		&response,
		SETTLE_MODE_COUNTER,
	)
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

func (swap *Swap) DoRequest(httpMethod, uri, reqBody string, response interface{}, settleMode int64) ([]byte, error) {
	header := map[string]string{
		"X-MBX-APIKEY": swap.config.ApiKey,
	}
	if httpMethod == http.MethodPost {
		header["Content-Type"] = "application/x-www-form-urlencoded"
	}

	var bnUrl = ""
	if settleMode == SETTLE_MODE_COUNTER {
		bnUrl = SWAP_COUNTER_ENDPOINT + uri
	} else {
		bnUrl = SWAP_BASIS_ENDPOINT + uri
	}
	resp, err := NewHttpRequest(
		swap.config.HttpClient,
		httpMethod,
		bnUrl,
		reqBody,
		header,
	)

	if err != nil {
		return nil, err
	} else {
		now := time.Now()
		if swap.LastKeepLiveTime.Before(now) {
			swap.LastKeepLiveTime = now
		}
		return resp, json.Unmarshal(resp, &response)
	}
}

func (swap *Swap) KeepAlive() {
	// last timestamp in 5s, no need to keep alive
	if (time.Now().Unix() - swap.LastKeepLiveTime.Unix()) < 5 {
		return
	}
	_, _ = swap.GetFundingFee(Pair{Basis: BTC, Counter: USDT})
}

func (swap *Swap) getContract(pair Pair) *SwapContract {
	defer swap.Unlock()
	swap.Lock()

	now := time.Now().In(swap.config.Location)
	if now.After(swap.nextUpdateContractTime) {
		_, err := swap.updateContracts()
		//重试三次
		for i := 0; err != nil && i < 3; i++ {
			time.Sleep(time.Second)
			_, err = swap.updateContracts()
		}
		// init fail at first time, get a default one.
		if swap.nextUpdateContractTime.IsZero() && err != nil {
			swap.swapContracts = SwapContracts{
				ContractNameKV: map[string]*SwapContract{
					"BTC-USD-SWAP": {
						Pair:            Pair{BTC, USD},
						Symbol:          "btc_usd",
						Exchange:        BINANCE,
						ContractName:    "BTC-USD-SWAP",
						SettleMode:      SETTLE_MODE_BASIS,
						UnitAmount:      100,
						PricePrecision:  1,
						AmountPrecision: 0,
					},
					"BTC-USDT-SWAP": {
						Pair:            Pair{BTC, USDT},
						Symbol:          "btc_usdt",
						Exchange:        BINANCE,
						ContractName:    "BTC-USDT-SWAP",
						SettleMode:      SETTLE_MODE_COUNTER,
						UnitAmount:      1,
						PricePrecision:  1,
						AmountPrecision: 2,
					},
					"ETH-USD-SWAP": {
						Pair:            Pair{ETH, USD},
						Symbol:          "eth_usd",
						Exchange:        BINANCE,
						ContractName:    "ETH-USD-SWAP",
						SettleMode:      SETTLE_MODE_BASIS,
						UnitAmount:      10,
						PricePrecision:  2,
						AmountPrecision: 0,
					},
					"ETH-USDT-SWAP": {
						Pair:            Pair{ETH, USDT},
						Symbol:          "eth_usdt",
						Exchange:        BINANCE,
						ContractName:    "ETH-USDT-SWAP",
						SettleMode:      SETTLE_MODE_COUNTER,
						UnitAmount:      1,
						PricePrecision:  2,
						AmountPrecision: 1,
					},
				},
			}
			swap.nextUpdateContractTime = now.Add(10 * time.Minute)
		}
	}
	return swap.swapContracts.ContractNameKV[pair.ToSwapContractName()]
}

func (swap *Swap) updateContracts() ([]byte, error) {

	var responseCounter struct {
		Symbols []struct {
			ContractType      string                   `json:"contractType"`
			BaseAsset         string                   `json:"baseAsset"`
			CounterAsset      string                   `json:"quoteAsset"`
			MarginAsset       string                   `json:"marginAsset"`
			PricePrecision    int64                    `json:"pricePrecision"`
			QuantityPrecision int64                    `json:"quantityPrecision"`
			Filters           []map[string]interface{} `json:"filters"`
		} `json:"symbols"`
	}

	var wg = sync.WaitGroup{}
	wg.Add(2)

	var respCounter []byte
	var errCounter error
	go func() {
		respCounter, errCounter = swap.DoRequest(
			http.MethodGet,
			"/fapi/v1/exchangeInfo",
			"",
			&responseCounter,
			SETTLE_MODE_COUNTER,
		)
		wg.Done()
	}()

	var responseBasis = struct {
		ServerTime int64 `json:"serverTime"`
		Symbols    []struct {
			ContractType      string                   `json:"contractType"`
			ContractSize      float64                  `json:"contractSize"`
			BaseAsset         string                   `json:"baseAsset"`
			CounterAsset      string                   `json:"quoteAsset"`
			MarginAsset       string                   `json:"marginAsset"`
			PricePrecision    int64                    `json:"pricePrecision"`
			QuantityPrecision int64                    `json:"quantityPrecision"`
			Filters           []map[string]interface{} `json:"filters"`
		} `json:"symbols"`
	}{}

	var respBasis []byte
	var errBasis error
	go func() {
		respBasis, errBasis = swap.DoRequest(
			http.MethodGet,
			"/dapi/v1/exchangeInfo",
			"",
			&responseBasis,
			SETTLE_MODE_BASIS,
		)
		wg.Done()
	}()

	wg.Wait()
	if errCounter != nil {
		return respCounter, errCounter
	}
	if errBasis != nil {
		return respBasis, errBasis
	}

	var swapContracts = SwapContracts{
		ContractNameKV: make(map[string]*SwapContract, 0),
	}

	for _, c := range responseCounter.Symbols {
		if c.ContractType != "PERPETUAL" {
			continue
		}

		pair := Pair{Basis: NewCurrency(c.BaseAsset, ""), Counter: NewCurrency(c.CounterAsset, "")}
		var stdSymbol = pair.ToSymbol("_", false)
		var stdContractName = pair.ToSwapContractName()
		var settleMode = SETTLE_MODE_BASIS
		if c.MarginAsset == c.CounterAsset {
			settleMode = SETTLE_MODE_COUNTER
		}

		var tickSize = float64(-1)
		for _, f := range c.Filters {
			var ft, isFt = f["filterType"]
			if !isFt || ft != "PRICE_FILTER" {
				continue
			}
			var ts, isTs = f["tickSize"]
			if isTs {
				tickSize = ToFloat64(ts)
				break
			}
		}

		contract := SwapContract{
			Pair:            pair,
			Symbol:          stdSymbol,
			Exchange:        BINANCE,
			ContractName:    stdContractName,
			SettleMode:      settleMode,
			UnitAmount:      1,
			TickSize:        tickSize,
			PricePrecision:  c.PricePrecision,
			AmountPrecision: c.QuantityPrecision,
		}

		swapContracts.ContractNameKV[stdContractName] = &contract
	}

	for _, c := range responseBasis.Symbols {
		if c.ContractType != "PERPETUAL" {
			continue
		}
		var pair = Pair{
			Basis:   NewCurrency(c.BaseAsset, ""),
			Counter: NewCurrency(c.CounterAsset, ""),
		}
		var stdSymbol = pair.ToSymbol("_", false)
		var stdContractName = pair.ToSwapContractName()
		var settleMode = SETTLE_MODE_BASIS
		if c.MarginAsset == c.CounterAsset {
			settleMode = SETTLE_MODE_COUNTER
		}

		var tickSize = float64(-1)
		for _, f := range c.Filters {
			var ft, isFt = f["filterType"]
			if !isFt || ft != "PRICE_FILTER" {
				continue
			}
			var ts, isTs = f["tickSize"]
			if isTs {
				tickSize = ToFloat64(ts)
				break
			}
		}

		contract := SwapContract{
			Pair:            pair,
			Symbol:          stdSymbol,
			Exchange:        BINANCE,
			ContractName:    stdContractName,
			SettleMode:      settleMode,
			UnitAmount:      c.ContractSize,
			TickSize:        tickSize,
			PricePrecision:  c.PricePrecision,
			AmountPrecision: c.QuantityPrecision,
		}
		swapContracts.ContractNameKV[stdContractName] = &contract
	}

	// setting next update time.
	var nowTime = time.Now().In(swap.config.Location)
	var nextUpdateTime = time.Date(
		nowTime.Year(), nowTime.Month(), nowTime.Day(),
		16, 0, 0, 0, swap.config.Location,
	)
	if nowTime.Hour() >= 16 {
		nextUpdateTime = nextUpdateTime.AddDate(0, 0, 1)
	}

	swap.nextUpdateContractTime = nextUpdateTime
	swap.swapContracts = swapContracts
	return respCounter, nil
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
