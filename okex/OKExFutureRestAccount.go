package okex

import (
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"

	. "github.com/deforceHK/goghostex"
)

func (future *Future) GetAccount() (*FutureAccount, []byte, error) {
	var response = struct {
		Code string `json:"code"`
		Msg  string `json:"msg"`
		Data []struct {
			UTime   int64 `json:"uTime,string"`
			Details []struct {
				Ccy       string `json:"ccy"`
				Eq        string `json:"eq"`
				CashBal   string `json:"cashBal"`
				AvailEq   string `json:"availEq"`
				FrozenBal string `json:"frozenBal"`
				OrdFrozen string `json:"ordFrozen"`
				MgnRatio  string `json:"mgnRatio"`
				Upl       string `json:"upl"`
			} `json:"details"`
		} `json:"data"`
	}{}

	var urlPath = "/api/v5/account/balance"
	resp, err := future.DoRequest(
		http.MethodGet,
		urlPath,
		"",
		&response,
	)
	if err != nil {
		return nil, nil, err
	}
	if response.Code != "0" {
		return nil, nil, errors.New(response.Msg)
	}

	acc := new(FutureAccount)
	acc.Exchange = future.GetExchangeName()
	acc.SubAccount = make(map[Currency]FutureSubAccount, 0)

	for _, detail := range response.Data[0].Details {
		currency := NewCurrency(detail.Ccy, "")
		acc.SubAccount[currency] = FutureSubAccount{
			Currency:       currency,
			Margin:         ToFloat64(detail.FrozenBal), //总体被占用的保证金，
			MarginDealed:   ToFloat64(detail.FrozenBal) - ToFloat64(detail.OrdFrozen),
			MarginUnDealed: ToFloat64(detail.OrdFrozen),
			MarginRate:     ToFloat64(detail.MgnRatio),
			BalanceTotal:   ToFloat64(detail.CashBal),
			BalanceNet:     ToFloat64(detail.Eq),
			BalanceAvail:   ToFloat64(detail.AvailEq),
			ProfitReal:     0,
			ProfitUnreal:   ToFloat64(detail.Upl),
		}
	}

	return acc, resp, nil
}

func (future *Future) GetPairFlow(pair Pair) ([]*FutureAccountItem, []byte, error) {
	var contract, errContract = future.GetContract(pair, QUARTER_CONTRACT)
	if errContract != nil {
		return nil, nil, errContract
	}

	var marginAsset = pair.Counter.String()
	if contract.SettleMode == SETTLE_MODE_BASIS {
		marginAsset = pair.Basis.String()
	}

	var params = url.Values{}
	params.Set("instType", "FUTURES")
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
	var uri = "/api/v5/account/bills?"
	var resp, err = future.DoRequest(
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

	var items = make([]*FutureAccountItem, 0)
	for _, item := range response.Data {
		if strings.Index(item.InstId, pair.ToSymbol("-", true)+"-") < 0 {
			continue
		}

		itemType, exist := _INERNAL_V5_FLOW_TYPE_CONVERTER[item.Type]
		if !exist {
			continue
		}

		var amount = ToFloat64(item.Pnl)
		var datetime = time.Unix(item.Ts/1000, 0).In(future.config.Location).Format(GO_BIRTHDAY)

		items = append(items, &FutureAccountItem{
			Pair:         pair,
			Exchange:     OKEX,
			Subject:      itemType,
			ContractName: item.InstId,

			SettleMode:     contract.SettleMode, // 1: basis 2: counter
			SettleCurrency: NewCurrency(item.Ccy, ""),
			Amount:         amount,
			Timestamp:      item.Ts,
			DateTime:       datetime,
			Info:           "",
		})

		if itemType == SUBJECT_SETTLE {
			items = append(items, &FutureAccountItem{
				Pair:         pair,
				Exchange:     OKEX,
				Subject:      SUBJECT_COMMISSION,
				ContractName: item.InstId,

				SettleMode:     contract.SettleMode, // 1: basis 2: counter
				SettleCurrency: NewCurrency(item.Ccy, ""),
				Amount:         ToFloat64(item.Fee),
				Timestamp:      item.Ts,
				DateTime:       datetime,
				Info:           "",
			})
		}
	}
	return items, resp, nil
}
