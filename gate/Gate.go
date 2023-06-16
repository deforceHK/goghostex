package gate

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	. "github.com/deforceHK/goghostex"
)

const (
	ACCEPT       = "Accept"
	CONTENT_TYPE = "Content-Type"

	APPLICATION_JSON      = "application/json"
	APPLICATION_JSON_UTF8 = "application/json; charset=UTF-8"

	ENDPOINT = "https://api.gateio.ws"
)

var _INERNAL_KLINE_PERIOD_CONVERTER = map[int]string{
	KLINE_PERIOD_1MIN:  "1m",
	KLINE_PERIOD_5MIN:  "5m",
	KLINE_PERIOD_15MIN: "15m",
	KLINE_PERIOD_30MIN: "30m",
	KLINE_PERIOD_60MIN: "1h",
	KLINE_PERIOD_4H:    "4h",
	KLINE_PERIOD_8H:    "8h",
	KLINE_PERIOD_1DAY:  "1d",
}

var GATE_PLACE_TYPE_CONVERTER = map[PlaceType]string{
	NORMAL:     "gtc",
	IOC:        "ioc",
	ONLY_MAKER: "poc",
}

var GATE_PLACE_TYPE_REVERTER = map[string]PlaceType{
	"gtc": NORMAL,
	"ioc": IOC,
	"poc": ONLY_MAKER,
}

type SwapOrderGate struct {
	Id           int64   `json:"id,omitempty"`
	User         int64   `json:"user,omitempty"`
	Contract     string  `json:"contract"`
	CreateTime   int64   `json:"create_time,omitempty"`
	Size         int64   `json:"size"`
	Left         int64   `json:"left,omitempty"`
	Price        float64 `json:"price,string"`
	FillPrice    float64 `json:"fill_price,string"`
	Status       string  `json:"status,omitempty"`
	Close        bool    `json:"close"`
	ReduceOnly   bool    `json:"reduce_only"`
	IsReduceOnly bool    `json:"is_reduce_only,omitempty"`
	Tif          string  `json:"tif"`
	Text         string  `json:"text"`
	FinishTime   int64   `json:"finish_time"`
	FinishAt     string  `json:"finish_at"`
}

func (sog *SwapOrderGate) Merge(order *SwapOrder) {
	placeType, exist := GATE_PLACE_TYPE_CONVERTER[order.PlaceType]
	if !exist {
		panic("not support the place type in gate. ")
	}

	sog.Contract = order.Pair.ToSymbol("_", true)
	sog.Size = int64(order.Price*order.Amount) / 1
	sog.Price = order.Price
	sog.Tif = placeType
	if order.Type == LIQUIDATE_LONG || order.Type == LIQUIDATE_SHORT {
		sog.ReduceOnly = true
	}
	if order.Type == LIQUIDATE_LONG || order.Type == OPEN_SHORT {
		sog.Size = -sog.Size
	}
}

func (sog *SwapOrderGate) New(loc *time.Location) *SwapOrder {
	placeTimestamp := sog.CreateTime * 1000
	placeDatetime := time.Unix(sog.CreateTime, 0).In(loc).Format(GO_BIRTHDAY)

	finishTimestamp := sog.FinishTime * 1000
	finishDatetime := time.Unix(sog.FinishTime, 0).In(loc).Format(GO_BIRTHDAY)

	status := ORDER_UNFINISH
	if sog.Status == "finished" {
		if sog.Left == 0 {
			status = ORDER_FINISH
		} else {
			status = ORDER_CANCEL
		}
	}

	orderType := OPEN_LONG
	positiveSize, positiveLeft := sog.Size, sog.Left
	if sog.Size < 0 {
		positiveSize = -sog.Size
	}
	if sog.Left < 0 {
		positiveLeft = -sog.Left
	}

	if sog.IsReduceOnly && sog.Size > 0 {
		orderType = LIQUIDATE_SHORT
	} else if sog.IsReduceOnly && sog.Size < 0 {
		orderType = LIQUIDATE_LONG
	} else if !sog.IsReduceOnly && sog.Size > 0 {
		orderType = OPEN_LONG
	} else if !sog.IsReduceOnly && sog.Size < 0 {
		orderType = OPEN_SHORT
	}

	sOrder := SwapOrder{
		Cid:     sog.Text,
		OrderId: fmt.Sprintf("%d", sog.Id),

		Amount:     float64(positiveSize),
		DealAmount: float64(positiveSize - positiveLeft),
		Price:      sog.Price,
		AvgPrice:   sog.FillPrice,

		PlaceTimestamp: placeTimestamp,
		PlaceDatetime:  placeDatetime,

		DealTimestamp: finishTimestamp,
		DealDatetime:  finishDatetime,
		Status:        status,
		PlaceType:     GATE_PLACE_TYPE_REVERTER[sog.Tif],

		Type:       orderType,
		MarginType: "",
		LeverRate:  0,
		Fee:        0,
		Pair:       NewPair(sog.Contract, "_"),
		Exchange:   GATE,
	}

	return &sOrder
}

type Gate struct {
	config *APIConfig
	Spot   *Spot
	Swap   *Swap
	//Future *Future
	//Margin *Margin
	//Wallet *Wallet
}

func (*Gate) GetExchangeName() string {
	return GATE
}

func New(config *APIConfig) *Gate {
	gate := &Gate{config: config}
	gate.Spot = &Spot{gate}
	gate.Swap = &Swap{gate}
	return gate
}

func (gate *Gate) DoRequest(
	httpMethod,
	uri,
	rawQuery string,
	reqBody string,
	response interface{},
) ([]byte, error) {
	url := ENDPOINT + uri
	if rawQuery != "" {
		url += fmt.Sprintf("?%s", rawQuery)
	}

	resp, err := NewHttpRequest(
		gate.config.HttpClient,
		httpMethod,
		url,
		reqBody,
		map[string]string{
			CONTENT_TYPE: APPLICATION_JSON_UTF8,
			ACCEPT:       APPLICATION_JSON,
		},
	)

	if err != nil {
		return nil, err
	} else {
		nowTimestamp := time.Now().Unix() * 1000
		if nowTimestamp > gate.config.LastTimestamp {
			gate.config.LastTimestamp = nowTimestamp
		}
		return resp, json.Unmarshal(resp, &response)
	}
}

func (gate *Gate) DoSignRequest(
	httpMethod,
	uri,
	rawQuery string,
	reqBody string,
	response interface{},
) ([]byte, error) {
	h := sha512.New()
	if reqBody != "" {
		h.Write([]byte(reqBody))
	}
	hashedPayload := hex.EncodeToString(h.Sum(nil))

	nowTS := strconv.FormatInt(time.Now().Unix(), 10)
	msg := fmt.Sprintf("%s\n%s\n%s\n%s\n%s", httpMethod, uri, rawQuery, hashedPayload, nowTS)
	mac := hmac.New(sha512.New, []byte(gate.config.ApiSecretKey))
	mac.Write([]byte(msg))

	sign := hex.EncodeToString(mac.Sum(nil))
	url := ENDPOINT + uri
	if rawQuery != "" {
		url += fmt.Sprintf("?%s", rawQuery)
	}

	resp, err := NewHttpRequest(
		gate.config.HttpClient,
		httpMethod,
		url,
		reqBody,
		map[string]string{
			"KEY":        gate.config.ApiKey,
			"SIGN":       sign,
			"Timestamp":  nowTS,
			CONTENT_TYPE: APPLICATION_JSON_UTF8,
			ACCEPT:       APPLICATION_JSON,
		},
	)

	if err != nil {
		return resp, err
	} else {
		nowTimestamp := time.Now().Unix() * 1000
		if nowTimestamp > gate.config.LastTimestamp {
			gate.config.LastTimestamp = nowTimestamp
		}
		return resp, json.Unmarshal(resp, &response)
	}
}
