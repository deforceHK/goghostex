package okex

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	. "github.com/deforceHK/goghostex"
)

const CONTRACT_STATUS_LIVE = "live"

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

// 通过contractType获取合约信息
func (future *Future) getFutureContract(pair Pair, contractType string) (*FutureContract, error) {
	future.Locker.Lock()
	defer future.Locker.Unlock()

	var now = time.Now().In(future.config.Location)
	if now.After(future.nextUpdateContractTime) {
		_, err := future.updateFutureContracts()
		if err != nil {
			return nil, err
		}
	}

	var currencies = strings.Split(pair.ToSymbol("_", false), "_")
	var contractTypeItem = fmt.Sprintf("%s,%s,%s", currencies[0], currencies[1], contractType)
	if cf, exist := future.Contracts.ContractTypeKV[contractTypeItem]; !exist {
		return nil, errors.New(fmt.Sprintf("Can not find the contract by contract_type %s. ", contractType))
	} else {
		return cf, nil
	}
}

// 通过dueTimestamp获取合约信息
func (future *Future) getContractByDueTimestamp(symbol string, dueTimestamp int64) (*FutureContract, error) {
	future.Locker.Lock()
	defer future.Locker.Unlock()

	var now = time.Now().In(future.config.Location)
	if now.After(future.nextUpdateContractTime) {
		_, err := future.updateFutureContracts()
		if err != nil {
			return nil, err
		}
	}
	var currencies = strings.Split(symbol, "_")
	var dueTimestampKey = fmt.Sprintf("%s,%s,%d", currencies[0], currencies[1], dueTimestamp)
	if contract, exist := future.Contracts.DueTimestampKV[dueTimestampKey]; !exist {
		return nil, errors.New(fmt.Sprintf("Can not find the contract by due_timestamp %d. ", dueTimestamp))
	} else {
		return contract, nil
	}
}

func (future *Future) updateFutureContracts() ([]byte, error) {
	var now = time.Now().In(future.config.Location)
	var contracts, resp, err = future.GetContracts()
	if err != nil {
		future.nextUpdateContractTime = now.Add(10 * time.Minute)
		return resp, err
	}

	var futureContracts = FutureContracts{
		ContractTypeKV: make(map[string]*FutureContract, 0),
		ContractNameKV: make(map[string]*FutureContract, 0),
		DueTimestampKV: make(map[string]*FutureContract, 0),
	}
	var isFreshNext10min = false
	for _, contract := range contracts {
		// need update at next 10 minutes
		if contract.Status != CONTRACT_STATUS_LIVE && !isFreshNext10min {
			isFreshNext10min = true
		}

		var currencies = strings.Split(contract.Symbol, "_")
		var contractTypeItem = fmt.Sprintf(
			"%s,%s,%s",
			currencies[0],
			currencies[1],
			contract.ContractType,
		)
		var contractNameItem = fmt.Sprintf(
			"%s,%s,%s",
			currencies[0],
			currencies[1],
			contract.ContractName,
		)
		var dueTimestampItem = fmt.Sprintf(
			"%s,%s,%d",
			currencies[0],
			currencies[1],
			contract.DueTimestamp,
		)

		futureContracts.ContractTypeKV[contractTypeItem] = contract
		futureContracts.ContractNameKV[contractNameItem] = contract
		futureContracts.DueTimestampKV[dueTimestampItem] = contract
	}
	future.Contracts = futureContracts

	if isFreshNext10min {
		future.nextUpdateContractTime = now.Add(10 * time.Minute)
	} else {
		var nextHour = time.Date(
			now.Year(), now.Month(), now.Day(), now.Hour(),
			0, 0, 0, future.config.Location,
		).Add(time.Hour)
		future.nextUpdateContractTime = nextHour
	}
	return resp, nil
}

//func (future *Future) updateFutureContracts() ([]byte, error) {
//	var response struct {
//		Code string `json:"code"`
//		Msg  string `json:"msg"`
//		Data []struct {
//			Alias    string  `json:"alias"`
//			CtVal    float64 `json:"ctVal,string"`
//			CtValCcy string  `json:"ctValCcy"`
//			//ExpTime  int64   `json:"expTime,string"`
//			InstId string `json:"instId"`
//			//ListTime  int64   `json:"listTime,string"`
//			SettleCcy string  `json:"settleCcy"`
//			TickSz    float64 `json:"tickSz,string"`
//			LotSz     float64 `json:"lotSz,string"`
//			Uly       string  `json:"uly"`
//			State     string  `json:"state"`
//		} `json:"data"`
//	}
//	resp, err := future.DoRequest(
//		http.MethodGet,
//		"/api/v5/public/instruments?instType=FUTURES",
//		"",
//		&response,
//	)
//
//	if err != nil {
//		return nil, err
//	}
//	if response.Code != "0" {
//		return nil, errors.New(response.Msg)
//	}
//	if len(response.Data) == 0 {
//		return nil, errors.New("The contract api not ready. ")
//	}
//
//	var nowTime = time.Now()
//	futureContracts := FutureContracts{
//		ContractTypeKV: make(map[string]*FutureContract, 0),
//		ContractNameKV: make(map[string]*FutureContract, 0),
//		DueTimestampKV: make(map[string]*FutureContract, 0),
//	}
//
//	var flag = (nowTime.Unix()*1000 - okTimestampFlags[0]) / (7 * 24 * 60 * 60 * 1000)
//	var isFreshNext10min = false
//
//	for _, item := range response.Data {
//		// 只要有合约状态不是live，那就是十分钟后更新
//		if isFreshNext10min == false && item.State != "live" {
//			isFreshNext10min = true
//		}
//
//		var contractType = item.Alias
//		// todo 加入本月次月合约情况
//		if contractType == "this_month" || contractType == "next_month" {
//			continue
//		}
//
//		var dueTimestamp = okDueTimestampBoard[contractType][flag]
//		var dueTime = time.Unix(dueTimestamp/1000, 0).In(future.config.Location)
//		var openTimestamp int64
//		if tmpTS, exist := okNextQuarterListReverseKV[dueTimestamp]; exist {
//			openTimestamp = tmpTS
//		} else {
//			openTimestamp = dueTimestamp - 14*24*60*60*1000
//		}
//		var openTime = time.Unix(openTimestamp/1000, 0).In(future.config.Location)
//
//		pair := NewPair(item.Uly, "-")
//		settleMode := SETTLE_MODE_BASIS
//		if item.SettleCcy != strings.Split(item.Uly, "-")[0] {
//			settleMode = SETTLE_MODE_COUNTER
//		}
//
//		var contract = &FutureContract{
//			Pair:         pair,
//			Symbol:       pair.ToSymbol("_", false),
//			Exchange:     OKEX,
//			ContractType: contractType,
//			ContractName: item.Uly + "-" + dueTime.Format("060102"),
//			SettleMode:   settleMode,
//
//			OpenTimestamp: openTime.UnixNano() / int64(time.Millisecond),
//			OpenDate:      openTime.Format(GO_BIRTHDAY),
//
//			DueTimestamp: dueTime.UnixNano() / int64(time.Millisecond),
//			DueDate:      dueTime.Format(GO_BIRTHDAY),
//
//			UnitAmount:      item.CtVal,
//			TickSize:        ToFloat64(item.TickSz),
//			PricePrecision:  GetPrecisionInt64(item.TickSz),
//			AmountPrecision: GetPrecisionInt64(item.LotSz),
//		}
//
//		currencies := strings.Split(contract.Symbol, "_")
//		contractTypeItem := fmt.Sprintf("%s,%s,%s", currencies[0], currencies[1], contract.ContractType)
//		contractNameItem := fmt.Sprintf("%s,%s,%s", currencies[0], currencies[1], contract.ContractName)
//		dueTimestampItem := fmt.Sprintf("%s,%s,%d", currencies[0], currencies[1], contract.DueTimestamp)
//		futureContracts.ContractTypeKV[contractTypeItem] = contract
//		futureContracts.ContractNameKV[contractNameItem] = contract
//		futureContracts.DueTimestampKV[dueTimestampItem] = contract
//	}
//	future.Contracts = futureContracts
//
//	var nextUpdateTime = time.Unix(okTimestampFlags[flag+1]/1000, 0).In(future.config.Location)
//	if isFreshNext10min || futureContracts.ContractTypeKV["btc,usd,this_week"] == nil {
//		nextUpdateTime = nowTime.Add(10 * time.Minute)
//	}
//	future.nextUpdateContractTime = nextUpdateTime
//	return resp, nil
//}

func (future *Future) GetExchangeName() string {
	return OKEX
}

// 获取instrument_id
func (future *Future) GetInstrumentId(pair Pair, contractAlias string) string {
	if contractAlias != NEXT_QUARTER_CONTRACT &&
		contractAlias != QUARTER_CONTRACT &&
		contractAlias != NEXT_WEEK_CONTRACT &&
		contractAlias != THIS_WEEK_CONTRACT {
		return contractAlias
	}
	fc, _ := future.getFutureContract(pair, contractAlias)
	return fc.ContractName
}

func (future *Future) GetContract(pair Pair, contractType string) (*FutureContract, error) {
	return future.getFutureContract(pair, contractType)
}
