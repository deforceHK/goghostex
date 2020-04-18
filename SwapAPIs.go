package goghostex

type SwapRestAPI interface {
	GetExchangeName() string

	GetTicker(pair Pair) (*SwapTicker, []byte, error)

	GetDepth(pair Pair, size int) (*SwapDepth, []byte, error)

	GetStdDepth(pair Pair, size int) (*SwapStdDepth, []byte, error)

	GetLimit(pair Pair) (float64, float64, error)

	GetKline(pair Pair, period, size, since int) ([]*SwapKline, []byte, error)

	GetOpenAmount(pair Pair) (float64, int64, []byte, error)

	GetFundingFees(pair Pair) ([][]interface{}, []byte, error)

	GetFee() (float64, error)

	GetAccount() (*SwapAccount, []byte, error)

	PlaceOrder(order *SwapOrder) ([]byte, error)

	CancelOrder(order *SwapOrder) ([]byte, error)

	GetOrders(pair Pair) ([]*SwapOrder, []byte, error)

	GetOrder(order *SwapOrder) ([]byte, error)

	GetUnFinishOrders(pair Pair) ([]*SwapOrder, []byte, error)

	GetPosition(pair Pair, openType FutureType) (*SwapPosition, []byte, error)

	AddMargin(pair Pair, openType FutureType, marginAmount float64) ([]byte, error)

	ReduceMargin(pair Pair, openType FutureType, marginAmount float64) ([]byte, error)
}
