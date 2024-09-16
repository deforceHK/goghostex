package kraken

import (
	"encoding/json"
	"sync"
	"time"

	. "github.com/deforceHK/goghostex"
)

const (
	SWAP_KRAKEN_ENDPOINT = "https://futures.kraken.com"

	SWAP_API_KEY        = ""
	SWAP_API_SECRETKEY  = ""
	SWAP_API_PASSPHRASE = ""
	PROXY_URL           = "socks5://127.0.0.1:1090"
)

type Swap struct {
	*Kraken
	sync.Locker
	swapContracts SwapContracts

	nextUpdateContractTime time.Time // 下一次更新交易所contract信息
	//LastKeepLiveTime       time.Time // 上一次keep live时间。
	lastRequestTS int64 // 最近一次请求时间戳
}

func (swap *Swap) DoRequest(httpMethod, uri, reqBody string, response interface{}) ([]byte, error) {
	resp, err := NewHttpRequest(
		swap.config.HttpClient,
		httpMethod,
		SWAP_KRAKEN_ENDPOINT+uri,
		reqBody,
		map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		},
	)

	if err != nil {
		return nil, err
	} else {
		swap.lastRequestTS = time.Now().UnixNano() / int64(time.Millisecond)
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

func (swap *Swap) PlaceOrder(order *SwapOrder) ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (swap *Swap) CancelOrder(order *SwapOrder) ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (swap *Swap) GetOrder(order *SwapOrder) ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (swap *Swap) GetOrders(pair Pair) ([]*SwapOrder, []byte, error) {
	//TODO implement me
	panic("implement me")
}

func (swap *Swap) GetUnFinishOrders(pair Pair) ([]*SwapOrder, []byte, error) {
	//TODO implement me
	panic("implement me")
}

func (swap *Swap) GetPosition(pair Pair, openType FutureType) (*SwapPosition, []byte, error) {
	//TODO implement me
	panic("implement me")
}

func (swap *Swap) AddMargin(pair Pair, openType FutureType, marginAmount float64) ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (swap *Swap) ReduceMargin(pair Pair, openType FutureType, marginAmount float64) ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (swap *Swap) KeepAlive() {
	//TODO implement me
	panic("implement me")
}
