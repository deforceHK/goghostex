package goghostex

type FutureRestAPI interface {

	// public api
	GetExchangeName() string
	GetExchangeRule(pair Pair) (*FutureRule, []byte, error)
	GetFutureEstimatedPrice(pair Pair) (float64, []byte, error)
	GetFutureTicker(pair Pair, contractType string) (*FutureTicker, []byte, error)
	GetFutureDepth(pair Pair, contractType string, size int) (*FutureDepth, []byte, error)
	GetFutureStdDepth(pair Pair, contractType string, size int) (*FutureStdDepth, []byte, error)
	GetFutureLimit(pair Pair, contractType string) (float64, float64, error)
	GetFutureIndex(pair Pair) (float64, []byte, error)
	GetFutureKlineRecords(contractType string, pair Pair, period, size, since int) ([]FutureKline, []byte, error)
	GetContractValue(pair Pair) (float64, error)
	/**
	 *获取交割时间 星期(0,1,2,3,4,5,6)，小时，分，秒
	 */
	GetDeliveryTime() (int, int, int, int)
	GetTrades(contractType string, pair Pair, since int64) ([]Trade, error)

	// private api
	GetFutureAccount() (*FutureAccount, []byte, error)
	PlaceFutureOrder(order *FutureOrder) ([]byte, error)
	CancelFutureOrder(order *FutureOrder) ([]byte, error)
	GetFuturePosition(pair Pair, contractType string) ([]FuturePosition, []byte, error)
	GetFutureOrders(orderIds []string, pair Pair, contractType string) ([]FutureOrder, []byte, error)
	GetFutureOrder(order *FutureOrder) ([]byte, error)
	GetUnFinishFutureOrders(pair Pair, contractType string) ([]FutureOrder, []byte, error)
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
