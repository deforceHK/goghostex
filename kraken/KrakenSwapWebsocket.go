package kraken

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"time"

	. "github.com/deforceHK/goghostex"
)

const (
	DEFAULT_WEBSOCKET_RESTART_SLEEP_SEC  = 30
	DEFAULT_WEBSOCKET_PING_SEC           = 20
	DEFAULT_WEBSOCKET_PENDING_SEC        = 100
	DERFAULT_WEBSOCKET_RESTART_LIMIT_NUM = 10
	DERFAULT_WEBSOCKET_RESTART_LIMIT_SEC = 300
)

type WSSwapKK struct {
	RecvHandler  func(string)
	ErrorHandler func(error)
	Config       *APIConfig

	conn   *websocket.Conn
	connId string

	restartSleepSec int
	restartLimitNum int // In X(restartLimitSec) seconds, the limit times(restartLimitNum) of restart
	restartLimitSec int // In X(restartLimitSec) seconds, the limit times(restartLimitNum) of restart

	restartTS map[int64]string

	subscribed []interface{}

	lastPingTS int64

	stopPingSign chan bool
	stopChecSign chan bool
}

func (this *WSSwapKK) Subscribe(v interface{}) {
	var err = this.conn.WriteJSON(v)
	if err != nil {
		this.ErrorHandler(err)
	}
}

func (this *WSSwapKK) Unsubscribe(v interface{}) {
	//TODO implement me
	panic("implement me")
}

func (this *WSSwapKK) Start() error {
	var conn, err = this.getConn("wss://futures.kraken.com/ws/v1")
	if err != nil {
		time.Sleep(time.Duration(this.restartSleepSec) * time.Second)
		return this.Start()
	}

	this.conn = conn

	var challenge = struct {
		Event  string `json:"event"`
		ApiKey string `json:"api_key"`
		Feed   string `json:"feed"`
	}{
		Event:  "challenge",
		Feed:   "heartbeat",
		ApiKey: this.Config.ApiKey,
	}

	err = this.conn.WriteJSON(challenge)
	if err != nil {
		this.ErrorHandler(err)
		this.restartTS[time.Now().Unix()] = this.connId
		if this.conn != nil {
			_ = this.conn.Close()
			log.Printf(
				"websocket conn %s will be restart in next %d seconds...",
				this.connId, this.restartSleepSec,
			)
			this.conn = nil
			this.connId = ""
		}
		time.Sleep(time.Duration(this.restartSleepSec) * time.Second)
		return this.Start()
	}

	for {
		var _, p, _ = conn.ReadMessage()
		var result = struct {
			Event   string `json:"event"`
			Message string `json:"message"`
		}{}

		_ = json.Unmarshal(p, &result)
		if result.Event != "challenge" {
			continue
		} else {
			this.connId = result.Message
			break
		}
	}

	go this.recvRoutine()

	return nil
}

func (this *WSSwapKK) Stop() {
	//TODO implement me
	panic("implement me")
}

func (this *WSSwapKK) Restart() {
	//TODO implement me
	panic("implement me")
}

func (this *WSSwapKK) startCheck() error {
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

func (this *WSSwapKK) getConn(wss string) (*websocket.Conn, error) {
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

func (this *WSSwapKK) initDefaultValue() {
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
	if this.restartSleepSec == 0 {
		this.restartSleepSec = DEFAULT_WEBSOCKET_RESTART_SLEEP_SEC
	}

	if this.restartLimitNum == 0 {
		this.restartLimitNum = DERFAULT_WEBSOCKET_RESTART_LIMIT_NUM
	}

	if this.restartLimitSec == 0 {
		this.restartLimitSec = DERFAULT_WEBSOCKET_RESTART_LIMIT_SEC
	}

	if this.restartTS == nil {
		this.restartTS = make(map[int64]string, 0)
	}

}

func (this *WSSwapKK) recvRoutine() {
	var conn = this.conn
	for {
		var msgType, msg, readErr = conn.ReadMessage()
		if readErr != nil {
			this.ErrorHandler(readErr)
			this.Restart()
			return
		}

		if msgType != websocket.TextMessage {
			continue
		}

		this.lastPingTS = time.Now().Unix()
		this.RecvHandler(string(msg))
	}
}
