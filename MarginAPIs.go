package goghostex

// api interface
type MarginAPI interface {
	Loan(record *LoanRecord) ([]byte, error)
	GetOneLoan(record *LoanRecord) ([]byte, error)
	Repay(record *LoanRecord) ([]byte, error)

	PlaceMarginOrder(*Order) ([]byte, error)
	CancelMarginOrder(*Order) ([]byte, error)
	GetMarginOneOrder(*Order) ([]byte, error)
	GetMarginUnFinishOrders(currency CurrencyPair) ([]Order, []byte, error)
	GetMarginAccount(currency CurrencyPair) (*MarginAccount, []byte, error)

	GetMarginTicker(currency CurrencyPair) (*Ticker, []byte, error)
	GetMarginDepth(size int, currency CurrencyPair) (*Depth, []byte, error)
	GetMarginKlineRecords(currency CurrencyPair, period, size, since int) ([]Kline, []byte, error)
	GetExchangeName() string
}
