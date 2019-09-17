package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/strengthening/goghostex/binance"
	"github.com/strengthening/goghostex/okex"

	. "github.com/strengthening/goghostex"
)

type Command struct {
	args []string

	currencyBasis   string
	currencyCounter string
	exchange        string
	tradeType       string

	proxy string
}

func (this *Command) New(args []string) {
	this.args = args

}

func (this *Command) Parse() error {

	cli := flag.NewFlagSet("command", flag.ExitOnError)
	cli.Usage = func() {
		fmt.Fprintf(os.Stderr, `ghostex golang version: 1.0.0\n`)
		cli.PrintDefaults()
	}

	cli.StringVar(
		&this.currencyBasis,
		"CB",
		"btc",
		"Set the currency of basis. Default is `btc`",
	)
	cli.StringVar(
		&this.currencyBasis,
		"currency-basis",
		"btc",
		"Set the currency of basis. Default is `btc`",
	)

	cli.StringVar(
		&this.currencyCounter,
		"CC",
		"usd",
		"Set the currency of counter. Default is `usd`",
	)
	cli.StringVar(
		&this.currencyCounter,
		"currency-counter",
		"usd",
		"Set the currency of counter. Default is `usd`",
	)

	cli.StringVar(
		&this.tradeType,
		"TT",
		"spot",
		"Set the trade type. Default is `SPOT`",
	)
	cli.StringVar(
		&this.tradeType,
		"trade-type",
		"spot",
		"Set the trade type. Default is `SPOT`",
	)

	cli.StringVar(
		&this.exchange,
		"EX",
		"okex",
		"Set the exchange. Default is `okex`",
	)
	cli.StringVar(
		&this.exchange,
		"exchange",
		"okex",
		"Set the exchange. Default is `okex`",
	)

	cli.StringVar(
		&this.proxy,
		"proxy",
		"",
		"Set the proxy url.Default is no proxy",
	)

	if err := cli.Parse(this.args[2:]); err != nil {
		panic(err)
	}

	if len(this.args) >= 2 {
		cmd := os.Args[1]
		switch {
		case cmd == "ticker" && len(os.Args) > 2:
			this.parseTicker()
		default:
			cli.Usage()
		}
	} else {
		cli.Usage()
	}

	return nil
}

var endPoints = map[string]string{
	"okex":    okex.ENDPOINT,
	"binance": binance.ENDPOINT,
}

func (this *Command) getNonAuthConfig() *APIConfig {
	config := &APIConfig{
		Endpoint:      endPoints[this.exchange],
		HttpClient:    http.DefaultClient,
		ApiKey:        "",
		ApiSecretKey:  "",
		ApiPassphrase: "",
		Location:      time.Now().Location(),
	}

	if this.proxy != "" {
		fmt.Println(this.proxy)
		config.HttpClient = &http.Client{
			Transport: &http.Transport{
				Proxy: func(req *http.Request) (*url.URL, error) {
					return url.Parse(this.proxy)
				},
			},
		}
	}
	return config
}

func (this *Command) getSponAPI() (API, error) {
	config := this.getNonAuthConfig()
	switch this.exchange {
	case "okex":
		return okex.New(config).Spot, nil
	case "binance":
		return binance.New(config).Spot, nil
	}
	return nil, nil
}

func (this *Command) parseTicker() {
	api, _ := this.getSponAPI()

	if ticker, _, err := api.GetTicker(CurrencyPair{
		CurrencyBasis: Currency{
			Symbol: strings.ToUpper(this.currencyBasis),
		},
		CurrencyCounter: Currency{
			Symbol: strings.ToUpper(this.currencyCounter),
		},
	}); err != nil {
		panic(err)
	} else {
		body, _ := json.Marshal(*ticker)
		fmt.Println(string(body))
	}

}
