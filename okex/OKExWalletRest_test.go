package okex

import (
	"net/http"
	"net/url"
	"testing"
	"time"

	. "github.com/strengthening/goghostex"
)

const (
	WALLET_API_KEY        = ""
	WALLET_API_SECRETKEY  = ""
	WALLET_API_PASSPHRASE = ""

	WALLET_PROXY_URL = "socks5://127.0.0.1:1090"
)

var (
	currency Currency = LTC
)

func TestWallet_GetAccount(t *testing.T) {

	config := &APIConfig{
		Endpoint: ENDPOINT,
		HttpClient: &http.Client{
			Transport: &http.Transport{
				Proxy: func(req *http.Request) (*url.URL, error) {
					return url.Parse(WALLET_PROXY_URL)
				},
			},
		},
		ApiKey:        WALLET_API_KEY,
		ApiSecretKey:  WALLET_API_SECRETKEY,
		ApiPassphrase: WALLET_API_PASSPHRASE,
		Location:      time.Now().Location(),
	}

	ok := New(config)
	acc, _, err := ok.Wallet.GetAccount()
	if err != nil {
		t.Error(err)
		return
	}

	ltcAcc, exist := acc.SubAccounts[currency]
	if !exist || ltcAcc.Amount < 1 {
		t.Error("不能继续测试下去了。")
		return
	}

	//tParam := TransferParameter{
	//	Currency:LTC.Symbol,
	//	From:WALLET,
	//	To:TIPS,
	//	Amount: 1,
	//}
	//
	//if err := ok.Wallet.Transfer(tParam);err!=nil{
	//	t.Error(err)
	//	return
	//}
	//
	//tParam = TransferParameter{
	//	Currency:LTC.Symbol,
	//	From:TIPS,
	//	To:WALLET,
	//	Amount: 1,
	//}
	//
	//if err := ok.Wallet.Transfer(tParam);err!=nil{
	//	t.Error(err)
	//	return
	//}

	tParam := TransferParameter{
		Currency:       LTC.Symbol,
		From:           WALLET,
		To:             FUTURE,
		Amount:         1,
		ToInstrumentId: NewCurrencyPair("ltc_usd").ToLower().ToSymbol("-"),
	}

	if err := ok.Wallet.Transfer(tParam); err != nil {
		t.Error(err)
		return
	}

	//fmt.Println(acc)
	//fmt.Println(string(body))
	//
	//acc1, body, err := ok.Future.GetFutureAccount()
	//if err != nil {
	//	t.Error(err)
	//	return
	//}
	//
	//fmt.Println(acc1)
	//fmt.Println(string(body))
	//
	//spotAcc, body, err := ok.Spot.GetAccount()
	//if err != nil {
	//	t.Error(err)
	//	return
	//}
	//
	//fmt.Println(spotAcc)
	//fmt.Println(string(body))
}
