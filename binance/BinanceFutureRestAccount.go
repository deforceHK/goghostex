package binance

import (
	"net/http"
	"net/url"
	"strings"
	"time"

	. "github.com/deforceHK/goghostex"
)

func (future *Future) GetPairFlow(pair Pair) ([]*FutureAccountItem, []byte, error) {

	var params = url.Values{}
	if err := future.buildParamsSigned(&params); err != nil {
		return nil, nil, err
	}

	var responses = make([]*struct {
		Symbol     string  `json:"symbol"`
		IncomeType string  `json:"incomeType"`
		Income     float64 `json:"income,string"`
		Asset      string  `json:"asset"`
		Info       string  `json:"info"`
		Time       int64   `json:"time"`
	}, 0)

	var resp, err = future.DoRequest(
		http.MethodGet,
		FUTURE_CM_ENDPOINT,
		FUTURE_INCOME_URI+params.Encode(),
		"",
		&responses,
	)
	if err != nil {
		return nil, resp, err
	}

	var items = make([]*FutureAccountItem, 0)
	for i := len(responses) - 1; i >= 0; i-- {
		var r = responses[i]
		if r.Symbol == "" || strings.Index(r.Symbol, "_PERP") > 0 {
			continue
		}

		// 不是这个pair的滤掉
		var symbolFilter = pair.ToSymbol("", true) + "_"
		if strings.Index(r.Symbol, symbolFilter) < 0 {
			continue
		}
		var dateTime = time.Unix(r.Time/1000, 0).In(future.config.Location).Format(GO_BIRTHDAY)
		var fai = &FutureAccountItem{
			Pair:           pair,
			Exchange:       BINANCE,
			ContractName:   pair.ToSymbol("-", true) + "-" + strings.Split(r.Symbol, "_")[1],
			Subject:        future.transferSubject(r.Income, r.IncomeType),
			SettleMode:     SETTLE_MODE_BASIS,
			SettleCurrency: NewCurrency(r.Asset, ""),
			Amount:         r.Income,
			Timestamp:      r.Time,
			DateTime:       dateTime,
			Info:           r.Info,
		}
		items = append(items, fai)
	}

	return items, resp, nil
}
