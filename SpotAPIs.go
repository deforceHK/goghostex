package goghostex

// api interface
type API interface {
	//just the orderid, get order info by func GetOneOrder

	LimitBuy(*Order) ([]byte, error)
	LimitSell(*Order) ([]byte, error)
	MarketBuy(*Order) ([]byte, error)
	MarketSell(*Order) ([]byte, error)
	CancelOrder(*Order) ([]byte, error)
	GetOneOrder(*Order) ([]byte, error)
	GetUnFinishOrders(currency CurrencyPair) ([]Order, []byte, error)
	GetOrderHistorys(currency CurrencyPair, currentPage, pageSize int) ([]Order, error)
	GetAccount() (*Account, []byte, error)

	GetTicker(currency CurrencyPair) (*Ticker, []byte, error)
	GetDepth(size int, currency CurrencyPair) (*Depth, []byte, error)
	GetKlineRecords(currency CurrencyPair, period, size, since int) ([]Kline, []byte, error)
	//非个人，整个交易所的交易记录
	GetTrades(currencyPair CurrencyPair, since int64) ([]Trade, error)

	GetExchangeName() string
}
