package main

import (
	"flag"
	"fmt"

	"github.com/strengthening/goghostex"
)

var cliExchange = flag.String("exchange", "coinbase", "Input the exchange name. ")
var cliPair = flag.String("pair", "btc_usd", "Input the pair. ")
var cliType = flag.String("type", "spot", "Input the type. ")
var cliProxy = flag.String("proxy", "", "Input the proxy. ")
var cliApiKey = flag.String("api-key", "", "Input the api-key. ")
var cliApiSecret = flag.String("api-secret", "", "Input the api-secret. ")
var cliApiPassphrase = flag.String("api-passphrase", "", "Input the api-passphrase. ")

var subCommand = map[string]string{
	"ticker": "exchange ticker api",
	"t2o":    "ticker to order time stat",
	"info":   "the exchange rule. ",
}

func main() {
	// 初始化变量 cliFlag
	//Init()
	flag.Parse()
	// flag.Args() 函数返回没有被解析的命令行参数
	// func NArg() 函数返回没有被解析的命令行参数的个数

	paramCount := flag.NArg()
	firstParam := ""
	if paramCount != 0 {
		firstParam = flag.Arg(0)
	}

	_, exist := subCommand[firstParam]
	if paramCount == 0 || !exist {
		flag.PrintDefaults()
	} else if firstParam == "ticker" {

		fs := flag.NewFlagSet("ticker", flag.ExitOnError)
		fs.StringVar(cliExchange, "exchange", "coinbase", "Input the exchange name. ")
		fs.StringVar(cliType, "type", "spot", "Input the type. Default is spot. ")
		fs.StringVar(cliPair, "pair", "btc_usd", "Input the pair. Default is btc_usd. ")
		fs.StringVar(cliProxy, "proxy", "", "Input the proxy. ")
		fs.StringVar(cliApiKey, "api-key", "", "Input the api-key. ")
		fs.StringVar(cliApiSecret, "api-secret", "", "Input the api-secret. ")
		fs.StringVar(cliApiPassphrase, "api-passphrase", "", "Input the api-passphrase. ")

		_ = fs.Parse(flag.Args()[1:])
		initClients(*cliProxy, *cliApiKey, *cliApiSecret, *cliApiPassphrase)

		pair := goghostex.NewPair(*cliPair, "_")
		if ticker, raw, err := spotClients[*cliExchange].GetTicker(pair); err != nil {
			fmt.Println(err)
		} else {
			fmt.Printf("%s %s @%s 24h ticker: \n", *cliPair, *cliType, *cliExchange)
			fmt.Println("Last price:", ticker.Last)
			fmt.Println("Bid price:", ticker.Buy)
			fmt.Println("Ask price:", ticker.Sell)
			fmt.Println("High in 24h:", ticker.High)
			fmt.Println("Low in 24h:", ticker.Low)
			fmt.Println("Volume in 24h:", ticker.Vol)
			fmt.Println("Datetime:", ticker.Date)
			fmt.Println("------------------- Raw Response From Exchange -------------------")
			fmt.Println(string(raw))
		}
	} else if firstParam == "t2o" {

	} else if firstParam == "info" {

		fs := flag.NewFlagSet("ticker", flag.ExitOnError)
		fs.StringVar(cliExchange, "exchange", "coinbase", "Input the exchange name. ")
		fs.StringVar(cliType, "type", "spot", "Input the type. Default is spot. ")
		fs.StringVar(cliPair, "pair", "btc_usd", "Input the pair. Default is btc_usd. ")
		fs.StringVar(cliProxy, "proxy", "", "Input the proxy. ")

		_ = fs.Parse(flag.Args()[1:])
		initClients(*cliProxy, *cliApiKey, *cliApiSecret, *cliApiPassphrase)

		pair := goghostex.NewPair(*cliPair, "_")
		if rule, resp, err := spotClients[*cliExchange].GetExchangeRule(pair); err != nil {
			fmt.Println(err)
		} else {
			fmt.Println(string(resp))
			fmt.Println(rule)
		}
	}
}
