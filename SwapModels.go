package goghostex

import "errors"

type SwapTicker struct {
	Pair      Pair    `json:"-"`
	Last      float64 `json:"last"`
	Buy       float64 `json:"buy"`
	Sell      float64 `json:"sell"`
	High      float64 `json:"high"`
	Low       float64 `json:"low"`
	Vol       float64 `json:"vol"`
	Timestamp int64   `json:"timestamp"` // unit:ms
	Date      string  `json:"date"`      // date: format yyyy-mm-dd HH:MM:SS, the timezone define in apiconfig
}

type SwapDepth struct {
	Pair      Pair
	Timestamp int64
	Sequence  int64 // The increasing sequence, cause the http return sequence is not sure.
	Date      string
	AskList   DepthRecords // Ascending order
	BidList   DepthRecords // Descending order
}

// Verify the depth data is right
func (depth *SwapDepth) Verify() error {
	AskCount := len(depth.AskList)
	BidCount := len(depth.BidList)

	if BidCount < 10 || AskCount < 10 {
		return errors.New("The ask_list or bid_list not enough! ")
	}

	for i := 1; i < AskCount; i++ {
		pre := depth.AskList[i-1]
		last := depth.AskList[i]
		if pre.Price >= last.Price {
			return errors.New("The ask_list is not ascending ordered! ")
		}
	}

	for i := 1; i < BidCount; i++ {
		pre := depth.BidList[i-1]
		last := depth.BidList[i]
		if pre.Price <= last.Price {
			return errors.New("The bid_list is not descending ordered! ")
		}
	}

	return nil
}

type SwapKline struct {
	Pair      Pair    `json:"-"`
	Exchange  string  `json:"exchange"`
	Timestamp int64   `json:"timestamp"`
	Date      string  `json:"date"`
	Open      float64 `json:"open"`
	Close     float64 `json:"close"`
	High      float64 `json:"high"`
	Low       float64 `json:"low"`
	Vol       float64 `json:"vol"`
}

type SwapOrder struct {
	// cid is important, when the order api return wrong, you can find it in unfinished api
	Cid            string
	OrderId        string
	Price          float64
	Amount         float64
	AvgPrice       float64
	DealAmount     float64
	PlaceTimestamp int64
	PlaceDatetime  string
	DealTimestamp  int64 // unit: ms
	DealDatetime   string
	Status         TradeStatus
	PlaceType      PlaceType  // place_type 0：NORMAL 1：MAKER_ONLY 2：FOK 3：IOC
	Type           FutureType // type 1：OPEN_LONG 2：OPEN_SHORT 3：LIQUIDATE_LONG 4： LIQUIDATE_SHORT
	MarginType     string     // margin_type 全仓：crossed 逐仓：isolated
	LeverRate      int64
	Fee            float64
	Pair           Pair
	Exchange       string
}

type SwapPosition struct {
	Pair           Pair
	Type           FutureType //open_long or open_short
	Amount         float64    // position amount
	Price          float64    // position price
	MarkPrice      float64
	LiquidatePrice float64
	MarginType     string
	MarginAmount   float64
	Leverage       int64
}

type SwapAccount struct {
	Exchange string
	// In swap, the usdt is default.
	Currency Currency
	// The future margin 期货保证金 == marginFilled+ marginUnFilled
	Margin float64
	// The future is filled 已经成交的订单占用的期货保证金
	MarginPosition float64
	// The future is unfilled 未成交的订单占用的保证金
	MarginOpen float64
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

	Positions []*SwapPosition
}

type SwapAccountItem struct {
	Pair     Pair
	Exchange string
	Subject  string

	SettleMode     int64 // 1: basis 2: counter
	SettleCurrency Currency
	Amount         float64
	Timestamp      int64
	DateTime       string
	Info           string
}

type SwapRule struct {
	Rule        Rule    `json:",-"`           // 按照Rule里面的规则进行。
	ContractVal float64 `json:"contract_val"` //合约一手价格
}
