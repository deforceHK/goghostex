package bitstamp

import (
	"encoding/json"

	. "github.com/strengthening/goghostex"
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

func (this *Bitstamp) DoRequest(httpMethod, uri, reqBody string, response interface{}) ([]byte, error) {
	resp, err := NewHttpRequest(
		this.config.HttpClient,
		httpMethod, this.config.Endpoint+uri, reqBody,
		nil,
	)

	if err != nil {
		return nil, err
	} else {
		return resp, json.Unmarshal(resp, &response)
	}
}
