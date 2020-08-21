package main

import (
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/strengthening/goghostex/binance"
	"github.com/strengthening/goghostex/coinbase"
	"github.com/strengthening/goghostex/okex"

	. "github.com/strengthening/goghostex"
)

var okexClient *okex.OKEx
var coinbaseClient *coinbase.Coinbase
var binanceClient *binance.Binance

var spotClients = map[string]SpotRestAPI{}
var swapClients = map[string]SwapRestAPI{}
var marginClients = map[string]MarginRestAPI{}
var futureClients = map[string]FutureRestAPI{}

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

type Command struct {
	Type         string
	Exchange     string
	Pair         string
	ContractType string

	Proxy string

	APIKey        string
	APISecret     string
	APIPassphrase string
}

func (c *Command) Init(subCommand string, args []string) {
	_, exist := sCommand[subCommand]

	fs := flag.NewFlagSet("ticker", flag.ExitOnError)
	fs.StringVar(&c.Exchange, "exchange", "coinbase", "Input the exchange name. ")
	fs.StringVar(&c.Type, "type", "spot", "Input the type. Default is spot. ")
	fs.StringVar(&c.Pair, "pair", "btc_usd", "Input the pair. Default is btc_usd. ")
	fs.StringVar(&c.ContractType, "contract-type", "", "Input the contract-type. It's nessary in future. ")
	fs.StringVar(&c.Proxy, "proxy", "", "Input the proxy. ")
	fs.StringVar(&c.APIKey, "api-key", "", "Input the api-key. ")
	fs.StringVar(&c.APISecret, "api-secret", "", "Input the api-secret. ")
	fs.StringVar(&c.APIPassphrase, "api-passphrase", "", "Input the api-passphrase. ")
	_ = fs.Parse(args)
	if !exist {
		fs.PrintDefaults()
		return
	}

	c.initClients()
	if subCommand == "ticker" {
		c.ticker()
	} else if subCommand == "co-ticker" {
		c.coTicker()
	} else if subCommand == "depth" {
		c.Depth()
	} else if subCommand == "co-depth" {
		c.coDepth()
	} else if subCommand == "info" {
		c.Info()
	} else if subCommand == "t2o" {
		//}else if subCommand == ""{
		//}else if subCommand == ""{
	}
}

func (c *Command) ticker() {
	ticker, response, delay, err := c.getTicker()
	if err != nil {
		fmt.Println(err)
	}

	fmt.Printf("%s %s @%s 24h ticker: \n", c.Pair, c.Type, c.Exchange)
	fmt.Println("Last price:", ticker.Last)
	fmt.Println("Bid price:", ticker.Buy)
	fmt.Println("Ask price:", ticker.Sell)
	fmt.Println("High in 24h:", ticker.High)
	fmt.Println("Low in 24h:", ticker.Low)
	fmt.Println("Volume in 24h:", ticker.Vol)
	fmt.Println("Datetime:", ticker.Date)
	fmt.Println("Request delay(ns):", delay)
	fmt.Println("------------------- Raw Response From Exchange -------------------")
	fmt.Println(string(response))

}

func (c *Command) coTicker() {
	receiveNum, errorNum := 0, 0
	receiveDelays := make([]int64, 0)
	totalDelays := int64(0)
	for i := 0; i < 5; i++ {
		_, _, delay, err := c.getTicker()
		if err != nil {
			errorNum += 1
			continue
		}
		receiveDelays = append(receiveDelays, delay)
		totalDelays += delay
	}
	fmt.Printf("%s %s @%s co-ticker: \n", c.Pair, c.Type, c.Exchange)
	fmt.Printf(
		"Request %d times, received %d times, errored %d times, avg delay is %.2f ns(nanosecond). \n",
		5, receiveNum, errorNum, float64(totalDelays)/5.0,
	)
	for i := 1; i <= len(receiveDelays); i++ {
		fmt.Printf("The %d sequence. The delay is %d ns. \n", i, receiveDelays[i-1])
	}
}

func (c *Command) getTicker() (*Ticker, []byte, int64, error) {

	p := NewPair(c.Pair, "_")

	var ticker Ticker
	var response []byte
	var delay int64
	switch c.Type {
	case "future":
		startTS := time.Now().UnixNano()
		t, resp, err := futureClients[c.Exchange].GetTicker(p, c.ContractType)
		if err != nil {
			return nil, nil, 0, err
		}
		finishTS := time.Now().UnixNano()
		ticker = t.Ticker
		response = resp
		delay = finishTS - startTS
	case "spot":
		startTS := time.Now().UnixNano()
		t, resp, err := spotClients[c.Exchange].GetTicker(p)
		if err != nil {
			return nil, nil, 0, err
		}
		finishTS := time.Now().UnixNano()
		ticker = *t
		response = resp
		delay = finishTS - startTS
	case "swap":
		startTS := time.Now().UnixNano()
		t, resp, err := swapClients[c.Exchange].GetTicker(p)
		if err != nil {
			return nil, nil, 0, err
		}
		finishTS := time.Now().UnixNano()
		ticker = Ticker{
			Pair: p, Last: t.Last, Buy: t.Buy, Sell: t.Sell,
			High: t.High, Low: t.Low, Vol: t.Vol, Timestamp: t.Timestamp, Date: t.Date,
		}
		response = resp
		delay = finishTS - startTS
	default:
		panic("imp it!")
	}

	return &ticker, response, delay, nil
}

func (c *Command) Depth() {
	depth, response, delay, err := c.getDepth()
	if err != nil {
		fmt.Println(err)
	}

	bidAmountCumsum, askAmountCumsum := float64(0), float64(0)
	for i := 0; i < len(depth.BidList); i++ {
		bidAmountCumsum += depth.BidList[i].Amount
	}
	for i := 0; i < len(depth.AskList); i++ {
		askAmountCumsum += depth.AskList[i].Amount
	}

	fmt.Printf("%s %s @%s 24h depth: \n", c.Pair, c.Type, c.Exchange)
	//fmt.Println("Last price:", ticker.Last)
	fmt.Println("First Bid price:", depth.BidList[0].Price)
	fmt.Println("Last Bid price:", depth.BidList[len(depth.BidList)-1].Price)
	fmt.Println("Bid cumsum amounts", bidAmountCumsum)

	fmt.Println("First Ask price:", depth.AskList[0].Price)
	fmt.Println("Last Ask price:", depth.AskList[len(depth.AskList)-1].Price)
	fmt.Println("Ask cumsum amounts", askAmountCumsum)
	fmt.Println("")
	fmt.Println("Sequence from remote:", depth.Sequence)
	fmt.Println("Datetime:", depth.Date)
	fmt.Println("Request delay(ns):", delay)
	fmt.Println("------------------- Raw Response From Exchange -------------------")
	fmt.Println(string(response))
}

func (c *Command) getDepth() (*Depth, []byte, int64, error) {

	p := NewPair(c.Pair, "_")

	var depth Depth
	var response []byte
	var delay int64
	switch c.Type {
	case "future":
		startTS := time.Now().UnixNano()
		fDepth, resp, err := futureClients[c.Exchange].GetDepth(p, c.ContractType, 200)
		if err != nil {
			return nil, nil, 0, err
		}
		finishTS := time.Now().UnixNano()
		depth = Depth{
			Pair:      p,
			Timestamp: fDepth.Timestamp,
			Sequence:  fDepth.Sequence,
			Date:      fDepth.Date,
			AskList:   fDepth.AskList,
			BidList:   fDepth.BidList,
		}
		response = resp
		delay = finishTS - startTS
	case "spot":
		startTS := time.Now().UnixNano()
		sDepth, resp, err := spotClients[c.Exchange].GetDepth(p, 200)
		if err != nil {
			return nil, nil, 0, err
		}
		finishTS := time.Now().UnixNano()
		depth = *sDepth
		response = resp
		delay = finishTS - startTS
	case "swap":
		startTS := time.Now().UnixNano()
		sDepth, resp, err := swapClients[c.Exchange].GetDepth(p, 200)
		if err != nil {
			return nil, nil, 0, err
		}
		finishTS := time.Now().UnixNano()
		depth = Depth{
			Pair:      p,
			Timestamp: sDepth.Timestamp,
			Sequence:  sDepth.Sequence,
			Date:      sDepth.Date,
			AskList:   sDepth.AskList,
			BidList:   sDepth.BidList,
		}
		response = resp
		delay = finishTS - startTS
	default:
		panic("imp it!")
	}

	return &depth, response, delay, nil
}

func (c *Command) coDepth() {

	receiveNum, errorNum := 0, 0
	receiveDelays := make([]int64, 0)
	totalDelays := int64(0)
	for i := 0; i < 5; i++ {
		_, _, delay, err := c.getDepth()
		if err != nil {
			errorNum += 1
			continue
		}
		receiveDelays = append(receiveDelays, delay)
		totalDelays += delay
	}
	fmt.Printf("%s %s @%s co-depth: \n", c.Pair, c.Type, c.Exchange)
	fmt.Printf(
		"Request %d times, received %d times, errored %d times, avg delay is %.2f ns(nanosecond). \n",
		5, receiveNum, errorNum, float64(totalDelays)/5.0,
	)
	for i := 1; i <= len(receiveDelays); i++ {
		fmt.Printf("The %d sequence. The delay is %d ns. \n", i, receiveDelays[i-1])
	}

}

func (c *Command) Info() {

	p := NewPair(c.Pair, "_")
	var rule *Rule
	var response []byte

	switch c.Type {
	case "future":
		fr, resp, err := futureClients[c.Exchange].GetExchangeRule(p)
		if err != nil {
			fmt.Println(err)
			return
		}
		rule = &fr.Rule
		response = resp
	case "spot":
		r, resp, err := spotClients[c.Exchange].GetExchangeRule(p)
		if err != nil {
			fmt.Println(err)
			return
		}
		rule = r
		response = resp
	case "swap":
		sr, resp, err := swapClients[c.Exchange].GetExchangeRule(p)
		if err != nil {
			fmt.Println(err)
			return
		}
		rule = &sr.Rule
		response = resp
	default:
		panic("imp it!")
	}

	fmt.Printf("%s %s @%s rule: \n", c.Pair, c.Type, c.Exchange)
	fmt.Printf("The min order amount: %f \n", rule.BaseMinSize)
	fmt.Printf("The order amount precision: %d \n", rule.BasePrecision)
	fmt.Printf("The order price precision: %d \n", rule.CounterPrecision)
	fmt.Println("------------------- Raw Response From Exchange -------------------")
	fmt.Println(string(response))
}

func (c *Command) initClients() {
	loc := time.Now().Location()
	okexClient = okex.New(
		&APIConfig{
			Endpoint:      okex.ENDPOINT,
			HttpClient:    getHttpClient(c.Proxy),
			ApiKey:        c.APIKey,
			ApiSecretKey:  c.APISecret,
			ApiPassphrase: c.APIPassphrase,
			Location:      loc,
		},
	)
	coinbaseClient = coinbase.New(
		&APIConfig{
			Endpoint:      coinbase.ENDPOINT,
			HttpClient:    getHttpClient(c.Proxy),
			ApiKey:        c.APIKey,
			ApiSecretKey:  c.APISecret,
			ApiPassphrase: c.APIPassphrase,
			Location:      loc,
		},
	)
	binanceClient = binance.New(
		&APIConfig{
			Endpoint:      binance.ENDPOINT,
			HttpClient:    getHttpClient(c.Proxy),
			ApiKey:        c.APIKey,
			ApiSecretKey:  c.APISecret,
			ApiPassphrase: c.APIPassphrase,
			Location:      loc,
		},
	)

	if c.Type == "spot" {
		spotClients[OKEX] = okexClient.Spot
		spotClients[COINBASE] = coinbaseClient.Spot
		spotClients[BINANCE] = binanceClient.Spot
	} else if c.Type == "future" {
		futureClients[OKEX] = okexClient.Future
	} else if c.Type == "swap" {
		swapClients[OKEX] = okexClient.Swap
		swapClients[BINANCE] = binanceClient.Swap
	} else {
		fmt.Printf("The command not support %s in %s. \n", c.Type, c.Exchange)
	}

}
