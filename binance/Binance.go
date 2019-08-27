package binance

import (
	"encoding/json"
	"net/url"
	"strconv"
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

func New(config *APIConfig) *Binance {
	binance := &Binance{config: config}
	binance.Spot = &Spot{
		Binance: binance,
	}
	return binance
}

type Binance struct {
	config *APIConfig
	Spot   *Spot
}

func (this *Binance) GetExchangeName() string {
	return BINANCE
}

func (this *Binance) buildParamsSigned(postForm *url.Values) error {
	tonce := strconv.FormatInt(time.Now().UnixNano()/1000000, 10)
	postForm.Set("recvWindow", "60000")
	postForm.Set("timestamp", tonce)
	payload := postForm.Encode()
	sign, _ := GetParamHmacSHA256Sign(this.config.ApiSecretKey, payload)
	postForm.Set("signature", sign)
	return nil
}

func (this *Binance) adaptCurrencyPair(pair CurrencyPair) CurrencyPair {
	if pair.CurrencyTarget.Eq(BCH) || pair.CurrencyTarget.Eq(BCC) {
		return NewCurrencyPair(NewCurrency("BCHABC", ""), pair.CurrencyBasis).AdaptUsdToUsdt()
	}

	return pair.AdaptUsdToUsdt()
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
		return resp, json.Unmarshal(resp, &response)
	}
}
