package goghostex

import (
	"errors"
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
	ContractType string // for future
	ContractName string // for future
	Pair         CurrencyPair
	Timestamp    int64
	// The increasing sequence, cause the http return sequence is not sure.
	Sequence int64
	Date     string
	AskList  FutureDepthRecords // Ascending order
	BidList  FutureDepthRecords // Descending order
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

type FutureStdDepthRecords []FutureStdDepthRecord

type FutureStdDepthRecord struct {
	Price  int64
	Amount int64
}

func (dr FutureStdDepthRecords) Len() int {
	return len(dr)
}

func (dr FutureStdDepthRecords) Swap(i, j int) {
	dr[i], dr[j] = dr[j], dr[i]
}

func (dr FutureStdDepthRecords) Less(i, j int) bool {
	return dr[i].Price < dr[j].Price
}

type FutureStdDepth struct {
	ContractType string // for future
	ContractName string // for future
	Pair         CurrencyPair
	Timestamp    int64
	// The increasing sequence, cause the http return sequence is not sure.
	Sequence int64
	Date     string
	AskList  FutureStdDepthRecords // Ascending order
	BidList  FutureStdDepthRecords // Descending order
}

// Do not trust the data from exchange, just verify it.
func (fd FutureStdDepth) Verify() error {

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
	Kline
	DueTimestamp int64
	DueDate      string
	Vol2         float64 //个数
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
	OrderTimestamp int64 // unit: ms
	OrderDate      string
	Status         TradeStatus
	PlaceType      PlaceType  // place order type 0：NORMAL 1：MAKER_ONLY 2：FOK 3：IOC
	Type           FutureType // type 1：OPEN_LONG 2：OPEN_SHORT 3：LIQUIDATE_LONG 4： LIQUIDATE_SHORT
	LeverRate      int
	Fee            float64
	Currency       CurrencyPair
	ContractType   string
	ContractName   string // for future
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
