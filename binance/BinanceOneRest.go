package binance

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	. "github.com/deforceHK/goghostex"
)

type One struct {
	*Binance
	sync.Locker

	Infos map[string]*OneInfo
}

func (o *One) GetTicker(productId string) (*OneTicker, []byte, error) {
	//TODO implement me
	panic("implement me")
}

func (o *One) GetDepth(productId string, size int) (*OneDepth, []byte, error) {
	//TODO implement me
	panic("implement me")
}

func (o *One) GetInfos() ([]*OneInfo, []byte, error) {
	var wg = sync.WaitGroup{}
	wg.Add(2)

	var cmInfos, umInfos []*OneInfo
	var cmResp, umResp []byte
	var cmErr, umErr error

	go func() {
		cmInfos, cmResp, cmErr = o.GetCMInfos()
		wg.Done()
	}()
	go func() {
		umInfos, umResp, umErr = o.GetUMInfos()
		wg.Done()
	}()
	wg.Wait()

	if cmErr != nil {
		return nil, cmResp, cmErr
	}
	if umErr != nil {
		return nil, umResp, umErr
	}

	return append(cmInfos, umInfos...), append(cmResp, umResp...), nil
}

func (o *One) GetCMInfos() ([]*OneInfo, []byte, error) {

	var responseBasis = struct {
		ServerTime int64 `json:"serverTime"`
		Symbols    []struct {
			Symbol         string  `json:"symbol"`
			ContractType   string  `json:"contractType"`
			ContractSize   float64 `json:"contractSize"`
			ContractStatus string  `json:"contractStatus"`

			BaseAsset         string `json:"baseAsset"`
			CounterAsset      string `json:"quoteAsset"`
			MarginAsset       string `json:"marginAsset"`
			PricePrecision    int64  `json:"pricePrecision"`
			QuantityPrecision int64  `json:"quantityPrecision"`

			DeliveryDate int64                    `json:"deliveryDate"`
			OnboardDate  int64                    `json:"onboardDate"`
			Filters      []map[string]interface{} `json:"filters"`
		} `json:"symbols"`
	}{}

	var respBasis, errBasis = o.Swap.DoRequest(
		http.MethodGet,
		"/dapi/v1/exchangeInfo",
		"",
		&responseBasis,
		SETTLE_MODE_BASIS,
	)
	if errBasis != nil {
		return nil, nil, errBasis
	}

	var infos = make([]*OneInfo, 0)
	for _, symbol := range responseBasis.Symbols {
		var status = symbol.ContractStatus
		if status != "TRADING" && status != "PENDING_TRADING" {
			continue
		}
		if status == "PENDING_TRADING" {
			status = "PENDING"
		}

		infos = append(infos, &OneInfo{
			Pair: Pair{
				Basis:   NewCurrency(symbol.BaseAsset, ""),
				Counter: NewCurrency(symbol.CounterAsset, ""),
			},
			Status:                  status,
			ProductId:               symbol.Symbol,
			SettleMode:              1, // BASIS
			Exchange:                BINANCE,
			ContractValue:           symbol.ContractSize,
			ContractType:            symbol.ContractType,
			ContractStartTimestamp:  symbol.OnboardDate,
			ContractFinishTimestamp: symbol.DeliveryDate,

			PricePrecision:  symbol.PricePrecision,
			AmountPrecision: symbol.QuantityPrecision,
		})
	}

	return infos, respBasis, errBasis
}

func (o *One) GetUMInfos() ([]*OneInfo, []byte, error) {

	var responseCounter struct {
		Symbols []struct {
			Symbol       string `json:"symbol"`
			ContractType string `json:"contractType"`
			Status       string `json:"status"`
			BaseAsset    string `json:"baseAsset"`
			CounterAsset string `json:"quoteAsset"`
			MarginAsset  string `json:"marginAsset"`

			DeliveryDate int64 `json:"deliveryDate"`
			OnboardDate  int64 `json:"onboardDate"`

			PricePrecision    int64 `json:"pricePrecision"`
			QuantityPrecision int64 `json:"quantityPrecision"`

			Filters []map[string]interface{} `json:"filters"`
		} `json:"symbols"`
	}

	var respCounter, errCounter = o.Swap.DoRequest(
		http.MethodGet,
		"/fapi/v1/exchangeInfo",
		"",
		&responseCounter,
		SETTLE_MODE_COUNTER,
	)

	var infos = make([]*OneInfo, 0)
	for _, symbol := range responseCounter.Symbols {
		var status = symbol.Status
		if status != "TRADING" && status != "PENDING_TRADING" {
			continue
		}
		if status == "PENDING_TRADING" {
			status = "PENDING"
		}
		infos = append(infos, &OneInfo{
			Pair: Pair{
				NewCurrency(symbol.BaseAsset, ""),
				NewCurrency(symbol.CounterAsset, ""),
			},
			Status:     status,
			ProductId:  symbol.Symbol,
			SettleMode: 2, // COUNTER
			Exchange:   BINANCE,

			ContractValue:           1.0, // USDT-M合约的合约价值通常为1 USDT
			ContractType:            symbol.ContractType,
			ContractStartTimestamp:  symbol.OnboardDate,
			ContractFinishTimestamp: symbol.DeliveryDate,

			PricePrecision:  symbol.PricePrecision,
			AmountPrecision: symbol.QuantityPrecision,
		})
	}

	return infos, respCounter, errCounter
}

func (o *One) getInfo(productId string) (*OneInfo, error) {
	if len(o.Infos) == 0 {
		var infos, _, err = o.GetInfos()
		for i := 0; err != nil && i < 4; i++ {
			time.Sleep(time.Second * 10)
			infos, _, err = o.GetInfos()
		}

		if err != nil {
			return nil, err
		}

		o.Lock()
		for _, info := range infos {
			o.Infos[info.ProductId] = info
		}
		o.Unlock()
	}

	var info, exist = o.Infos[productId]
	if !exist {
		return nil, fmt.Errorf("info for product %s not found", productId)
	}

	return info, nil
}

func (o *One) PlaceOrder(order *OneOrder) ([]byte, error) {
	if order == nil {
		return nil, errors.New("order param is nil")
	}

	var info, err = o.getInfo(order.ProductId)
	if err != nil {
		return nil, fmt.Errorf("get info for product %s failed: %w", order.ProductId, err)
	}

	// 获取下单类型和方向
	var side, positionSide, placeType = "", "", ""
	var exist = false

	if side, exist = sideRelation[order.Type]; !exist {
		return nil, errors.New("swap type not found")
	}
	if positionSide, exist = positionSideRelation[order.Type]; !exist {
		return nil, errors.New("swap type not found")
	}
	if placeType, exist = placeTypeRelation[order.PlaceType]; !exist {
		return nil, errors.New("place type not found")
	}

	// 设置订单参数
	var param = url.Values{}
	param.Set("symbol", order.ProductId)
	param.Set("side", side)
	param.Set("positionSide", positionSide)

	// 处理订单类型
	if order.PlaceType == MARKET {
		param.Set("type", "MARKET")
		// 市价单以数量下单
		param.Set("quantity", FloatToString(order.Amount, info.AmountPrecision)) // 假设精度为8，实际应该从合约信息获取
	} else {
		param.Set("type", "LIMIT")
		param.Set("price", FloatToPrice(order.Price, 8, 0.0001)) // 假设精度为8，tickSize为0.0001
		param.Set("quantity", FloatToString(order.Amount, info.AmountPrecision))
		// 设置时效类型
		param.Set("timeInForce", placeType)
	}

	// 设置客户端订单ID
	if order.Cid != "" {
		param.Set("newClientOrderId", order.Cid)
	}

	// 添加签名
	if err := o.buildParamsSigned(&param); err != nil {
		return nil, err
	}

	// 定义响应结构
	var response struct {
		Cid        string  `json:"clientOrderId"`
		Status     string  `json:"status"`
		CumQuote   float64 `json:"cumQuote,string"`
		DealAmount float64 `json:"executedQty,string"`
		OrderId    int64   `json:"orderId"`
		UpdateTime int64   `json:"updateTime"`
		Price      float64 `json:"price,string"`
		Amount     float64 `json:"origQty,string"`
	}

	// 记录当前时间
	now := time.Now()

	var uri = "/papi/v1/um/order?"
	if order.SettleMode == 1 {
		uri = "/papi/v1/cm/order?"
	}
	// 发送API请求
	resp, err := o.DoRequest(
		http.MethodPost,
		uri+param.Encode(),
		"",
		&response,
	)

	if err != nil {
		return nil, err
	}

	// 更新订单信息
	orderTime := time.Unix(response.UpdateTime/1000, 0)
	order.OrderId = fmt.Sprintf("%d", response.OrderId)
	order.PlaceTimestamp = now.UnixNano() / int64(time.Millisecond)
	order.PlaceDatetime = now.In(o.config.Location).Format(GO_BIRTHDAY)
	order.DealTimestamp = response.UpdateTime
	order.DealDatetime = orderTime.In(o.config.Location).Format(GO_BIRTHDAY)

	// 更新订单状态
	if status, exists := statusRelation[response.Status]; exists {
		order.Status = status
	} else {
		order.Status = ORDER_UNFINISH
	}

	// 更新价格和数量信息
	order.Price = response.Price
	order.Amount = response.Amount

	// 计算成交均价
	if response.DealAmount > 0 {
		order.AvgPrice = response.CumQuote / response.DealAmount
		order.DealAmount = response.DealAmount
	}

	return resp, nil
}

func (o *One) CancelOrder(order *OneOrder) ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (o *One) GetOrder(order *OneOrder) ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (o *One) KeepAlive() {
	//TODO implement me
	panic("implement me")
}
