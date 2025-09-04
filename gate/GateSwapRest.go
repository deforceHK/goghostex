package gate

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	. "github.com/deforceHK/goghostex"
)

type Swap struct {
	*Gate
}

//func (swap *Swap) GetExchangeRule(pair Pair) (*SwapRule, []byte, error) {
//	uri := "/api/v4/futures/%s/contracts/%s"
//	symbol := pair.ToSymbol("_", true)
//	if strings.Index(symbol, "_USDT") > 0 {
//		uri = fmt.Sprintf(uri, strings.ToLower(pair.Counter.Symbol), symbol)
//	} else {
//		uri = fmt.Sprintf(uri, strings.ToLower(pair.Basis.Symbol), symbol)
//	}
//
//	r := struct {
//		OrderSizeMin    int     `json:"order_size_min"`
//		OrderPriceRound float64 `json:"order_price_round,string"`
//	}{}
//	if resp, err := swap.DoRequest(
//		http.MethodGet,
//		uri,
//		"",
//		"",
//		&r,
//	); err != nil {
//		return nil, resp, err
//	} else {
//		return &SwapRule{
//			Rule: Rule{
//				Pair:             pair,
//				Base:             pair.Basis,
//				Counter:          pair.Counter,
//				BaseMinSize:      float64(r.OrderSizeMin),
//				BasePrecision:    GetPrecision(r.OrderPriceRound),
//				CounterPrecision: GetPrecision(float64(r.OrderSizeMin)),
//			},
//			ContractVal: 1,
//		}, resp, nil
//	}
//}

func (swap *Swap) GetTicker(pair Pair) (*SwapTicker, []byte, error) {
	uri := "/api/v4/futures/%s/tickers"
	symbol := pair.ToSymbol("_", true)
	settle := ""
	if strings.Index(symbol, "_USDT") > 0 {
		settle = strings.ToLower(pair.Counter.Symbol)
	} else {
		settle = strings.ToLower(pair.Basis.Symbol)
	}

	params := url.Values{}
	params.Add("settle", settle)
	params.Add("contract", symbol)

	r := make([]*struct {
		High24H   float64 `json:"high_24h,string"`
		Low24H    float64 `json:"low_24h,string"`
		Last      float64 `json:"last,string"`
		MarkPrice float64 `json:"mark_price,string"`
		Volume24H float64 `json:"volume_24h,string"`
	}, 0)

	if resp, err := swap.DoRequest(
		http.MethodGet,
		fmt.Sprintf(uri, settle),
		params.Encode(),
		"",
		&r,
	); err != nil {
		return nil, resp, err
	} else {
		now := time.Now()
		ts := now.Unix() * 1000
		date := now.In(swap.config.Location).Format(GO_BIRTHDAY)
		return &SwapTicker{
			Pair:      pair,
			Last:      r[0].Last,
			Buy:       r[0].Last,
			Sell:      r[0].Last,
			High:      r[0].High24H,
			Low:       r[0].Low24H,
			Vol:       r[0].Volume24H,
			Timestamp: ts,
			Date:      date,
		}, resp, nil
	}

}

func (swap *Swap) GetDepth(pair Pair, size int) (*SwapDepth, []byte, error) {
	uri := "/api/v4/futures/%s/order_book"
	symbol := pair.ToSymbol("_", true)
	settle := ""
	if strings.Index(symbol, "_USDT") > 0 {
		settle = strings.ToLower(pair.Counter.Symbol)
	} else {
		settle = strings.ToLower(pair.Basis.Symbol)
	}

	params := url.Values{}
	params.Add("settle", settle)
	params.Add("contract", symbol)
	params.Add("limit", fmt.Sprintf("%d", size))
	r := struct {
		Asks []*struct {
			P float64 `json:"p,string"`
			S float64 `json:"s"`
		} `json:"asks"`
		Bids []*struct {
			P float64 `json:"p,string"`
			S float64 `json:"s"`
		} `json:"bids"`
	}{}

	if resp, err := swap.DoRequest(
		http.MethodGet,
		fmt.Sprintf(uri, settle),
		params.Encode(),
		"",
		&r,
	); err != nil {
		return nil, resp, err
	} else {
		depth := new(SwapDepth)
		depth.Pair = pair
		now := time.Now()
		depth.Timestamp = now.UnixNano() / int64(time.Millisecond)
		depth.Date = now.In(swap.config.Location).Format(GO_BIRTHDAY)
		depth.Sequence = depth.Timestamp
		for _, bid := range r.Bids {
			price, amount := bid.P, bid.S
			depthItem := DepthRecord{Price: price, Amount: amount}
			depth.BidList = append(depth.BidList, depthItem)
		}

		for _, ask := range r.Asks {
			price, amount := ask.P, ask.S
			depthItem := DepthRecord{Price: price, Amount: amount}
			depth.AskList = append(depth.AskList, depthItem)
		}
		return depth, resp, nil
	}
}

func (swap *Swap) GetContract(pair Pair) *SwapContract {
	panic("implement me")
}

func (swap *Swap) GetLimit(pair Pair) (float64, float64, error) {
	panic("implement me")
}

func (swap *Swap) GetKline(pair Pair, period, size, since int) ([]*SwapKline, []byte, error) {
	uri := "/api/v4/futures/%s/candlesticks"
	symbol := pair.ToSymbol("_", true)
	settle := ""
	if strings.Index(symbol, "_USDT") > 0 {
		settle = strings.ToLower(pair.Counter.Symbol)
	} else {
		settle = strings.ToLower(pair.Basis.Symbol)
	}

	if _, exist := _INERNAL_KLINE_PERIOD_CONVERTER[period]; !exist {
		return nil, nil, errors.New("Can not get the period kline data. ")
	}

	params := url.Values{}
	params.Add("settle", settle)
	params.Add("contract", symbol)
	params.Add("interval", _INERNAL_KLINE_PERIOD_CONVERTER[period])
	params.Add("limit", fmt.Sprintf("%d", size))

	rawResp := make([]*struct {
		T int64   `json:"t"`
		V float64 `json:"v,int"`
		C float64 `json:"c,string"`
		H float64 `json:"h,string"`
		L float64 `json:"l,string"`
		O float64 `json:"o,string"`
	}, 0)

	if resp, err := swap.DoRequest(
		http.MethodGet,
		fmt.Sprintf(uri, settle),
		params.Encode(),
		"",
		&rawResp,
	); err != nil {
		return nil, resp, err
	} else {
		var swapKlines = make([]*SwapKline, 0)
		for _, item := range rawResp {
			t := time.Unix(item.T, 0)
			swapKlines = append(swapKlines, &SwapKline{
				Exchange:  GATE,
				Timestamp: t.UnixNano() / int64(time.Millisecond),
				Date:      t.In(swap.config.Location).Format(GO_BIRTHDAY),
				Pair:      pair,
				Open:      item.O,
				High:      item.H,
				Low:       item.L,
				Close:     item.C,
				Vol:       item.V,
			})
		}
		return GetAscSwapKline(swapKlines), resp, nil
	}
}

func (swap *Swap) GetOpenAmount(pair Pair) (float64, int64, []byte, error) {
	panic("implement me")
}

func (swap *Swap) GetFundingFees(pair Pair) ([][]interface{}, []byte, error) {
	uri := "/api/v4/futures/%s/funding_rate"
	symbol, settle := pair.ToSymbol("_", true), ""
	if strings.Index(symbol, "_USDT") > 0 {
		settle = strings.ToLower(pair.Counter.Symbol)
	} else {
		settle = strings.ToLower(pair.Basis.Symbol)
	}

	params := url.Values{}
	params.Add("settle", settle)
	params.Add("contract", symbol)
	rawResp := make([]*struct {
		T int64   `json:"t"`
		R float64 `json:"r,string"`
	}, 0)
	if resp, err := swap.DoRequest(
		http.MethodGet,
		fmt.Sprintf(uri, settle),
		params.Encode(),
		"",
		&rawResp,
	); err != nil {
		return nil, resp, err
	} else {
		fees := make([][]interface{}, 0)
		for _, r := range rawResp {
			fee := []interface{}{r.R, r.T * 1000}
			fees = append(fees, fee)
		}
		return fees, resp, nil
	}
}

func (swap *Swap) GetFundingFee(pair Pair) (float64, error) {
	uri := "/api/v4/futures/%s/funding_rate"
	symbol, settle := pair.ToSymbol("_", true), ""
	if strings.Index(symbol, "_USDT") > 0 {
		settle = strings.ToLower(pair.Counter.Symbol)
	} else {
		settle = strings.ToLower(pair.Basis.Symbol)
	}

	params := url.Values{}
	params.Add("settle", settle)
	params.Add("contract", symbol)
	params.Add("limit", "1")
	rawResp := make([]*struct {
		T int64   `json:"t"`
		R float64 `json:"r,string"`
	}, 0)
	if _, err := swap.DoRequest(
		http.MethodGet,
		fmt.Sprintf(uri, settle),
		params.Encode(),
		"",
		&rawResp,
	); err != nil {
		return 0, err
	} else {
		return rawResp[0].R, nil
	}
}

func (swap *Swap) GetAccount() (*SwapAccount, []byte, error) {
	uri := "/api/v4/futures/usdt/accounts"
	rawResp := struct {
		Total          float64 `json:"total,string"`
		UnrealisedPnl  float64 `json:"unrealised_pnl,string"`
		Available      float64 `json:"available,string"`
		OrderMargin    float64 `json:"order_margin,string"`
		PositionMargin float64 `json:"position_margin,string"`
		Point          float64 `json:"point,string"`
		Currency       string  `json:"currency"`
	}{}

	if resp, err := swap.DoSignRequest(
		http.MethodGet,
		uri,
		"",
		"",
		&rawResp,
	); err != nil {
		return nil, resp, err
	} else {
		totalMargin := rawResp.OrderMargin + rawResp.PositionMargin
		swapAccount := SwapAccount{
			Exchange:       GATE,
			Currency:       USDT,
			Margin:         totalMargin,
			MarginPosition: rawResp.PositionMargin,
			MarginOpen:     rawResp.OrderMargin,
			MarginRate:     totalMargin / rawResp.Total,
			BalanceTotal:   rawResp.Total,
			BalanceAvail:   rawResp.Available,
			BalanceNet:     rawResp.Total + rawResp.UnrealisedPnl,
			ProfitReal:     0,
			ProfitUnreal:   rawResp.UnrealisedPnl,
		}
		return &swapAccount, resp, nil
	}
}

func (swap *Swap) PlaceOrder(order *SwapOrder) ([]byte, error) {
	uri := "/api/v4/futures/%s/orders"
	symbol := order.Pair.ToSymbol("_", true)
	settle := strings.ToLower(order.Pair.Basis.Symbol)
	if strings.Index(symbol, "_USDT") > 0 {
		settle = strings.ToLower(order.Pair.Counter.Symbol)
	}

	sog := &SwapOrderGate{}
	sog.Merge(order)
	reqBody, err := json.Marshal(sog)
	if err != nil {
		return nil, err
	}
	response := struct {
		CreateTime int64 `json:"create_time"`
		Id         int64 `json:"id"`
	}{}

	resp, err := swap.DoSignRequest(
		http.MethodPost,
		fmt.Sprintf(uri, settle),
		"",
		string(reqBody),
		&response,
	)
	placeDate := time.Unix(response.CreateTime, 0).In(swap.config.Location).Format(GO_BIRTHDAY)
	order.OrderId = fmt.Sprintf("%d", response.Id)
	order.PlaceTimestamp = response.CreateTime * 1000
	order.PlaceDatetime = placeDate
	return resp, nil
}

func (swap *Swap) CancelOrder(order *SwapOrder) ([]byte, error) {
	uri := "/api/v4/futures/%s/orders/%s"
	symbol := order.Pair.ToSymbol("_", true)
	settle := strings.ToLower(order.Pair.Basis.Symbol)
	if strings.Index(symbol, "_USDT") > 0 {
		settle = strings.ToLower(order.Pair.Counter.Symbol)
	}

	response := struct {
		FinishTime int64   `json:"finish_time"`
		FinishAs   string  `json:"finish_as"`
		Status     string  `json:"status"`
		FillPrice  float64 `json:"fill_price,string"`
		Left       int64   `json:"left"`
		Size       int64   `json:"size"`
	}{}

	resp, err := swap.DoSignRequest(
		http.MethodDelete,
		fmt.Sprintf(uri, settle, order.OrderId),
		"",
		"",
		&response,
	)
	if err != nil {
		return resp, err
	}

	order.DealAmount = float64(response.Size - response.Left)
	if response.Status == "finished" {
		order.Status = ORDER_CANCEL
		order.AvgPrice = response.FillPrice
		placeDate := time.Unix(response.FinishTime, 0).In(swap.config.Location).Format(GO_BIRTHDAY)
		order.DealTimestamp = response.FinishTime * 1000
		order.DealDatetime = placeDate
	}

	return resp, nil
}

func (swap *Swap) GetOrder(order *SwapOrder) ([]byte, error) {
	uri := "/api/v4/futures/%s/orders/%s"
	symbol := order.Pair.ToSymbol("_", true)
	settle := strings.ToLower(order.Pair.Basis.Symbol)
	if strings.Index(symbol, "_USDT") > 0 {
		settle = strings.ToLower(order.Pair.Counter.Symbol)
	}

	response := struct {
		Id        int64   `json:"id"`
		Left      int64   `json:"left"`
		Size      int64   `json:"size"`
		Price     float64 `json:"price,string"`
		FillPrice float64 `json:"fill_price,string"`

		FinishTime int64  `json:"finish_time"`
		FinishAs   string `json:"finish_as"`
		Status     string `json:"status"`
	}{}
	resp, err := swap.DoSignRequest(
		http.MethodGet,
		fmt.Sprintf(uri, settle, order.OrderId),
		"",
		"",
		&response,
	)
	if err != nil {
		return resp, err
	}

	order.DealAmount = float64(response.Size - response.Left)
	if response.Status == "finished" {
		order.Status = ORDER_FINISH
		order.AvgPrice = response.FillPrice
		placeDate := time.Unix(response.FinishTime, 0).In(swap.config.Location).Format(GO_BIRTHDAY)
		order.DealTimestamp = response.FinishTime * 1000
		order.DealDatetime = placeDate
	} else if response.Status == "open" {
		order.Status = ORDER_UNFINISH
		if response.Left < response.Size {
			order.Status = ORDER_PART_FINISH
			order.AvgPrice = response.FillPrice
		}
	}
	return resp, nil
}

func (swap *Swap) GetOrders(pair Pair) ([]*SwapOrder, []byte, error) {

	uri := "/api/v4/futures/%s/orders"
	symbol := pair.ToSymbol("_", true)
	settle := strings.ToLower(pair.Basis.Symbol)
	if strings.Index(symbol, "_USDT") > 0 {
		settle = strings.ToLower(pair.Counter.Symbol)
	}

	params := url.Values{}
	params.Add("contract", pair.ToSymbol("_", true))
	params.Add("status", "finished")

	response := make([]*SwapOrderGate, 0)
	resp, err := swap.DoSignRequest(
		http.MethodGet,
		fmt.Sprintf(uri, settle),
		params.Encode(),
		"",
		&response,
	)
	orders := make([]*SwapOrder, 0)
	if err != nil {
		return orders, resp, err
	}

	for _, o := range response {
		so := o.New(swap.config.Location)
		orders = append(orders, so)
	}
	return orders, resp, nil
}

func (swap *Swap) GetUnFinishOrders(pair Pair) ([]*SwapOrder, []byte, error) {
	uri := "/api/v4/futures/%s/orders"
	symbol := pair.ToSymbol("_", true)
	settle := strings.ToLower(pair.Basis.Symbol)
	if strings.Index(symbol, "_USDT") > 0 {
		settle = strings.ToLower(pair.Counter.Symbol)
	}

	params := url.Values{}
	params.Add("contract", pair.ToSymbol("_", true))
	params.Add("status", "open")

	response := make([]*SwapOrderGate, 0)
	resp, err := swap.DoSignRequest(
		http.MethodGet,
		fmt.Sprintf(uri, settle),
		params.Encode(),
		"",
		&response,
	)
	orders := make([]*SwapOrder, 0)
	if err != nil {
		return orders, resp, err
	}

	for _, o := range response {
		so := o.New(swap.config.Location)
		orders = append(orders, so)
	}
	return orders, resp, nil
}

func (swap *Swap) GetPosition(pair Pair, openType FutureType) (*SwapPosition, []byte, error) {
	panic("implement me")
}

func (swap *Swap) GetAccountFlow() ([]*SwapAccountItem, []byte, error) {
	panic("implement me")
}

func (swap *Swap) KeepAlive() {
	_, _ = swap.GetFundingFee(Pair{Basis: BTC, Counter: USDT})
}
