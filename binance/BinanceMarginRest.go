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

func (this *Margin) GetMarginUnFinishOrders(pair Pair) ([]Order, []byte, error) {
	panic("implement me")
}

func (this *Margin) GetMarginAccount(pair Pair) (*MarginAccount, []byte, error) {
	panic("implement me")
}

func (this *Margin) GetMarginTicker(pair Pair) (*Ticker, []byte, error) {
	panic("implement me")
}

func (this *Margin) GetMarginDepth(size int, pair Pair) (*Depth, []byte, error) {
	panic("implement me")
}

func (this *Margin) GetMarginKlineRecords(pair Pair, period, size, since int) ([]Kline, []byte, error) {
	panic("implement me")
}

func (this *Margin) GetExchangeRule(pair Pair) (*Rule, []byte, error) {
	return nil, nil, nil
}
