package gate

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	//"time"

	. "github.com/strengthening/goghostex"
)

type Swap struct {
	*Gate
}

func (swap *Swap) GetExchangeRule(pair Pair) (*SwapRule, []byte, error) {
	uri := "https://api.gateio.ws/api/v4/futures/%s/contracts/%s"
	symbol := pair.ToSymbol("_", true)

	if strings.Index(symbol, "_USDT") > 0 {
		uri = fmt.Sprintf(uri, strings.ToLower(pair.Counter.Symbol), symbol)
	} else {
		uri = fmt.Sprintf(uri, strings.ToLower(pair.Basis.Symbol), symbol)
	}

	r := struct {
		OrderSizeMin    int     `json:"order_size_min"`
		OrderPriceRound float64 `json:"order_price_round,string"`
	}{}

	resp, err := swap.DoRequest(
		http.MethodGet,
		uri,
		"",
		&r,
	)

	if err != nil {
		return nil, resp, err
	}

	return &SwapRule{
		Rule: Rule{
			Pair:             pair,
			Base:             pair.Basis,
			Counter:          pair.Counter,
			BaseMinSize:      float64(r.OrderSizeMin),
			BasePrecision:    GetPrecision(r.OrderPriceRound),
			CounterPrecision: GetPrecision(float64(r.OrderSizeMin)),
		},
		ContractVal: 1,
	}, resp, nil
}

func (swap *Swap) GetTicker(pair Pair) (*SwapTicker, []byte, error) {
	uri := "https://api.gateio.ws/api/v4/futures/%s/tickers?"
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

	resp, err := swap.DoRequest(
		http.MethodGet,
		fmt.Sprintf(uri, settle)+params.Encode(),
		"",
		&r,
	)

	if err != nil {
		return nil, nil, err
	}

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

func (swap *Swap) GetDepth(pair Pair, size int) (*SwapDepth, []byte, error) {
	uri := "https://api.gateio.ws/api/v4/futures/%s/order_book?"
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

	resp, err := swap.DoRequest(
		http.MethodGet,
		fmt.Sprintf(uri, settle)+params.Encode(),
		"",
		&r,
	)

	if err != nil {
		return nil, nil, err
	}

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

func (swap *Swap) GetLimit(pair Pair) (float64, float64, error) {
	panic("implement me")
}

func (swap *Swap) GetKline(pair Pair, period, size, since int) ([]*SwapKline, []byte, error) {
	uri := "https://api.gateio.ws/api/v4/futures/%s/candlesticks?"
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

	resp, err := swap.DoRequest(
		http.MethodGet,
		fmt.Sprintf(uri, settle)+params.Encode(),
		"",
		&rawResp,
	)
	if err != nil {
		return nil, resp, err
	}

	var klines = make([]*SwapKline, 0)
	for _, item := range rawResp {
		t := time.Unix(item.T, 0)
		klines = append(klines, &SwapKline{
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

	return GetAscSwapKline(klines), resp, nil
}

func (swap *Swap) GetOpenAmount(pair Pair) (float64, int64, []byte, error) {
	panic("implement me")
}

func (swap *Swap) GetFundingFees(pair Pair) ([][]interface{}, []byte, error) {
	panic("implement me")
}

func (swap *Swap) GetFundingFee(pair Pair) (float64, error) {
	panic("implement me")
}

func (swap *Swap) GetAccount() (*SwapAccount, []byte, error) {
	panic("implement me")
}

func (swap *Swap) PlaceOrder(order *SwapOrder) ([]byte, error) {

	panic("implement me")
}

func (swap *Swap) CancelOrder(order *SwapOrder) ([]byte, error) {
	panic("implement me")
}

func (swap *Swap) GetOrder(order *SwapOrder) ([]byte, error) {
	panic("implement me")
}

func (swap *Swap) GetOrders(pair Pair) ([]*SwapOrder, []byte, error) {
	panic("implement me")
}

func (swap *Swap) GetUnFinishOrders(pair Pair) ([]*SwapOrder, []byte, error) {
	panic("implement me")
}

func (swap *Swap) GetPosition(pair Pair, openType FutureType) (*SwapPosition, []byte, error) {
	panic("implement me")
}

func (swap *Swap) AddMargin(pair Pair, openType FutureType, marginAmount float64) ([]byte, error) {
	panic("implement me")
}

func (swap *Swap) ReduceMargin(pair Pair, openType FutureType, marginAmount float64) ([]byte, error) {
	panic("implement me")
}

func (swap *Swap) GetAccountFlow() ([]*SwapAccountItem, []byte, error) {
	panic("implement me")
}

func (swap *Swap) KeepAlive() {
	panic("implement me")
}
