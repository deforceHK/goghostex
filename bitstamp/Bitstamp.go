package bitstamp

import (
	"encoding/json"
	"time"

	. "github.com/deforceHK/goghostex"
)

var (
	ENDPOINT = "https://www.bitstamp.net"
)

type Bitstamp struct {
	config *APIConfig

	Spot *Spot
}

func New(config *APIConfig) *Bitstamp {
	bitstamp := &Bitstamp{config: config}
	bitstamp.Spot = &Spot{bitstamp}
	return bitstamp
}

func (bitstamp *Bitstamp) DoRequest(httpMethod, uri, reqBody string, response interface{}) ([]byte, error) {
	resp, err := NewHttpRequest(
		bitstamp.config.HttpClient,
		httpMethod, bitstamp.config.Endpoint+uri, reqBody,
		nil,
	)

	if err != nil {
		return nil, err
	} else {
		nowTimestamp := time.Now().Unix() * 1000
		if bitstamp.config.LastTimestamp < nowTimestamp {
			bitstamp.config.LastTimestamp = nowTimestamp
		}
		return resp, json.Unmarshal(resp, &response)
	}
}
