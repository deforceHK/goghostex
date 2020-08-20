package okex

import (
	. "github.com/strengthening/goghostex"
)

type Swap struct {
	*OKEx
}

func (swap *Swap) GetTicker(pair Pair) (*SwapTicker, []byte, error) {
	panic("implement me")
}

func (swap *Swap) GetDepth(pair Pair, size int) (*SwapDepth, []byte, error) {
	panic("implement me")
}

func (swap *Swap) GetLimit(pair Pair) (float64, float64, error) {
	panic("implement me")
}

func (swap *Swap) GetKline(pair Pair, period, size, since int) ([]*SwapKline, []byte, error) {
	panic("implement me")
}

func (swap *Swap) GetOpenAmount(pair Pair) (float64, int64, []byte, error) {
	panic("implement me")
}

func (swap *Swap) GetFundingFees(pair Pair) ([][]interface{}, []byte, error) {
	panic("implement me")
}

func (swap *Swap) GetFundingFee(pair Pair) (float64, error) {
	panic("implement me")
}

func (swap *Swap) GetAccount() (*SwapAccount, []byte, error) {
	panic("implement me")
}

func (swap *Swap) PlaceOrder(order *SwapOrder) ([]byte, error) {
	panic("implement me")
}

func (swap *Swap) CancelOrder(order *SwapOrder) ([]byte, error) {
	panic("implement me")
}

func (swap *Swap) GetOrders(pair Pair) ([]*SwapOrder, []byte, error) {
	panic("implement me")
}

func (swap *Swap) GetOrder(order *SwapOrder) ([]byte, error) {
	panic("implement me")
}

func (swap *Swap) GetUnFinishOrders(pair Pair) ([]*SwapOrder, []byte, error) {
	panic("implement me")
}

func (swap *Swap) GetPosition(pair Pair, openType FutureType) (*SwapPosition, []byte, error) {
	panic("implement me")
}

func (swap *Swap) AddMargin(pair Pair, openType FutureType, marginAmount float64) ([]byte, error) {
	panic("implement me")
}

func (swap *Swap) ReduceMargin(pair Pair, openType FutureType, marginAmount float64) ([]byte, error) {
	panic("implement me")
}

func (swap *Swap) GetAccountFlow() ([]*SwapAccountItem, []byte, error) {
	panic("implement me")
}

func (swap *Swap) GetExchangeRule(pair Pair) (*SwapRule, []byte, error) {
	panic("implement me")
	//return nil,nil,nil
}

func (swap *Swap) KeepAlive() {
	panic("implement me")
}
