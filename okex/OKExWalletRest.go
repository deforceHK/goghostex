package okex

import (
	"errors"
	"fmt"
	"strings"
	"time"

	. "github.com/deforceHK/goghostex"
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

func (ok *Wallet) GetAccount() (*Account, []byte, error) {
	var response []struct {
		Balance   float64 `json:"balance,string"`
		Available float64 `json:"available,string"`
		Currency  string  `json:"currency"`
		Hold      float64 `json:"hold,string"`
	}
	resp, err := ok.DoRequest(
		"GET",
		"/api/account/v3/wallet",
		"",
		&response,
	)
	if err != nil {
		return nil, resp, err
	}

	var acc Account
	acc.SubAccounts = make(map[string]SubAccount, 0)
	acc.Exchange = OKEX
	for _, itm := range response {
		currency := NewCurrency(itm.Currency, "")
		acc.SubAccounts[strings.ToUpper(itm.Currency)] = SubAccount{
			Currency:     currency,
			Amount:       itm.Balance,
			AmountFrozen: itm.Hold,
		}
	}
	return &acc, resp, nil
}

//func (ok *Wallet) GetAsset() ([]byte, error) {
//	//struct {
//	//	AccoutType string `json:"accout_type"`
//	//
//	//}{}
//
//	//reqBody, _, _ := ok.BuildRequestBody()
//	resp, err := ok.DoRequest(
//		"GET",
//		"/api/account/v3/asset-valuation",
//		"",
//		nil,
//	)
//
//	return resp, err
//}

/*
 解释说明
from或to指定为0时，sub_account为必填项。
当from为0时，to只能填6，即子账户的资金账户只能转到母账户的资金账户。
当from指定为6，to指定为1-9，且sub_account填写子账户名时，可从母账户直接划转至子账户对应的币币、合约等账户。
from或to指定为5时，instrument_id为必填项。
*/
func (ok *Wallet) Transfer(param TransferParameter) error {
	var response struct {
		Result       bool   `json:"result"`
		ErrorCode    string `json:"code"`
		ErrorMessage string `json:"message"`
	}
	reqBody, _, _ := ok.BuildRequestBody(param)
	//println(reqBody)
	_, err := ok.DoRequest("POST", "/api/account/v3/transfer", reqBody, &response)
	if err != nil {
		return err
	}

	if !response.Result {
		return errors.New(response.ErrorMessage)
	}
	return nil
}

/*
 认证过的数字货币地址、邮箱或手机号。某些数字货币地址格式为:地址+标签，例："ARDOR-7JF3-8F2E-QUWZ-CAN7F：123456"
*/
func (ok *Wallet) Withdrawal(param WithdrawParameter) (withdrawId string, resp []byte, err error) {
	var response struct {
		Result       bool   `json:"result"`
		WithdrawId   string `json:"withdraw_id"`
		ErrorCode    string `json:"code"`
		ErrorMessage string `json:"message"`
	}
	reqBody, _, _ := ok.BuildRequestBody(param)
	resp, err = ok.DoRequest("POST", "/api/account/v3/withdrawal", reqBody, &response) //
	if err != nil {
		return
	}
	if !response.Result {
		err = errors.New(response.ErrorMessage)
		return
	}
	withdrawId = response.WithdrawId
	return
}

type DepositAddress struct {
	Address     string `json:"address"`
	Tag         string `json:"tag"`
	PaymentId   string `json:"payment_id"`
	Currency    string `json:"currency"`
	CanDeposit  int    `json:"can_deposit"`
	CanWithdraw int    `json:"can_withdraw"`
	Memo        string `json:"memo"` //eos need
}

func (ok *Wallet) GetDepositAddress(currency Currency) ([]DepositAddress, []byte, error) {
	urlPath := fmt.Sprintf("/api/account/v3/deposit/address?currency=%s", currency.Symbol)
	var response []DepositAddress
	resp, err := ok.DoRequest("GET", urlPath, "", &response)
	if err != nil {
		return nil, resp, err
	}
	return response, resp, nil
}

type WithdrawFee struct {
	Currency string `json:"currency"`
	MaxFee   string `json:"max_fee"`
	MinFee   string `json:"min_fee"`
}

func (ok *Wallet) GetWithDrawalFee(currency *Currency) ([]WithdrawFee, []byte, error) {
	urlPath := "/api/account/v3/withdrawal/fee"
	if currency != nil && *currency != UNKNOWN {
		urlPath += "?currency=" + currency.Symbol
	}
	var response []WithdrawFee
	resp, err := ok.DoRequest("GET", urlPath, "", &response)
	if err != nil {
		return nil, resp, err
	}
	return response, resp, nil
}

type DepositWithDrawHistory struct {
	WithdrawalId string    `json:"withdrawal_id,omitempty"`
	Currency     string    `json:"currency"`
	Txid         string    `json:"txid"`
	Amount       float64   `json:"amount,string"`
	From         string    `json:"from,omitempty"`
	To           string    `json:"to"`
	Memo         string    `json:"memo,omitempty"`
	Fee          string    `json:"fee"`
	Status       int       `json:"status,string"`
	Timestamp    time.Time `json:"timestamp"`
}

func (ok *Wallet) GetWithDrawalHistory(currency *Currency) ([]DepositWithDrawHistory, []byte, error) {
	urlPath := "/api/account/v3/withdrawal/history"
	if currency != nil && *currency != UNKNOWN {
		urlPath += "/" + currency.Symbol
	}
	var response []DepositWithDrawHistory
	resp, err := ok.DoRequest("GET", urlPath, "", &response)
	return response, resp, err
}

func (ok *Wallet) GetDepositHistory(currency *Currency) ([]DepositWithDrawHistory, []byte, error) {
	urlPath := "/api/account/v3/deposit/history"
	if currency != nil && *currency != UNKNOWN {
		urlPath += "/" + currency.Symbol
	}
	var response []DepositWithDrawHistory
	resp, err := ok.DoRequest("GET", urlPath, "", &response)
	return response, resp, err
}
