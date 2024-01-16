package okex

import (
	"net/http"
	"net/url"
	"testing"
	"time"

	. "github.com/deforceHK/goghostex"
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
	//tAcc, resp, err := ok.Wallet.GetTradingAccount()
	//if err != nil {
	//	t.Error(err)
	//	return
	//}
	//
	//t.Log(string(resp))
	//t.Log(tAcc)

	//tAcc, resp, err := ok.Wallet.GetCurrencyChainInfo("ARB", "ARB-Arbitrum One")
	//if err != nil {
	//	t.Error(err)
	//	return
	//}
	//
	//t.Log(string(resp))
	//t.Log(*tAcc)

	var wi = WithdrawalInfo{
		Ccy:   "ARB",
		Chain: "ARB-Arbitrum One",
		Amt:   600,
		Addr:  "0x06BD042145bbdCE44654116be11FBE104adA0488",
	}
	result, resp, err := ok.Wallet.Withdrawal(wi)
	if err != nil {
		t.Error(err)
		return
	}

	t.Log(string(resp))
	t.Log(result)

}
