package goghostex

import (
	"errors"
	"time"
)

/**
 *
 * models about account
 *
 **/

type FutureSubAccount struct {
	Currency Currency
	// The future margin 期货保证金 == marginFilled+ marginUnFilled
	Margin float64
	// The future is filled 已经成交的订单占用的期货保证金
	MarginDealed float64
	// The future is unfilled 未成交的订单占用的保证金
	MarginUnDealed float64
	// 保证金率
	MarginRate float64
	// 总值
	BalanceTotal float64
	// 净值
	// BalanceNet = BalanceTotal + ProfitUnreal + ProfitReal
	BalanceNet float64
	// 可提取
	// BalanceAvail = BalanceNet - Margin
	BalanceAvail float64
	//已实现盈亏
	ProfitReal float64
	// 未实现盈亏
	ProfitUnreal float64
}

type FutureAccount struct {
	SubAccount map[Currency]FutureSubAccount
	Exchange   string
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
}

type FutureDepth struct {
	ContractType string // for future
	ContractName string // for future
	Pair         Pair
	Timestamp    int64
	DueTimestamp int64
	// The increasing sequence, cause the http return sequence is not sure.
	Sequence int64
	Date     string
	AskList  DepthRecords // Ascending order
	BidList  DepthRecords // Descending order
}

// Do not trust the data from exchange, just verify it.
func (fd FutureDepth) Verify() error {

	AskCount := len(fd.AskList)
	BidCount := len(fd.BidList)

	if BidCount < 10 || AskCount < 10 {
		return errors.New("The ask_list or bid_list not enough! ")
	}

	for i := 1; i < AskCount; i++ {
		pre := fd.AskList[i-1]
		last := fd.AskList[i]
		if pre.Price >= last.Price {
			return errors.New("The ask_list is not ascending ordered! ")
		}
	}

	for i := 1; i < BidCount; i++ {
		pre := fd.BidList[i-1]
		last := fd.BidList[i]
		if pre.Price <= last.Price {
			return errors.New("The bid_list is not descending ordered! ")
		}
	}

	return nil
}

type FutureKline struct {
	Kline        `json:",-"` // 按照kline中的字段进行解析。
	DueTimestamp int64       `json:"due_timestamp"`
	DueDate      string      `json:"due_date"`
	Vol2         float64     `json:"vol_2"` //个数
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
	Amount         int64
	AvgPrice       float64
	DealAmount     int64
	PlaceTimestamp int64
	PlaceDatetime  string
	DealTimestamp  int64 // unit: ms
	DealDatetime   string
	Status         TradeStatus
	PlaceType      PlaceType  // place order type 0：NORMAL 1：MAKER_ONLY 2：FOK 3：IOC
	Type           FutureType // type 1：OPEN_LONG 2：OPEN_SHORT 3：LIQUIDATE_LONG 4： LIQUIDATE_SHORT
	LeverRate      int64
	Fee            float64
	Pair           Pair
	ContractType   string
	ContractName   string // for future
	Exchange       string
}

type FuturePosition struct {
	BuyAmount           float64
	BuyAvailable        float64
	BuyPriceAvg         float64
	BuyPriceCost        float64
	BuyProfitReal       float64
	CreateDate          int64
	LeverRate           int
	SellAmount          float64
	SellAvailable       float64
	SellPriceAvg        float64
	SellPriceCost       float64
	SellProfitReal      float64
	Symbol              Pair //btc_usd:比特币,ltc_usd:莱特币
	ContractType        string
	ContractId          int64
	ForceLiquidatePrice float64 //预估爆仓价
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

type FutureRule struct {
	Rule        `json:",-"` // 按照Rule里面的规则进行。
	ContractVal float64     `json:"contract_val"` //合约一手价格
}

type FutureContract struct {
	Pair         Pair   `json:"-"`
	Symbol       string `json:"symbol"`
	Exchange     string `json:"exchange"`
	ContractType string `json:"contract_type"` // eg: this_week next_week quarter next_quarter
	ContractName string `json:"contract_name"` // eg: BTC-USD-201025
	SettleMode   int64  `json:"settle_mode"`   // 1: BASIS 2: COUNTER

	OpenTimestamp int64  `json:"open_timestamp"`
	OpenDate      string `json:"open_date"`
	DueTimestamp  int64  `json:"due_timestamp"`
	DueDate       string `json:"due_date"`

	UnitAmount      float64 `json:"unit_amount"`
	PricePrecision  int64   `json:"price_precision"`
	AmountPrecision int64   `json:"amount_precision"`

	MaxScalePriceLimit float64 `json:"max_scale_price_limit"`
	MinScalePriceLimit float64 `json:"min_scale_price_limit"`
}

type FutureContracts struct {
	ContractTypeKV map[string]*FutureContract `json:"contract_type_kv"`
	ContractNameKV map[string]*FutureContract `json:"contract_name_kv"`
	DueTimestampKV map[string]*FutureContract `json:"due_timestamp_kv"`

	SyncTime time.Time // sync from remote service time
}
