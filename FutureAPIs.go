package goghostex

type FutureRestAPI interface {

	// public api
	GetExchangeName() string
	GetExchangeRule(pair Pair) (*FutureRule, []byte, error)
	GetEstimatedPrice(pair Pair) (float64, []byte, error)
	GetContract(pair Pair, contractType string) (*FutureContract, error)
	GetTicker(pair Pair, contractType string) (*FutureTicker, []byte, error)
	GetDepth(pair Pair, contractType string, size int) (*FutureDepth, []byte, error)
	GetLimit(pair Pair, contractType string) (float64, float64, error)
	GetIndex(pair Pair) (float64, []byte, error)
	GetKlineRecords(contractType string, pair Pair, period, size, since int) ([]*FutureKline, []byte, error)
	GetTrades(pair Pair, contractType string) ([]*Trade, []byte, error)

	// private api
	GetAccount() (*FutureAccount, []byte, error)
	PlaceOrder(order *FutureOrder) ([]byte, error)
	CancelOrder(order *FutureOrder) ([]byte, error)
	GetPosition(pair Pair, contractType string) ([]*FuturePosition, []byte, error)
	GetOrders(pair Pair, contractType string) ([]*FutureOrder, []byte, error)
	GetOrder(order *FutureOrder) ([]byte, error)
	GetUnFinishOrders(pair Pair, contractType string) ([]*FutureOrder, []byte, error)

	// util api
	KeepAlive()
}

type FutureWebsocketAPI interface {
	Init()

	//Login(config *APIConfig) error

	Subscribe(channel string) error

	//Unsubscribe(channel string) error

	Start()

	Close()
}
