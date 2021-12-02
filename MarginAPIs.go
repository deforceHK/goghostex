package goghostex

// api interface
type MarginRestAPI interface {
	// public api
	GetTicker(pair Pair) (*Ticker, []byte, error)
	GetDepth(pair Pair, size int) (*Depth, []byte, error)
	GetKlineRecords(pair Pair, period, size, since int) ([]*Kline, []byte, error)
	GetExchangeName() string
	//GetExchangeRule(pair Pair) (*Rule, []byte, error)

	// private api
	GetAccount(pair Pair) (*MarginAccount, []byte, error)
	PlaceOrder(order *Order) ([]byte, error)
	CancelOrder(order *Order) ([]byte, error)
	GetOrder(order *Order) ([]byte, error)
	GetOrders(pair Pair) ([]*Order, []byte, error) // all orders desc recently
	GetUnFinishOrders(pair Pair) ([]*Order, []byte, error)
	PlaceLoan(loan *Loan) ([]byte, error)
	GetLoan(loan *Loan) ([]byte, error)
	ReturnLoan(loan *Loan) ([]byte, error)

	// util api
	KeepAlive()
}
