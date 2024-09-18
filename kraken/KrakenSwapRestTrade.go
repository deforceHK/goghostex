package kraken

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

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

var statusRelation = map[string]TradeStatus{
	"placed":          ORDER_UNFINISH,
	"partiallyFilled": ORDER_PART_FINISH,
	"filled":          ORDER_FINISH,
	"cancelled":       ORDER_CANCEL,
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
		SendStatus struct {
			CliOrdId     string `json:"cliOrdId"`
			Status       string `json:"status"`
			ReceivedTime string `json:"receivedTime"`
			OrderId      string `json:"order_id"`
			OrderEvents  []struct {
				Order struct {
					OrderId             string  `json:"orderId"`
					CliOrdId            string  `json:"cliOrdId"`
					Type                string  `json:"type"`
					Symbol              string  `json:"symbol"`
					Side                string  `json:"side"`
					Quantity            float64 `json:"quantity"`
					Filled              float64 `json:"filled"`
					LimitPrice          float64 `json:"limitPrice"`
					ReduceOnly          bool    `json:"reduceOnly"`
					Timestamp           string  `json:"timestamp"`
					LastUpdateTimestamp string  `json:"lastUpdateTimestamp"`
				} `json:"order"`
				ReduceQuantity string `json:"reduceQuantity"`
				Type           string `json:"type"`
			} `json:"orderEvents"`
		} `json:"sendStatus"`
	}
	var uri = "/api/v3/sendorder"
	if resp, err := swap.DoRequest(
		http.MethodPost,
		uri,
		param.Encode(),
		&response,
	); err != nil {
		return resp, err
	} else {
		if response.Result != "success" ||
			len(response.SendStatus.OrderEvents) == 0 {
			return resp, errors.New(string(resp))
		}
		if orderStatus, exist := statusRelation[response.SendStatus.Status]; !exist {
			order.Status = ORDER_FAIL
			return resp, errors.New(string(resp))
		} else {

			order.Status = orderStatus
		}

		if orderTime, err := time.Parse(time.RFC3339, response.SendStatus.ReceivedTime); err != nil {
			return resp, errors.New(string(resp))
		} else {
			order.PlaceTimestamp = orderTime.UnixMilli()
			order.PlaceDatetime = orderTime.In(swap.config.Location).Format(GO_BIRTHDAY)
		}
		order.OrderId = response.SendStatus.OrderEvents[0].Order.OrderId
		return resp, nil
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
}

func (swap *Swap) CancelOrder(order *SwapOrder) ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (swap *Swap) GetOrder(order *SwapOrder) ([]byte, error) {
	var param = url.Values{}
	param.Set("orderIds", fmt.Sprintf("[\"%s\"]", order.OrderId))
	if order.Cid != "" {
		param.Set("cliOrdIds", fmt.Sprintf("[\"%s\"]", order.Cid))
	}

	var response struct {
		ServerTime string `json:"serverTime"`
		Result     string `json:"result"`
		Orders     []struct {
			Status string `json:"status"`
			Error  string `json:"error"`
		} `json:"orders"`
	}

	var uri = "/api/v3/orders/status"
	if resp, err := swap.DoRequest(
		http.MethodPost,
		uri,
		param.Encode(),
		&response,
	); err != nil {
		return resp, err
	} else {
		return resp, nil
	}
}

func (swap *Swap) GetOrders(pair Pair) ([]*SwapOrder, []byte, error) {
	//TODO implement me
	panic("implement me")
}

func (swap *Swap) GetUnFinishOrders(pair Pair) ([]*SwapOrder, []byte, error) {
	//TODO implement me
	panic("implement me")
}
