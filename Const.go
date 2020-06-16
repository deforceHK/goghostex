package goghostex

const (
	GO_BIRTHDAY = "2006-01-02 15:04:05"
)

type TradeSide int64

const (
	BUY = 1 + iota
	SELL
	BUY_MARKET
	SELL_MARKET
)

func (ts TradeSide) String() string {
	switch ts {
	case 1:
		return "buy"
	case 2:
		return "sell"
	case 3:
		return "buy_market"
	case 4:
		return "sell_market"
	default:
		return "unknown"
	}
}

type TradeStatus int64

func (ts TradeStatus) String() string {
	return tradeStatusSymbol[ts]
}

var tradeStatusSymbol = [...]string{"unfinish", "part_finish", "finish", "cancel", "reject", "canceling", "fail"}

const (
	ORDER_UNFINISH TradeStatus = iota
	ORDER_PART_FINISH
	ORDER_FINISH
	ORDER_CANCEL
	ORDER_REJECT
	ORDER_CANCEL_ING
	ORDER_FAIL
)

//k线周期
const (
	KLINE_PERIOD_1MIN = 1 + iota
	KLINE_PERIOD_3MIN
	KLINE_PERIOD_5MIN
	KLINE_PERIOD_15MIN
	KLINE_PERIOD_30MIN
	KLINE_PERIOD_60MIN
	KLINE_PERIOD_1H
	KLINE_PERIOD_2H
	KLINE_PERIOD_4H
	KLINE_PERIOD_6H
	KLINE_PERIOD_8H
	KLINE_PERIOD_12H
	KLINE_PERIOD_1DAY
	KLINE_PERIOD_3DAY
	KLINE_PERIOD_1WEEK
	KLINE_PERIOD_1MONTH
	KLINE_PERIOD_1YEAR
)

const (
	THIS_WEEK_CONTRACT    = "this_week"    //周合约
	NEXT_WEEK_CONTRACT    = "next_week"    //次周合约
	QUARTER_CONTRACT      = "quarter"      //季度合约
	NEXT_QUARTER_CONTRACT = "next_quarter" //次季合约
	SWAP_CONTRACT         = "swap"         //永续合约
)

//exchanges const
const (
	OKEX      = "okex"
	OKEX_SWAP = "okex.com_swap"
	HUOBI     = "huobi.com"
	HUOBI_PRO = "huobi.pro"
	BITFINEX  = "bitfinex"
	BINANCE   = "binance"
	BITMEX    = "bitmex.com"
	HBDM      = "hbdm.com"
	COINBASE  = "coinbase"
	BITSTAMP  = "bitstamp"
)

var orderTypeSymbol = [...]string{"NORMAL", "ONLY_MAKER", "FOK", "IOC"}

type PlaceType int

const (
	NORMAL     PlaceType = iota // normal order, need to cancel (GTC)
	ONLY_MAKER                  // only maker
	FOK                         // fill or kill
	IOC                         // Immediate or Cancel
)

func (ot PlaceType) String() string {
	return orderTypeSymbol[ot]
}

var futureTypeSymbol = [...]string{"", "OPEN_LONG", "OPEN_SHORT", "LIQUIDATE_LONG", "LIQUIDATE_SHORT"}

type FutureType int64

const (
	OPEN_LONG       FutureType = 1 + iota //开多
	OPEN_SHORT                            //开空
	LIQUIDATE_LONG                        //平多
	LIQUIDATE_SHORT                       //平空
)

func (ft FutureType) String() string {
	return futureTypeSymbol[ft]
}

const (
	CROSS    = "cross"
	ISOLATED = "isolated"
)
