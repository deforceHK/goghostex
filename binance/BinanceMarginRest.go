package binance

import (
	. "github.com/strengthening/goghostex"
)

type Margin struct {
	*Binance
}

func (this *Margin) Loan(record *LoanRecord) ([]byte, error) {
	panic("implement me")
}

func (this *Margin) GetOneLoan(record *LoanRecord) ([]byte, error) {
	panic("implement me")
}

func (this *Margin) Repay(record *LoanRecord) ([]byte, error) {
	panic("implement me")
}

func (this *Margin) PlaceMarginOrder(*Order) ([]byte, error) {
	panic("implement me")
}

func (this *Margin) CancelMarginOrder(*Order) ([]byte, error) {
	panic("implement me")
}

func (this *Margin) GetMarginOneOrder(*Order) ([]byte, error) {
	panic("implement me")
}

func (this *Margin) GetMarginUnFinishOrders(currency CurrencyPair) ([]Order, []byte, error) {
	panic("implement me")
}

func (this *Margin) GetMarginAccount(currency CurrencyPair) (*MarginAccount, []byte, error) {
	panic("implement me")
}

func (this *Margin) GetMarginTicker(currency CurrencyPair) (*Ticker, []byte, error) {
	panic("implement me")
}

func (this *Margin) GetMarginDepth(size int, currency CurrencyPair) (*Depth, []byte, error) {
	panic("implement me")
}

func (this *Margin) GetMarginKlineRecords(currency CurrencyPair, period, size, since int) ([]Kline, []byte, error) {
	panic("implement me")
}
