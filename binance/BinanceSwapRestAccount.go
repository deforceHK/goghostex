package binance

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	. "github.com/deforceHK/goghostex"
)

func (swap *Swap) GetAccount() (*SwapAccount, []byte, error) {
	var params = url.Values{}
	params.Add(
		"timestamp",
		fmt.Sprintf("%d", time.Now().UnixNano()/int64(time.Millisecond)),
	)

	if err := swap.buildParamsSigned(&params); err != nil {
		return nil, nil, err
	}

	var response = struct {
		FeeTier      int64 `json:"feeTier"`
		CanTrade     bool  `json:"canTrade"`
		CanDeposit   bool  `json:"canDeposit"`
		CanWithdraw  bool  `json:"canWithdraw"`
		UpdateTime   int64 `json:"updateTime"`
		TradeGroupId int64 `json:"tradeGroupId"`

		MultiAssetsMargin bool `json:"multiAssetsMargin"`

		TotalInitialMargin          float64 `json:"totalInitialMargin,string"`
		TotalMaintMargin            float64 `json:"totalMaintMargin,string"`
		TotalWalletBalance          float64 `json:"totalWalletBalance,string"`
		TotalUnrealizedProfit       float64 `json:"totalUnrealizedProfit,string"`
		TotalMarginBalance          float64 `json:"totalMarginBalance,string"`
		TotalPositionInitialMargin  float64 `json:"totalPositionInitialMargin,string"`
		TotalOpenOrderInitialMargin float64 `json:"totalOpenOrderInitialMargin,string"`
		TotalCrossWalletBalance     float64 `json:"totalCrossWalletBalance,string"`
		TotalCrossUnPnl             float64 `json:"totalCrossUnPnl,string"`
		AvailableBalance            float64 `json:"availableBalance,string"`
		MaxWithdrawAmount           float64 `json:"maxWithdrawAmount,string"`

		Assets []struct {
			Asset                  string  `json:"asset"`
			WalletBalance          float64 `json:"walletBalance,string"`
			UnrealizedProfit       float64 `json:"unrealizedProfit,string"`
			MarginBalance          float64 `json:"marginBalance,string"`
			MaintMargin            float64 `json:"maintMargin,string"`
			InitialMargin          float64 `json:"initialMargin,string"`
			PositionInitialMargin  float64 `json:"positionInitialMargin,string"`
			OpenOrderInitialMargin float64 `json:"openOrderInitialMargin,string"`
			CrossWalletBalance     float64 `json:"crossWalletBalance,string"`
			CrossUnPnl             float64 `json:"crossUnPnl,string"`
			AvailableBalance       float64 `json:"availableBalance,string"`
			MaxWithdrawAmount      float64 `json:"maxWithdrawAmount,string"`
			MarginAvailable        bool    `json:"marginAvailable"`
			UpdateTime             int64   `json:"updateTime"`
		} `json:"assets"`

		Positions []struct {
			Symbol           string  `json:"symbol"`
			InitialMargin    float64 `json:"initialMargin,string"`
			MaintMargin      float64 `json:"maintMargin,string"`
			UnrealizedProfit float64 `json:"unrealizedProfit,string"`
			PositionInitial  float64 `json:"positionInitialMargin,string"`
			OpenOrderInitial float64 `json:"openOrderInitialMargin,string"`
			Leverage         float64 `json:"leverage,string"`
			Isolated         bool    `json:"isolated"`
			EntryPrice       float64 `json:"entryPrice,string"`
			MaxNotional      float64 `json:"maxNotional,string"`
			BidNotional      float64 `json:"bidNotional,string"`
			AskNotional      float64 `json:"askNotional,string"`

			PositionSide string  `json:"positionSide"`
			PositionAmt  float64 `json:"positionAmt,string"`
		} `json:"positions"`
	}{}

	resp, err := swap.DoRequest(
		http.MethodGet,
		"/fapi/v2/account?"+params.Encode(),
		"",
		&response,
		SETTLE_MODE_COUNTER,
	)
	if err != nil {
		return nil, nil, err
	}

	var account = &SwapAccount{
		Exchange:       BINANCE,
		Currency:       USDT,
		Margin:         response.TotalMaintMargin,
		MarginPosition: response.TotalPositionInitialMargin,
		MarginOpen:     response.TotalOpenOrderInitialMargin,
		BalanceTotal:   response.TotalMarginBalance,
		BalanceNet:     response.TotalWalletBalance,
		BalanceAvail:   response.AvailableBalance,
		ProfitReal:     0,
		ProfitUnreal:   response.TotalUnrealizedProfit,
		Positions:      make([]*SwapPosition, 0),
	}

	for _, p := range response.Positions {
		// do not support binance, the one side trade.
		if p.PositionSide == "BOTH" {
			continue
		}
		// There don't have position.
		if p.InitialMargin == 0.0 {
			continue
		}

		pair := Pair{
			Basis:   NewCurrency(p.Symbol[0:len(p.Symbol)-4], ""),
			Counter: NewCurrency(p.Symbol[len(p.Symbol)-4:len(p.Symbol)], ""),
		}
		futureType := OPEN_LONG
		if p.PositionSide == "SHORT" {
			futureType = OPEN_SHORT
		}

		marginType := CROSS
		if p.Isolated {
			marginType = ISOLATED
		}

		sp := &SwapPosition{
			Pair:         pair,
			Type:         futureType,
			Price:        p.EntryPrice,
			MarginType:   marginType,
			MarginAmount: p.InitialMargin,
			Leverage:     int64(p.Leverage),
		}

		account.Positions = append(account.Positions, sp)
	}

	return account, resp, nil

}

var subjectKV = map[string]string{
	"COMMISSION":   SUBJECT_COMMISSION,
	"REALIZED_PNL": SUBJECT_SETTLE,
	"FUNDING_FEE":  SUBJECT_FUNDING_FEE,
}

// only have funding_fee commision settle
func (swap *Swap) GetAccountFlow() ([]*SwapAccountItem, []byte, error) {
	var cItems, cResp, cErr = swap.counterAccountFlow()
	if cErr != nil {
		return nil, cResp, cErr
	}

	var bItems, bResp, bErr = swap.basisAccountFlow()
	if bErr != nil {
		return nil, bResp, bErr
	}

	var items = make([]*SwapAccountItem, 0)
	for _, item := range cItems {
		items = append(items, item)
	}
	for _, item := range bItems {
		items = append(items, item)
	}

	var resp, _ = json.Marshal([]string{string(cResp), string(bResp)})
	return items, resp, nil
}

func (swap *Swap) counterAccountFlow() ([]*SwapAccountItem, []byte, error) {

	var params = url.Values{}
	if err := swap.buildParamsSigned(&params); err != nil {
		return nil, nil, err
	}

	var responses = make([]*struct {
		Symbol     string  `json:"symbol"`
		IncomeType string  `json:"incomeType"`
		Income     float64 `json:"income,string"`
		Asset      string  `json:"asset"`
		Info       string  `json:"info"`
		Time       int64   `json:"time"`
		TranId     int64   `json:"tranId"`
		TradeId    string  `json:"tradeId"`
	}, 0)

	resp, err := swap.DoRequest(
		http.MethodGet,
		SWAP_COUNTER_INCOME_URI+params.Encode(),
		"",
		&responses,
		SETTLE_MODE_COUNTER,
	)
	if err != nil {
		return nil, resp, err
	}

	var items = make([]*SwapAccountItem, 0)
	for i := len(responses) - 1; i >= 0; i-- {
		var r = responses[i]
		var itemType, exist = subjectKV[r.IncomeType]
		if !exist || r.Symbol == "" || strings.Index(r.Symbol, "_") > 0 {
			continue
		}

		var spiltNum = strings.Index(r.Symbol, r.Asset)
		var dateTime = time.Unix(r.Time/1000, 0).In(swap.config.Location).Format(GO_BIRTHDAY)
		var pair = Pair{
			NewCurrency(r.Symbol[:spiltNum], ""), NewCurrency(r.Asset, ""),
		}
		var info, _ = json.Marshal(r)
		items = append(items, &SwapAccountItem{
			Pair:           pair,
			Exchange:       BINANCE,
			Subject:        itemType,
			SettleMode:     2,
			SettleCurrency: NewCurrency(r.Asset, ""),
			Amount:         r.Income,
			Timestamp:      r.Time,
			DateTime:       dateTime,
			Info:           string(info),
			Id:             strconv.FormatInt(r.TranId, 10),
		})
	}

	return items, resp, nil
}

func (swap *Swap) basisAccountFlow() ([]*SwapAccountItem, []byte, error) {
	var params = url.Values{}
	if err := swap.buildParamsSigned(&params); err != nil {
		return nil, nil, err
	}

	var responses = make([]*struct {
		Symbol     string  `json:"symbol"`
		IncomeType string  `json:"incomeType"`
		Income     float64 `json:"income,string"`
		Asset      string  `json:"asset"`
		Info       string  `json:"info"`
		Time       int64   `json:"time"`
		TranId     int64   `json:"tranId"`
		TradeId    string  `json:"tradeId"`
	}, 0)

	resp, err := swap.DoRequest(
		http.MethodGet,
		SWAP_BASIS_INCOME_URI+params.Encode(),
		"",
		&responses,
		SETTLE_MODE_BASIS,
	)
	if err != nil {
		return nil, resp, err
	}

	var items = make([]*SwapAccountItem, 0)
	for i := len(responses) - 1; i >= 0; i-- {
		var r = responses[i]
		var itemType, exist = subjectKV[r.IncomeType]
		if !exist || r.Symbol == "" || strings.Index(r.Symbol, "_PERP") < 0 {
			continue
		}

		var dateTime = time.Unix(r.Time/1000, 0).In(swap.config.Location).Format(GO_BIRTHDAY)
		var pair = Pair{
			NewCurrency(r.Asset, ""), USD,
		}
		var info, _ = json.Marshal(r)
		items = append(items, &SwapAccountItem{
			Pair:           pair,
			Exchange:       BINANCE,
			Subject:        itemType,
			SettleMode:     SETTLE_MODE_BASIS,
			SettleCurrency: NewCurrency(r.Asset, ""),
			Amount:         r.Income,
			Timestamp:      r.Time,
			DateTime:       dateTime,
			Info:           string(info),
			Id:             strconv.FormatInt(r.TranId, 10),
		})
	}

	return items, resp, nil
}

func (swap *Swap) GetPairFlow(pair Pair) ([]*SwapAccountItem, []byte, error) {
	var contract = swap.GetContract(pair)
	var paramSymbol = pair.ToSymbol("", true)
	var uri = SWAP_COUNTER_INCOME_URI
	if contract.SettleMode == SETTLE_MODE_BASIS {
		paramSymbol += "_PERP"
		uri = SWAP_BASIS_INCOME_URI
	}

	var params = url.Values{}
	params.Set("symbol", paramSymbol)
	if err := swap.buildParamsSigned(&params); err != nil {
		return nil, nil, err
	}

	var responses = make([]*struct {
		Symbol     string  `json:"symbol"`
		IncomeType string  `json:"incomeType"`
		Income     float64 `json:"income,string"`
		Asset      string  `json:"asset"`
		Info       string  `json:"info"`
		Time       int64   `json:"time"`
		TranId     int64   `json:"tranId"`
	}, 0)

	var resp, err = swap.DoRequest(
		http.MethodGet,
		uri+params.Encode(),
		"",
		&responses,
		contract.SettleMode,
	)
	if err != nil {
		return nil, resp, err
	}

	items := make([]*SwapAccountItem, 0)
	for i := len(responses) - 1; i >= 0; i-- {
		var itemType, exist = subjectKV[responses[i].IncomeType]
		if !exist {
			continue
		}

		var r = responses[i]
		var dateTime = time.Unix(r.Time/1000, 0).In(swap.config.Location).Format(GO_BIRTHDAY)
		var info, _ = json.Marshal(r)
		items = append(items, &SwapAccountItem{
			Pair:           pair,
			Exchange:       BINANCE,
			Subject:        itemType,
			SettleMode:     contract.SettleMode,
			SettleCurrency: NewCurrency(r.Asset, ""),
			Amount:         r.Income,
			Timestamp:      r.Time,
			DateTime:       dateTime,
			Info:           string(info),
			Id:             strconv.FormatInt(r.TranId, 10),
		})
	}

	return items, resp, nil
}
