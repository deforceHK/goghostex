package binance

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	. "github.com/strengthening/goghostex"
)

const (
	ENDPOINT = "https://api.binance.com"
	API_V1   = "/api/v1/"
	API_V3   = "/api/v3/"

	TICKER_URI             = "ticker/24hr?symbol=%s"
	TICKERS_URI            = "ticker/allBookTickers"
	DEPTH_URI              = "depth?symbol=%s&limit=%d"
	ACCOUNT_URI            = "account?"
	ORDER_URI              = "order?"
	UNFINISHED_ORDERS_INFO = "openOrders?"
	KLINE_URI              = "klines"
	SERVER_TIME_URL        = "api/v1/time"
)

var _INTERNAL_ORDER_STATUS_REVERSE_CONVERTER = map[string]TradeStatus{
	"NEW":              ORDER_UNFINISH,
	"PARTIALLY_FILLED": ORDER_PART_FINISH,
	"FILLED":           ORDER_FINISH,
	"CANCELED":         ORDER_CANCEL,
	"REJECTED":         ORDER_FAIL,
	"EXPIRED":          ORDER_FAIL,
}

var _INTERNAL_PLACE_TYPE_REVERSE_CONVERTER = map[string]PlaceType{
	"GTC": NORMAL,
	"GTX": ONLY_MAKER,
	"FOK": FOK,
	"IOC": IOC,
}

var _INERNAL_KLINE_PERIOD_CONVERTER = map[int]string{
	KLINE_PERIOD_1MIN:   "1m",
	KLINE_PERIOD_3MIN:   "3m",
	KLINE_PERIOD_5MIN:   "5m",
	KLINE_PERIOD_15MIN:  "15m",
	KLINE_PERIOD_30MIN:  "30m",
	KLINE_PERIOD_60MIN:  "1h",
	KLINE_PERIOD_2H:     "2h",
	KLINE_PERIOD_4H:     "4h",
	KLINE_PERIOD_6H:     "6h",
	KLINE_PERIOD_8H:     "8h",
	KLINE_PERIOD_12H:    "12h",
	KLINE_PERIOD_1DAY:   "1d",
	KLINE_PERIOD_3DAY:   "3d",
	KLINE_PERIOD_1WEEK:  "1w",
	KLINE_PERIOD_1MONTH: "1M",
}

var _INERNAL_KLINE_SECOND_CONVERTER = map[int]int{
	KLINE_PERIOD_1MIN:   60 * 1000,
	KLINE_PERIOD_3MIN:   3 * 60 * 1000,
	KLINE_PERIOD_5MIN:   5 * 60 * 1000,
	KLINE_PERIOD_15MIN:  15 * 60 * 1000,
	KLINE_PERIOD_30MIN:  30 * 60 * 1000,
	KLINE_PERIOD_60MIN:  60 * 60 * 1000,
	KLINE_PERIOD_2H:     2 * 60 * 60 * 1000,
	KLINE_PERIOD_4H:     4 * 60 * 60 * 1000,
	KLINE_PERIOD_6H:     6 * 60 * 60 * 1000,
	KLINE_PERIOD_8H:     8 * 60 * 60 * 1000,
	KLINE_PERIOD_12H:    12 * 60 * 60 * 1000,
	KLINE_PERIOD_1DAY:   24 * 60 * 60 * 1000,
	KLINE_PERIOD_3DAY:   3 * 24 * 60 * 60 * 1000,
	KLINE_PERIOD_1WEEK:  7 * 24 * 60 * 60 * 1000,
	KLINE_PERIOD_1MONTH: 30.5 * 24 * 60 * 60 * 1000,
}

func New(config *APIConfig) *Binance {
	binance := &Binance{config: config}
	binance.Spot = &Spot{Binance: binance}
	binance.Margin = &Margin{
		Binance: binance,
	}
	binance.Swap = &Swap{
		Binance:       binance,
		Locker:        new(sync.Mutex),
		swapContracts: SwapContracts{},
	}
	binance.Future = &Future{
		Binance: binance,
		Locker:  new(sync.Mutex),
	}
	return binance
}

type Binance struct {
	config *APIConfig
	Spot   *Spot
	Margin *Margin
	Swap   *Swap
	Future *Future
}

func (this *Binance) GetExchangeName() string {
	return BINANCE
}

func (this *Binance) buildParamsSigned(postForm *url.Values) error {
	timestamp := fmt.Sprintf("%d", time.Now().UnixNano()/int64(time.Millisecond))
	postForm.Set("timestamp", timestamp)
	postForm.Set("recvWindow", "60000")
	payload := postForm.Encode()
	sign, _ := GetParamHmacSHA256Sign(this.config.ApiSecretKey, payload)
	postForm.Set("signature", sign)
	return nil
}

func (this *Binance) DoRequest(httpMethod, uri, reqBody string, response interface{}) ([]byte, error) {
	resp, err := NewHttpRequest(
		this.config.HttpClient,
		httpMethod,
		this.config.Endpoint+uri,
		reqBody,
		map[string]string{
			"X-MBX-APIKEY": this.config.ApiKey,
		},
	)

	if err != nil {
		return nil, err
	} else {
		nowTimestamp := time.Now().Unix() * 1000
		if this.config.LastTimestamp < nowTimestamp {
			this.config.LastTimestamp = nowTimestamp
		}
		return resp, json.Unmarshal(resp, &response)
	}
}

func (this *Binance) ExchangeInfo() ([]byte, error) {

	body, err := this.DoRequest(
		http.MethodGet,
		API_V3+"exchangeInfo",
		"",
		nil,
	)

	if err != nil {
		return nil, err
	} else {
		return body, nil
	}
}
