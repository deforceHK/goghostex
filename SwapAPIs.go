package goghostex

type SwapRestAPI interface {
	// public api
	GetExchangeName() string
	GetTicker(pair Pair) (*SwapTicker, []byte, error)
	GetDepth(pair Pair, size int) (*SwapDepth, []byte, error)
	GetContract(pair Pair) *SwapContract
	GetLimit(pair Pair) (float64, float64, error)
	GetKline(pair Pair, period, size, since int) ([]*SwapKline, []byte, error)
	GetOpenAmount(pair Pair) (float64, int64, []byte, error)
	GetFundingFees(pair Pair) ([][]interface{}, []byte, error)
	GetFundingFee(pair Pair) (float64, error)

	// private api
	GetAccount() (*SwapAccount, []byte, error)
	PlaceOrder(order *SwapOrder) ([]byte, error)
	CancelOrder(order *SwapOrder) ([]byte, error)
	GetOrder(order *SwapOrder) ([]byte, error)
	GetOrders(pair Pair) ([]*SwapOrder, []byte, error)
	GetUnFinishOrders(pair Pair) ([]*SwapOrder, []byte, error)
	GetPosition(pair Pair, openType FutureType) (*SwapPosition, []byte, error)
	GetAccountFlow() ([]*SwapAccountItem, []byte, error)
	GetPairFlow(pair Pair) ([]*SwapAccountItem, []byte, error)

	// util api
	KeepAlive()
}
