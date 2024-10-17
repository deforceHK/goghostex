package kraken

import (
	. "github.com/deforceHK/goghostex"
)

func (swap *Swap) GetAccountFlow() ([]*SwapAccountItem, []byte, error) {
	//TODO implement me
	panic("implement me")
}

func (swap *Swap) GetPairFlow(pair Pair) ([]*SwapAccountItem, []byte, error) {
	// todo sync fee
	return make([]*SwapAccountItem, 0), []byte(""), nil
}
