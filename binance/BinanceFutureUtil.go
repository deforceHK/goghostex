package binance

import (
	"time"

	. "github.com/strengthening/goghostex"
)

var bnTimestampFlags = []int64{
	1420185600000,
}

var bnDueTimestampBoard = map[string][]int64{
	QUARTER_CONTRACT:      {1427443200000},
	NEXT_QUARTER_CONTRACT: {1435305600000},
}

var bnNextQuarterListKV = make(map[int64]int64, 0)        //k list_timestamp v due_timestamp
var bnNextQuarterListReverseKV = make(map[int64]int64, 0) // k due_timestamp v list_timestamp

func init() {
	// 从2015年开始以后的 每个季度最后一个周五的交割时间。
	var fridays = getLastFridayQuarter()
	for i := 2; i < len(fridays); i++ {
		var listTS = fridays[i-2].Unix() * 1000
		var nextQuarterTS = fridays[i].Unix() * 1000

		bnNextQuarterListKV[listTS] = nextQuarterTS
		bnNextQuarterListReverseKV[nextQuarterTS] = listTS
	}

	// 计算2500周，大概50年
	for i := int64(1); i < 2500; i++ {
		var ts = bnTimestampFlags[0] + i*7*24*60*60*1000
		bnTimestampFlags = append(bnTimestampFlags, ts)
		if nextQuarterTS, exist := bnNextQuarterListKV[ts]; exist {
			// 次季生成日上一个次季变成本季
			bnDueTimestampBoard[QUARTER_CONTRACT] = append(
				bnDueTimestampBoard[QUARTER_CONTRACT],
				bnDueTimestampBoard[NEXT_QUARTER_CONTRACT][len(bnDueTimestampBoard[NEXT_QUARTER_CONTRACT])-1],
			)
			bnDueTimestampBoard[NEXT_QUARTER_CONTRACT] = append(
				bnDueTimestampBoard[NEXT_QUARTER_CONTRACT],
				nextQuarterTS,
			)
		} else {
			bnDueTimestampBoard[QUARTER_CONTRACT] = append(
				bnDueTimestampBoard[QUARTER_CONTRACT],
				bnDueTimestampBoard[QUARTER_CONTRACT][len(bnDueTimestampBoard[QUARTER_CONTRACT])-1],
			)
			bnDueTimestampBoard[NEXT_QUARTER_CONTRACT] = append(
				bnDueTimestampBoard[NEXT_QUARTER_CONTRACT],
				bnDueTimestampBoard[NEXT_QUARTER_CONTRACT][len(bnDueTimestampBoard[NEXT_QUARTER_CONTRACT])-1],
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

func GetDueTimestamp(timestamp int64) (flag int, dueTimestamp map[string]int64) {
	flag = int((timestamp - bnTimestampFlags[0]) / (7 * 24 * 60 * 60 * 1000))
	dueTimestamp = map[string]int64{
		QUARTER_CONTRACT:      bnDueTimestampBoard[QUARTER_CONTRACT][flag],
		NEXT_QUARTER_CONTRACT: bnDueTimestampBoard[NEXT_QUARTER_CONTRACT][flag],
	}
	return flag, dueTimestamp
}
