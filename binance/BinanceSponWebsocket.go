package binance

import (
	"github.com/gorilla/websocket"
	"log"
	"time"

	. "github.com/strengthening/goghostex"
)

type BinanceFutureWebsocket struct {
	ws       *WsConn
	proxyUrl string

	msg     chan []byte
	receive func([]byte) error
}

func (this *BinanceFutureWebsocket) Init() {

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
		"wss://stream.binance.com:9443/ws/bnbbtc@depth",
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
	}).Build()

}

func (this *BinanceFutureWebsocket) Start() {

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

func (this *BinanceFutureWebsocket) Close() {
	this.ws.CloseWs()
}
