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

type CurrencyPair struct {
	//The target currency, you want to buy or long
	//目标货币，你想买或者做多的。
	CurrencyBasis Currency
	//The counter currency, you use it to buy or to mortgage
	//计数货币，你想用它来购买或者用它来做为抵押物。
	CurrencyCounter Currency
}

var (
	UNKNOWN = Currency{"UNKNOWN", ""}

	USD = Currency{"USD", ""}
	EUR = Currency{"EUR", ""}
	CNY = Currency{"CNY", ""}
	KRW = Currency{"KRW", ""}
	JPY = Currency{"JPY", ""}

	USDT = Currency{"USDT", ""}
	USDC = Currency{"USDC", "https://www.centre.io/"}
	PAX  = Currency{"PAX", "https://www.paxos.com/"}
	DAI  = Currency{"DAI", ""}
	BUSD = Currency{"BUSD", ""}

	BTC  = Currency{"BTC", "https://bitcoin.org/"}
	XBT  = Currency{"XBT", ""}
	BCC  = Currency{"BCC", ""}
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
	BTG  = Currency{"BTG", ""}
	BCD  = Currency{"BCD", ""}
	NEO  = Currency{"NEO", ""}
	HSR  = Currency{"HSR", ""}
	BSV  = Currency{"BSV", ""}
	LINK = Currency{"LINK", ""}
	XTZ  = Currency{"XTZ", ""}
	DASH = Currency{"DASH", ""}
	ADA  = Currency{"ADA", ""}

	OKB = Currency{"OKB", "OKB is a global utility token issued by OK Blockchain Foundation"}
	HT  = Currency{"HT", "HuoBi Token"}
	BNB = Currency{"BNB", "BNB, or Binance Coin, is a cryptocurrency created by Binance."}

	// crypto to national currency pair
	BTC_USD = CurrencyPair{BTC, USD}
	LTC_USD = CurrencyPair{LTC, USD}
	ETH_USD = CurrencyPair{ETH, USD}
	ETC_USD = CurrencyPair{ETC, USD}
	BCH_USD = CurrencyPair{BCH, USD}
	BCC_USD = CurrencyPair{BCC, USD}
	XRP_USD = CurrencyPair{XRP, USD}
	BCD_USD = CurrencyPair{BCD, USD}
	EOS_USD = CurrencyPair{EOS, USD}
	BTG_USD = CurrencyPair{BTG, USD}
	BSV_USD = CurrencyPair{BSV, USD}

	BTC_CNY  = CurrencyPair{BTC, CNY}
	LTC_CNY  = CurrencyPair{LTC, CNY}
	BCC_CNY  = CurrencyPair{BCC, CNY}
	ETH_CNY  = CurrencyPair{ETH, CNY}
	ETC_CNY  = CurrencyPair{ETC, CNY}
	EOS_CNY  = CurrencyPair{EOS, CNY}
	BTS_CNY  = CurrencyPair{BTS, CNY}
	QTUM_CNY = CurrencyPair{QTUM, CNY}
	SC_CNY   = CurrencyPair{SC, CNY}
	ANS_CNY  = CurrencyPair{ANS, CNY}
	ZEC_CNY  = CurrencyPair{ZEC, CNY}

	BTC_KRW = CurrencyPair{BTC, KRW}
	ETH_KRW = CurrencyPair{ETH, KRW}
	ETC_KRW = CurrencyPair{ETC, KRW}
	LTC_KRW = CurrencyPair{LTC, KRW}
	BCH_KRW = CurrencyPair{BCH, KRW}

	XRP_EUR = CurrencyPair{XRP, EUR}

	BTC_JPY = CurrencyPair{BTC, JPY}
	LTC_JPY = CurrencyPair{LTC, JPY}
	ETH_JPY = CurrencyPair{ETH, JPY}
	ETC_JPY = CurrencyPair{ETC, JPY}
	BCH_JPY = CurrencyPair{BCH, JPY}

	// crypto to crypto currency pair
	BTC_USDT = CurrencyPair{BTC, USDT}
	LTC_USDT = CurrencyPair{LTC, USDT}
	BCH_USDT = CurrencyPair{BCH, USDT}
	BCC_USDT = CurrencyPair{BCC, USDT}
	ETC_USDT = CurrencyPair{ETC, USDT}
	ETH_USDT = CurrencyPair{ETH, USDT}
	BCD_USDT = CurrencyPair{BCD, USDT}
	NEO_USDT = CurrencyPair{NEO, USDT}
	EOS_USDT = CurrencyPair{EOS, USDT}
	XRP_USDT = CurrencyPair{XRP, USDT}
	HSR_USDT = CurrencyPair{HSR, USDT}
	BSV_USDT = CurrencyPair{BSV, USDT}
	OKB_USDT = CurrencyPair{OKB, USDT}
	HT_USDT  = CurrencyPair{HT, USDT}
	BNB_USDT = CurrencyPair{BNB, USDT}

	LTC_BTC = CurrencyPair{LTC, BTC}
	ETH_BTC = CurrencyPair{ETH, BTC}
	ETC_BTC = CurrencyPair{ETC, BTC}
	BCC_BTC = CurrencyPair{BCC, BTC}
	BCH_BTC = CurrencyPair{BCH, BTC}
	DCR_BTC = CurrencyPair{DCR, BTC}
	XRP_BTC = CurrencyPair{XRP, BTC}
	BTG_BTC = CurrencyPair{BTG, BTC}
	BCD_BTC = CurrencyPair{BCD, BTC}
	NEO_BTC = CurrencyPair{NEO, BTC}
	EOS_BTC = CurrencyPair{EOS, BTC}
	HSR_BTC = CurrencyPair{HSR, BTC}
	BSV_BTC = CurrencyPair{BSV, BTC}
	OKB_BTC = CurrencyPair{OKB, BTC}
	HT_BTC  = CurrencyPair{HT, BTC}
	BNB_BTC = CurrencyPair{BNB, BTC}

	ETC_ETH = CurrencyPair{ETC, ETH}
	EOS_ETH = CurrencyPair{EOS, ETH}
	ZEC_ETH = CurrencyPair{ZEC, ETH}
	NEO_ETH = CurrencyPair{NEO, ETH}
	HSR_ETH = CurrencyPair{HSR, ETH}
	LTC_ETH = CurrencyPair{LTC, ETH}

	UNKNOWN_PAIR = CurrencyPair{UNKNOWN, UNKNOWN}
)

func (c CurrencyPair) String() string {
	return c.ToSymbol("_")
}

func (c CurrencyPair) Eq(c2 CurrencyPair) bool {
	return c.String() == c2.String()
}

func (c Currency) AdaptBchToBcc() Currency {
	if c.Symbol == "BCH" || c.Symbol == "bch" {
		return BCC
	}
	return c
}

func (c Currency) AdaptBccToBch() Currency {
	if c.Symbol == "BCC" || c.Symbol == "bcc" {
		return BCH
	}
	return c
}

var currencyRelation = map[string]Currency{
	//fiat currency
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

	//stable coin
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

	//crypto currency
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
	"ltc":  LTC,
	"LTC":  LTC,
	"ans":  ANS,
	"ANS":  ANS,
	"neo":  NEO,
	"NEO":  NEO,
	"link": LINK,
	"LINK": LINK,

	//exchange coin
	"okb": OKB,
	"OKB": OKB,
	"ht":  HT,
	"HT":  HT,
	"bnb": BNB,
	"BNB": BNB,
}

func NewCurrency(symbol, desc string) Currency {
	currency, exist := currencyRelation[symbol]
	if exist {
		return currency
	}
	return Currency{strings.ToUpper(symbol), desc}
}

func newCurrencyPair(CurrencyBasis Currency, CurrencyCounter Currency) CurrencyPair {
	return CurrencyPair{CurrencyBasis, CurrencyCounter}
}

func NewCurrencyPair(currencyPairSymbol string) CurrencyPair {
	currencys := strings.Split(currencyPairSymbol, "_")
	if len(currencys) == 2 {
		return newCurrencyPair(NewCurrency(currencys[0], ""), NewCurrency(currencys[1], ""))
	}
	return UNKNOWN_PAIR
}

func (pair CurrencyPair) ToSymbol(joinChar string) string {
	return strings.Join([]string{pair.CurrencyBasis.Symbol, pair.CurrencyCounter.Symbol}, joinChar)
}

func (pair CurrencyPair) AdaptUsdtToUsd() CurrencyPair {
	CurrencyCounter := pair.CurrencyCounter
	if pair.CurrencyCounter.Eq(USDT) {
		CurrencyCounter = USD
	}
	return CurrencyPair{pair.CurrencyBasis, CurrencyCounter}
}

func (pair CurrencyPair) AdaptUsdToUsdt() CurrencyPair {
	CurrencyCounter := pair.CurrencyCounter
	if pair.CurrencyCounter.Eq(USD) {
		CurrencyCounter = USDT
	}
	return CurrencyPair{pair.CurrencyBasis, CurrencyCounter}
}

//for to symbol lower , Not practical '==' operation method
func (pair CurrencyPair) ToLower() CurrencyPair {
	return CurrencyPair{
		Currency{strings.ToLower(pair.CurrencyBasis.Symbol), ""},
		Currency{strings.ToLower(pair.CurrencyCounter.Symbol), ""},
	}
}

func (pair CurrencyPair) Reverse() CurrencyPair {
	return CurrencyPair{
		pair.CurrencyCounter,
		pair.CurrencyBasis,
	}
}

func (pair CurrencyPair) MarshalJSON() ([]byte, error) {
	return []byte(`"` + pair.ToLower().ToSymbol("_") + `"`), nil
}

// Important ! uncomment here cause the func can not chanage the pair value. so the Currencypair just can used in MarshalJSON
// you can use Symbol string column instead the CurrencyPair when MarshaJSON.

//func (pair CurrencyPair) UnmarshalJSON(input []byte) error {
//	symbol := string(input)[1:]
//	symbol = symbol[0: len(symbol)-1]
//	currencies := strings.Split(symbol, "_")
//
//	if len(currencies) != 2 {
//		return errors.New("The symbol can not bind currency pair. ")
//	}
//
//	pair.CurrencyBasis = NewCurrency(currencies[0],"")
//	pair.CurrencyCounter = NewCurrency(currencies[1],"")
//	fmt.Println(pair)
//	return nil
//}
