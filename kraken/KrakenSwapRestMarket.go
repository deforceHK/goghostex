package kraken

import (
	"fmt"
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
	// TODO implement me
	panic("implement me")
}

func (swap *Swap) GetDepth(pair Pair, size int) (*SwapDepth, []byte, error) {
	// TODO implement me
	panic("implement me")
}

func (swap *Swap) GetContract(pair Pair) *SwapContract {
	// TODO implement me
	panic("implement me")
}

func (swap *Swap) GetLimit(pair Pair) (float64, float64, error) {
	// TODO implement me
	panic("implement me")
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

	if resp, err := swap.DoGetRequest(
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
				).Format("2006-01-02 15:04:05"),
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
