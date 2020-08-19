package goghostex

// api interface
type MarginRestAPI interface {
	Loan(record *LoanRecord) ([]byte, error)
	GetOneLoan(record *LoanRecord) ([]byte, error)
	Repay(record *LoanRecord) ([]byte, error)

	PlaceMarginOrder(*Order) ([]byte, error)
	CancelMarginOrder(*Order) ([]byte, error)
	GetMarginOneOrder(*Order) ([]byte, error)
	GetMarginUnFinishOrders(pair Pair) ([]*Order, []byte, error)
	GetMarginAccount(pair Pair) (*MarginAccount, []byte, error)

	GetMarginTicker(pair Pair) (*Ticker, []byte, error)
	GetMarginDepth(size int, pair Pair) (*Depth, []byte, error)
	GetMarginKlineRecords(pair Pair, period, size, since int) ([]*Kline, []byte, error)
	GetExchangeName() string
	GetExchangeRule(pair Pair) (*Rule, []byte, error)
}
