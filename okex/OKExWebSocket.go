package okex

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/gorilla/websocket"

	. "github.com/deforceHK/goghostex"
)

type OKExWSArg struct {
	Channel string `json:"channel"`
	InstId  string `json:"instId,omitempty"`
	InstType  string `json:"instType,omitempty"`
}

type OKExWSOp struct {
	Op   string   `json:"op"`
	Args []*OKExWSArg `json:"args"`
}


type OKExWSRes struct {
	Event string `json:"event"`
	Arg *OKExWSArg `json:"arg,omitempty"`
	Code string `json:"code,omitempty"`
	Msg string `json:"msg,omitempty"`
	Data json.RawMessage `json:"data,omitempty"`

}

//type OKExTradeWebSocket struct {
//	WS *okexWebSocket
//}
//
//type OKExMarketWebSocket struct {
//	*okexWebSocket
//}

type OKExTradeWebSocket struct {
	ws       *WsConn
	proxyUrl string

	msg     chan []byte
	receive func([]byte) error
}

func (this *OKExTradeWebSocket) Init() {
	// the default buffer channel
	if this.msg == nil {
		this.msg = make(chan []byte, 10)
	}

	if this.receive == nil {
		this.receive = func(data []byte) error {
			log.Println(string(data))
			return nil
		}
	}

	this.ws = NewWsBuilder().WsUrl(
		"wss://ws.okx.com:8443/ws/v5/public",
	).ProxyUrl(
		this.proxyUrl,
	).Heartbeat(
		[]byte("ping"),
		websocket.TextMessage,
		5*time.Second,
	).UnCompressFunc(
		FlateUnCompress,
	).ProtoHandleFunc(func(data []byte) error {
		this.ws.UpdateActiveTime()
		this.msg <- data
		return nil
	}).Build()

}

func (this *OKExTradeWebSocket) Login(config *APIConfig) error {

	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	sign, err := GetParamHmacSHA256Base64Sign(
		config.ApiSecretKey,
		fmt.Sprintf("%sGET/users/self/verify", timestamp),
	)

	if err != nil {
		return err
	}

	param := struct {
		Op   string   `json:"op"`
		Args []string `json:"args"`
	}{Op: "login", Args: []string{config.ApiKey, config.ApiPassphrase, timestamp, sign}}

	req, err := json.Marshal(param)
	if err != nil {
		return err
	}

	return this.ws.WriteMessage(websocket.TextMessage, req)
}

func (this *OKExTradeWebSocket) Subscribe(channel string) error {
	return this.ws.WriteMessage(websocket.TextMessage, []byte(channel))
}

func (this *OKExTradeWebSocket) Unsubscribe(channel string) error {
	return this.ws.WriteMessage(websocket.TextMessage, []byte(channel))
}

func (this *OKExTradeWebSocket) Start() {
	this.ws.ReceiveMessage()
	for {
		data := <-this.msg
		if err := this.receive(data); err != nil {
			log.Fatal(err)
			this.Close()
			return
		}
	}
}

func (this *OKExTradeWebSocket) Close() {
	this.ws.CloseWs()
}
