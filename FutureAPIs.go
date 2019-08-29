package goghostex

type FutureRestAPI interface {
	GetExchangeName() string
	GetFutureEstimatedPrice(currencyPair CurrencyPair) (float64, []byte, error)
	GetFutureTicker(currencyPair CurrencyPair, contractType string) (*FutureTicker, []byte, error)
	GetFutureDepth(currencyPair CurrencyPair, contractType string, size int) (*FutureDepth, []byte, error)
	GetFutureIndex(currencyPair CurrencyPair) (float64, []byte, error)
	GetFutureUserinfo() (*FutureAccount, []byte, error)

	/**
	 *
	 * 期货下单
	 * @param currencyPair   btc_usd:比特币    ltc_usd :莱特币
	 * @param contractType   合约类型: this_week:当周   next_week:下周   month:当月   quarter:季度
	 * @param price  价格
	 * @param amount  委托数量
	 * @param openType   1:开多   2:开空   3:平多   4:平空
	 * @param matchPrice  是否为对手价 0:不是    1:是   ,当取值为1时,price无效
	 *
	 */
	PlaceFutureOrder(order *FutureOrder) ([]byte, error)

	/**
	 * 取消订单
	 * @param symbol   btc_usd:比特币    ltc_usd :莱特币
	 * @param contractType    合约类型: this_week:当周   next_week:下周   month:当月   quarter:季度
	 * @param orderId   订单ID
	 */
	CancelFutureOrder(order *FutureOrder) ([]byte, error)

	/**
	 * 用户持仓查询
	 * @param symbol   btc_usd:比特币    ltc_usd :莱特币
	 * @param contractType   合约类型: this_week:当周   next_week:下周   month:当月   quarter:季度
	 * @return
	 */
	GetFuturePosition(currencyPair CurrencyPair, contractType string) ([]FuturePosition, []byte, error)

	/**
	 *获取订单信息
	 */
	GetFutureOrders(orderIds []string, currencyPair CurrencyPair, contractType string) ([]FutureOrder, []byte, error)

	/**
	 *获取单个订单信息
	 */
	GetFutureOrder(orderId string, currencyPair CurrencyPair, contractType string) (*FutureOrder, []byte, error)

	/**
	 *获取未完成订单信息
	 */
	GetUnFinishFutureOrders(currencyPair CurrencyPair, contractType string) ([]FutureOrder, []byte, error)

	/**
	 *获取交易费
	 */
	GetFee() (float64, error)

	/**
	 *获取交易所的美元人民币汇率
	 */
	//GetExchangeRate() (float64, error)

	/**
	 *获取每张合约价值
	 */
	GetContractValue(currencyPair CurrencyPair) (float64, error)

	/**
	 *获取交割时间 星期(0,1,2,3,4,5,6)，小时，分，秒
	 */
	GetDeliveryTime() (int, int, int, int)

	/**
	 * 获取K线数据
	 */
	GetKlineRecords(contract_type string, currency CurrencyPair, period, size, since int) ([]FutureKline, []byte, error)

	/**
	 * 获取Trade数据
	 */
	GetTrades(contract_type string, currencyPair CurrencyPair, since int64) ([]Trade, error)
}

type FutureWebsocketAPI interface {
	Init()

	//Login(config *APIConfig) error

	Subscribe(channel string) error

	//Unsubscribe(channel string) error

	Start()

	Close()
}
