package gate

import (
	"encoding/json"
	"time"

	. "github.com/strengthening/goghostex"
)

const (
	ACCEPT       = "Accept"
	CONTENT_TYPE = "Content-Type"

	APPLICATION_JSON      = "application/json"
	APPLICATION_JSON_UTF8 = "application/json; charset=UTF-8"
)

var _INERNAL_KLINE_PERIOD_CONVERTER = map[int]string{
	KLINE_PERIOD_1MIN:  "1m",
	KLINE_PERIOD_5MIN:  "5m",
	KLINE_PERIOD_15MIN: "15m",
	KLINE_PERIOD_30MIN: "30m",
	KLINE_PERIOD_60MIN: "1h",
	KLINE_PERIOD_4H:    "4h",
	KLINE_PERIOD_8H:    "8h",
	KLINE_PERIOD_1DAY:  "1d",
}

type Gate struct {
	config *APIConfig
	Spot   *Spot
	Swap   *Swap
	//Future *Future
	//Margin *Margin
	//Wallet *Wallet
}

func (*Gate) GetExchangeName() string {
	return GATE
}

func New(config *APIConfig) *Gate {
	gate := &Gate{config: config}
	gate.Spot = &Spot{gate}
	gate.Swap = &Swap{gate}
	return gate
}

func (gate *Gate) DoRequest(
	httpMethod,
	url,
	reqBody string,
	response interface{},
) ([]byte, error) {

	resp, err := NewHttpRequest(
		gate.config.HttpClient,
		httpMethod,
		url,
		reqBody,
		map[string]string{
			CONTENT_TYPE: APPLICATION_JSON_UTF8,
			ACCEPT:       APPLICATION_JSON,
		},
	)

	if err != nil {
		return nil, err
	} else {
		nowTimestamp := time.Now().Unix() * 1000
		if nowTimestamp > gate.config.LastTimestamp {
			gate.config.LastTimestamp = nowTimestamp
		}
		return resp, json.Unmarshal(resp, &response)
	}
}
