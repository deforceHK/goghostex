package main

import (
	"flag"
)

var cliExchange = flag.String("exchange", "coinbase", "Input the exchange name. ")
var cliPair = flag.String("pair", "btc_usd", "Input the pair. ")
var cliType = flag.String("type", "spot", "Input the type. ")

var sCommand = map[string]string{
	"ticker":    "exchange ticker api",
	"co-ticker": "co-location info of exchange ticker api",
	"depth":     "exchange depth api",
	"co-depth":  "co-location info of exchange depth api",
	"info":      "the exchange rule. ",
	"t2o":       "ticker to order time stat",
}

func main() {
	flag.Parse()
	paramCount := flag.NArg()
	firstParam := ""
	if paramCount != 0 {
		firstParam = flag.Arg(0)
	}

	_, exist := sCommand[firstParam]
	if paramCount == 0 || !exist {
		flag.PrintDefaults()
	} else {
		c := &Command{}
		c.Init(firstParam, flag.Args()[1:])
	}
}
