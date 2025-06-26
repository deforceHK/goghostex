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
	MARKET:     "mkt",
}

var statusRelation = map[string]TradeStatus{
	"placed":          ORDER_UNFINISH,
	"partiallyFilled": ORDER_PART_FINISH,
	"filled":          ORDER_FINISH,
	"cancelled":       ORDER_CANCEL,
}

var getOrderStatusRelation = map[string]TradeStatus{
	"ENTERED_BOOK":   ORDER_UNFINISH,
	"FULLY_EXECUTED": ORDER_FINISH,
	"REJECTED":       ORDER_FAIL,
	"CANCELLED":      ORDER_CANCEL,
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

	var param = url.Values{}
	param.Set("symbol", symbol)
	param.Set("orderType", placeType)
	param.Set("side", side)
	param.Set("size", fmt.Sprintf("%v", order.Amount))
	param.Set("reduceOnly", "false")
	if order.PlaceType != MARKET {
		param.Set("limitPrice", FloatToPrice(order.Price, contract.PricePrecision, contract.TickSize))
	}
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
	if resp, err := swap.DoAuthRequest(
		http.MethodPost,
		uri,
		param.Encode(),
		&response,
	); err != nil {
		return resp, err
	} else {
		if response.Result != "success" || len(response.SendStatus.OrderEvents) == 0 {
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
		order.OrderId = response.SendStatus.OrderId
		return resp, nil
	}
}

func (swap *Swap) CancelOrder(order *SwapOrder) ([]byte, error) {
	var param = url.Values{}
	param.Set("order_id", order.OrderId)
	if order.Cid != "" {
		param.Set("cliOrdId", order.Cid)
	}
	var uri = "/api/v3/cancelorder"
	var response struct {
		Result       string `json:"result"`
		CancelStatus struct {
			Status       string `json:"status"`
			CliOrdId     string `json:"cliOrdId"`
			ReceivedTime string `json:"receivedTime"`
			OrderId      string `json:"order_id"`
		} `json:"cancelStatus"`
	}

	var resp, err = swap.DoAuthRequest(http.MethodPost, uri, param.Encode(), &response)
	if err != nil {
		return resp, err
	} else {
		if response.Result != "success" {
			return resp, errors.New(string(resp))
		}
		if orderStatus, exist := statusRelation[response.CancelStatus.Status]; !exist {
			return resp, errors.New(string(resp))
		} else {
			order.Status = orderStatus
		}
		return resp, nil
	}

}

func (swap *Swap) GetOrder(order *SwapOrder) ([]byte, error) {
	var param = url.Values{}
	param.Set("orderIds", order.OrderId)
	if order.Cid != "" {
		param.Set("cliOrdIds", order.Cid)
	}

	var response struct {
		ServerTime string `json:"serverTime"`
		Result     string `json:"result"`
		Orders     []struct {
			Status       string `json:"status"`
			Error        string `json:"error"`
			UpdateReason string `json:"updateReason"`
			Order        struct {
				Type                string  `json:"type"`
				OrderId             string  `json:"orderId"`
				CliOrdId            string  `json:"cliOrdId"`
				Symbol              string  `json:"symbol"`
				Side                string  `json:"side"`
				Quantity            float64 `json:"quantity"`
				Filled              float64 `json:"filled"`
				LimitPrice          float64 `json:"limitPrice"`
				ReduceOnly          bool    `json:"reduceOnly"`
				Timestamp           string  `json:"timestamp"`
				LastUpdateTimestamp string  `json:"lastUpdateTimestamp"`
			} `json:"order"`
		} `json:"orders"`
	}

	var uri = "/api/v3/orders/status"
	if resp, err := swap.DoAuthRequest(
		http.MethodPost,
		uri,
		param.Encode(),
		&response,
	); err != nil {
		return resp, err
	} else {
		if orderStatus, exist := getOrderStatusRelation[response.Orders[0].Status]; !exist {
			return resp, errors.New(string(resp))
		} else {
			order.Status = orderStatus
		}
		order.DealAmount = response.Orders[0].Order.Filled
		if order.Status != ORDER_FINISH {
			return resp, nil
		}

		// the order is completed, and get it fill info.
		if fills, _, err := swap.GetOrders(order.Pair); err != nil {
			return resp, err
		} else {
			var dealAmount, dealValue float64 = 0, 0
			// fill detail in fills, so one order may have multiple fills
			for _, fill := range fills {
				if fill.OrderId != order.OrderId {
					continue
				} else {
					dealAmount += fill.DealAmount
					dealValue += fill.DealAmount * fill.AvgPrice

					order.DealTimestamp = fill.DealTimestamp
					order.DealDatetime = fill.DealDatetime
				}
			}

			if dealValue > 0 {
				order.AvgPrice = dealValue / dealAmount
				order.DealAmount = dealAmount
			}
		}
		return resp, nil
	}
}

func (swap *Swap) GetOrders(pair Pair) ([]*SwapOrder, []byte, error) {
	var contract = swap.getContract(pair)
	var param = url.Values{}
	var response struct {
		ServerTime string `json:"serverTime"`
		Result     string `json:"result"`
		Fills      []struct {
			CliOrdId string  `json:"cliOrdId"`
			FillTime string  `json:"fillTime"`
			FillType string  `json:"fillType"`
			FillId   string  `json:"fill_id"`
			OrderId  string  `json:"order_id"`
			Price    float64 `json:"price"`
			Side     string  `json:"side"`
			Size     float64 `json:"size"`
			Symbol   string  `json:"symbol"`
		} `json:"fills"`
	}
	var uri = "/api/v3/fills"
	if resp, err := swap.DoAuthRequest(
		http.MethodGet,
		uri,
		param.Encode(),
		&response,
	); err != nil {
		return nil, resp, err
	} else {
		if response.Result != "success" {
			return nil, resp, errors.New(string(resp))
		}
		var orders = make([]*SwapOrder, 0)
		for _, fill := range response.Fills {
			if contract.ContractName != fill.Symbol {
				continue
			}

			var fillTime, _ = time.Parse(time.RFC3339, fill.FillTime)
			orders = append(orders, &SwapOrder{
				Cid:           fill.CliOrdId,
				OrderId:       fill.OrderId,
				AvgPrice:      fill.Price,
				DealAmount:    fill.Size,
				DealDatetime:  fillTime.Format(GO_BIRTHDAY),
				DealTimestamp: fillTime.UnixMilli(),
			})
		}
		return orders, resp, nil
	}
}

func (swap *Swap) GetUnFinishOrders(pair Pair) ([]*SwapOrder, []byte, error) {

	panic("implement me")
}
