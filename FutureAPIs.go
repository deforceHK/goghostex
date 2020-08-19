package goghostex

type FutureRestAPI interface {

	// public api
	GetExchangeName() string
	GetExchangeRule(pair Pair) (*FutureRule, []byte, error)
	GetEstimatedPrice(pair Pair) (float64, []byte, error)
	GetTicker(pair Pair, contractType string) (*FutureTicker, []byte, error)
	GetDepth(pair Pair, contractType string, size int) (*FutureDepth, []byte, error)
	GetLimit(pair Pair, contractType string) (float64, float64, error)
	GetIndex(pair Pair) (float64, []byte, error)
	GetKlineRecords(contractType string, pair Pair, period, size, since int) ([]*FutureKline, []byte, error)
	GetContractValue(pair Pair) (float64, error)
	GetDeliveryTime() (int, int, int, int) //获取交割时间 星期(0,1,2,3,4,5,6)，小时，分，秒
	GetTrades(contractType string, pair Pair, since int64) ([]*Trade, error)

	// private api
	GetAccount() (*FutureAccount, []byte, error)
	PlaceOrder(order *FutureOrder) ([]byte, error)
	CancelOrder(order *FutureOrder) ([]byte, error)
	GetPosition(pair Pair, contractType string) ([]*FuturePosition, []byte, error)
	GetOrders(orderIds []string, pair Pair, contractType string) ([]*FutureOrder, []byte, error)
	GetOrder(order *FutureOrder) ([]byte, error)
	GetUnFinishOrders(pair Pair, contractType string) ([]*FutureOrder, []byte, error)
	GetFee() (float64, error)
}

type FutureWebsocketAPI interface {
	Init()

	//Login(config *APIConfig) error

	Subscribe(channel string) error

	//Unsubscribe(channel string) error

	Start()

	Close()
}
