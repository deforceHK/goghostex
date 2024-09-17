package kraken

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"

	. "github.com/deforceHK/goghostex"
)

var sideRelation = map[FutureType]string{
	OPEN_LONG:       "buy",
	OPEN_SHORT:      "sell",
	LIQUIDATE_LONG:  "sell",
	LIQUIDATE_SHORT: "buy",
}

var placeTypeRelation = map[PlaceType]string{
	NORMAL:     "lmt",
	ONLY_MAKER: "post",
	IOC:        "ioc",
}

func (swap *Swap) PlaceOrder(order *SwapOrder) ([]byte, error) {
	if order == nil {
		return nil, errors.New("order param is nil")
	}

	var side, placeType = "", ""
	var exist = false

	if side, exist = sideRelation[order.Type]; !exist {
		return nil, errors.New("swap side not found. ")
	}

	if placeType, exist = placeTypeRelation[order.PlaceType]; !exist {
		return nil, errors.New("place type not found. ")
	}

	var contract = swap.getContract(order.Pair)
	var symbol = contract.ContractName
	var reduceOnly = "false"
	if order.Type == LIQUIDATE_LONG || order.Type == LIQUIDATE_SHORT {
		reduceOnly = "true"
	}

	var param = url.Values{}
	param.Set("symbol", symbol)
	param.Set("orderType", placeType)
	param.Set("limitPrice", FloatToPrice(order.Price, contract.PricePrecision, contract.TickSize))
	param.Set("side", side)
	param.Set("size", fmt.Sprintf("%v", order.Amount))
	param.Set("reduceOnly", reduceOnly)
	if order.Cid != "" {
		param.Set("cliOrdId", order.Cid)
	}

	var response struct {
		ServerTime string `json:"serverTime"`
		Result     string `json:"result"`
		//SendStatus []json.RawMessage `json:"sendStatus"`
	}
	var uri = "/api/v3/sendorder"
	resp, err := swap.DoRequest(
		http.MethodPost,
		uri,
		param.Encode(),
		&response,
	)

	if err != nil {
		return nil, err
	}
	//orderTime := time.Unix(response.UpdateTime/1000, 0)
	//order.OrderId = fmt.Sprintf("%d", response.OrderId)
	//order.PlaceTimestamp = now.UnixNano() / int64(time.Millisecond)
	//order.PlaceDatetime = now.In(swap.config.Location).Format(GO_BIRTHDAY)
	//order.DealTimestamp = response.UpdateTime
	//order.DealDatetime = orderTime.In(swap.config.Location).Format(GO_BIRTHDAY)
	//order.Status = statusRelation[response.Status]
	//order.Price = response.Price
	//order.Amount = response.Amount
	//if response.DealAmount > 0 {
	//	order.AvgPrice = response.CumQuote / response.DealAmount
	//	order.DealAmount = response.DealAmount
	//}
	return resp, nil
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
