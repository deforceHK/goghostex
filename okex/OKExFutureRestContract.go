package okex

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	. "github.com/deforceHK/goghostex"
)

func (future *Future) GetContracts() ([]*FutureContract, []byte, error) {
	var contracts = make([]*FutureContract, 0)
	var response struct {
		Code string `json:"code"`
		Msg  string `json:"msg"`
		Data []struct {
			Alias     string  `json:"alias"`
			CtVal     float64 `json:"ctVal,string"`
			CtValCcy  string  `json:"ctValCcy"`
			ExpTime   int64   `json:"expTime,string"`
			InstId    string  `json:"instId"`
			ListTime  int64   `json:"listTime,string"`
			SettleCcy string  `json:"settleCcy"`
			TickSz    float64 `json:"tickSz,string"`
			LotSz     float64 `json:"lotSz,string"`
			Uly       string  `json:"uly"`
			State     string  `json:"state"`
			CtType    string  `json:"ctType"`
		} `json:"data"`
	}
	resp, err := future.DoRequest(
		http.MethodGet,
		"/api/v5/public/instruments?instType=FUTURES",
		"",
		&response,
	)

	if err != nil {
		return nil, resp, err
	}
	if response.Code != "0" {
		return nil, resp, errors.New(response.Msg)
	}
	if len(response.Data) == 0 {
		return nil, resp, errors.New("The contract api not ready. ")
	}

	for _, item := range response.Data {

		var dueTimestamp = item.ExpTime
		var dueTime = time.Unix(dueTimestamp/1000, 0).In(future.config.Location)
		var openTimestamp = item.ListTime
		var openTime = time.Unix(openTimestamp/1000, 0).In(future.config.Location)
		var listTimestamp = item.ListTime
		var listTime = time.Unix(listTimestamp/1000, 0).In(future.config.Location)

		var pair = NewPair(item.Uly, "-")
		var settleMode = SETTLE_MODE_BASIS
		if item.SettleCcy != strings.Split(item.Uly, "-")[0] {
			settleMode = SETTLE_MODE_COUNTER
		}
		var rawData, _ = json.Marshal(item)

		var contract = &FutureContract{
			Pair:         pair,
			Symbol:       pair.ToSymbol("_", false),
			Exchange:     OKEX,
			ContractType: item.Alias,
			ContractName: item.InstId,
			SettleMode:   settleMode,
			Status:       item.State,
			Type:         item.CtType,

			OpenTimestamp: openTime.UnixNano() / int64(time.Millisecond),
			OpenDate:      openTime.Format(GO_BIRTHDAY),

			ListTimestamp: listTimestamp,
			ListDate:      listTime.Format(GO_BIRTHDAY),

			DueTimestamp: dueTime.UnixNano() / int64(time.Millisecond),
			DueDate:      dueTime.Format(GO_BIRTHDAY),

			UnitAmount:      item.CtVal,
			TickSize:        ToFloat64(item.TickSz),
			PricePrecision:  GetPrecisionInt64(item.TickSz),
			AmountPrecision: GetPrecisionInt64(item.LotSz),
			RawData:         string(rawData),
		}

		contracts = append(contracts, contract)
	}
	return contracts, resp, nil
}
