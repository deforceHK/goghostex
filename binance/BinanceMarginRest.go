package binance

import (
	. "github.com/strengthening/goghostex"
)

type Margin struct {
	*Binance
}

func (Margin) Loan(parameter BorrowParameter) (borrowId string, err error) {
	panic("implement me")
}

func (Margin) Repay(parameter RepaymentParameter) (repaymentId string, err error) {
	panic("implement me")
}

func (Margin) PlaceMarginOrder(*Order) ([]byte, error) {
	panic("implement me")
}

func (Margin) CancelMarginOrder(*Order) ([]byte, error) {
	panic("implement me")
}

func (Margin) GetMarginOneOrder(*Order) ([]byte, error) {
	panic("implement me")
}

func (Margin) GetMarginUnFinishOrders(pair CurrencyPair) ([]Order, []byte, error) {
	panic("implement me")
}

func (Margin) GetMarginAccount(pair CurrencyPair) (*MarginAccount, []byte, error) {
	panic("implement me")
}

func (Margin) GetMarginTicker(currency CurrencyPair) (*Ticker, []byte, error) {
	panic("implement me")
}

func (Margin) GetMarginDepth(size int, currency CurrencyPair) (*Depth, []byte, error) {
	panic("implement me")
}

func (Margin) GetMarginKlineRecords(currency CurrencyPair, period, size, since int) ([]Kline, []byte, error) {
	panic("implement me")
}
