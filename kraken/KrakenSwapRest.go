package kraken

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	. "github.com/deforceHK/goghostex"
)

const (
	SWAP_KRAKEN_ENDPOINT = "https://futures.kraken.com/derivatives"

	SWAP_BASE_MODE_CHART = "https://futures.kraken.com"
)

type Swap struct {
	*Kraken
	sync.Locker
	swapContracts SwapContracts

	nextUpdateContractTime time.Time // 下一次更新交易所contract信息
	//LastKeepLiveTime       time.Time // 上一次keep live时间。
	lastRequestTS int64 // 最近一次请求时间戳
}

func (swap *Swap) DoRequest(baseUrl, httpMethod, uri, reqBody string, response interface{}) ([]byte, error) {
	var resp, err = NewHttpRequest(
		swap.config.HttpClient,
		httpMethod,
		baseUrl+uri,
		reqBody,
		map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
			"Accept":       "application/json",
		},
	)
	if err != nil {
		return resp, err
	} else {
		swap.lastRequestTS = time.Now().UnixMilli()
		return resp, json.Unmarshal(resp, &response)
	}
}

func (swap *Swap) DoAuthRequest(httpMethod, uri, reqBody string, response interface{}) ([]byte, error) {
	var aut = ""
	var nonce = fmt.Sprintf("%d", time.Now().UnixNano())
	aut = reqBody + nonce + uri
	var sha256Hash = sha256.New()
	sha256Hash.Write([]byte(aut))
	var sha256AUT = sha256Hash.Sum(nil)
	if decodedSecret, err := base64.StdEncoding.DecodeString(swap.config.ApiSecretKey); err != nil {
		return nil, err
	} else {
		var hmacHash = hmac.New(sha512.New, decodedSecret)
		hmacHash.Write(sha256AUT)
		var hmacAUT = hmacHash.Sum(nil)
		aut = base64.StdEncoding.EncodeToString(hmacAUT)
	}

	resp, err := NewHttpRequest(
		swap.config.HttpClient,
		httpMethod,
		SWAP_KRAKEN_ENDPOINT+uri,
		reqBody,
		map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
			"Accept":       "application/json",
			"APIKey":       swap.config.ApiKey,
			"Authent":      aut,
			"Nonce":        nonce,
		},
	)

	if err != nil {
		return nil, err
	} else {
		swap.lastRequestTS = time.Now().UnixMilli()
		return resp, json.Unmarshal(resp, &response)
	}

}

func (swap *Swap) GetOpenAmount(pair Pair) (float64, int64, []byte, error) {
	//TODO implement me
	panic("implement me")
}

func (swap *Swap) GetFundingFees(pair Pair) ([][]interface{}, []byte, error) {
	//TODO implement me
	panic("implement me")
}

func (swap *Swap) GetFundingFee(pair Pair) (float64, error) {
	//TODO implement me
	panic("implement me")
}

func (swap *Swap) GetAccount() (*SwapAccount, []byte, error) {

	//TODO implement me
	panic("implement me")
}

func (swap *Swap) GetPosition(pair Pair, openType FutureType) (*SwapPosition, []byte, error) {
	//TODO implement me
	panic("implement me")
}

func (swap *Swap) KeepAlive() {
	if time.Now().UnixMilli()-swap.lastRequestTS < 5*1000 {
		return
	}
	_, _, _ = swap.GetTicker(Pair{BTC, USD})
	swap.lastRequestTS = time.Now().UnixMilli()
}
