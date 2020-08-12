package goghostex

// api interface
type SpotAPI interface {
	//just the orderid, get order info by func GetOneOrder

	LimitBuy(*Order) ([]byte, error)
	LimitSell(*Order) ([]byte, error)
	MarketBuy(*Order) ([]byte, error)
	MarketSell(*Order) ([]byte, error)
	CancelOrder(*Order) ([]byte, error)
	GetOneOrder(*Order) ([]byte, error)
	GetUnFinishOrders(pair Pair) ([]Order, []byte, error)
	GetOrderHistorys(pair Pair, currentPage, pageSize int) ([]Order, error)
	GetAccount() (*Account, []byte, error)

	GetTicker(pair Pair) (*Ticker, []byte, error)
	GetDepth(size int, pair Pair) (*Depth, []byte, error)
	GetKlineRecords(pair Pair, period, size, since int) ([]Kline, []byte, error)
	//非个人，整个交易所的交易记录
	GetTrades(pair Pair, since int64) ([]Trade, error)

	GetExchangeName() string
}
