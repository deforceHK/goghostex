package okex

import (
	"errors"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	. "github.com/deforceHK/goghostex"
)

type Swap struct {
	*OKEx
	sync.Locker
	swapContracts SwapContracts

	nextUpdateContractTime time.Time // 下一次更新交易所contract信息
}

func (swap *Swap) getAmount(price, amountContract float64, contract *SwapContract) float64 {
	var amount float64 = 0
	if contract.SettleMode == SETTLE_MODE_BASIS {
		amount = amountContract * contract.UnitAmount / price
	} else {
		amount = amountContract * contract.UnitAmount
	}
	return amount
}

func (swap *Swap) GetAccount() (*SwapAccount, []byte, error) {
	panic("implement me")
}

var _INERNAL_V5_FUTURE_TYPE_CONVERTER = map[FutureType][]string{
	OPEN_LONG:       {"buy", "long"},
	OPEN_SHORT:      {"sell", "short"},
	LIQUIDATE_LONG:  {"sell", "long"},
	LIQUIDATE_SHORT: {"buy", "short"},
}

var _INERNAL_V5_FUTURE_PLACE_TYPE_CONVERTER = map[PlaceType]string{
	NORMAL:     "limit",
	ONLY_MAKER: "post_only",
	FOK:        "fok",
	IOC:        "ioc",
	MARKET:     "market",
}

func (swap *Swap) PlaceOrder(order *SwapOrder) ([]byte, error) {
	var contract = swap.getContract(order.Pair)
	var request = struct {
		InstId  string `json:"instId"`
		TdMode  string `json:"tdMode"`
		Side    string `json:"side"`
		PosSide string `json:"posSide,omitempty"`
		OrdType string `json:"ordType"`
		Sz      string `json:"sz"`
		Px      string `json:"px"`
		ClOrdId string `json:"clOrdId,omitempty"`
	}{}

	request.InstId = order.Pair.ToSymbol("-", true) + "-SWAP"
	request.TdMode = "cross" // todo 目前写死全仓，后续调整成可逐仓
	sideInfo, _ := _INERNAL_V5_FUTURE_TYPE_CONVERTER[order.Type]
	request.Side = sideInfo[0]
	request.PosSide = sideInfo[1]
	placeInfo, _ := _INERNAL_V5_FUTURE_PLACE_TYPE_CONVERTER[order.PlaceType]
	request.OrdType = placeInfo
	request.Sz = FloatToString(order.Amount, contract.AmountPrecision)
	request.Px = FloatToPrice(order.Price, contract.PricePrecision, contract.TickSize)
	request.ClOrdId = order.Cid

	var response = struct {
		Code string `json:"code"`
		Msg  string `json:"msg"`
		Data []struct {
			ClOrdId string `json:"clOrdId"`
			OrdId   string `json:"ordId"`
			SCode   string `json:"sCode"`
			SMsg    string `json:"sMsg"`
		} `json:"data"`
	}{}
	var uri = "/api/v5/trade/order"

	now := time.Now()
	order.PlaceTimestamp = now.UnixNano() / int64(time.Millisecond)
	order.PlaceDatetime = now.In(swap.config.Location).Format(GO_BIRTHDAY)
	reqBody, _, _ := swap.BuildRequestBody(request)
	resp, err := swap.DoRequest(
		http.MethodPost,
		uri,
		reqBody,
		&response,
	)

	if err != nil {
		return resp, err
	}
	if len(response.Data) > 0 && response.Data[0].SCode != "0" {
		return resp, errors.New(string(resp)) // very important cause it has the error code
	}
	if response.Code != "0" {
		return resp, errors.New(string(resp)) // very important cause it has the error code
	}

	now = time.Now()
	order.DealTimestamp = now.UnixNano() / int64(time.Millisecond)
	order.DealDatetime = now.In(swap.config.Location).Format(GO_BIRTHDAY)
	order.OrderId = response.Data[0].OrdId
	return resp, nil
}

func (swap *Swap) CancelOrder(order *SwapOrder) ([]byte, error) {

	var request = struct {
		InstId string `json:"instId"`
		OrdId  string `json:"ordId"`
	}{
		order.Pair.ToSymbol("-", true) + "-SWAP",
		order.OrderId,
	}

	var response = struct {
		Code string `json:"code"`
		Msg  string `json:"msg"`
		Data []struct {
			ClOrdId string `json:"clOrdId"`
			OrdId   string `json:"ordId"`
			SCode   string `json:"sCode"`
			SMsg    string `json:"sMsg"`
		} `json:"data"`
	}{}

	var uri = "/api/v5/trade/cancel-order"
	reqBody, _, _ := swap.BuildRequestBody(request)

	resp, err := swap.DoRequest(
		http.MethodPost,
		uri,
		reqBody,
		&response,
	)
	if err != nil {
		return resp, err
	}
	if len(response.Data) == 0 {
		return resp, errors.New("request lack the data. ")
	}
	if len(response.Data) != 0 && response.Data[0].SCode != "0" {
		return resp, errors.New(response.Data[0].SMsg)
	}

	return resp, nil

}

func (swap *Swap) GetOrders(pair Pair) ([]*SwapOrder, []byte, error) {
	panic("implement me")
}

var _INERNAL_V5_FUTURE_ORDER_STATUE_CONVERTER = map[string]TradeStatus{
	"canceled":         ORDER_CANCEL,
	"live":             ORDER_UNFINISH,
	"partially_filled": ORDER_PART_FINISH,
	"filled":           ORDER_FINISH,
}

func (swap *Swap) GetOrder(order *SwapOrder) ([]byte, error) {

	var params = url.Values{}
	params.Set("instId", order.Pair.ToSymbol("-", true)+"-SWAP")
	params.Set("ordId", order.OrderId)

	var response = struct {
		Code string `json:"code"`
		Msg  string `json:"msg"`
		Data []struct {
			ClOrdId   string  `json:"clOrdId"`
			OrdId     string  `json:"ordId"`
			Px        float64 `json:"px,string"`
			Sz        float64 `json:"sz,string"`
			AvgPx     string  `json:"avgPx"`
			AccFillSz float64 `json:"accFillSz,string"`
			State     string  `json:"state"`
			Lever     float64 `json:"lever,string"`
			Fee       float64 `json:"fee,string"`
			UTime     int64   `json:"uTime,string"`
			CTime     int64   `json:"cTime,string"`
		} `json:"data"`
	}{}
	var uri = "/api/v5/trade/order?"

	resp, err := swap.DoRequest(
		http.MethodGet,
		uri+params.Encode(),
		"",
		&response,
	)

	if err != nil {
		return resp, err
	}
	if response.Code != "0" {
		return resp, errors.New(response.Msg)
	}
	if len(response.Data) == 0 || response.Data[0].State == "live" {
		return resp, nil
	}

	if status, exist := _INERNAL_V5_FUTURE_ORDER_STATUE_CONVERTER[response.Data[0].State]; exist {
		order.Status = status
	}
	order.Price = response.Data[0].Px
	order.Amount = response.Data[0].Sz

	order.AvgPrice = ToFloat64(response.Data[0].AvgPx)
	order.DealAmount = response.Data[0].AccFillSz
	order.LeverRate = ToInt64(response.Data[0].Lever)
	order.Fee = response.Data[0].Fee

	order.DealTimestamp = response.Data[0].UTime
	order.DealDatetime = time.Unix(
		order.DealTimestamp/1000, 0,
	).In(swap.config.Location).Format(GO_BIRTHDAY)

	order.PlaceTimestamp = response.Data[0].CTime
	order.PlaceDatetime = time.Unix(
		order.PlaceTimestamp/1000, 0,
	).In(swap.config.Location).Format(GO_BIRTHDAY)
	return resp, err

}

func (swap *Swap) GetUnFinishOrders(pair Pair) ([]*SwapOrder, []byte, error) {
	panic("implement me")
}

func (swap *Swap) GetPosition(pair Pair, openType FutureType) (*SwapPosition, []byte, error) {
	panic("implement me")
}

func (swap *Swap) getContract(pair Pair) *SwapContract {
	defer swap.Unlock()
	swap.Lock()

	var now = time.Now().In(swap.config.Location)
	if now.After(swap.nextUpdateContractTime) {
		_, err := swap.updateContracts()
		//重试三次
		for i := 0; err != nil && i < 3; i++ {
			time.Sleep(time.Second)
			_, err = swap.updateContracts()
		}
		// 初次启动必须可以吧。
		if swap.nextUpdateContractTime.IsZero() && err != nil {
			panic(err)
		}

	}
	return swap.swapContracts.ContractNameKV[pair.ToSwapContractName()]
}

func (swap *Swap) updateContracts() ([]byte, error) {
	var params = url.Values{}
	params.Set("instType", "SWAP")

	var uri = "/api/v5/public/instruments?"
	var response = struct {
		Code string `json:"code"`
		Msg  string `json:"msg"`
		Data []struct {
			InstId    string `json:"instId"`
			Uly       string `json:"uly"`
			SettleCcy string `json:"settleCcy"`
			CtVal     string `json:"ctVal"`
			TickSz    string `json:"tickSz"`
			LotSz     string `json:"lotSz"`
			MinSz     string `json:"minSz"`
		} `json:"data"`
	}{}
	resp, err := swap.DoRequestMarket(
		http.MethodGet,
		uri+params.Encode(),
		"",
		&response,
	)
	if err != nil {
		return resp, err
	}
	if response.Code != "0" {
		return resp, errors.New(response.Msg)
	}
	if len(response.Data) == 0 {
		return resp, errors.New("The api data not ready. ")
	}

	var swapContracts = SwapContracts{
		ContractNameKV: make(map[string]*SwapContract, 0),
	}
	for _, item := range response.Data {
		var pair = NewPair(item.Uly, "-")
		var symbol = pair.ToSymbol("_", false)
		var stdContractName = pair.ToSwapContractName()
		var firstCurrency = strings.Split(symbol, "_")[0]
		var settleMode = SETTLE_MODE_BASIS
		if firstCurrency != strings.ToLower(item.SettleCcy) {
			settleMode = SETTLE_MODE_COUNTER
		}

		var contract = SwapContract{
			Pair:         pair,
			Symbol:       pair.ToSymbol("_", false),
			Exchange:     OKEX,
			ContractName: stdContractName,
			SettleMode:   settleMode,

			UnitAmount:      ToFloat64(item.CtVal),
			TickSize:        ToFloat64(item.TickSz),
			PricePrecision:  GetPrecisionInt64(ToFloat64(item.TickSz)),
			AmountPrecision: GetPrecisionInt64(ToFloat64(item.LotSz)),
		}

		swapContracts.ContractNameKV[stdContractName] = &contract
	}

	// setting next update time.
	var nowTime = time.Now().In(swap.config.Location)
	var nextUpdateTime = time.Date(
		nowTime.Year(), nowTime.Month(), nowTime.Day(),
		16, 0, 0, 0, swap.config.Location,
	)
	if nowTime.Hour() >= 16 {
		nextUpdateTime = nextUpdateTime.AddDate(0, 0, 1)
	}

	swap.nextUpdateContractTime = nextUpdateTime
	swap.swapContracts = swapContracts
	return resp, nil
}

func (swap *Swap) KeepAlive() {
	nowTimestamp := time.Now().Unix() * 1000
	// last in 5s, no need to keep alive.
	if (nowTimestamp - swap.config.LastTimestamp) < 5*1000 {
		return
	}

	_, _, _ = swap.GetDepth(Pair{BTC, USDT}, 2)
	swap.config.LastTimestamp = nowTimestamp
}
