package kraken

import (
	//"sync"
	//"time"

	. "github.com/deforceHK/goghostex"
)

func (swap *Swap) GetTicker(pair Pair) (*SwapTicker, []byte, error) {
	// TODO implement me
	panic("implement me")
}

func (swap *Swap) GetDepth(pair Pair, size int) (*SwapDepth, []byte, error) {
	// TODO implement me
	panic("implement me")
}

func (swap *Swap) GetContract(pair Pair) *SwapContract {
	// TODO implement me
	panic("implement me")
}

func (swap *Swap) GetLimit(pair Pair) (float64, float64, error) {
	// TODO implement me
	panic("implement me")
}

func (swap *Swap) GetKline(pair Pair, period, size, since int) ([]*SwapKline, []byte, error) {
	// TODO implement me
	panic("implement me")
}
