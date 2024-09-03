package binance

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"

	. "github.com/deforceHK/goghostex"
)

const (
	DEFAULT_WEBSOCKET_RESTART_SEC        = 60
	DEFAULT_WEBSOCKET_PING_SEC           = 20
	DEFAULT_WEBSOCKET_PENDING_SEC        = 100
	DERFAULT_WEBSOCKET_RESTART_LIMIT_NUM = 10
	DERFAULT_WEBSOCKET_RESTART_LIMIT_SEC = 300
)

type WSTradeUMBN struct {
	RecvHandler  func(string)
	ErrorHandler func(error)
	Config       *APIConfig

	conn   *websocket.Conn
	connId string

	restartSec      int
	restartLimitNum int // In X seconds, the times of restart
	restartLimitSec int // During the seconds, the restart limit

	restartTS map[int64]string

	subscribed []interface{}

	listenKey  string
	lastPingTS int64

	stopPingSign chan bool
	stopRecvSign chan bool
}

type WSMethodBN struct {
	Id     string   `json:"id"`
	Param  []string `json:"params"`
	Method string   `json:"method"`
}

func (this *WSTradeUMBN) Subscribe(v interface{}) {
	if item, ok := v.(string); ok {
		fmt.Println("sdfsf")
		this.subscribed = append(this.subscribed, item)
		var req = WSMethodBN{
			this.connId,
			[]string{fmt.Sprintf("%s@%s", this.listenKey, item)},
			"REQUEST",
		}
		if err := this.conn.WriteJSON(req); err != nil {
			this.ErrorHandler(err)
		}
	}
}

func (this *WSTradeUMBN) Unsubscribe(v interface{}) {
	this.subscribed = append(this.subscribed, v)
	if err := this.conn.WriteJSON(v); err != nil {
		this.ErrorHandler(err)
	}
}

func (this *WSTradeUMBN) Start() error {
	// it will stop the ws if the restart limit is reached
	if err := this.startCheck(); err != nil {
		return err
	}

	var conn, err = this.loginConn()
	if err != nil {
		this.ErrorHandler(err)
		time.Sleep(time.Duration(this.restartSec) * time.Second)
		return err
	}
	this.conn = conn

	var req = WSMethodBN{
		this.connId, []string{"userDataStream.start"}, "REQUEST",
	}
	//json.Marshal(req)
	err = conn.WriteJSON(req)
	if err != nil {
		this.ErrorHandler(err)
		this.restartTS[time.Now().Unix()] = this.connId
		if this.conn != nil {
			_ = this.conn.Close()
			log.Printf(
				"websocket conn %s will be restart in next %d seconds...",
				this.connId, this.restartSec,
			)
			this.conn = nil
			this.connId = ""
		}
		time.Sleep(time.Duration(this.restartSec) * time.Second)
		return this.Start()
	}

	var _, p, readErr = conn.ReadMessage()
	if readErr != nil {
		this.ErrorHandler(readErr)
		this.restartTS[time.Now().Unix()] = this.connId
		if this.conn != nil {
			_ = this.conn.Close()
			log.Printf(
				"websocket conn %s will be restart in next %d seconds...",
				this.connId, this.restartSec,
			)
			this.conn = nil
			this.connId = ""
		}
		time.Sleep(time.Duration(this.restartSec) * time.Second)
		return this.Start()
	}

	var result = struct {
		Id string `json:"id"`
	}{}

	var jsonErr = json.Unmarshal(p, &result)
	if jsonErr != nil {
		this.ErrorHandler(jsonErr)
		return jsonErr
	}

	go this.pingRoutine()
	go this.recvRoutine()

	return nil
}

func (this *WSTradeUMBN) pingRoutine() {
	var stopPingChn = make(chan bool, 1)
	this.stopPingSign = stopPingChn
	var ticker = time.NewTicker(DEFAULT_WEBSOCKET_PING_SEC * time.Second)
	defer ticker.Stop()
	var conn = this.conn
	for {
		select {
		case <-ticker.C:
			if this.conn == nil {
				continue
			}
			_ = conn.WriteMessage(websocket.PingMessage, nil)
		case <-stopPingChn:
			close(stopPingChn)
			return
		}
	}
}

func (this *WSTradeUMBN) recvRoutine() {
	var stopRecvChn = make(chan bool, 1)
	this.stopRecvSign = stopRecvChn
	var ticker = time.NewTicker(DEFAULT_WEBSOCKET_PENDING_SEC * time.Second)
	defer ticker.Stop()
	var conn = this.conn

	for {
		select {
		case <-ticker.C:
			// 超过x秒没有收到消息，重新连接，如果超出重连次数，ws将停止。
			if time.Now().Unix()-this.lastPingTS > DEFAULT_WEBSOCKET_PENDING_SEC {
				this.ErrorHandler(fmt.Errorf("ping timeout, last ping ts: %d", this.lastPingTS))
				this.Restart()
				continue
			}
		case <-stopRecvChn:
			close(stopRecvChn)
			return
		default:
			var msgType, msg, readErr = conn.ReadMessage()
			if readErr != nil {
				this.ErrorHandler(readErr)
				this.Restart()
				continue
			}

			if msgType != websocket.TextMessage {
				fmt.Println("msgType != websocket.TextMessage")
				continue
			}

			this.lastPingTS = time.Now().Unix()
			//var msgStr = string(msg)
			//if msgStr != "pong" {
			this.RecvHandler(string(msg))
			//}
		}
	}

}

func (this *WSTradeUMBN) Stop() {
	if this.stopPingSign != nil {
		this.stopPingSign <- true
	}

	if this.stopRecvSign != nil {
		this.stopRecvSign <- true
	}

	if this.conn != nil {
		_ = this.conn.Close()
		this.conn = nil
	}

}

func (this *WSTradeUMBN) Restart() {
	this.ErrorHandler(
		&WSRestartError{Msg: fmt.Sprintf("websocket will restart in next %d seconds...", this.restartSec)},
	)
	this.restartTS[time.Now().Unix()] = this.connId
	this.Stop()

	time.Sleep(time.Duration(this.restartSec) * time.Second)
	if err := this.Start(); err != nil {
		this.ErrorHandler(err)
		return
	}

	var conn = this.conn
	// subscribe unsubscribe the channel
	for _, v := range this.subscribed {
		var err = conn.WriteJSON(v)
		if err != nil {
			this.ErrorHandler(err)
			var errMsg, _ = json.Marshal(v)
			this.ErrorHandler(fmt.Errorf("subscribe error: %s", string(errMsg)))
		}
	}

}

func (this *WSTradeUMBN) initDefaultValue() {
	if this.RecvHandler == nil {
		this.RecvHandler = func(msg string) {
			log.Println(msg)
		}
	}
	if this.ErrorHandler == nil {
		this.ErrorHandler = func(err error) {
			log.Println(err)
		}
	}
	if this.restartSec == 0 {
		this.restartSec = DEFAULT_WEBSOCKET_RESTART_SEC
	}

	if this.restartLimitNum == 0 {
		this.restartLimitSec = DERFAULT_WEBSOCKET_RESTART_LIMIT_NUM
	}

	if this.restartLimitSec == 0 {
		this.restartLimitSec = DERFAULT_WEBSOCKET_RESTART_LIMIT_SEC
	}

	if this.restartTS == nil {
		this.restartTS = make(map[int64]string, 0)
	}

}

func (this *WSTradeUMBN) startCheck() error {
	var restartNum, limitTS = 0, time.Now().Unix() - int64(this.restartLimitSec)
	for ts, _ := range this.restartTS {
		if ts > limitTS {
			restartNum++
		}
	}
	if restartNum > this.restartLimitNum {
		var wsErr = &WSStopError{
			Msg: fmt.Sprintf(
				"The ws restarted %d times in %d seconds, stop the ws",
				restartNum, this.restartLimitSec,
			),
		}
		return wsErr
	}
	return nil
}

func (this *WSTradeUMBN) loginConn() (*websocket.Conn, error) {
	this.initDefaultValue()

	var bn = New(this.Config)
	var response = struct {
		ListenKey string `json:"listenKey"`
	}{}
	if _, err := bn.Swap.DoRequest(
		http.MethodPost,
		"/fapi/v1/listenKey",
		"",
		&response,
		SETTLE_MODE_COUNTER,
	); err != nil {
		return nil, err
	}

	var conn, _, err = websocket.DefaultDialer.Dial(
		fmt.Sprintf("wss://fstream.binance.com/ws/%s", response.ListenKey),
		nil,
	)
	if err != nil {
		this.restartTS[time.Now().Unix()] = this.connId
		if this.conn != nil {
			_ = this.conn.Close()
			this.conn = nil
			this.connId = ""
		}
		return nil, err
	}

	this.connId = UUID()
	this.listenKey = response.ListenKey

	return conn, nil
}

func (this *WSTradeUMBN) noLoginConn(wss string) (*websocket.Conn, error) {
	this.initDefaultValue()
	var conn, _, err = websocket.DefaultDialer.Dial(
		wss,
		nil,
	)
	if err != nil {
		this.restartTS[time.Now().Unix()] = this.connId
		if this.conn != nil {
			_ = this.conn.Close()
			this.conn = nil
			this.connId = ""
		}
		return nil, err
	}
	return conn, nil
}