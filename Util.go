package goghostex

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

func ToFloat64(v interface{}) float64 {
	if v == nil {
		return 0.0
	}

	switch v.(type) {
	case float64:
		return v.(float64)
	case string:
		vStr := v.(string)
		vF, _ := strconv.ParseFloat(vStr, 64)
		return vF
	default:
		panic("to float64 error.")
	}
}

func ToInt(v interface{}) int {
	if v == nil {
		return 0
	}

	switch v.(type) {
	case string:
		vStr := v.(string)
		vInt, _ := strconv.Atoi(vStr)
		return vInt
	case int:
		return v.(int)
	case float64:
		vF := v.(float64)
		return int(vF)
	default:
		panic("to int error.")
	}
}

func ToUint64(v interface{}) uint64 {
	if v == nil {
		return 0
	}

	switch v.(type) {
	case int:
		return uint64(v.(int))
	case float64:
		return uint64((v.(float64)))
	case string:
		uV, _ := strconv.ParseUint(v.(string), 10, 64)
		return uV
	default:
		panic("to uint64 error.")
	}
}

func ToInt64(v interface{}) int64 {
	if v == nil {
		return 0
	}

	switch v.(type) {
	case float64:
		return int64(v.(float64))
	default:
		vv := fmt.Sprint(v)

		if vv == "" {
			return 0
		}

		vvv, err := strconv.ParseInt(vv, 0, 64)
		if err != nil {
			return 0
		}

		return vvv
	}
}

// FloatToString n :保留的小数点位数,去除末尾多余的0(StripTrailingZeros)
func FloatToString(v float64, n int64) string {
	theN := int(n)
	ret := strconv.FormatFloat(v, 'f', theN, 64)
	return strconv.FormatFloat(ToFloat64(ret), 'f', -1, 64) //StripTrailingZeros
}

// FloatToPrice n :保留的小数点位数,去除末尾多余的0(StripTrailingZeros)，并加入ticksize
func FloatToPrice(v float64, n int64, tickSize float64) string {
	if tickSize <= 0 {
		return FloatToString(v, n)
	}
	var intPart = float64(int64(v))
	var floatPart = float64(int64((v-intPart)/tickSize)) * tickSize
	var ret = strconv.FormatFloat(intPart+floatPart, 'f', int(n), 64)
	return strconv.FormatFloat(ToFloat64(ret), 'f', -1, 64) //StripTrailingZeros
}

func ValuesToJson(v url.Values) ([]byte, error) {
	parammap := make(map[string]interface{})
	for k, vv := range v {
		if len(vv) == 1 {
			parammap[k] = vv[0]
		} else {
			parammap[k] = vv
		}
	}
	return json.Marshal(parammap)
}

func GzipUnCompress(data []byte) ([]byte, error) {
	r, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	return ioutil.ReadAll(r)
}

func FlateUnCompress(data []byte) ([]byte, error) {
	reader := flate.NewReader(bytes.NewReader(data))
	defer func() { _ = reader.Close() }()

	return ioutil.ReadAll(reader)
}

func UUID() string {
	return strings.Replace(uuid.New().String(), "-", "", 32)
}

func GetPrecision(minSize float64) int {
	if minSize < 0.0000000001 {
		return 10
	}

	for i := 0; ; i++ {
		if minSize >= 1 {
			return i
		}
		minSize *= 10
	}
}

func GetPrecisionInt64(minSize float64) int64 {

	if minSize < 0.0000000001 {
		return 10
	}

	for i := int64(0); ; i++ {
		if minSize >= 1 {
			return i
		}
		minSize *= 10
	}
}

func GetAscKline(klines []*Kline) []*Kline {
	if len(klines) <= 1 {
		return klines
	}

	// asc seq
	if klines[0].Timestamp < klines[1].Timestamp {
		return klines
	}

	ascKlines := make([]*Kline, 0)
	for i := len(klines) - 1; i >= 0; i-- {
		ascKlines = append(ascKlines, klines[i])
	}
	return ascKlines
}

func GetAscFutureKline(klines []*FutureKline) []*FutureKline {
	if len(klines) <= 1 {
		return klines
	}

	// asc seq
	if klines[0].Timestamp < klines[1].Timestamp {
		return klines
	}

	ascKlines := make([]*FutureKline, 0)
	for i := len(klines) - 1; i >= 0; i-- {
		ascKlines = append(ascKlines, klines[i])
	}
	return ascKlines
}

func GetAscFutureCandle(candles []*FutureCandle) []*FutureCandle {
	if len(candles) <= 1 {
		return candles
	}

	// asc seq
	if candles[0].Timestamp < candles[1].Timestamp {
		return candles
	}

	ascCandles := make([]*FutureCandle, 0)
	for i := len(candles) - 1; i >= 0; i-- {
		ascCandles = append(ascCandles, candles[i])
	}
	return ascCandles
}

func GetAscSwapKline(klines []*SwapKline) []*SwapKline {
	if len(klines) <= 1 {
		return klines
	}

	// asc seq
	if klines[0].Timestamp < klines[1].Timestamp {
		return klines
	}

	ascKlines := make([]*SwapKline, 0)
	for i := len(klines) - 1; i >= 0; i-- {
		ascKlines = append(ascKlines, klines[i])
	}
	return ascKlines
}
