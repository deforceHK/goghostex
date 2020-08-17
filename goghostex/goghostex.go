package main

import (
	"flag"
)

var cliExchange = flag.String("exchange", "coinbase", "Input the exchange name. ")
var cliPair = flag.String("pair", "btc_usd", "Input the pair. ")
var cliType = flag.String("type", "spot", "Input the type. ")

var subCommand = map[string]string{
	"ticker":    "exchange ticker api",
	"co-ticker": "co-location info of exchange ticker api",
	"t2o":       "ticker to order time stat",
	"info":      "the exchange rule. ",
}

func main() {
	flag.Parse()
	paramCount := flag.NArg()
	firstParam := ""
	if paramCount != 0 {
		firstParam = flag.Arg(0)
	}

	_, exist := subCommand[firstParam]
	if paramCount == 0 || !exist {
		flag.PrintDefaults()
	} else {
		c := &Command{}
		c.Init(firstParam, flag.Args()[1:])
	}
}
