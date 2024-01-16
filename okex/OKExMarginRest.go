package okex

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	. "github.com/deforceHK/goghostex"
)

type Margin struct {
	*OKEx
}

// public api
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
	return nil, nil, nil
}

// private api
func (margin *Margin) GetAccount(pair Pair) (*MarginAccount, []byte, error) {

	uri := fmt.Sprintf("/api/margin/v3/accounts/%s", pair.ToSymbol("-", true))
	var response map[string]interface{}
	resp, err := margin.DoRequest(
		"GET",
		uri,
		"",
		&response,
	)
	if err != nil {
		return nil, nil, err
	}

	acc := MarginAccount{
		Pair: pair,
	}
	acc.SubAccount = make(map[string]MarginSubAccount, 0)
	acc.LiquidationPrice = ToFloat64(response["liquidation_price"])
	acc.RiskRate = ToFloat64(response["risk_rate"])
	acc.MarginRatio = ToFloat64(response["margin_ratio"])

	for k, v := range response {
		if strings.Contains(k, "currency") {
			c := NewCurrency(strings.Split(k, ":")[1], "")
			info := v.(map[string]interface{})
			acc.SubAccount[strings.ToUpper(c.Symbol)] = MarginSubAccount{
				Currency:     c,
				Amount:       ToFloat64(info["balance"]),
				AmountFrozen: ToFloat64(info["frozen"]),
				AmountAvail:  ToFloat64(info["available"]),
				AmountLoaned: ToFloat64(info["borrowed"]),
				LoaningFee:   ToFloat64(info["lending_fee"])}
		}
	}

	return &acc, resp, nil
}

func (margin *Margin) PlaceOrder(order *Order) ([]byte, error) {

	param := PlaceOrderParam{
		InstrumentId:  order.Pair.ToSymbol("-", true),
		MarginTrading: "2",
	}

	if order.Cid == "" {
		order.Cid = UUID()
	}
	param.ClientOid = order.Cid

	switch order.Side {
	case BUY, SELL:
		param.Side = strings.ToLower(order.Side.String())
		param.Price = order.Price
		param.Size = order.Amount
		param.Type = "limit"
		param.OrderType = _INTERNAL_ORDER_TYPE_CONVERTER[order.OrderType]
	case SELL_MARKET:
		param.Side = "sell"
		param.Type = "market"
		param.Size = order.Amount
	case BUY_MARKET:
		param.Side = "buy"
		param.Type = "market"
		param.Notional = order.Price
	default:
		param.Size = order.Amount
		param.Price = order.Price
	}

	response := remoteOrder{}
	jsonStr, _, _ := margin.BuildRequestBody(param)
	resp, err := margin.DoRequest(
		"POST",
		"/api/margin/v3/orders",
		jsonStr,
		&response,
	)
	if err != nil {
		return nil, err
	}
	if err := response.Merge(order); err != nil {
		return nil, err
	}
	return resp, nil
}

func (margin *Margin) CancelOrder(order *Order) ([]byte, error) {

	uri := "/api/margin/v3/cancel_orders/" + order.OrderId
	param := struct {
		InstrumentId string `json:"instrument_id"`
	}{
		order.Pair.ToSymbol("-", true),
	}
	reqBody, _, _ := margin.BuildRequestBody(param)
	var response struct {
		ClientOid string `json:"client_oid"`
		OrderId   string `json:"order_id"`
		Result    bool   `json:"result"`
	}

	resp, err := margin.DoRequest(
		"POST",
		uri,
		reqBody,
		&response,
	)
	if err != nil {
		return nil, err
	}
	if response.Result {
		return resp, nil
	}
	return resp, NewError(400, "cancel fail, unknown error")
}

func (margin *Margin) GetOrder(order *Order) ([]byte, error) {
	uri := "/api/margin/v3/orders/" + order.OrderId + "?instrument_id=" + order.Pair.ToSymbol("-", true)
	var response OrderResponse
	resp, err := margin.DoRequest(
		"GET",
		uri,
		"",
		&response,
	)

	if err != nil {
		return nil, err
	}

	if err := margin.adaptOrder(order, &response); err != nil {
		return nil, err
	}
	return resp, nil
}

func (margin *Margin) GetOrders(pair Pair) ([]*Order, error) {
	panic("impl me. ")
}

func (margin *Margin) GetUnFinishOrders(pair Pair) ([]*Order, []byte, error) {
	uri := fmt.Sprintf(
		"/api/margin/v3/orders_pending?instrument_id=%s",
		pair.ToSymbol("-", true),
	)
	var response []OrderResponse
	resp, err := margin.DoRequest(
		"GET",
		uri,
		"",
		&response,
	)
	if err != nil {
		return nil, nil, err
	}

	var orders []*Order
	for _, itm := range response {
		order := Order{Pair: pair}
		if err := margin.adaptOrder(&order, &itm); err != nil {
			return nil, nil, err
		}
		orders = append(orders, &order)
	}

	return orders, resp, nil
}

func (margin *Margin) adaptOrder(order *Order, response *OrderResponse) error {

	order.Cid = response.ClientOid
	order.OrderId = response.OrderId
	order.Price = response.Price
	order.Amount = response.Size
	order.AvgPrice = ToFloat64(response.PriceAvg)
	order.DealAmount = ToFloat64(response.FilledSize)
	order.Status = margin.adaptOrderState(response.State)

	switch response.Side {
	case "buy":
		if response.Type == "market" {
			order.Side = BUY_MARKET
			order.DealAmount = ToFloat64(response.Notional) //成交金额
		} else {
			order.Side = BUY
		}
	case "sell":
		if response.Type == "market" {
			order.Side = SELL_MARKET
			order.DealAmount = ToFloat64(response.Notional) //成交数量
		} else {
			order.Side = SELL
		}
	}

	switch response.OrderType {
	case 0:
		order.OrderType = NORMAL
	case 1:
		order.OrderType = ONLY_MAKER
	case 2:
		order.OrderType = FOK
	case 3:
		order.OrderType = IOC
	default:
		order.OrderType = NORMAL
	}

	if date, err := time.Parse(time.RFC3339, response.Timestamp); err != nil {
		return err
	} else {
		order.OrderTimestamp = date.UnixNano() / int64(time.Millisecond)
		order.OrderDate = date.In(margin.config.Location).Format(GO_BIRTHDAY)
		return nil
	}
}

func (margin *Margin) PlaceLoan(loan *Loan) ([]byte, error) {
	var param = struct {
		InstrumentId string `json:"instrument_id"`
		Currency     string `json:"currency"`
		Amount       string `json:"amount"`
	}{
		InstrumentId: loan.Pair.ToSymbol("-", true),
		Currency:     loan.Currency.Symbol,
		Amount:       FloatToString(loan.Amount, 8),
	}

	reqBody, _, _ := margin.BuildRequestBody(param)
	var response struct {
		BorrowId     string `json:"borrow_id"`
		ClientOid    string `json:"client_oid"`
		Result       bool   `json:"result"`
		ErrorCode    string `json:"code"`
		ErrorMessage string `json:"message"`
	}
	resp, err := margin.DoRequest(
		"POST",
		"/api/margin/v3/accounts/borrow",
		reqBody,
		&response,
	)
	if err != nil {
		return nil, err
	}
	if response.ErrorMessage != "" {
		loan.Status = LOAN_FAIL
		return nil, errors.New(string(resp))
	}
	loan.LoanId = response.BorrowId
	loan.Status = LOAN_FINISH
	loan.AmountLoaned = loan.Amount
	now := time.Now()
	loan.LoanTimestamp = now.UnixNano() / int64(time.Millisecond)
	loan.LoanDate = now.In(margin.config.Location).Format(GO_BIRTHDAY)
	return resp, nil
}

func (margin *Margin) GetLoan(loan *Loan) ([]byte, error) {
	if loan.LoanId == "" {
		return nil, errors.New("The loan_id can not be empty! ")
	}
	// retry 5 times max.
	return margin.getLoan(loan, 0, 0)
}

func (margin *Margin) getLoan(loan *Loan, from int64, retry int) ([]byte, error) {
	if retry > 5 {
		return nil, errors.New("retry too many times to find the loan record")
	}

	params := url.Values{}
	params.Add("instrument_id", loan.Pair.ToSymbol("-", true))
	params.Add("status", "0") // find the loan not repay
	if from != 0 {
		params.Add("from", fmt.Sprintf("%d", from))
	}

	uri := fmt.Sprintf("/api/margin/v3/accounts/%s/borrowed?", loan.Pair.ToSymbol("-", true))
	remoteRecords := make([]struct {
		BorrowId       string  `json:"borrow_id"`
		ClientOid      string  `json:"client_oid"`
		Result         bool    `json:"result"`
		ErrorCode      string  `json:"code"`
		ErrorMessage   string  `json:"message"`
		Interest       float64 `json:"interest,string"`
		Amount         float64 `json:"amount,string"`
		ForceRepayTime string  `json:"force_repay_time"`
		CreatedAt      string  `json:"created_at"`
	}, 0)

	if resp, err := margin.DoRequest(
		"GET",
		uri+params.Encode(),
		"",
		&remoteRecords,
	); err != nil {
		return nil, err
	} else if len(remoteRecords) == 0 {
		return nil, errors.New("Can not find the borrowId. ")
	} else {
		minLoanId, err := strconv.ParseInt(remoteRecords[0].BorrowId, 10, 64)
		if err != nil {
			return nil, err
		}

		for _, remoteRecord := range remoteRecords {
			if remoteRecord.BorrowId == loan.LoanId {
				loan.AmountInterest = remoteRecord.Interest
				t, _ := time.Parse(time.RFC3339, remoteRecord.ForceRepayTime)
				loan.RepayDeadlineDate = t.In(margin.config.Location).Format(GO_BIRTHDAY)
				t, _ = time.Parse(time.RFC3339, remoteRecord.CreatedAt)
				loan.LoanTimestamp = t.UnixNano() / int64(time.Millisecond)
				loan.LoanDate = t.In(margin.config.Location).Format(GO_BIRTHDAY)
				loan.Amount = remoteRecord.Amount
				loan.AmountLoaned = remoteRecord.Amount
				return resp, nil
			} else {
				if loanId, err := strconv.ParseInt(remoteRecord.BorrowId, 10, 64); err != nil {
					return nil, err
				} else {
					if loanId < minLoanId {
						minLoanId = loanId
					}
				}
			}
		}
		return margin.getLoan(loan, minLoanId, retry+1)
	}
}

func (margin *Margin) ReturnLoan(loan *Loan) ([]byte, error) {

	urlPath := "/api/margin/v3/accounts/repayment"
	param := struct {
		BorrowId     string `json:"borrow_id,omitempty"`
		InstrumentId string `json:"instrument_id"`
		Currency     string `json:"currency"`
		Amount       string `json:"amount"`
	}{
		loan.LoanId,
		loan.Pair.ToSymbol("-", true),
		loan.Currency.Symbol,
		FloatToString(loan.AmountLoaned+loan.AmountInterest, 8),
	}

	reqBody, _, _ := margin.BuildRequestBody(param)
	var response struct {
		RepaymentId string `json:"repayment_id"`
		Result      bool   `json:"result"`
		Code        string `json:"code"`
		Message     string `json:"message"`
	}
	resp, err := margin.DoRequest("POST", urlPath, reqBody, &response)
	if err != nil {
		return nil, err
	}

	if !response.Result {
		return nil, errors.New(string(resp))
	}

	now := time.Now()
	loan.Status = LOAN_REPAY
	loan.RepayId = response.RepaymentId
	loan.RepayDate = now.In(margin.config.Location).Format(GO_BIRTHDAY)
	loan.RepayTimestamp = now.UnixNano() / int64(time.Millisecond)
	return resp, nil
}

// util api
func (margin *Margin) KeepAlive() {
	nowTimestamp := time.Now().Unix() * 1000
	if (nowTimestamp - margin.config.LastTimestamp) < 5*1000 {
		return
	}
	_, _, _ = margin.Spot.GetTicker(Pair{Basis: BTC, Counter: USDT})
}
