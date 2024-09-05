package kraken

import (
	"encoding/json"
	"sync"
	"time"

	. "github.com/deforceHK/goghostex"
)

const (
	ENDPOINT = "https://api.kraken.com"
	API_V1   = "/0/public/"

	KLINE_URI = "OHLC"
)

var _INERNAL_KLINE_PERIOD_CONVERTER = map[int]string{
	KLINE_PERIOD_1MIN:  "1",
	KLINE_PERIOD_5MIN:  "5",
	KLINE_PERIOD_15MIN: "15",
	KLINE_PERIOD_30MIN: "30",
	KLINE_PERIOD_1H:    "60",
	KLINE_PERIOD_4H:    "240",
	KLINE_PERIOD_1DAY:  "1440",
}

func New(config *APIConfig) *Kraken {
	var k = &Kraken{config: config}
	k.Spot = &Spot{k}
	k.Swap = &Swap{
		Kraken:        k,
		Locker:        new(sync.Mutex),
		swapContracts: SwapContracts{},
	}
	return k
}

type Kraken struct {
	config *APIConfig
	Spot   *Spot
	Swap   *Swap
	//Margin *Margin
	//Future *Future
}

func (k *Kraken) GetExchangeName() string {
	return KRAKEN
}

func (k *Kraken) DoRequest(httpMethod, uri, reqBody string, response interface{}) ([]byte, error) {
	resp, err := NewHttpRequest(
		k.config.HttpClient,
		httpMethod,
		k.config.Endpoint+uri,
		reqBody,
		map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		},
	)

	if err != nil {
		return nil, err
	} else {
		nowTimestamp := time.Now().Unix() * 1000
		if k.config.LastTimestamp < nowTimestamp {
			k.config.LastTimestamp = nowTimestamp
		}
		return resp, json.Unmarshal(resp, &response)
	}
}
