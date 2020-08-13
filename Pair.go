package goghostex

import (
	"strings"
)

var (
	BTC_USD = Pair{Basis: BTC, Counter: USD}
	LTC_USD = Pair{Basis: LTC, Counter: USD}
	ETH_USD = Pair{Basis: ETH, Counter: USD}
	EOS_USD = Pair{Basis: EOS, Counter: USD}

	BTC_USDT = Pair{Basis: BTC, Counter: USDT}
	ETH_USDT = Pair{Basis: ETH, Counter: USDT}
)

type Pair struct {
	//The target currency, you want to buy or long
	//目标货币，你想买或者做多的。
	Basis Currency
	//The counter currency, you use it to buy or to mortgage
	//计数货币，你想用它来购买或者用它来做为抵押物。
	Counter Currency
}

func newPair(basis Currency, counter Currency) Pair {
	return Pair{basis, counter}
}

func NewPair(symbol string, sepChar string) Pair {
	currencys := strings.Split(symbol, sepChar)
	if len(currencys) == 2 {
		return newPair(NewCurrency(currencys[0], ""), NewCurrency(currencys[1], ""))
	}
	return Pair{UNKNOWN, UNKNOWN}
}

func (pair Pair) String() string {
	return pair.ToSymbol("_", false)
}

func (pair Pair) Eq(otherPair Pair) bool {
	return pair.String() == otherPair.String()
}

func (pair Pair) ToSymbol(joinChar string, isUpper bool) string {
	rawSymbol := strings.Join([]string{pair.Basis.Symbol, pair.Counter.Symbol}, joinChar)
	if isUpper {
		return strings.ToUpper(rawSymbol)
	}
	return strings.ToLower(rawSymbol)
}
