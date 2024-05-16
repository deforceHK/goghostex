package okex

import (
	"encoding/json"
	"errors"
	"fmt"
)

const (
	SUB_ACCOUNT = iota //子账户
	SPOT               // 币币交易
	_
	FUTURE      //交割合约
	C2C         //法币
	SPOT_MARGIN //币币杠杆交易
	WALLET      // 资金账户
	_
	TIPS //余币宝
	SWAP //永续合约
	_
	_
	OPTION //期权
)

const (
	WITHDRAWAL_OKCOIN int = 2 //提币到okcoin国际站
	WITHDRAWAL_OKEx       = 3 //提币到okex，站内提币
	WITHDRAWAL_COIN       = 4 //提币到数字货币地址，跨平台提币或者提到自己钱包
)

type TradingAccountDetail struct {
	Ccy string `json:"ccy"`

	CashBal  float64 `json:"cashBal,string"`
	AvailBal float64 `json:"availBal,string"`
	AvailEq  float64 `json:"availEq,string"`

	FrozenBal  float64 `json:"frozenBal,string"`
	BorrowFroz float64 `json:"borrowFroz,string"`
	OrdFrozen  float64 `json:"ordFrozen,string"`
}

type FundingAccountDetail struct {
	Ccy       string  `json:"ccy"`
	AvailBal  float64 `json:"availBal,string"`
	CashBal   float64 `json:"bal,string"`
	FrozenBal float64 `json:"frozenBal,string"`
}

type TransferParameter struct {
	Currency       string  `json:"currency"`
	From           int     `json:"from"`
	To             int     `json:"to"`
	Amount         float64 `json:"amount"`
	SubAccount     string  `json:"sub_account"`
	InstrumentId   string  `json:"instrument_id"`
	ToInstrumentId string  `json:"to_instrument_id"`
}

type WithdrawParameter struct {
	Currency    string  `json:"currency"`
	Amount      float64 `json:"amount,string"`
	Destination int     `json:"destination"` //提币到(2:OKCoin国际 3:OKEx 4:数字货币地址)
	ToAddress   string  `json:"to_address"`
	TradePwd    string  `json:"trade_pwd"`
	Fee         string  `json:"fee"`
}

type Wallet struct {
	*OKEx
}

type CurrencyStatus struct {
	Ccy    string  `json:"ccy"`
	Chain  string  `json:"chain"`
	MinFee float64 `json:"minFee,string"`
	MaxFee float64 `json:"maxFee,string"`
}

func (ok *Wallet) GetCurrencyChainInfo(ccy, chain string) (*CurrencyStatus, []byte, error) {
	var response struct {
		Code int64             `json:"code,string"`
		Msg  string            `json:"msg"`
		Data []*CurrencyStatus `json:"data"`
	}

	resp, err := ok.DoRequest(
		"GET",
		"/api/v5/asset/currencies?ccy="+ccy,
		"",
		&response,
	)
	if err != nil {
		return nil, resp, err
	}
	if response.Code != 0 || len(response.Data) == 0 {
		return nil, resp, errors.New(response.Msg)
	}

	var supportedChain = ""
	for _, detail := range response.Data {
		supportedChain += detail.Chain + ","
		if detail.Chain == chain {
			return detail, resp, nil
		}
	}

	return nil, resp, errors.New(fmt.Sprintf("not supported chain %s, supported chain %s", chain, supportedChain))
}

func (ok *Wallet) GetTradingAccount() (map[string]*TradingAccountDetail, []byte, error) {
	var response struct {
		Code int64                   `json:"code,string"`
		Msg  string                  `json:"msg"`
		Data []*TradingAccountDetail `json:"data"`
	}

	var resp, err = ok.DoRequest(
		"GET",
		"/api/v5/account/balance",
		"",
		&response,
	)
	if err != nil {
		return nil, resp, err
	}

	var details = make(map[string]*TradingAccountDetail, 0)
	for _, detail := range response.Data {
		details[detail.Ccy] = detail
	}

	return details, resp, nil
}

func (ok *Wallet) GetFundingAccount() (map[string]*FundingAccountDetail, []byte, error) {
	var response struct {
		Code int64                   `json:"code,string"`
		Msg  string                  `json:"msg"`
		Data []*FundingAccountDetail `json:"data"`
	}

	var resp, err = ok.DoRequest(
		"GET",
		"/api/v5/asset/balances",
		"",
		&response,
	)
	if err != nil {
		return nil, resp, err
	}

	fmt.Println(string(resp))

	var details = make(map[string]*FundingAccountDetail, 0)
	for _, detail := range response.Data {
		details[detail.Ccy] = detail
	}

	return details, resp, nil
}

type WithdrawalInfo struct {
	Ccy   string  `json:"ccy"`
	Chain string  `json:"chain"`
	Amt   float64 `json:"amt,string"`
	Addr  string  `json:"addr"`
}

func (ok *Wallet) Withdrawal(info WithdrawalInfo) (withdrawId string, resp []byte, err error) {
	var withdrawFee = 0.0

	var ccyStatus, ccyResp, ccyErr = ok.GetCurrencyChainInfo(info.Ccy, info.Chain)
	if ccyErr != nil {
		return "", ccyResp, ccyErr
	} else {
		withdrawFee = ccyStatus.MinFee
	}

	var response struct {
		Code int64           `json:"code,string"`
		Msg  string          `json:"msg"`
		Data json.RawMessage `json:"data"`
	}

	var request = struct {
		Ccy    string `json:"ccy"`
		Amt    string `json:"amt"`
		Dest   string `json:"dest"`
		ToAddr string `json:"toAddr"`
		Fee    string `json:"fee"`
		Chain  string `json:"chain"`
	}{
		info.Ccy, fmt.Sprintf("%f", info.Amt-withdrawFee),
		fmt.Sprintf("%d", WITHDRAWAL_COIN),
		info.Addr, fmt.Sprintf("%f", withdrawFee), info.Chain,
	}

	reqBody, _, _ := ok.BuildRequestBody(request)
	resp, err = ok.DoRequest("POST", "/api/v5/asset/withdrawal", reqBody, &response) //
	if err != nil {
		return
	}
	if response.Code != 0 {
		err = errors.New(response.Msg)
		return
	}
	withdrawId = string(response.Data)
	return
}
