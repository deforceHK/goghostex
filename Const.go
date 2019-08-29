package goghostex

const (
	GO_BIRTHDAY = "2006-01-02 15:04:05"
)

type TradeSide int

const (
	BUY = 1 + iota
	SELL
	BUY_MARKET
	SELL_MARKET
)

func (ts TradeSide) String() string {
	switch ts {
	case 1:
		return "BUY"
	case 2:
		return "SELL"
	case 3:
		return "BUY_MARKET"
	case 4:
		return "SELL_MARKET"
	default:
		return "UNKNOWN"
	}
}

type TradeStatus int

func (ts TradeStatus) String() string {
	return tradeStatusSymbol[ts]
}

var tradeStatusSymbol = [...]string{"UNFINISH", "PART_FINISH", "FINISH", "CANCEL", "REJECT", "CANCEL_ING", "FAIL"}

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
	THIS_WEEK_CONTRACT = "this_week" //周合约
	NEXT_WEEK_CONTRACT = "next_week" //次周合约
	QUARTER_CONTRACT   = "quarter"   //季度合约
	SWAP_CONTRACT      = "swap"      //永续合约
)

//exchanges const
const (
	OKCOIN_CN   = "okcoin.cn"
	OKCOIN_COM  = "okcoin.com"
	OKEX        = "okex.com"
	OKEX_FUTURE = "okex.com_future"
	OKEX_SWAP   = "okex.com_swap"
	HUOBI       = "huobi.com"
	HUOBI_PRO   = "huobi.pro"
	BITFINEX    = "bitfinex.com"
	BINANCE     = "binance.com"
	BITMEX      = "bitmex.com"
	HBDM        = "hbdm.com"
)

var orderTypeSymbol = [...]string{"NORMAL", "ONLY_MAKER", "FOK", "IOC"}

type OrderType int

const (
	NORMAL     OrderType = iota // normal order, need to cancel (GTC)
	ONLY_MAKER                  // only maker
	FOK                         // fill or kill
	IOC                         // Immediate or Cancel
)

func (ot OrderType) String() string {
	return orderTypeSymbol[ot]
}

var futureTypeSymbol = [...]string{"", "OPEN_LONG", "OPEN_SHORT", "LIQUIDATE_LONG", "LIQUIDATE_SHORT"}

type FutureType int

const (
	OPEN_LONG       FutureType = 1 + iota //开多
	OPEN_SHORT                            //开空
	LIQUIDATE_LONG                        //平多
	LIQUIDATE_SHORT                       //平空
)

func (ft FutureType) String() string {
	return futureTypeSymbol[ft]
}
