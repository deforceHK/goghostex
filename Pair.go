package goghostex

import (
	"strings"
)

type Pair struct {
	//The target currency, you want to buy or long
	//目标货币，你想买或者做多的。
	Basis Currency
	//The counter currency, you use it to buy or to mortgage
	//计数货币，你想用它来购买或者用它来做为抵押物。
	Counter Currency
}

func newPair(basis Currency, counter Currency) CurrencyPair {
	return CurrencyPair{basis, counter}
}

func NewPair(symbol string, sepChar string) CurrencyPair {
	currencys := strings.Split(symbol, sepChar)
	if len(currencys) == 2 {
		return newPair(NewCurrency(currencys[0], ""), NewCurrency(currencys[1], ""))
	}
	return UNKNOWN_PAIR
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
