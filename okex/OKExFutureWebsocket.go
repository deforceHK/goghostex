package okex

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/gorilla/websocket"

	. "github.com/deforceHK/goghostex"
)

type OKexFutureWebsocket struct {
	ws       *WsConn
	proxyUrl string

	msg     chan []byte
	receive func([]byte) error
}

func (this *OKexFutureWebsocket) Init() {
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

	this.ws = NewWsBuilder().Dump().WsUrl(
		"wss://real.okex.com:8443/ws/v3",
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

func (this *OKexFutureWebsocket) Login(config *APIConfig) error {

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

func (this *OKexFutureWebsocket) Subscribe(channel string) error {
	return this.ws.WriteMessage(websocket.TextMessage, []byte(channel))
}

func (this *OKexFutureWebsocket) Unsubscribe(channel string) error {
	return this.ws.WriteMessage(websocket.TextMessage, []byte(channel))
}

func (this *OKexFutureWebsocket) Start() {
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

func (this *OKexFutureWebsocket) Close() {
	this.ws.CloseWs()
}
