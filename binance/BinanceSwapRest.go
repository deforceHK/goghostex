package binance

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	. "github.com/strengthening/goghostex"
)

const (
	SWAP_ENDPOINT = "https://fapi.binance.com"

	SWAP_TICKER_URI = "/fapi/v1/ticker/24hr?symbol=%s"
	SWAP_DEPTH_URI  = "/fapi/v1/depth?symbol=%s&limit=%d"
	SWAP_KLINE_URI  = "/fapi/v1/klines"
)

type Swap struct {
	*Binance
	sync.Locker
	swapContracts SwapContracts
	uTime         time.Time
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
		tickerRaw, tickerErr = swap.DoRequest(
			"GET",
			fmt.Sprintf(SWAP_TICKER_URI, pair.ToSymbol("", true)),
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

	depth := new(SwapDepth)
	depth.Pair = pair
	now := time.Now()
	depth.Timestamp = now.UnixNano() / int64(time.Millisecond)
	depth.Date = now.In(swap.config.Location).Format(GO_BIRTHDAY)
	depth.Sequence = response.LastUpdateId

	for _, bid := range response.Bids {
		price := ToFloat64(bid[0])
		amount := ToFloat64(bid[1])
		depthItem := DepthItem{price, amount}
		depth.BidList = append(depth.BidList, depthItem)
	}

	for _, ask := range response.Asks {
		price := ToFloat64(ask[0])
		amount := ToFloat64(ask[1])
		depthItem := DepthItem{price, amount}
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
		dsi := DepthStdItem{price, amount}
		depth.BidList = append(depth.BidList, dsi)
	}

	for _, ask := range response.Asks {
		price := int64(math.Floor(ToFloat64(ask[0])*100000000 + 0.5))
		amount := ToFloat64(ask[1])
		dsi := DepthStdItem{price, amount}
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

func (swap *Swap) GetAccount() (*SwapAccount, []byte, error) {
	panic("implement me")
}

func (swap *Swap) PlaceOrder(order *SwapOrder) ([]byte, error) {
	panic("implement me")
}

func (swap *Swap) CancelOrder(order *SwapOrder) ([]byte, error) {
	panic("implement me")
}

func (swap *Swap) GetPosition(pair Pair) (*SwapPosition, []byte, error) {
	panic("implement me")
}

func (swap *Swap) GetOrders(orderIds []string, pair Pair) ([]*SwapOrder, []byte, error) {
	panic("implement me")
}

func (swap *Swap) GetOrder(order *SwapOrder) ([]byte, error) {
	panic("implement me")
}

func (swap *Swap) GetUnFinishOrders(pair Pair) ([]*SwapOrder, []byte, error) {
	panic("implement me")
}

func (swap *Swap) GetFee() (float64, error) {
	panic("implement me")
}

func (swap *Swap) DoRequest(httpMethod, uri, reqBody string, response interface{}) ([]byte, error) {
	resp, err := NewHttpRequest(
		swap.config.HttpClient,
		httpMethod,
		SWAP_ENDPOINT+uri,
		reqBody,
		map[string]string{
			"X-MBX-APIKEY": swap.config.ApiKey,
		},
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
		pair := Pair{NewCurrency(c.BaseAsset, ""), NewCurrency(c.CounterAsset, "")}
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
