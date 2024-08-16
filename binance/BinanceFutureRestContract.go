package binance

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	. "github.com/deforceHK/goghostex"
)

var __CONTRACT_STATUS_TRANS = map[string]string{
	"PENDING_TRADING": CONTRACT_STATUS_PREPARE,
	"TRADING":         CONTRACT_STATUS_LIVE,
	"PRE_DELIVERING":  CONTRACT_STATUS_SUSPEND,
	"DELIVERING":      CONTRACT_STATUS_SUSPEND,
	"DELIVERED":       CONTRACT_STATUS_CLOSE,
	"PRE_SETTLE":      CONTRACT_STATUS_CLOSE,
	"SETTLING":        CONTRACT_STATUS_CLOSE,
	"SETTLED":         CONTRACT_STATUS_CLOSE,
}

var __CONTRACT_TYPE_TRANS = map[string]string{
	"CURRENT_QUARTER": QUARTER_CONTRACT,
	"NEXT_QUARTER":    NEXT_QUARTER_CONTRACT,
}

var __CONTRACT_TYPE_REVERSE = map[string]string{
	QUARTER_CONTRACT:      "CURRENT_QUARTER",
	NEXT_QUARTER_CONTRACT: "NEXT_QUARTER",
}

// get the future contract info.
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
	var contractTypeItem = fmt.Sprintf(
		"%s,%s,%s",
		currencies[0],
		currencies[1],
		contractType,
	)
	var cf, exist = future.Contracts.ContractTypeKV[contractTypeItem]
	if !exist {
		return nil, errors.New(fmt.Sprintf("Can not find the contract by contract_type %s. ", contractType))
	}
	return cf, nil
}

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
	var dueTimestampItem = fmt.Sprintf(
		"%s,%s,%d",
		currencies[0],
		currencies[1],
		dueTimestamp,
	)

	var contract, exist = future.Contracts.DueTimestampKV[dueTimestampItem]
	if !exist {
		return nil, errors.New(fmt.Sprintf(
			"Can not find the contract by dueTimestamp %d. ", dueTimestamp,
		))
	}
	return contract, nil
}

// update the future contracts info.
//func (future *Future) updateFutureContracts() ([]byte, error) {
//
//	var response = struct {
//		Symbols []struct {
//			Symbol      string `json:"symbol"`
//			Pair        string `json:"pair"`
//			BaseAsset   string `json:"baseAsset"`
//			QuoteAsset  string `json:"quoteAsset"`
//			MarginAsset string `json:"marginAsset"`
//
//			ContractType      string  `json:"contractType"`
//			DeliveryDate      int64   `json:"deliveryDate"`
//			OnboardDate       int64   `json:"onboardDate"`
//			ContractStatus    string  `json:"contractStatus"`
//			ContractSize      float64 `json:"contractSize"`
//			PricePrecision    int64   `json:"pricePrecision"`
//			QuantityPrecision int64   `json:"quantityPrecision"`
//
//			Filters []map[string]interface{} `json:"filters"`
//		} `json:"symbols"`
//		ServerTime int64 `json:"serverTime"`
//	}{}
//
//	var resp, err = future.DoRequest(
//		http.MethodGet,
//		FUTURE_CM_ENDPOINT,
//		FUTURE_EXCHANGE_INFO_URI,
//		"",
//		&response,
//	)
//	if err != nil {
//		return nil, err
//	}
//
//	var contracts = FutureContracts{
//		ContractTypeKV: make(map[string]*FutureContract, 0),
//		ContractNameKV: make(map[string]*FutureContract, 0),
//		DueTimestampKV: make(map[string]*FutureContract, 0),
//	}
//
//	for _, item := range response.Symbols {
//		// it is not future , it's swap in this project.
//		if item.ContractType == "PERPETUAL" {
//			continue
//		}
//
//		if item.ContractType != "CURRENT_QUARTER" && item.ContractType != "NEXT_QUARTER" {
//			continue
//		}
//
//		var contractType = ""
//		if item.ContractType == "CURRENT_QUARTER" {
//			contractType = QUARTER_CONTRACT
//		} else if item.ContractType == "NEXT_QUARTER" {
//			contractType = NEXT_QUARTER_CONTRACT
//		} else {
//			continue
//		}
//
//		settleMode := SETTLE_MODE_BASIS
//		if item.MarginAsset == item.QuoteAsset {
//			settleMode = SETTLE_MODE_COUNTER
//		}
//
//		var priceMaxScale, priceMinScale = float64(1.2), float64(0.8)
//		var tickSize = float64(-1)
//		for _, filter := range item.Filters {
//			if value, ok := filter["filterType"].(string); ok && value == "PERCENT_PRICE" {
//				priceMaxScale = ToFloat64(filter["multiplierUp"])
//				priceMinScale = ToFloat64(filter["multiplierDown"])
//			}
//
//			if value, ok := filter["filterType"].(string); ok && value == "PRICE_FILTER" {
//				tickSize = ToFloat64(filter["tickSize"])
//			}
//		}
//
//		dueTime := time.Unix(item.DeliveryDate/1000, 0).In(future.config.Location)
//		openTime := time.Unix(item.OnboardDate/1000, 0).In(future.config.Location)
//
//		pair := Pair{
//			Basis:   NewCurrency(item.BaseAsset, ""),
//			Counter: NewCurrency(item.QuoteAsset, ""),
//		}
//
//		var contractNameInfo = strings.Split(item.Symbol, "_")
//		contract := &FutureContract{
//			Pair:         pair,
//			Symbol:       pair.ToSymbol("_", false),
//			Exchange:     BINANCE,
//			ContractType: contractType,
//			ContractName: pair.ToSymbol("-", true) + "-" + contractNameInfo[1],
//
//			SettleMode:    settleMode,
//			OpenTimestamp: openTime.UnixNano() / int64(time.Millisecond),
//			OpenDate:      openTime.Format(GO_BIRTHDAY),
//			DueTimestamp:  dueTime.UnixNano() / int64(time.Millisecond),
//			DueDate:       dueTime.Format(GO_BIRTHDAY),
//
//			UnitAmount:      item.ContractSize,
//			TickSize:        tickSize,
//			PricePrecision:  item.PricePrecision,
//			AmountPrecision: item.QuantityPrecision,
//
//			MaxScalePriceLimit: priceMaxScale,
//			MinScalePriceLimit: priceMinScale,
//		}
//
//		currencies := strings.Split(contract.Symbol, "_")
//		contractTypeItem := fmt.Sprintf("%s,%s,%s", currencies[0], currencies[1], contractType)
//		contractNameItem := fmt.Sprintf("%s,%s,%s", currencies[0], currencies[1], contract.ContractName)
//		dueTimestampItem := fmt.Sprintf("%s,%s,%d", currencies[0], currencies[1], contract.DueTimestamp)
//
//		contracts.ContractTypeKV[contractTypeItem] = contract
//		contracts.ContractNameKV[contractNameItem] = contract
//		contracts.DueTimestampKV[dueTimestampItem] = contract
//	}
//
//	future.Contracts = contracts
//	// setting next update time.
//	var nowTime = time.Now().In(future.config.Location)
//	var nextUpdateTime = time.Date(
//		nowTime.Year(), nowTime.Month(), nowTime.Day(),
//		16, 0, 0, 0, future.config.Location,
//	)
//	if nowTime.Hour() >= 16 {
//		nextUpdateTime = nextUpdateTime.AddDate(0, 0, 1)
//	}
//	future.nextUpdateContractTime = nextUpdateTime
//	return resp, nil
//}

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

	// setting next hour update contract.
	var nextUpdateTime = time.Date(
		now.Year(), now.Month(), now.Day(), now.Hour(), 0, 0, 0, future.config.Location,
	).Add(time.Hour)
	for _, contract := range contracts {
		// 如果有一个合约的状态不是live，那么十分钟后再更新，此外每个整点更新一次
		if contract.Status != CONTRACT_STATUS_LIVE {
			nextUpdateTime = now.Add(10 * time.Minute)
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
	future.nextUpdateContractTime = nextUpdateTime
	return resp, nil
}

func (future *Future) GetContract(pair Pair, contractType string) (*FutureContract, error) {
	return future.getFutureContract(pair, contractType)
}

func (future *Future) GetContracts() ([]*FutureContract, []byte, error) {
	var contracts = make([]*FutureContract, 0)
	var cmContracts, umContracts []*FutureContract
	var cmResp, umResp []byte
	var cmErr, umErr error

	var wg = sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()
		cmContracts, cmResp, cmErr = future.getCMContracts()
	}()

	go func() {
		defer wg.Done()
		umContracts, umResp, umErr = future.getUMContracts()
	}()

	wg.Wait()

	if cmErr != nil {
		return nil, cmResp, cmErr
	}
	if umErr != nil {
		return nil, umResp, umErr
	}

	contracts = append(contracts, cmContracts...)
	contracts = append(contracts, umContracts...)

	return contracts, []byte(fmt.Sprintf("[%s,%s]", string(cmResp), string(umResp))), nil
}

func (future *Future) getCMContracts() ([]*FutureContract, []byte, error) {

	var contracts = make([]*FutureContract, 0)
	var nowTimestamp = time.Now().UnixNano() / int64(time.Millisecond)
	var respCm = struct {
		Symbols []*struct {
			Symbol      string `json:"symbol"`
			Pair        string `json:"pair"`
			BaseAsset   string `json:"baseAsset"`
			QuoteAsset  string `json:"quoteAsset"`
			MarginAsset string `json:"marginAsset"`

			ContractType      string  `json:"contractType"`
			DeliveryDate      int64   `json:"deliveryDate"`
			OnboardDate       int64   `json:"onboardDate"`
			ContractStatus    string  `json:"contractStatus"`
			ContractSize      float64 `json:"contractSize"`
			PricePrecision    int64   `json:"pricePrecision"`
			QuantityPrecision int64   `json:"quantityPrecision"`

			Filters []map[string]interface{} `json:"filters"`
		} `json:"symbols"`
		ServerTime int64 `json:"serverTime"`
	}{}

	var resp, errCm = future.DoRequest(
		http.MethodGet,
		FUTURE_CM_ENDPOINT,
		FUTURE_EXCHANGE_INFO_URI,
		"",
		&respCm,
	)
	if errCm != nil {
		return nil, resp, errCm
	}
	for _, item := range respCm.Symbols {
		// it is not future , it's swap in this project.
		if strings.Contains(item.ContractType, "PERPETUAL") ||
			item.DeliveryDate > (nowTimestamp+5*365*24*60*60*1000) {
			continue
		}

		var rawData, _ = json.Marshal(item)
		var priceMaxScale, priceMinScale float64 = 1.2, 0.8
		var tickSize float64 = -1
		for _, filter := range item.Filters {
			if value, ok := filter["filterType"].(string); ok && value == "PERCENT_PRICE" {
				priceMaxScale = ToFloat64(filter["multiplierUp"])
				priceMinScale = ToFloat64(filter["multiplierDown"])
			}

			if value, ok := filter["filterType"].(string); ok && value == "PRICE_FILTER" {
				tickSize = ToFloat64(filter["tickSize"])
			}
		}

		var dueTime = time.Unix(item.DeliveryDate/1000, 0).In(future.config.Location)
		var openTime = time.Unix(item.OnboardDate/1000, 0).In(future.config.Location)
		var listTime = time.Unix(item.OnboardDate/1000, 0).In(future.config.Location)

		var pair = Pair{
			Basis:   NewCurrency(item.BaseAsset, ""),
			Counter: NewCurrency(item.QuoteAsset, ""),
		}

		var contractStatus = __CONTRACT_STATUS_TRANS[item.ContractStatus]
		var contractType = __CONTRACT_TYPE_TRANS[item.ContractType]
		if contractStatus == "" || contractType == "" {
			continue
		}

		var contract = &FutureContract{
			Pair:         pair,
			Symbol:       pair.ToSymbol("_", false),
			Exchange:     BINANCE,
			ContractType: contractType,
			ContractName: item.Symbol,
			Type:         FUTURE_TYPE_INVERSER, // "inverse", "linear

			SettleMode:    SETTLE_MODE_BASIS,
			Status:        contractStatus,
			OpenTimestamp: openTime.UnixNano() / int64(time.Millisecond),
			OpenDate:      openTime.Format(GO_BIRTHDAY),
			ListTimestamp: listTime.UnixNano() / int64(time.Millisecond),
			ListDate:      listTime.Format(GO_BIRTHDAY),
			DueTimestamp:  dueTime.UnixNano() / int64(time.Millisecond),
			DueDate:       dueTime.Format(GO_BIRTHDAY),

			UnitAmount:      item.ContractSize,
			TickSize:        tickSize,
			PricePrecision:  item.PricePrecision,
			AmountPrecision: item.QuantityPrecision,

			MaxScalePriceLimit: priceMaxScale,
			MinScalePriceLimit: priceMinScale,
			RawData:            string(rawData),
		}

		contracts = append(contracts, contract)
	}
	return contracts, resp, nil
}

func (future *Future) getUMContracts() ([]*FutureContract, []byte, error) {

	var contracts = make([]*FutureContract, 0)
	var nowTimestamp = time.Now().UnixNano() / int64(time.Millisecond)

	var respUm = struct {
		Symbols []*struct {
			Symbol       string `json:"symbol"`
			Pair         string `json:"pair"`
			ContractType string `json:"contractType"`
			DeliveryDate int64  `json:"deliveryDate"`
			OnboardDate  int64  `json:"onboardDate"`
			Status       string `json:"status"`
			BaseAsset    string `json:"baseAsset"`
			QuoteAsset   string `json:"quoteAsset"`
			MarginAsset  string `json:"marginAsset"`

			PricePrecision    int64 `json:"pricePrecision"`
			QuantityPrecision int64 `json:"quantityPrecision"`

			Filters []map[string]interface{} `json:"filters"`
		} `json:"symbols"`
		ServerTime int64 `json:"serverTime"`
	}{}

	var resp, errUm = future.DoRequest(
		http.MethodGet,
		FUTURE_UM_ENDPOINT,
		FUTURE_UM_EXCHANGE_INFO_URI,
		"",
		&respUm,
	)
	if errUm != nil {
		return nil, resp, errUm
	}

	for _, item := range respUm.Symbols {
		if strings.Contains(item.ContractType, "PERPETUAL") ||
			item.ContractType == "" ||
			item.DeliveryDate > (nowTimestamp+5*365*24*60*60*1000) {
			continue
		}

		var rawData, _ = json.Marshal(item)

		var priceMaxScale, priceMinScale float64 = 1.2, 0.8
		var tickSize float64 = -1
		for _, filter := range item.Filters {
			if value, ok := filter["filterType"].(string); ok && value == "PERCENT_PRICE" {
				priceMaxScale = ToFloat64(filter["multiplierUp"])
				priceMinScale = ToFloat64(filter["multiplierDown"])
			}

			if value, ok := filter["filterType"].(string); ok && value == "PRICE_FILTER" {
				tickSize = ToFloat64(filter["tickSize"])
			}
		}

		var dueTime = time.Unix(item.DeliveryDate/1000, 0).In(future.config.Location)
		var openTime = time.Unix(item.OnboardDate/1000, 0).In(future.config.Location)
		var listTime = time.Unix(item.OnboardDate/1000, 0).In(future.config.Location)

		var pair = Pair{
			Basis:   NewCurrency(item.BaseAsset, ""),
			Counter: NewCurrency(item.QuoteAsset, ""),
		}

		var contractStatus, exist = __CONTRACT_STATUS_TRANS[item.Status]
		if !exist {
			contractStatus = CONTRACT_STATUS_SUSPEND
		}

		var contractType = __CONTRACT_TYPE_TRANS[item.ContractType]
		if contractType == "" {
			continue
		}

		var contract = &FutureContract{
			Pair:         pair,
			Symbol:       pair.ToSymbol("_", false),
			Exchange:     BINANCE,
			ContractType: contractType,
			ContractName: item.Symbol,
			Type:         FUTURE_TYPE_LINEAR, // "inverse", "linear

			SettleMode: SETTLE_MODE_COUNTER,
			Status:     contractStatus,

			OpenTimestamp: openTime.UnixNano() / int64(time.Millisecond),
			OpenDate:      openTime.Format(GO_BIRTHDAY),

			ListTimestamp: listTime.UnixNano() / int64(time.Millisecond),
			ListDate:      listTime.Format(GO_BIRTHDAY),
			DueTimestamp:  dueTime.UnixNano() / int64(time.Millisecond),
			DueDate:       dueTime.Format(GO_BIRTHDAY),

			UnitAmount:      1,
			TickSize:        tickSize,
			PricePrecision:  item.PricePrecision,
			AmountPrecision: item.QuantityPrecision,

			MaxScalePriceLimit: priceMaxScale,
			MinScalePriceLimit: priceMinScale,
			RawData:            string(rawData),
		}

		contracts = append(contracts, contract)
	}

	return contracts, resp, nil
}
