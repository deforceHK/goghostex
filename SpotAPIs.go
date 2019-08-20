package goghostex

// api interface
type API interface {
	LimitBuy(amount, price string, currency CurrencyPair) (*Order, []byte, error)
	LimitSell(amount, price string, currency CurrencyPair) (*Order, []byte, error)
	MarketBuy(amount, price string, currency CurrencyPair) (*Order, []byte, error)
	MarketSell(amount, price string, currency CurrencyPair) (*Order, []byte, error)
	CancelOrder(orderId string, currency CurrencyPair) (bool, []byte, error)
	GetOneOrder(orderId string, currency CurrencyPair) (*Order, []byte, error)
	GetUnfinishOrders(currency CurrencyPair) ([]Order, []byte, error)
	GetOrderHistorys(currency CurrencyPair, currentPage, pageSize int) ([]Order, error)
	GetAccount() (*Account, []byte, error)

	GetTicker(currency CurrencyPair) (*Ticker, []byte, error)
	GetDepth(size int, currency CurrencyPair) (*Depth, []byte, error)
	GetKlineRecords(currency CurrencyPair, period, size, since int) ([]Kline, []byte, error)
	//非个人，整个交易所的交易记录
	GetTrades(currencyPair CurrencyPair, since int64) ([]Trade, error)

	GetExchangeName() string
}
