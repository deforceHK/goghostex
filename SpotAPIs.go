package goghostex

// api interface
type SpotRestAPI interface {

	// public api
	GetExchangeName() string
	GetExchangeRule(pair Pair) (*Rule, []byte, error)
	GetTicker(pair Pair) (*Ticker, []byte, error)
	GetDepth(size int, pair Pair) (*Depth, []byte, error)
	GetKlineRecords(pair Pair, period, size, since int) ([]*Kline, []byte, error)
	GetTrades(pair Pair, since int64) ([]*Trade, error)

	// private api
	LimitBuy(*Order) ([]byte, error)
	LimitSell(*Order) ([]byte, error)
	MarketBuy(*Order) ([]byte, error)
	MarketSell(*Order) ([]byte, error)
	CancelOrder(*Order) ([]byte, error)
	GetOneOrder(*Order) ([]byte, error)
	GetUnFinishOrders(pair Pair) ([]*Order, []byte, error)
	GetHistoryOrders(pair Pair, currentPage, pageSize int) ([]*Order, error)
	GetAccount() (*Account, []byte, error)

	// util api
	KeepAlive()
}
