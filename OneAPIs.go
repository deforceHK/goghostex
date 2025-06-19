package goghostex

type OneRestAPI interface {
	// public api
	GetTicker(productId string) (*OneTicker, []byte, error)
	GetDepth(productId string, size int) (*OneDepth, []byte, error)
	GetInfos() ([]*OneInfo, []byte, error)

	// private api
	PlaceOrder(order *OneOrder) ([]byte, error)
	CancelOrder(order *OneOrder) ([]byte, error)
	GetOrder(order *OneOrder) ([]byte, error)

	// util api
	KeepAlive()
}
