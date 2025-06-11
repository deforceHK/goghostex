package binance

import (
	. "github.com/deforceHK/goghostex"
)

type Derivatives struct {
	*Binance
}

func (o *Derivatives) GetTicker(pair Pair) (*SwapTicker, []byte, error) {
	//TODO implement me
	panic("implement me")
}

func (o *Derivatives) GetDepth(pair Pair, size int) (*SwapDepth, []byte, error) {
	//TODO implement me
	panic("implement me")
}

func (o *Derivatives) GetContract(pair Pair) *SwapContract {
	// TODO implement me
	panic("implement me")
}

func (o *Derivatives) GetLimit(pair Pair) (float64, float64, error) {
	//TODO implement me
	panic("implement me")
}

func (o *Derivatives) GetKline(pair Pair, period, size, since int) ([]*SwapKline, []byte, error) {
	//TODO implement me
	panic("implement me")
}

func (o *Derivatives) GetOpenAmount(pair Pair) (float64, int64, []byte, error) {
	//TODO implement me
	panic("implement me")
}

func (o *Derivatives) GetFundingFees(pair Pair) ([][]interface{}, []byte, error) {
	//TODO implement me
	panic("implement me")
}

func (o *Derivatives) GetFundingFee(pair Pair) (float64, error) {
	//TODO implement me
	panic("implement me")
}

func (o *Derivatives) GetAccount() (*SwapAccount, []byte, error) {
	//TODO implement me
	panic("implement me")
}

func (o *Derivatives) PlaceOrder(order *SwapOrder) ([]byte, error) {

	panic("implement me")
}

func (o *Derivatives) CancelOrder(order *SwapOrder) ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (o *Derivatives) GetOrder(order *SwapOrder) ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (o *Derivatives) GetOrders(pair Pair) ([]*SwapOrder, []byte, error) {
	//TODO implement me
	panic("implement me")
}

func (o *Derivatives) GetUnFinishOrders(pair Pair) ([]*SwapOrder, []byte, error) {
	//TODO implement me
	panic("implement me")
}

func (o *Derivatives) GetPosition(pair Pair, openType FutureType) (*SwapPosition, []byte, error) {
	//TODO implement me
	panic("implement me")
}

func (o *Derivatives) AddMargin(pair Pair, openType FutureType, marginAmount float64) ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (o *Derivatives) ReduceMargin(pair Pair, openType FutureType, marginAmount float64) ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (o *Derivatives) GetAccountFlow() ([]*SwapAccountItem, []byte, error) {
	//TODO implement me
	panic("implement me")
}

func (o *Derivatives) GetPairFlow(pair Pair) ([]*SwapAccountItem, []byte, error) {
	//TODO implement me
	panic("implement me")
}

func (o *Derivatives) KeepAlive() {
	//TODO implement me
	panic("implement me")
}
