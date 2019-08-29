package goghostex

/**
 *
 * models about account
 *
 **/

type FutureSubAccount struct {
	Currency      Currency
	AccountRights float64 //账户权益
	KeepDeposit   float64 //保证金
	ProfitReal    float64 //已实现盈亏
	ProfitUnreal  float64
	RiskRate      float64 //保证金率
}

type FutureAccount struct {
	FutureSubAccounts map[Currency]FutureSubAccount
}

/**
 *
 * models about market
 *
 **/

type FutureTicker struct {
	Ticker
	ContractType string `json:"contract_type"`
	ContractName string `json:"contract_name"`
	//LimitHigh    float64 `json:"limit_high,string"`
	//LimitLow     float64 `json:"limit_low,string"`
	//HoldAmount   float64 `json:"hold_amount,string"`
	//UnitAmount   float64 `json:"unit_amount,string"`
}

type FutureDepthRecords []FutureDepthRecord

type FutureDepthRecord struct {
	Price  float64
	Amount int64
}

func (dr FutureDepthRecords) Len() int {
	return len(dr)
}

func (dr FutureDepthRecords) Swap(i, j int) {
	dr[i], dr[j] = dr[j], dr[i]
}

func (dr FutureDepthRecords) Less(i, j int) bool {
	return dr[i].Price < dr[j].Price
}

type FutureDepth struct {
	ContractType string //for future
	ContractName string //for future
	Pair         CurrencyPair
	Timestamp    uint64
	Date         string
	AskList      FutureDepthRecords // Ascending order
	BidList      FutureDepthRecords // Descending order
}

type FutureKline struct {
	Kline
	Vol2 float64 //个数
}

/**
 *
 * models about trade
 *
 **/

type FutureOrder struct {
	// cid is important, when the order api return wrong, you can find it in unfinished api
	Cid            string
	OrderId        string
	Price          float64
	Amount         float64
	AvgPrice       float64
	DealAmount     float64
	OrderTimestamp uint64 // unit: ms
	OrderDate      string
	Status         TradeStatus
	OrderType      OrderType  //0：NORMAL 1：MAKER_ONLY 2：FOK 3：IOC
	Type           FutureType //1：OPEN_LONG 2：OPEN_SHORT 3：LIQUIDATE_LONG 4： LIQUIDATE_SHORT
	LeverRate      int
	Fee            float64
	Currency       CurrencyPair
	ContractType   string
	Exchange       string
	MatchPrice     int // some exchange need
}

type FuturePosition struct {
	BuyAmount      float64
	BuyAvailable   float64
	BuyPriceAvg    float64
	BuyPriceCost   float64
	BuyProfitReal  float64
	CreateDate     int64
	LeverRate      int
	SellAmount     float64
	SellAvailable  float64
	SellPriceAvg   float64
	SellPriceCost  float64
	SellProfitReal float64
	Symbol         CurrencyPair //btc_usd:比特币,ltc_usd:莱特币
	ContractType   string
	ContractId     int64
	ForceLiquPrice float64 //预估爆仓价
}

/**
 *
 * models about API config
 *
 */
type FutureAPIConfig struct {
	APIConfig
	Lever int // lever number , for future
}
