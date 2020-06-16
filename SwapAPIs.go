package goghostex

type SwapRestAPI interface {
	GetExchangeName() string
	// public data
	GetTicker(pair Pair) (*SwapTicker, []byte, error)
	// public data
	GetDepth(pair Pair, size int) (*SwapDepth, []byte, error)
	// public data
	GetStdDepth(pair Pair, size int) (*SwapStdDepth, []byte, error)
	// public data
	GetLimit(pair Pair) (float64, float64, error)
	// public data
	GetKline(pair Pair, period, size, since int) ([]*SwapKline, []byte, error)
	// public data
	GetOpenAmount(pair Pair) (float64, int64, []byte, error)
	// public data
	GetFundingFees(pair Pair) ([][]interface{}, []byte, error)
	// public data
	GetFee() (float64, error)

	// private api
	GetAccount() (*SwapAccount, []byte, error)
	// private api
	PlaceOrder(order *SwapOrder) ([]byte, error)
	// private api
	CancelOrder(order *SwapOrder) ([]byte, error)
	// private api
	GetOrders(pair Pair) ([]*SwapOrder, []byte, error)
	// private api
	GetOrder(order *SwapOrder) ([]byte, error)
	// private api
	GetUnFinishOrders(pair Pair) ([]*SwapOrder, []byte, error)
	// private api
	GetPosition(pair Pair, openType FutureType) (*SwapPosition, []byte, error)
	// private api
	AddMargin(pair Pair, openType FutureType, marginAmount float64) ([]byte, error)
	// private api
	ReduceMargin(pair Pair, openType FutureType, marginAmount float64) ([]byte, error)
	// private api, desc
	GetAccountFlow() ([]*SwapAccountItem, []byte, error)
}
