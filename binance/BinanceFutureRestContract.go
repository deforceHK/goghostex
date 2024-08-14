package binance

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	. "github.com/deforceHK/goghostex"
)

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

		var contract = &FutureContract{
			Pair:         pair,
			Symbol:       pair.ToSymbol("_", false),
			Exchange:     BINANCE,
			ContractType: item.ContractType,
			ContractName: item.Symbol,
			Type:         FUTURE_TYPE_INVERSER, // "inverse", "linear

			SettleMode:    SETTLE_MODE_BASIS,
			Status:        item.ContractStatus,
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

		var contract = &FutureContract{
			Pair:         pair,
			Symbol:       pair.ToSymbol("_", false),
			Exchange:     BINANCE,
			ContractType: item.ContractType,
			ContractName: item.Symbol,
			Type:         FUTURE_TYPE_LINEAR, // "inverse", "linear

			SettleMode: SETTLE_MODE_COUNTER,
			Status:     item.Status,

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
