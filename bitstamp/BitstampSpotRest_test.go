package bitstamp

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"testing"
	"time"

	. "github.com/strengthening/goghostex"
)

func TestSpot_GetTicker(t *testing.T) {

	config := &APIConfig{
		Endpoint: ENDPOINT,
		HttpClient: &http.Client{
			Transport: &http.Transport{
				Proxy: func(req *http.Request) (*url.URL, error) {
					return url.Parse("socks5://127.0.0.1:1090")
				},
			},
		},
		ApiKey:        "",
		ApiSecretKey:  "",
		ApiPassphrase: "",
		Location:      time.Now().Location(),
	}

	b := New(config)
	ticker, resp, err := b.Spot.GetTicker(CurrencyPair{BTC, USD})
	if err != nil {
		t.Error(err)
		return
	}

	fmt.Println(string(resp))
	tickerResponse, _ := json.Marshal(*ticker)
	fmt.Println(string(tickerResponse))
}

func TestSpot_GetDepth(t *testing.T) {

	config := &APIConfig{
		Endpoint: ENDPOINT,
		HttpClient: &http.Client{
			Transport: &http.Transport{
				Proxy: func(req *http.Request) (*url.URL, error) {
					return url.Parse("socks5://127.0.0.1:1090")
				},
			},
		},
		ApiKey:        "",
		ApiSecretKey:  "",
		ApiPassphrase: "",
		Location:      time.Now().Location(),
	}

	b := New(config)
	_, resp, err := b.Spot.GetDepth(10, CurrencyPair{BTC, USD})
	if err != nil {
		t.Error(err)
		return
	}

	fmt.Println(string(resp))
}

func TestSpot_GetKlineRecords(t *testing.T) {

	config := &APIConfig{
		Endpoint: ENDPOINT,
		HttpClient: &http.Client{
			Transport: &http.Transport{
				Proxy: func(req *http.Request) (*url.URL, error) {
					return url.Parse("socks5://127.0.0.1:1090")
				},
			},
		},
		ApiKey:        "",
		ApiSecretKey:  "",
		ApiPassphrase: "",
		Location:      time.Now().Location(),
	}

	b := New(config)
	klines, _, err := b.Spot.GetKlineRecords(
		CurrencyPair{BTC, USD},
		KLINE_PERIOD_1MIN,
		0,
		0,
	)
	if err != nil {
		t.Error(err)
		return
	}

	body, err := json.Marshal(klines)
	if err != nil {
		t.Error(err)
		return
	}

	fmt.Println(string(body))
}
