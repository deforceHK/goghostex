package goghostex

import "errors"

type SwapTicker struct {
	//Ticker        `json:",-"` // 按照kline中的字段进行解析。
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

type DepthItem struct {
	Price  float64
	Amount float64
}

type DepthItems []DepthItem

func (dr DepthItems) Len() int {
	return len(dr)
}

func (dr DepthItems) Swap(i, j int) {
	dr[i], dr[j] = dr[j], dr[i]
}

func (dr DepthItems) Less(i, j int) bool {
	return dr[i].Price < dr[j].Price
}

type SwapDepth struct {
	Pair      Pair
	Timestamp int64
	Sequence  int64 // The increasing sequence, cause the http return sequence is not sure.
	Date      string
	AskList   DepthItems // Ascending order
	BidList   DepthItems // Descending order
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

type DepthStdItem struct {
	Price  int64
	Amount float64
}

type DepthStdItems []DepthStdItem

func (dsi DepthStdItems) Len() int {
	return len(dsi)
}

func (dsi DepthStdItems) Swap(i, j int) {
	dsi[i], dsi[j] = dsi[j], dsi[i]
}

func (dsi DepthStdItems) Less(i, j int) bool {
	return dsi[i].Price < dsi[j].Price
}

type SwapStdDepth struct {
	Pair      Pair
	Timestamp int64
	Sequence  int64 // The increasing sequence, cause the http return sequence is not sure.
	Date      string
	AskList   DepthStdItems // Ascending order
	BidList   DepthStdItems // Descending order
}

// Verify the depth data is right
func (depth *SwapStdDepth) Verify() error {
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

type SwapAccount struct {
}

type SwapKline struct {
	//Kline `json:",-"` // 按照kline中的字段进行解析。
	Pair      Pair    `json:"symbol"`
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
	Amount         int64
	AvgPrice       float64
	DealAmount     int64
	PlaceTimestamp int64
	PlaceDatetime  string
	OrderTimestamp int64 // unit: ms
	OrderDate      string
	Status         TradeStatus
	PlaceType      PlaceType  // place order type 0：NORMAL 1：MAKER_ONLY 2：FOK 3：IOC
	Type           FutureType // type 1：OPEN_LONG 2：OPEN_SHORT 3：LIQUIDATE_LONG 4： LIQUIDATE_SHORT
	LeverRate      int64
	Fee            float64
	Currency       CurrencyPair
	Exchange       string
	MatchPrice     int64 // some exchange need
}

type SwapPosition struct {
}
