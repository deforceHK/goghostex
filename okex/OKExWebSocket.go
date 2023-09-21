package okex

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/gorilla/websocket"

	. "github.com/deforceHK/goghostex"
)

type WSArgOKEx struct {
	Channel  string `json:"channel"`
	InstId   string `json:"instId,omitempty"`
	InstType string `json:"instType,omitempty"`
}

type WSOpOKEx struct {
	Op   string              `json:"op"`
	Args []map[string]string `json:"args"`
}

type WSResOKEx struct {
	Event string          `json:"event"`
	Arg   *WSArgOKEx      `json:"arg,omitempty"`
	Code  string          `json:"code,omitempty"`
	Msg   string          `json:"msg,omitempty"`
	Data  json.RawMessage `json:"data,omitempty"`
}

type WSTradeOKEx struct {
	RecvHandler  func(string)
	ErrorHandler func(error)
	Config     *APIConfig

	conn       *websocket.Conn
	lastPingTS int64

	stopPingSign chan bool
	stopRecvSign chan bool
}

func (this *WSTradeOKEx) Subscribe(v interface{}) {
	this.conn.WriteJSON(v)
}

func (this *WSTradeOKEx) Unsubscribe(v interface{}) {
	this.conn.WriteJSON(v)
}

func (this *WSTradeOKEx) Start() {
	if this.RecvHandler == nil {
		this.RecvHandler = func(msg string) {
			log.Println(msg)
		}
	}
	if this.ErrorHandler == nil {
		this.ErrorHandler = func(err error) {
			log.Fatalln(err)
		}
	}

	var conn, _, err = websocket.DefaultDialer.Dial(
		"wss://ws.okx.com:8443/ws/v5/private",
		nil,
	)
	if err != nil {
		this.ErrorHandler(err)
		return
	}
	this.conn = conn

	var ts = fmt.Sprintf("%d", time.Now().Unix())
	var sign, _ = GetParamHmacSHA256Base64Sign(
		this.Config.ApiSecretKey,
		fmt.Sprintf("%sGET/users/self/verify", ts),
	)
	var login = WSOpOKEx{
		Op: "login",
		Args: []map[string]string{
			{
				"apiKey":     this.Config.ApiKey,
				"passphrase": this.Config.ApiPassphrase,
				"timestamp":  ts,
				"sign":       sign,
			},
		},
	}

	this.conn.WriteJSON(login)
	var messageType, p, readErr = this.conn.ReadMessage()
	if readErr != nil {
		this.ErrorHandler(readErr)
		return
	}
	if messageType != websocket.TextMessage {
		this.ErrorHandler(fmt.Errorf("message type error"))
		return
	}

	var result = struct {
		Event string `json:"event"`
		Code  string `json:"code"`
		Msg   string `json:"msg"`
	}{}

	var jsonErr = json.Unmarshal(p, &result)
	if jsonErr != nil {
		this.ErrorHandler(jsonErr)
		return
	}
	if result.Code != "0" {
		this.ErrorHandler(fmt.Errorf("login error: %s", result.Msg))
		return
	}

	go this.pingRoutine()
	go this.recvRoutine()
	log.Println("okex trade websocket start")

}

func (this *WSTradeOKEx) pingRoutine() {
	this.stopPingSign = make(chan bool, 1)
	var ticker = time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := this.conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				this.ErrorHandler(err)
				this.Stop()
			}
		case <-this.stopPingSign:
			return
		}
	}
}

func (this *WSTradeOKEx) recvRoutine() {
	this.stopRecvSign = make(chan bool, 1)
	var ticker = time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// 超过5分钟没有收到消息，关闭连接
			if time.Now().Unix()-this.lastPingTS > 5*60 {
				this.ErrorHandler(fmt.Errorf("ping timeout"))
				this.Stop()
			}
		case <-this.stopRecvSign:
			return
		default:
			var msgType, msg, readErr = this.conn.ReadMessage()
			if readErr != nil {
				this.ErrorHandler(readErr)
				this.Stop()
			}

			if msgType != websocket.TextMessage {
				continue
			}

			this.lastPingTS = time.Now().Unix()
			var msgStr = string(msg)
			if msgStr != "pong" {
				this.RecvHandler(msgStr)
			}
		}
	}

}

func (this *WSTradeOKEx) Stop() {
	this.stopPingSign <- true
	this.stopRecvSign <- true
	this.conn.Close()

	close(this.stopPingSign)
	close(this.stopRecvSign)
}
