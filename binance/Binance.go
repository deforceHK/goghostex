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

type Filter struct {
	FilterType string `json:"filterType"`
	MaxPrice   string `json:"maxPrice"`
	MinPrice   string `json:"minPrice"`
	TickSize   string `json:"tickSize"`
}

type RateLimit struct {
	Interval      string `json:"interval"`
	IntervalNum   int    `json:"intervalNum"`
	Limit         int    `json:"limit"`
	RateLimitType string `json:"rateLimitType"`
}

type TradeSymbol struct {
	BaseAsset              string   `json:"baseAsset"`
	BaseAssetPrecision     int      `json:"baseAssetPrecision"`
	Filters                []Filter `json:"filters"`
	IcebergAllowed         bool     `json:"icebergAllowed"`
	IsMarginTradingAllowed bool     `json:"isMarginTradingAllowed"`
	IsSpotTradingAllowed   bool     `json:"isSpotTradingAllowed"`
	OcoAllowed             bool     `json:"ocoAllowed"`
	OrderTypes             []string `json:"orderTypes"`
	QuoteAsset             string   `json:"quoteAsset"`
	QuotePrecision         int      `json:"quotePrecision"`
	Status                 string   `json:"status"`
	Symbol                 string   `json:"symbol"`
}

type ExchangeInfo struct {
	Timezone        string        `json:"timezone"`
	ServerTime      int           `json:"serverTime"`
	ExchangeFilters []interface{} `json:"exchangeFilters,omitempty"`
	RateLimits      []RateLimit   `json:"rateLimits"`
	Symbols         []TradeSymbol `json:"symbols"`
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

	Spot *Spot
	//accessKey,
	//secretKey string
	//httpClient   *http.Client
	//timeoffset   int64 //nanosecond
	//tradeSymbols []TradeSymbol
}

//func (this *Binance) New(config *APIConfig) {
//	this.config = config
//
//	binance := &Binance{config: config}
//	this.Spot = &Spot{binance}
//
//}

//func (this *Binance) buildParamsSigned(postForm *url.Values) error {
//	postForm.Set("recvWindow", "60000")
//	tonce := strconv.FormatInt(time.Now().UnixNano()+this.timeoffset, 10)[0:13]
//	postForm.Set("timestamp", tonce)
//	payload := postForm.Encode()
//	sign, _ := GetParamHmacSHA256Sign(this.secretKey, payload)
//	postForm.Set("signature", sign)
//	return nil
//}

//func New(client *http.Client, api_key, secret_key string) *Binance {
//	bn := &Binance{accessKey: api_key, secretKey: secret_key, httpClient: client}
//	bn.setTimeOffset()
//	return bn
//}

func (this *Binance) GetExchangeName() string {
	return BINANCE
}

func (this *Binance) buildParamsSigned(postForm *url.Values) error {
	postForm.Set("recvWindow", "60000")
	tonce := strconv.FormatInt(time.Now().UnixNano(), 10)[0:13]
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
		httpMethod, this.config.Endpoint+uri, reqBody, map[string]string{
			"X-MBX-APIKEY": this.config.ApiKey,
		},
	)

	if err != nil {
		return nil, err
	} else {
		return resp, json.Unmarshal(resp, &response)
	}
}
