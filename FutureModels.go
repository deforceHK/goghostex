package goghostex

/*
	models about account
*/
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

/*
	models about market
*/

type FutureTicker struct {
	Ticker
	ContractType string  `json:"omitempty"`
	ContractId   int     `json:"contract_id"`
	LimitHigh    float64 `json:"limit_high,string"`
	LimitLow     float64 `json:"limit_low,string"`
	HoldAmount   float64 `json:"hold_amount,string"`
	UnitAmount   float64 `json:"unit_amount,string"`
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
	Pair         CurrencyPair
	Timestamp    uint64
	Date         string
	AskList      FutureDepthRecords // Descending order
	BidList      FutureDepthRecords // Descending order
}

type FutureKline struct {
	Kline
	Vol2 float64 //个数
}

/*
	models about trade
*/

type FutureOrder struct {
	ClientOid  string //自定义ID，GoEx内部自动生成
	OrderId    string
	Price      float64
	Amount     float64
	AvgPrice   float64
	DealAmount float64

	OrderTime      int64
	OrderTimestamp uint64
	OrderDate      string
	Status         TradeStatus
	Currency       CurrencyPair
	OrderType      int     //ORDINARY=0 POST_ONLY=1 FOK= 2 IOC= 3
	OType          int     //1：开多 2：开空 3：平多 4： 平空
	LeverRate      int     //倍数
	Fee            float64 //手续费
	ContractName   string
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

/*
	models about API config
*/
type FutureAPIConfig struct {
	APIConfig
	Lever int // lever number , for future
}
