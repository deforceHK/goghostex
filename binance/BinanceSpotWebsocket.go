package binance

import (
	"fmt"
	"log"
	"time"

	"github.com/gorilla/websocket"

	. "github.com/strengthening/goghostex"
)

type BinanceWebsocket struct {
	ws       *WsConn
	wsBuild  *WsBuilder
	wsUrl    string
	proxyUrl string

	msg     chan []byte
	receive func([]byte) error
}

func (this *BinanceWebsocket) Init() {
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

	if this.wsUrl == "" {
		this.wsUrl = "wss://stream.binance.com:9443/stream?streams="
	}

	this.wsBuild = NewWsBuilder().Dump().WsUrl(
		this.wsUrl,
	).ProxyUrl(
		this.proxyUrl,
	).Heartbeat(
		[]byte("--heartbeat--"),
		websocket.PongMessage,
		5*time.Second,
	).UnCompressFunc(
		FlateUnCompress,
	).ProtoHandleFunc(func(data []byte) error {
		this.ws.UpdateActiveTime()
		this.msg <- data
		return nil
	})

}

func (this *BinanceWebsocket) Subscribe(channel string) error {
	this.wsUrl += fmt.Sprintf("%s/", channel)
	this.wsBuild.WsUrl(this.wsUrl[:(len(this.wsUrl) - 1)])
	return nil
}

func (this *BinanceWebsocket) Start() {
	this.ws = this.wsBuild.Build()
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

func (this *BinanceWebsocket) Close() {
	this.ws.CloseWs()
}
