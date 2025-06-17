package goghostex

type OneRestAPI interface {
	// public api
	GetExchangeName() string
	GetTicker(pair Pair) (*Ticker, []byte, error)
	GetDepth(pair Pair, size int) (*Depth, []byte, error)

	GetKline(pair Pair, period, size, since int) ([]*Kline, []byte, error)

	// private api
	PlaceOrder(order *SwapOrder) ([]byte, error)
	CancelOrder(order *SwapOrder) ([]byte, error)
	GetOrder(order *SwapOrder) ([]byte, error)
	GetOrders(pair Pair) ([]*SwapOrder, []byte, error)

	GetPosition(pair Pair, openType FutureType) (*SwapPosition, []byte, error)

	// util api
	KeepAlive()
}
