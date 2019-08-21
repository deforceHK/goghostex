package binance

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
					return &url.URL{
						Scheme: "socks5",
						Host:   "127.0.0.1:1090"}, nil
				},
			},
		},
		ApiKey:        "",
		ApiSecretKey:  "",
		ApiPassphrase: "",
		Location:      time.Now().Location(),
	}

	b := New(config)
	if ticker, resp, err := b.Spot.GetTicker(CurrencyPair{
		CurrencyTarget: Currency{"btc", ""},
		CurrencyBasis:  Currency{"usdt", ""},
	}); err != nil {
		t.Error(err)
		return
	} else {
		body, _ := json.Marshal(*ticker)
		fmt.Println(string(body))
		fmt.Println(string(resp))
	}
}

func TestSpot_GetDepth(t *testing.T) {

	config := &APIConfig{
		Endpoint: ENDPOINT,
		HttpClient: &http.Client{
			Transport: &http.Transport{
				Proxy: func(req *http.Request) (*url.URL, error) {
					return &url.URL{
						Scheme: "socks5",
						Host:   "127.0.0.1:1090"}, nil
				},
			},
		},
		ApiKey:        "",
		ApiSecretKey:  "",
		ApiPassphrase: "",
		Location:      time.Now().Location(),
	}

	b := New(config)
	if depth, _, err := b.Spot.GetDepth(
		50,
		CurrencyPair{BTC, USDT}); err != nil {
		t.Error(err)
		return
	} else {
		body, _ := json.Marshal(*depth)
		fmt.Println(string(body))
		//fmt.Println(string(resp))
	}
}

func TestSpot_GetKlineRecords(t *testing.T) {
	config := &APIConfig{
		Endpoint: ENDPOINT,
		HttpClient: &http.Client{
			Transport: &http.Transport{
				Proxy: func(req *http.Request) (*url.URL, error) {
					return &url.URL{
						Scheme: "socks5",
						Host:   "127.0.0.1:1090"}, nil
				},
			},
		},
		ApiKey:        "",
		ApiSecretKey:  "",
		ApiPassphrase: "",
		Location:      time.Now().Location(),
	}

	b := New(config)
	if klines, resp, err := b.Spot.GetKlineRecords(
		CurrencyPair{
			CurrencyTarget: Currency{"btc", ""},
			CurrencyBasis:  Currency{"usdt", ""},
		},
		KLINE_PERIOD_1MIN,
		50,
		int(time.Now().Add(-2*24*time.Hour).UnixNano()),
	); err != nil {
		t.Error(err)
		return
	} else {
		body, _ := json.Marshal(klines)
		fmt.Println(string(body))
		fmt.Println(string(resp))
	}
}
