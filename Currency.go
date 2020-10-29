package goghostex

import (
	"strings"
)

type Currency struct {
	Symbol string `json:"symbol"`
	Desc   string `json:"-"`
}

func (c Currency) String() string {
	return c.Symbol
}

func (c Currency) Eq(c2 Currency) bool {
	return c.Symbol == c2.Symbol
}

var (
	UNKNOWN = Currency{"UNKNOWN", ""}

	USD = Currency{"USD", ""}
	EUR = Currency{"EUR", ""}
	CNY = Currency{"CNY", ""}
	KRW = Currency{"KRW", ""}
	JPY = Currency{"JPY", ""}
	SGD = Currency{"SGD", ""}
	HKD = Currency{"HKD", ""}

	USDT = Currency{"USDT", ""}
	USDC = Currency{"USDC", "https://www.centre.io/"}
	PAX  = Currency{"PAX", "https://www.paxos.com/"}
	DAI  = Currency{"DAI", ""}
	BUSD = Currency{"BUSD", ""}

	BTC  = Currency{"BTC", "https://bitcoin.org/"}
	XBT  = Currency{"XBT", ""}
	BCH  = Currency{"BCH", ""}
	LTC  = Currency{"LTC", ""}
	ETH  = Currency{"ETH", ""}
	ETC  = Currency{"ETC", ""}
	EOS  = Currency{"EOS", ""}
	BTS  = Currency{"BTS", ""}
	QTUM = Currency{"QTUM", ""}
	SC   = Currency{"SC", ""}
	ANS  = Currency{"ANS", ""}
	ZEC  = Currency{"ZEC", ""}
	DCR  = Currency{"DCR", ""}
	XRP  = Currency{"XRP", ""}
	NEO  = Currency{"NEO", ""}
	BSV  = Currency{"BSV", ""}
	LINK = Currency{"LINK", ""}
	XTZ  = Currency{"XTZ", ""}
	DASH = Currency{"DASH", ""}
	ADA  = Currency{"ADA", ""}
	DOT  = Currency{"DOT", ""}
	FIL  = Currency{"FIL", ""}

	UNI   = Currency{"UNI", ""}
	SUSHI = Currency{"SUSHI", ""}
	AAVE  = Currency{"AAVE", ""}
	COMP  = Currency{"COMP", ""}
	YFI   = Currency{"YFI", ""}
	YFII  = Currency{"YFII", ""}

	OKB = Currency{"OKB", "OKB is a global utility token issued by OK Blockchain Foundation. "}
	HT  = Currency{"HT", "HuoBi Token. "}
	BNB = Currency{"BNB", "BNB, or Binance Coin, is a cryptocurrency created by Binance. "}

	SHIT = Currency{"SHIT", "SHIT, There are some many shit coin in the market, we make the currency for dev."}
)

var currencyRelation = map[string]Currency{
	// fiat currency
	"usd": USD,
	"USD": USD,
	"eur": EUR,
	"EUR": EUR,
	"cny": CNY,
	"CNY": CNY,
	"jpy": JPY,
	"JPY": JPY,
	"krw": KRW,
	"KRW": KRW,
	"hkd": HKD,
	"HKD": HKD,
	"sgd": SGD,
	"SGD": SGD,

	// stable coin
	"usdt": USDT,
	"USDT": USDT,
	"usdc": USDC,
	"USDC": USDC,
	"pax":  PAX,
	"PAX":  PAX,
	"dai":  DAI,
	"DAI":  DAI,
	"busd": BUSD,
	"BUSD": BUSD,

	// crypto currency
	"btc":  BTC,
	"BTC":  BTC,
	"xbt":  XBT,
	"XBT":  XBT,
	"eth":  ETH,
	"ETH":  ETH,
	"eos":  EOS,
	"EOS":  EOS,
	"bch":  BCH,
	"BCH":  BCH,
	"bsv":  BSV,
	"BSV":  BSV,
	"ltc":  LTC,
	"LTC":  LTC,
	"ans":  ANS,
	"ANS":  ANS,
	"neo":  NEO,
	"NEO":  NEO,
	"link": LINK,
	"LINK": LINK,
	"dot":  DOT,
	"DOT":  DOT,
	"ada":  ADA,
	"ADA":  ADA,
	"fil":  FIL,
	"FIL":  FIL,

	// defi coin
	"uni":   UNI,
	"UNI":   UNI,
	"sushi": SUSHI,
	"SUSHI": SUSHI,
	"aave":  AAVE,
	"AAVE":  AAVE,
	"comp":  COMP,
	"COMP":  COMP,
	"YFI":   YFI,
	"yfi":   YFI,
	"YFII":  YFII,
	"yfii":  YFII,

	// exchange coin
	"okb": OKB,
	"OKB": OKB,
	"ht":  HT,
	"HT":  HT,
	"bnb": BNB,
	"BNB": BNB,

	// shit coin
	"shit": SHIT,
	"SHIT": SHIT,
}

func NewCurrency(symbol, desc string) Currency {
	currency, exist := currencyRelation[symbol]
	if exist {
		return currency
	}
	return Currency{strings.ToUpper(symbol), desc}
}
