package gate

import (
	. "github.com/strengthening/goghostex"
)

type Spot struct {
	*Gate
}

func (spot *Spot) GetExchangeRule(pair Pair) (*Rule, []byte, error) {
	panic("implement me")
}

func (spot *Spot) GetTicker(pair Pair) (*Ticker, []byte, error) {
	panic("implement me")
}

func (spot *Spot) GetDepth(pair Pair, size int) (*Depth, []byte, error) {
	panic("implement me")
}

func (spot *Spot) GetKlineRecords(pair Pair, period, size, since int) ([]*Kline, []byte, error) {
	panic("implement me")
}

func (spot *Spot) GetTrades(pair Pair, since int64) ([]*Trade, error) {
	panic("implement me")
}

func (spot *Spot) GetAccount() (*Account, []byte, error) {
	panic("implement me")
}

func (spot *Spot) PlaceOrder(order *Order) ([]byte, error) {
	panic("implement me")
}

func (spot *Spot) CancelOrder(order *Order) ([]byte, error) {
	panic("implement me")
}

func (spot *Spot) GetOrder(order *Order) ([]byte, error) {
	panic("implement me")
}

func (spot *Spot) GetOrders(pair Pair) ([]*Order, error) {
	panic("implement me")
}

func (spot *Spot) GetUnFinishOrders(pair Pair) ([]*Order, []byte, error) {
	panic("implement me")
}

func (spot *Spot) KeepAlive() {
	panic("implement me")
}
