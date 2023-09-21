package goghostex

type FutureRestAPI interface {

	// public api
	GetExchangeName() string
	GetContract(pair Pair, contractType string) (*FutureContract, error)
	GetTicker(pair Pair, contractType string) (*FutureTicker, []byte, error)
	GetDepth(pair Pair, contractType string, size int) (*FutureDepth, []byte, error)
	GetLimit(pair Pair, contractType string) (float64, float64, error)
	GetIndex(pair Pair) (float64, []byte, error)
	GetMark(pair Pair, contractType string) (float64, []byte, error)
	GetKlineRecords(contractType string, pair Pair, period, size, since int) ([]*FutureKline, []byte, error)
	GetTrades(pair Pair, contractType string) ([]*Trade, []byte, error)

	// private api
	GetAccount() (*FutureAccount, []byte, error)
	PlaceOrder(order *FutureOrder) ([]byte, error)
	CancelOrder(order *FutureOrder) ([]byte, error)
	GetOrders(pair Pair, contractType string) ([]*FutureOrder, []byte, error)
	GetOrder(order *FutureOrder) ([]byte, error)
	GetPairFlow(pair Pair) ([]*FutureAccountItem, []byte, error)

	// util api
	KeepAlive()
}

type FutureWebsocketAPI interface {
	Subscribe(v interface{})

	Unsubscribe(v interface{})

	Start()

	Stop()
}
