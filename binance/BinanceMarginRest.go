package binance

import (
	. "github.com/strengthening/goghostex"
)

type Margin struct {
	*Binance
}

func (margin *Margin) Loan(record *LoanRecord) ([]byte, error) {
	panic("implement me")
}

func (margin *Margin) GetOneLoan(record *LoanRecord) ([]byte, error) {
	panic("implement me")
}

func (margin *Margin) Repay(record *LoanRecord) ([]byte, error) {
	panic("implement me")
}

func (margin *Margin) PlaceMarginOrder(*Order) ([]byte, error) {
	panic("implement me")
}

func (margin *Margin) CancelMarginOrder(*Order) ([]byte, error) {
	panic("implement me")
}

func (margin *Margin) GetMarginOneOrder(*Order) ([]byte, error) {
	panic("implement me")
}

func (margin *Margin) GetMarginUnFinishOrders(pair Pair) ([]*Order, []byte, error) {
	panic("implement me")
}

func (margin *Margin) GetMarginAccount(pair Pair) (*MarginAccount, []byte, error) {
	panic("implement me")
}

func (margin *Margin) GetMarginTicker(pair Pair) (*Ticker, []byte, error) {
	panic("implement me")
}

func (margin *Margin) GetMarginDepth(size int, pair Pair) (*Depth, []byte, error) {
	panic("implement me")
}

func (margin *Margin) GetMarginKlineRecords(pair Pair, period, size, since int) ([]*Kline, []byte, error) {
	panic("implement me")
}

func (margin *Margin) GetExchangeRule(pair Pair) (*Rule, []byte, error) {
	return nil, nil, nil
}

func (margin *Margin) KeepAlive() {

}
