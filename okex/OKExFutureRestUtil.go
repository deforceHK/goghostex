package okex

import (
	"time"

	. "github.com/deforceHK/goghostex"
)

/*
	因为okex在返回无论是最近的kline，还是历史kline都有个奇葩的返回逻辑，故这里处理okex这个奇葩逻辑的各种问题。

	K线接替规则如下：

	非次季度合约生成日：老次周轮替为新当周，老当周对接新次周， 季度、次季度不变。
	次季度合约生成日：老次季度轮替为新当季，老当季轮替为新次周 ，老次周轮替为新当周，老当周对接新次季。
*/

var okTimestampFlags = []int64{
	1420185600000,
}

var okDueTimestampBoard = map[string][]int64{
	THIS_WEEK_CONTRACT:    {1420790400000},
	NEXT_WEEK_CONTRACT:    {1421395200000},
	QUARTER_CONTRACT:      {1427443200000},
	NEXT_QUARTER_CONTRACT: {1435305600000},
}

var okNextQuarterListKV = make(map[int64]int64, 0)        //k list_timestamp v due_timestamp
var okNextQuarterListReverseKV = make(map[int64]int64, 0) // k due_timestamp v list_timestamp

func init() {
	// 从2015年开始以后的 每个季度最后一个周五的交割时间。
	fridays := getLastFridayQuarter()
	for i := 2; i < len(fridays); i++ {
		var listTS = fridays[i-2].AddDate(0, 0, -14).Unix() * 1000
		var nextQuarterTS = fridays[i].Unix() * 1000
		okNextQuarterListKV[listTS] = nextQuarterTS
		okNextQuarterListReverseKV[nextQuarterTS] = listTS
	}

	// 计算2500周，大概50年
	for i := int64(1); i < 2500; i++ {
		var ts = okTimestampFlags[0] + i*7*24*60*60*1000
		okTimestampFlags = append(okTimestampFlags, ts)
		if nextQuarterTS, exist := okNextQuarterListKV[ts]; exist {

			okDueTimestampBoard[THIS_WEEK_CONTRACT] = append(
				okDueTimestampBoard[THIS_WEEK_CONTRACT],
				okDueTimestampBoard[NEXT_WEEK_CONTRACT][len(okDueTimestampBoard[NEXT_WEEK_CONTRACT])-1],
			)
			okDueTimestampBoard[NEXT_WEEK_CONTRACT] = append(
				okDueTimestampBoard[NEXT_WEEK_CONTRACT],
				okDueTimestampBoard[QUARTER_CONTRACT][len(okDueTimestampBoard[QUARTER_CONTRACT])-1],
			)
			okDueTimestampBoard[QUARTER_CONTRACT] = append(
				okDueTimestampBoard[QUARTER_CONTRACT],
				okDueTimestampBoard[NEXT_QUARTER_CONTRACT][len(okDueTimestampBoard[NEXT_QUARTER_CONTRACT])-1],
			)
			okDueTimestampBoard[NEXT_QUARTER_CONTRACT] = append(
				okDueTimestampBoard[NEXT_QUARTER_CONTRACT],
				nextQuarterTS,
			)
		} else {
			okDueTimestampBoard[THIS_WEEK_CONTRACT] = append(
				okDueTimestampBoard[THIS_WEEK_CONTRACT],
				okDueTimestampBoard[NEXT_WEEK_CONTRACT][len(okDueTimestampBoard[NEXT_WEEK_CONTRACT])-1],
			)
			okDueTimestampBoard[NEXT_WEEK_CONTRACT] = append(
				okDueTimestampBoard[NEXT_WEEK_CONTRACT],
				okDueTimestampBoard[THIS_WEEK_CONTRACT][len(okDueTimestampBoard[THIS_WEEK_CONTRACT])-1]+7*24*60*60*1000,
			)
			okDueTimestampBoard[QUARTER_CONTRACT] = append(
				okDueTimestampBoard[QUARTER_CONTRACT],
				okDueTimestampBoard[QUARTER_CONTRACT][len(okDueTimestampBoard[QUARTER_CONTRACT])-1],
			)
			okDueTimestampBoard[NEXT_QUARTER_CONTRACT] = append(
				okDueTimestampBoard[NEXT_QUARTER_CONTRACT],
				okDueTimestampBoard[NEXT_QUARTER_CONTRACT][len(okDueTimestampBoard[NEXT_QUARTER_CONTRACT])-1],
			)
		}
	}
}

func getLastFridayQuarter() []time.Time {
	var lastFridayQuarter = make([]time.Time, 0)
	var loc, _ = time.LoadLocation("Asia/Shanghai")
	var months = []time.Month{time.March, time.June, time.September, time.December}
	for year := 2015; year < 2050; year += 1 {

		for _, month := range months {
			var lastFriday = time.Date(
				year, month, 1,
				16, 0, 0, 0, loc,
			)
			// 月末最后一天
			lastFriday = lastFriday.AddDate(0, 1, 0).AddDate(0, 0, -1)
			for lastFriday.Weekday() != time.Friday {
				lastFriday = lastFriday.AddDate(0, 0, -1)
			}
			lastFridayQuarter = append(lastFridayQuarter, lastFriday)
		}
	}
	return lastFridayQuarter
}

func GetRealContractTypeBoard() map[string][]string {
	var nowTS = time.Now().Unix() * 1000
	var flag = int((nowTS - okTimestampFlags[0]) / (7 * 24 * 60 * 60 * 1000))

	var board = map[string][]string{
		THIS_WEEK_CONTRACT:    make([]string, flag+1),
		NEXT_WEEK_CONTRACT:    make([]string, flag+1),
		QUARTER_CONTRACT:      make([]string, flag+1),
		NEXT_QUARTER_CONTRACT: make([]string, flag+1),
	}

	var tmpKV = map[string]string{
		THIS_WEEK_CONTRACT:    THIS_WEEK_CONTRACT,
		NEXT_WEEK_CONTRACT:    NEXT_WEEK_CONTRACT,
		QUARTER_CONTRACT:      QUARTER_CONTRACT,
		NEXT_QUARTER_CONTRACT: NEXT_QUARTER_CONTRACT,
	}

	// 次季合约生成日，规则
	var listKV = map[string]string{
		THIS_WEEK_CONTRACT:    NEXT_WEEK_CONTRACT,
		NEXT_WEEK_CONTRACT:    QUARTER_CONTRACT,
		QUARTER_CONTRACT:      NEXT_QUARTER_CONTRACT,
		NEXT_QUARTER_CONTRACT: THIS_WEEK_CONTRACT,
	}

	// 非次季合约生成日，规则
	var nonListKV = map[string]string{
		THIS_WEEK_CONTRACT:    NEXT_WEEK_CONTRACT,
		NEXT_WEEK_CONTRACT:    THIS_WEEK_CONTRACT,
		QUARTER_CONTRACT:      QUARTER_CONTRACT,
		NEXT_QUARTER_CONTRACT: NEXT_QUARTER_CONTRACT,
	}

	for ; flag >= 0; flag-- {
		board[THIS_WEEK_CONTRACT][flag] = tmpKV[THIS_WEEK_CONTRACT]
		board[NEXT_WEEK_CONTRACT][flag] = tmpKV[NEXT_WEEK_CONTRACT]
		board[QUARTER_CONTRACT][flag] = tmpKV[QUARTER_CONTRACT]
		board[NEXT_QUARTER_CONTRACT][flag] = tmpKV[NEXT_QUARTER_CONTRACT]

		var timestamp = okTimestampFlags[flag]
		if _, exist := okNextQuarterListKV[timestamp]; exist {
			for k, v := range tmpKV {
				tmpKV[k] = listKV[v]
			}
		} else {
			for k, v := range tmpKV {
				tmpKV[k] = nonListKV[v]
			}
		}
	}

	return board
}

func GetDueTimestamp(timestamp int64) (flag int, dueTimestamp map[string]int64) {
	flag = int((timestamp - okTimestampFlags[0]) / (7 * 24 * 60 * 60 * 1000))
	dueTimestamp = map[string]int64{
		THIS_WEEK_CONTRACT:    okDueTimestampBoard[THIS_WEEK_CONTRACT][flag],
		NEXT_WEEK_CONTRACT:    okDueTimestampBoard[NEXT_WEEK_CONTRACT][flag],
		QUARTER_CONTRACT:      okDueTimestampBoard[QUARTER_CONTRACT][flag],
		NEXT_QUARTER_CONTRACT: okDueTimestampBoard[NEXT_QUARTER_CONTRACT][flag],
	}
	return flag, dueTimestamp
}
