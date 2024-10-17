package kraken

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	. "github.com/deforceHK/goghostex"
)

var SWAP_KRAKEN_PERIOD_TRANS = map[int]string{
	KLINE_PERIOD_1MIN:  "1m",
	KLINE_PERIOD_5MIN:  "5m",
	KLINE_PERIOD_15MIN: "15m",
	KLINE_PERIOD_30MIN: "30m",
	KLINE_PERIOD_1H:    "1H",
	KLINE_PERIOD_4H:    "4H",
	KLINE_PERIOD_12H:   "12H",
	KLINE_PERIOD_1DAY:  "1D",
	KLINE_PERIOD_1WEEK: "1W",
}

var SWAP_KRAKEN_PERIOD_INTERVAL = map[int]int{
	KLINE_PERIOD_1MIN:  1,
	KLINE_PERIOD_5MIN:  5,
	KLINE_PERIOD_15MIN: 15,
	KLINE_PERIOD_30MIN: 30,
	KLINE_PERIOD_1H:    60,
	KLINE_PERIOD_4H:    240,
	KLINE_PERIOD_12H:   720,
	KLINE_PERIOD_1DAY:  1440,
	KLINE_PERIOD_1WEEK: 10080,
}

func (swap *Swap) GetTicker(pair Pair) (*SwapTicker, []byte, error) {
	var contract = swap.getContract(pair)
	var uri = fmt.Sprintf("/api/v3/tickers/%s", contract.ContractName)
	var response = struct {
		ServerTime string `json:"serverTime"`
		Result     string `json:"result"`
		Ticker     struct {
			Last     float64 `json:"last"`
			LastTime string  `json:"lastTime"`
			Bid      float64 `json:"bid"`
			Ask      float64 `json:"ask"`
			Vol24h   float64 `json:"vol24h"`
			High24h  float64 `json:"high24h"`
			Low24h   float64 `json:"low24h"`
		} `json:"ticker"`
	}{}

	if resp, err := swap.DoRequest(
		SWAP_KRAKEN_ENDPOINT,
		http.MethodGet,
		uri, "",
		&response,
	); err != nil || response.Result != "success" {
		return nil, resp, err
	} else {
		var serverTime, err = time.Parse(time.RFC3339, response.ServerTime)
		if err != nil {
			return nil, resp, err
		}

		return &SwapTicker{
			Pair:      pair,
			Last:      response.Ticker.Last,
			High:      response.Ticker.High24h,
			Low:       response.Ticker.Low24h,
			Vol:       response.Ticker.Vol24h,
			Buy:       response.Ticker.Bid,
			Sell:      response.Ticker.Ask,
			Timestamp: serverTime.UnixMilli(),
			Date:      serverTime.In(swap.config.Location).Format(GO_BIRTHDAY),
		}, resp, nil
	}
}

func (swap *Swap) GetDepth(pair Pair, size int) (*SwapDepth, []byte, error) {
	var contract = swap.getContract(pair)
	var uri = fmt.Sprintf("/api/v3/orderbook?symbol=%s", contract.ContractName)
	var response = struct {
		ServerTime string `json:"serverTime"`
		Result     string `json:"result"`
		OrderBook  struct {
			Asks [][2]float64 `json:"asks"`
			Bids [][2]float64 `json:"bids"`
		} `json:"orderBook"`
	}{}

	if resp, err := swap.DoRequest(
		SWAP_KRAKEN_ENDPOINT,
		http.MethodGet,
		uri,
		"",
		&response,
	); err != nil || response.Result != "success" {
		return nil, resp, err
	} else {
		var serverTime, err = time.Parse(time.RFC3339, response.ServerTime)
		if err != nil {
			return nil, resp, err
		}
		var depth = SwapDepth{
			Pair:      pair,
			Timestamp: serverTime.UnixMilli(),
			Date:      serverTime.In(swap.config.Location).Format(GO_BIRTHDAY),
			Sequence:  serverTime.UnixMilli(),
			BidList:   make(DepthRecords, 0),
			AskList:   make(DepthRecords, 0),
		}

		for _, bid := range response.OrderBook.Bids {
			depth.BidList = append(depth.BidList, DepthRecord{bid[0], bid[1]})
		}

		for _, ask := range response.OrderBook.Asks {
			depth.AskList = append(depth.AskList, DepthRecord{ask[0], ask[1]})
		}

		return &depth, resp, nil
	}
}

func (swap *Swap) GetContract(pair Pair) *SwapContract {
	return swap.getContract(pair)
}

func (swap *Swap) GetLimit(pair Pair) (float64, float64, error) {
	var ticker, _, err = swap.GetTicker(pair)
	if err != nil {
		return 0, 0, err
	}
	return ticker.Sell * 1.2, ticker.Buy * 0.8, nil
}

func (swap *Swap) GetKline(pair Pair, period, size, since int) ([]*SwapKline, []byte, error) {
	var symbol = pair.ToSymbol("", true)
	if symbol == "BTCUSD" {
		symbol = "XBTUSD"
	}
	symbol = fmt.Sprintf("PF_%s", symbol)

	var params = url.Values{}
	var reqBody = ""
	if since > 0 {
		var sinceTS = since / 1000
		params.Set("from", fmt.Sprintf("%d", sinceTS))
		if size > 0 {
			var finishTS = sinceTS + SWAP_KRAKEN_PERIOD_INTERVAL[period]*size*60
			params.Set("to", fmt.Sprintf("%d", finishTS))
		}
		reqBody = "?" + params.Encode()
	}
	var candles = struct {
		Candles []struct {
			Time   int64   `json:"time"`
			Open   float64 `json:"open,string"`
			High   float64 `json:"high,string"`
			Low    float64 `json:"low,string"`
			Close  float64 `json:"close,string"`
			Volume float64 `json:"volume,string"`
		} `json:"candles"`
	}{}

	if resp, err := swap.DoRequest(
		SWAP_BASE_MODE_CHART,
		http.MethodGet,
		fmt.Sprintf("/api/charts/v1/trade/%s/%s", symbol, SWAP_KRAKEN_PERIOD_TRANS[period])+reqBody,
		"", &candles,
	); err != nil {
		return nil, resp, err
	} else {
		var klines = make([]*SwapKline, 0)
		for _, candle := range candles.Candles {
			klines = append(klines, &SwapKline{
				Pair:      pair,
				Exchange:  KRAKEN,
				Timestamp: candle.Time,
				Date: time.Unix(candle.Time/1000, 0).In(
					swap.config.Location,
				).Format(GO_BIRTHDAY),
				Open:  candle.Open,
				Close: candle.Close,
				High:  candle.High,
				Low:   candle.Low,
				Vol:   candle.Volume,
			})
		}

		return GetAscSwapKline(klines), resp, nil
	}
}
