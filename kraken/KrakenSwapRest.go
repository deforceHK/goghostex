package kraken

import (
	"sync"
	"time"

	. "github.com/deforceHK/goghostex"
)

type Swap struct {
	*Kraken
	sync.Locker
	swapContracts SwapContracts

	nextUpdateContractTime time.Time // 下一次更新交易所contract信息
	LastKeepLiveTime       time.Time // 上一次keep live时间。
}

func (swap *Swap) GetOpenAmount(pair Pair) (float64, int64, []byte, error) {
	//TODO implement me
	panic("implement me")
}

func (swap *Swap) GetFundingFees(pair Pair) ([][]interface{}, []byte, error) {
	//TODO implement me
	panic("implement me")
}

func (swap *Swap) GetFundingFee(pair Pair) (float64, error) {
	//TODO implement me
	panic("implement me")
}

func (swap *Swap) GetAccount() (*SwapAccount, []byte, error) {

	//TODO implement me
	panic("implement me")
}

func (swap *Swap) PlaceOrder(order *SwapOrder) ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (swap *Swap) CancelOrder(order *SwapOrder) ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (swap *Swap) GetOrder(order *SwapOrder) ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (swap *Swap) GetOrders(pair Pair) ([]*SwapOrder, []byte, error) {
	//TODO implement me
	panic("implement me")
}

func (swap *Swap) GetUnFinishOrders(pair Pair) ([]*SwapOrder, []byte, error) {
	//TODO implement me
	panic("implement me")
}

func (swap *Swap) GetPosition(pair Pair, openType FutureType) (*SwapPosition, []byte, error) {
	//TODO implement me
	panic("implement me")
}

func (swap *Swap) AddMargin(pair Pair, openType FutureType, marginAmount float64) ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (swap *Swap) ReduceMargin(pair Pair, openType FutureType, marginAmount float64) ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (swap *Swap) KeepAlive() {
	//TODO implement me
	panic("implement me")
}
