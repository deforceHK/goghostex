package okex

import "sync"

type OrderBookArbi struct {
	*WSMarketOKEx

	SpotPrices []int64
	SpotData   map[int64]float64

	BidData      map[string]map[int64]float64
	AskData      map[string]map[int64]float64
	SeqData      map[string]int64
	TsData       map[string]int64
	OrderBookMux *sync.RWMutex

	// if the channel is not nil, send the update message to the channel. User should read the channel in the loop.
	UpdateChan chan string
}
