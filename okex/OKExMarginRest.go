package okex

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	. "github.com/strengthening/goghostex"
)

type Margin struct {
	*OKEx
}

func (this *Margin) Loan(record *LoanRecord) ([]byte, error) {
	var param = struct {
		InstrumentId string `json:"instrument_id"`
		Currency     string `json:"currency"`
		Amount       string `json:"amount"`
	}{
		InstrumentId: record.CurrencyPair.ToSymbol("-"),
		Currency:     record.Currency.Symbol,
		Amount:       FloatToString(record.Amount, 8),
	}

	reqBody, _, _ := this.BuildRequestBody(param)
	var response struct {
		BorrowId     string `json:"borrow_id"`
		ClientOid    string `json:"client_oid"`
		Result       bool   `json:"result"`
		ErrorCode    string `json:"code"`
		ErrorMessage string `json:"message"`
	}
	resp, err := this.DoRequest(
		"POST",
		"/api/margin/v3/accounts/borrow",
		reqBody,
		&response,
	)
	if err != nil {
		return nil, err
	}
	if response.ErrorMessage != "" {
		record.Status = LOAN_FAIL
		return nil, errors.New(string(resp))
	}
	record.LoanId = response.BorrowId
	record.Status = LOAN_FINISH
	record.AmountLoaned = record.Amount
	now := time.Now()
	record.LoanTimestamp = now.UnixNano() / int64(time.Millisecond)
	record.LoanDate = now.In(this.config.Location).Format(GO_BIRTHDAY)
	return resp, nil
}

func (this *Margin) GetOneLoan(record *LoanRecord) ([]byte, error) {
	if record.LoanId == "" {
		return nil, errors.New("The loan_id can not be empty! ")
	}
	// retry 5 times max.
	return this.getOneLoan(record, 0, 0)
}

func (this *Margin) getOneLoan(record *LoanRecord, from int64, retry int) ([]byte, error) {
	if retry > 5 {
		return nil, errors.New("retry too many times to find the loan record")
	}

	params := url.Values{}
	params.Add("instrument_id", record.CurrencyPair.ToSymbol("-"))
	params.Add("status", "0") // find the loan not repay
	if from != 0 {
		params.Add("from", fmt.Sprintf("%d", from))
	}

	uri := fmt.Sprintf("/api/margin/v3/accounts/%s/borrowed?", record.CurrencyPair.ToSymbol("-"))
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

	if resp, err := this.DoRequest(
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
			if remoteRecord.BorrowId == record.LoanId {
				record.AmountInterest = remoteRecord.Interest
				t, _ := time.Parse(time.RFC3339, remoteRecord.ForceRepayTime)
				record.RepayDeadlineDate = t.In(this.config.Location).Format(GO_BIRTHDAY)
				t, _ = time.Parse(time.RFC3339, remoteRecord.CreatedAt)
				record.LoanTimestamp = t.UnixNano() / int64(time.Millisecond)
				record.LoanDate = t.In(this.config.Location).Format(GO_BIRTHDAY)
				record.Amount = remoteRecord.Amount
				record.AmountLoaned = remoteRecord.Amount
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
		return this.getOneLoan(record, minLoanId, retry+1)
	}
}

func (this *Margin) Repay(record *LoanRecord) ([]byte, error) {

	urlPath := "/api/margin/v3/accounts/repayment"
	param := struct {
		BorrowId     string `json:"borrow_id,omitempty"`
		InstrumentId string `json:"instrument_id"`
		Currency     string `json:"currency"`
		Amount       string `json:"amount"`
	}{
		record.LoanId,
		record.CurrencyPair.ToSymbol("-"),
		record.Currency.Symbol,
		FloatToString(record.AmountLoaned+record.AmountInterest, 8),
	}

	reqBody, _, _ := this.BuildRequestBody(param)
	var response struct {
		RepaymentId string `json:"repayment_id"`
		Result      bool   `json:"result"`
		Code        string `json:"code"`
		Message     string `json:"message"`
	}
	resp, err := this.DoRequest("POST", urlPath, reqBody, &response)
	if err != nil {
		return nil, err
	}

	if !response.Result {
		return nil, errors.New(string(resp))
	}

	now := time.Now()
	record.Status = LOAN_REPAY
	record.RepayId = response.RepaymentId
	record.RepayDate = now.In(this.config.Location).Format(GO_BIRTHDAY)
	record.RepayTimestamp = now.UnixNano() / int64(time.Millisecond)
	return resp, nil
}

func (this *Margin) PlaceMarginOrder(order *Order) ([]byte, error) {

	param := PlaceOrderParam{
		InstrumentId:  order.Currency.AdaptUsdToUsdt().ToLower().ToSymbol("-"),
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
	jsonStr, _, _ := this.BuildRequestBody(param)
	resp, err := this.DoRequest(
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

func (this *Margin) CancelMarginOrder(order *Order) ([]byte, error) {

	uri := "/api/margin/v3/cancel_orders/" + order.OrderId
	param := struct {
		InstrumentId string `json:"instrument_id"`
	}{
		order.Currency.AdaptUsdToUsdt().ToLower().ToSymbol("-"),
	}
	reqBody, _, _ := this.BuildRequestBody(param)
	var response struct {
		ClientOid string `json:"client_oid"`
		OrderId   string `json:"order_id"`
		Result    bool   `json:"result"`
	}

	resp, err := this.DoRequest(
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

func (this *Margin) GetMarginOneOrder(order *Order) ([]byte, error) {
	uri := "/api/margin/v3/orders/" + order.OrderId + "?instrument_id=" + order.Currency.AdaptUsdToUsdt().ToSymbol("-")
	var response OrderResponse
	resp, err := this.DoRequest(
		"GET",
		uri,
		"",
		&response,
	)

	if err != nil {
		return nil, err
	}

	if err := this.adaptOrder(order, &response); err != nil {
		return nil, err
	}
	return resp, nil
}

func (this *Margin) adaptOrder(order *Order, response *OrderResponse) error {

	order.Cid = response.ClientOid
	order.OrderId = response.OrderId
	order.Price = response.Price
	order.Amount = response.Size
	order.AvgPrice = ToFloat64(response.PriceAvg)
	order.DealAmount = ToFloat64(response.FilledSize)
	order.Status = this.adaptOrderState(response.State)

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
		order.OrderDate = date.In(this.config.Location).Format(GO_BIRTHDAY)
		return nil
	}
}

func (this *Margin) GetMarginUnFinishOrders(currency CurrencyPair) ([]Order, []byte, error) {
	uri := fmt.Sprintf(
		"/api/margin/v3/orders_pending?instrument_id=%s",
		currency.AdaptUsdToUsdt().ToSymbol("-"),
	)
	var response []OrderResponse
	resp, err := this.DoRequest(
		"GET",
		uri,
		"",
		&response,
	)
	if err != nil {
		return nil, nil, err
	}

	var orders []Order
	for _, itm := range response {
		order := Order{Currency: currency}
		if err := this.adaptOrder(&order, &itm); err != nil {
			return nil, nil, err
		}
		orders = append(orders, order)
	}

	return orders, resp, nil
}

func (this *Margin) GetMarginAccount(pair CurrencyPair) (*MarginAccount, []byte, error) {

	uri := fmt.Sprintf("/api/margin/v3/accounts/%s", pair.ToSymbol("-"))
	var response map[string]interface{}
	resp, err := this.DoRequest(
		"GET",
		uri,
		"",
		&response,
	)
	if err != nil {
		return nil, nil, err
	}

	acc := MarginAccount{
		CurrencyPair: pair,
	}
	acc.SubAccount = make(map[Currency]MarginSubAccount, 0)
	acc.LiquidationPrice = ToFloat64(response["liquidation_price"])
	acc.RiskRate = ToFloat64(response["risk_rate"])
	acc.MarginRatio = ToFloat64(response["margin_ratio"])

	for k, v := range response {
		if strings.Contains(k, "currency") {
			c := NewCurrency(strings.Split(k, ":")[1], "")
			info := v.(map[string]interface{})
			acc.SubAccount[c] = MarginSubAccount{
				Currency:      c,
				BalanceTotal:  ToFloat64(info["balance"]),
				BalanceFrozen: ToFloat64(info["frozen"]),
				BalanceAvail:  ToFloat64(info["available"]),
				Loaned:        ToFloat64(info["borrowed"]),
				LoaningFee:    ToFloat64(info["lending_fee"])}
		}
	}

	return &acc, resp, nil
}

func (this *Margin) GetMarginInfo(pair CurrencyPair) ([]byte, error) {
	uri := fmt.Sprintf(
		"/api/margin/v3/accounts/%s/availability",
		pair.ToSymbol("-"),
	)
	var response []map[string]interface{}

	resp, err := this.DoRequest(
		"GET",
		uri,
		"",
		&response,
	)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (this *Margin) GetMarginTicker(pair CurrencyPair) (*Ticker, []byte, error) {
	return this.Spot.GetTicker(pair)
}

func (this *Margin) GetMarginDepth(size int, pair CurrencyPair) (*Depth, []byte, error) {
	return this.Spot.GetDepth(size, pair)
}

func (this *Margin) GetMarginKlineRecords(pair CurrencyPair, period, size, since int) ([]Kline, []byte, error) {
	return this.Spot.GetKlineRecords(pair, period, size, since)
}
