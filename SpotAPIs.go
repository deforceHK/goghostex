package goghostex

// api interface
type SpotRestAPI interface {

	// public api
	GetExchangeName() string
	//GetExchangeRule(pair Pair) (*Rule, []byte, error)
	GetTicker(pair Pair) (*Ticker, []byte, error)
	GetDepth(pair Pair, size int) (*Depth, []byte, error)
	GetKlineRecords(pair Pair, period, size, since int) ([]*Kline, []byte, error)
	GetTrades(pair Pair, since int64) ([]*Trade, error)

	// private api
	GetAccount() (*Account, []byte, error)
	PlaceOrder(order *Order) ([]byte, error)
	CancelOrder(order *Order) ([]byte, error)
	GetOrder(order *Order) ([]byte, error)
	GetOrders(pair Pair) ([]*Order, error) // dealed orders
	GetUnFinishOrders(pair Pair) ([]*Order, []byte, error)

	// util api
	KeepAlive()
}
