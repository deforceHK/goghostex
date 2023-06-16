package binance

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	. "github.com/strengthening/goghostex"
)

type Margin struct {
	*Binance
}

//public api
func (margin *Margin) GetTicker(pair Pair) (*Ticker, []byte, error) {
	return margin.Spot.GetTicker(pair)
}

func (margin *Margin) GetDepth(pair Pair, size int) (*Depth, []byte, error) {
	return margin.Spot.GetDepth(pair, size)
}

func (margin *Margin) GetKlineRecords(pair Pair, period, size, since int) ([]*Kline, []byte, error) {
	return margin.Spot.GetKlineRecords(pair, period, size, since)
}

func (margin *Margin) GetExchangeRule(pair Pair) (*Rule, []byte, error) {
	return margin.Spot.GetExchangeRule(pair)
}

// private api
func (margin *Margin) PlaceOrder(order *Order) ([]byte, error) {
	uri := "/sapi/v1/margin/order?"
	if order.Cid == "" {
		order.Cid = UUID()
	}

	orderSide := ""
	orderType := ""

	switch order.Side {
	case BUY:
		orderSide = "BUY"
		orderType = "LIMIT"
	case SELL:
		orderSide = "SELL"
		orderType = "LIMIT"
	case BUY_MARKET:
		orderSide = "BUY"
		orderType = "MARKET"
	case SELL_MARKET:
		orderSide = "SELL"
		orderType = "MARKET"
	default:
		return nil, errors.New("Can not deal the order side. ")
	}

	params := url.Values{}
	params.Set("symbol", order.Pair.ToSymbol("", true))
	params.Set("side", orderSide)
	params.Set("type", orderType)
	params.Set("quantity", fmt.Sprintf("%f", order.Amount))
	params.Set("newClientOrderId", order.Cid)

	switch order.OrderType {
	case NORMAL, ONLY_MAKER:
		params.Set("timeInForce", "GTC")
	case FOK:
		params.Set("timeInForce", "FOK")
	case IOC:
		params.Set("timeInForce", "IOC")
	default:
		params.Set("timeInForce", "GTC")
	}

	switch orderType {
	case "LIMIT":
		params.Set("price", fmt.Sprintf("%f", order.Price))
	}

	if err := margin.buildParamsSigned(&params); err != nil {
		return nil, err
	}

	response := remoteOrder{}
	resp, err := margin.DoRequest(
		http.MethodPost,
		uri,
		params.Encode(),
		&response,
	)
	if err != nil {
		return nil, err
	}

	if response.OrderId <= 0 {
		return nil, errors.New(string(resp))
	}
	response.merge(order, margin.config.Location)
	return resp, nil
}

func (margin *Margin) CancelOrder(order *Order) ([]byte, error) {
	if order.OrderId == "" {
		return nil, errors.New("You must get the order_id. ")
	}

	uri := "/sapi/v1/margin/order?"
	params := url.Values{}
	params.Set("symbol", order.Pair.ToSymbol("", true))
	params.Set("orderId", order.OrderId)
	if err := margin.buildParamsSigned(&params); err != nil {
		return nil, err
	}

	response := remoteOrder{}
	resp, err := margin.DoRequest(
		http.MethodDelete,
		uri,
		params.Encode(),
		&response,
	)

	if err != nil {
		return nil, err
	}
	response.merge(order, margin.config.Location)
	return resp, nil
}

func (margin *Margin) GetOrder(order *Order) ([]byte, error) {
	if order.OrderId == "" {
		return nil, errors.New("You must get the order_id. ")
	}

	params := url.Values{}
	params.Set("symbol", order.Pair.ToSymbol("", true))
	params.Set("orderId", order.OrderId)
	if err := margin.buildParamsSigned(&params); err != nil {
		return nil, err
	}

	uri := "/sapi/v1/margin/order?"
	response := remoteOrder{}
	resp, err := margin.DoRequest(
		http.MethodGet,
		uri,
		params.Encode(),
		&response,
	)
	if err != nil {
		return nil, err
	}
	if response.OrderId <= 0 {
		return nil, errors.New(string(resp))
	}
	response.merge(order, margin.config.Location)
	return resp, nil
}

// get all orders in desc.
func (margin *Margin) GetOrders(pair Pair) ([]*Order, []byte, error) {
	uri := "/sapi/v1/margin/allOrders?"
	params := url.Values{}
	params.Set("symbol", pair.ToSymbol("", true))
	if err := margin.buildParamsSigned(&params); err != nil {
		return nil, nil, err
	}

	rawOrders := make([]remoteOrder, 0)
	resp, err := margin.DoRequest(
		http.MethodGet,
		uri,
		params.Encode(),
		&rawOrders,
	)
	if err != nil {
		return nil, resp, err
	}

	orders := make([]*Order, 0)
	for _, r := range rawOrders {
		o := &Order{
			Pair: pair,
		}
		r.merge(o, margin.config.Location)
		orders = append(orders, o)
	}
	return orders, resp, nil
}

func (margin *Margin) GetUnFinishOrders(pair Pair) ([]*Order, []byte, error) {
	params := url.Values{}
	params.Set("symbol", pair.ToSymbol("", true))
	if err := margin.buildParamsSigned(&params); err != nil {
		return nil, nil, err
	}

	uri := "/sapi/v1/margin/openOrders?"
	remoteOrders := make([]*remoteOrder, 0)
	resp, err := margin.DoRequest(
		http.MethodGet,
		uri,
		params.Encode(),
		&remoteOrders,
	)
	if err != nil {
		return nil, nil, err
	}

	orders := make([]*Order, 0)
	for _, remoteOrder := range remoteOrders {
		order := Order{}
		remoteOrder.merge(&order, margin.config.Location)
		orders = append(orders, &order)
	}

	return orders, resp, nil
}

func (margin *Margin) GetAccount(pair Pair) (*MarginAccount, []byte, error) {
	uri := "/sapi/v1/margin/account?"
	params := url.Values{}
	if err := margin.buildParamsSigned(&params); err != nil {
		return nil, nil, err
	}

	remoteAsset := struct {
		UserAssets []*struct {
			Asset    string  `json:"asset"`
			Borrowed float64 `json:"borrowed,string"`
			Free     float64 `json:"free,string"`
			Interest float64 `json:"interest,string"`
			Locked   float64 `json:"locked,string"`
			NetAsset float64 `json:"netAsset,string"`
		} `json:"userAssets"`
	}{}

	resp, err := margin.DoRequest(
		http.MethodGet,
		uri,
		params.Encode(),
		&remoteAsset,
	)
	if err != nil {
		return nil, resp, err
	}

	marginAccount := &MarginAccount{
		Pair:       pair,
		SubAccount: make(map[string]MarginSubAccount, 0),
	}
	basisCurrency := strings.ToUpper(pair.Basis.Symbol)
	counterCurrency := strings.ToUpper(pair.Counter.Symbol)

	for _, asset := range remoteAsset.UserAssets {
		if asset.Asset == basisCurrency {
			marginAccount.SubAccount[basisCurrency] = MarginSubAccount{
				Currency:     pair.Basis,
				Amount:       asset.Borrowed + asset.NetAsset,
				AmountAvail:  asset.Free,
				AmountFrozen: asset.Locked,
				AmountLoaned: asset.Borrowed,
				LoaningFee:   asset.Interest,
			}
		} else if asset.Asset == counterCurrency {
			marginAccount.SubAccount[counterCurrency] = MarginSubAccount{
				Currency:     pair.Counter,
				Amount:       asset.Borrowed + asset.NetAsset,
				AmountAvail:  asset.Free,
				AmountFrozen: asset.Locked,
				AmountLoaned: asset.Borrowed,
				LoaningFee:   asset.Interest,
			}
		}
	}

	return marginAccount, resp, nil
}

func (margin *Margin) PlaceLoan(loan *Loan) ([]byte, error) {
	uri := "/sapi/v1/margin/loan?"
	params := url.Values{}
	params.Set("asset", loan.Currency.Symbol)
	params.Set("amount", fmt.Sprintf("%f", loan.Amount))
	if err := margin.buildParamsSigned(&params); err != nil {
		return nil, err
	}

	rawLoan := struct {
		TranId string `json:"tranId,int"`
	}{}
	resp, err := margin.DoRequest(
		http.MethodPost,
		uri,
		params.Encode(),
		&rawLoan,
	)
	if err != nil {
		return resp, err
	}
	now := time.Now()
	loan.LoanId = rawLoan.TranId
	loan.AmountLoaned = loan.Amount
	loan.LoanTimestamp = now.Unix() * 1000
	loan.LoanDate = now.In(margin.config.Location).Format(GO_BIRTHDAY)
	return resp, err
}

func (margin *Margin) GetLoan(loan *Loan) ([]byte, error) {
	panic("implement me")
}

func (margin *Margin) ReturnLoan(loan *Loan) ([]byte, error) {
	uri := "/sapi/v1/margin/repay?"

	params := url.Values{}
	params.Set("asset", loan.Currency.Symbol)
	params.Set("amount", fmt.Sprintf("%f", loan.AmountLoaned))

	rawLoan := struct {
		TranId string `json:"tranId,int"`
	}{}

	resp, err := margin.DoRequest(
		http.MethodPost,
		uri,
		params.Encode(),
		&rawLoan,
	)
	if err != nil {
		return resp, err
	}
	now := time.Now()
	loan.RepayId = rawLoan.TranId
	loan.RepayTimestamp = now.Unix() * 1000
	loan.RepayDate = now.In(margin.config.Location).Format(GO_BIRTHDAY)

	return resp, nil
}

//util api
func (margin *Margin) KeepAlive() {

}
