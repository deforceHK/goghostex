package goghostex

type SwapRestAPI interface {
	GetExchangeName() string

	GetTicker(pair Pair) (*SwapTicker, []byte, error)

	GetDepth(pair Pair, size int) (*SwapDepth, []byte, error)

	GetStdDepth(pair Pair, size int) (*SwapStdDepth, []byte, error)

	GetLimit(pair Pair) (float64, float64, error)

	GetKline(pair Pair, period, size, since int) ([]*SwapKline, []byte, error)

	GetAccount() (*SwapAccount, []byte, error)

	PlaceOrder(order *SwapOrder) ([]byte, error)

	CancelOrder(order *SwapOrder) ([]byte, error)

	GetPosition(pair Pair) (*SwapPosition, []byte, error)

	GetOrders(orderIds []string, pair Pair) ([]*SwapOrder, []byte, error)

	GetOrder(order *SwapOrder) ([]byte, error)

	GetUnFinishOrders(pair Pair) ([]*SwapOrder, []byte, error)

	GetFee() (float64, error)
}
