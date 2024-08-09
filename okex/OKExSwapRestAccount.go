package okex

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"

	. "github.com/deforceHK/goghostex"
)

var _INERNAL_V5_FLOW_TYPE_CONVERTER = map[string]string{
	"2": SUBJECT_SETTLE,
	"8": SUBJECT_FUNDING_FEE,
}

func (swap *Swap) GetAccountFlow() ([]*SwapAccountItem, []byte, error) {
	var params = url.Values{}
	params.Set("instType", "SWAP")
	var response = struct {
		Code string `json:"code"`
		Msg  string `json:"msg"`
		Data []struct {
			Bal     string `json:"bal"`
			BalChg  string `json:"balChg"`
			BillId  string `json:"billId"`
			Ccy     string `json:"ccy"`
			Fee     string `json:"fee"`
			InstId  string `json:"instId"`
			SubType string `json:"subType"`
			Pnl     string `json:"pnl"`
			Type    string `json:"type"`
			Sz      string `json:"sz"`
			Ts      int64  `json:"ts,string"`
		} `json:"data"`
	}{}
	var uri = "/api/v5/account/bills?"
	resp, err := swap.DoRequest(
		http.MethodGet,
		uri+params.Encode(),
		"",
		&response,
	)

	if err != nil {
		return nil, resp, err
	}
	if response.Code != "0" {
		return nil, resp, errors.New(response.Msg)
	}

	var items = make([]*SwapAccountItem, 0)
	for _, item := range response.Data {
		var pairInfo = strings.Split(item.InstId, "-")

		itemType, exist := _INERNAL_V5_FLOW_TYPE_CONVERTER[item.Type]
		if !exist {
			continue
		}

		var settleMode = SETTLE_MODE_BASIS
		if pairInfo[1] == item.Ccy {
			settleMode = SETTLE_MODE_COUNTER
		}

		var amount = ToFloat64(item.Pnl)
		var datetime = time.Unix(item.Ts/1000, 0).In(swap.config.Location).Format(GO_BIRTHDAY)
		var info, _ = json.Marshal(item)
		items = append(items, &SwapAccountItem{
			Pair:     NewPair(pairInfo[0]+"-"+pairInfo[1], "-"),
			Exchange: OKEX,
			Subject:  itemType,

			SettleMode:     settleMode, // 1: basis 2: counter
			SettleCurrency: NewCurrency(item.Ccy, ""),
			Amount:         amount,
			Timestamp:      item.Ts,
			DateTime:       datetime,
			Info:           string(info),
			Id:             item.BillId,
		})

		if itemType == SUBJECT_SETTLE {
			items = append(items, &SwapAccountItem{
				Pair:     NewPair(pairInfo[0]+"-"+pairInfo[1], "-"),
				Exchange: OKEX,
				Subject:  SUBJECT_COMMISSION,

				SettleMode:     settleMode, // 1: basis 2: counter
				SettleCurrency: NewCurrency(item.Ccy, ""),
				Amount:         ToFloat64(item.Fee),
				Timestamp:      item.Ts,
				DateTime:       datetime,
				Info:           string(info),
				Id:             item.BillId,
			})
		}
	}
	return items, resp, nil
}

func (swap *Swap) GetPairFlow(pair Pair) ([]*SwapAccountItem, []byte, error) {
	var contract = swap.GetContract(pair)
	var marginAsset = pair.Counter.String()
	if contract.SettleMode == SETTLE_MODE_BASIS {
		marginAsset = pair.Basis.String()
	}

	var params = url.Values{}
	params.Set("instType", "SWAP")
	params.Set("instId", pair.ToSymbol("-", true)+"-SWAP")
	params.Set("ccy", marginAsset)
	var response = struct {
		Code string `json:"code"`
		Msg  string `json:"msg"`
		Data []struct {
			Bal     string `json:"bal"`
			BalChg  string `json:"balChg"`
			BillId  string `json:"billId"`
			Ccy     string `json:"ccy"`
			Fee     string `json:"fee"`
			InstId  string `json:"instId"`
			SubType string `json:"subType"`
			Pnl     string `json:"pnl"`
			Type    string `json:"type"`
			Sz      string `json:"sz"`
			Ts      int64  `json:"ts,string"`
		} `json:"data"`
	}{}
	var uri = "/api/v5/account/bills-archive?"
	var resp, err = swap.DoRequest(
		http.MethodGet,
		uri+params.Encode(),
		"",
		&response,
	)

	if err != nil {
		return nil, resp, err
	}
	if response.Code != "0" {
		return nil, resp, errors.New(response.Msg)
	}

	var items = make([]*SwapAccountItem, 0)
	for _, item := range response.Data {
		pairInfo := strings.Split(item.InstId, "-")
		var itemPair = NewPair(pairInfo[0]+"-"+pairInfo[1], "-")
		if itemPair.ToSwapContractName() != pair.ToSwapContractName() {
			continue
		}

		itemType, exist := _INERNAL_V5_FLOW_TYPE_CONVERTER[item.Type]
		if !exist {
			continue
		}

		var amount = ToFloat64(item.Pnl)
		var datetime = time.Unix(item.Ts/1000, 0).In(swap.config.Location).Format(GO_BIRTHDAY)
		var info, _ = json.Marshal(item)
		items = append(items, &SwapAccountItem{
			Pair:     pair,
			Exchange: OKEX,
			Subject:  itemType,

			SettleMode:     contract.SettleMode, // 1: basis 2: counter
			SettleCurrency: NewCurrency(item.Ccy, ""),
			Amount:         amount,
			Timestamp:      item.Ts,
			DateTime:       datetime,
			Info:           string(info),
			Id:             item.BillId,
		})

		if itemType == SUBJECT_SETTLE {
			items = append(items, &SwapAccountItem{
				Pair:     pair,
				Exchange: OKEX,
				Subject:  SUBJECT_COMMISSION,

				SettleMode:     contract.SettleMode, // 1: basis 2: counter
				SettleCurrency: NewCurrency(item.Ccy, ""),
				Amount:         ToFloat64(item.Fee),
				Timestamp:      item.Ts,
				DateTime:       datetime,
				Info:           "",
				Id:             item.BillId,
			})
		}
	}
	return items, resp, nil
}
