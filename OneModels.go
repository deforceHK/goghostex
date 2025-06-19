package goghostex

import "errors"

type OneTicker struct {
	Pair       Pair    `json:"-"`
	ProductId  string  `json:"product_id"`
	SettleMode int     `json:"settle_mode"` // 1: BASIS 币本位 2: COUNTER U本位 3: NONE 无
	Last       float64 `json:"last"`
	Buy        float64 `json:"buy"`
	Sell       float64 `json:"sell"`
	High       float64 `json:"high"`
	Low        float64 `json:"low"`
	Vol        float64 `json:"vol"`
	Timestamp  int64   `json:"timestamp"` // unit:ms
	Date       string  `json:"date"`      // date: format yyyy-mm-dd HH:MM:SS, the timezone define in api config
}

type OneDepth struct {
	Pair       Pair
	ProductId  string `json:"product_id"`
	SettleMode int    `json:"settle_mode"` // 1: BASIS 币本位 2: COUNTER U本位 3: NONE 无
	Timestamp  int64
	Sequence   int64 // The increasing sequence, cause the http return sequence is not sure.
	Date       string
	AskList    DepthRecords // Ascending order
	BidList    DepthRecords // Descending order
}

// Verify the depth data is right
func (depth *OneDepth) Verify() error {
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

type OneOrder struct {
	Pair           Pair
	SettleMode     int // 1: BASIS 币本位 2: COUNTER U本位 3: NONE 无
	ProductId      string
	OrderId        string
	Cid            string
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
	Exchange       string
}

type OneInfo struct {
	Pair       Pair   `json:"-"`
	ProductId  string `json:"product_id"`
	Status     string `json:"status"`      // PENDING, TRADING,
	SettleMode int    `json:"settle_mode"` // 1: BASIS 币本位 2: COUNTER U本位 3: NONE 无
	Exchange   string `json:"exchange"`

	ContractValue           float64 `json:"contract_value"`
	ContractType            string  `json:"contract_type"`
	ContractStartTimestamp  int64   `json:"contract_start_timestamp"`
	ContractFinishTimestamp int64   `json:"contract_finish_timestamp"`

	TickSize        float64 `json:"tick_size"`
	PricePrecision  int64   `json:"price_precision"`
	AmountPrecision int64   `json:"amount_precision"`
}
