package main

import (
	"github.com/strengthening/goghostex/binance"
	"net/http"
	"net/url"
	"time"

	"github.com/strengthening/goghostex/coinbase"
	"github.com/strengthening/goghostex/okex"

	. "github.com/strengthening/goghostex"
)

var okexClient *okex.OKEx
var coinbaseClient *coinbase.Coinbase
var binanceClient *binance.Binance

var spotClients = map[string]SpotAPI{}

func initClients(proxy, apiKey, apiSecretKey, apiPassPhrase string) {
	loc := time.Now().Location()
	okexClient = okex.New(
		&APIConfig{
			Endpoint:      okex.ENDPOINT,
			HttpClient:    getHttpClient(proxy),
			ApiKey:        apiKey,
			ApiSecretKey:  apiSecretKey,
			ApiPassphrase: apiPassPhrase,
			Location:      loc,
		},
	)
	coinbaseClient = coinbase.New(
		&APIConfig{
			Endpoint:      coinbase.ENDPOINT,
			HttpClient:    getHttpClient(proxy),
			ApiKey:        apiKey,
			ApiSecretKey:  apiSecretKey,
			ApiPassphrase: apiPassPhrase,
			Location:      loc,
		},
	)
	binanceClient = binance.New(
		&APIConfig{
			Endpoint:      binance.ENDPOINT,
			HttpClient:    getHttpClient(proxy),
			ApiKey:        apiKey,
			ApiSecretKey:  apiSecretKey,
			ApiPassphrase: apiPassPhrase,
			Location:      loc,
		},
	)

	spotClients[COINBASE] = coinbaseClient.Spot
	spotClients[OKEX] = okexClient.Spot
	spotClients[BINANCE] = binanceClient.Spot
}

func getHttpClient(proxyUrl string) *http.Client {
	if proxyUrl == "" {
		return &http.Client{
			Timeout: 15 * time.Second,
		}
	}

	return &http.Client{
		Transport: &http.Transport{
			Proxy: func(req *http.Request) (*url.URL, error) {
				return url.Parse(proxyUrl)
			},
		},
		Timeout: 15 * time.Second,
	}
}
