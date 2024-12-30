package kraken

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gorilla/websocket"

	. "github.com/deforceHK/goghostex"
)

type WSSpotTradeKK struct {
	RecvHandler  func(string)
	ErrorHandler func(error)
	Config       *APIConfig

	conn       *websocket.Conn
	connId     string
	subscribed []interface{}

	restartSleepSec int
	restartLimitNum int // In X(restartLimitSec) seconds, the limit times(restartLimitNum) of restart
	restartLimitSec int // In X(restartLimitSec) seconds, the limit times(restartLimitNum) of restart

	lastPingTS int64
	restartTS  map[int64]string

	stopChecSign chan bool
	stopPingSign chan bool
}

func (this *WSSpotTradeKK) Subscribe(v interface{}) {
	var channel, isStr = v.(string)
	if !isStr {
		this.ErrorHandler(fmt.Errorf("the subscribe param must be string"))
		return
	}

	var err = this.Write(map[string]string{
		"event":              "subscribe",
		"feed":               channel,
		"api_key":            this.Config.ApiKey,
		"original_challenge": this.connId,
		"signed_challenge":   hashChallenge(this.Config.ApiSecretKey, this.connId),
	})
	if err != nil {
		this.ErrorHandler(err)
		return
	}
	this.subscribed = append(this.subscribed, v)
}

func (this *WSSpotTradeKK) Unsubscribe(v interface{}) {
	var channel, isStr = v.(string)
	if !isStr {
		this.ErrorHandler(fmt.Errorf("the unsubscribe param must be string"))
		return
	}

	var err = this.Write(map[string]string{
		"event":              "unsubscribe",
		"feed":               channel,
		"api_key":            this.Config.ApiKey,
		"original_challenge": this.connId,
		"signed_challenge":   hashChallenge(this.Config.ApiSecretKey, this.connId),
	})
	if err != nil {
		this.ErrorHandler(err)
		return
	}

	var newSub = make([]interface{}, 0)
	for _, subCh := range this.subscribed {
		if subCh.(string) != channel {
			newSub = append(newSub, subCh)
		}
	}
	this.subscribed = newSub
}

func (this *WSSpotTradeKK) Write(v interface{}) error {
	if err := this.conn.WriteJSON(v); err != nil {
		return err
	}
	return nil
}

func (this *WSSpotTradeKK) Start() error {
	var stopErr = this.startCheck()
	if stopErr != nil {
		return stopErr
	}

	var conn, err = this.getLoginConn("wss://ws-auth.kraken.com/v2")
	if err != nil {
		if len(this.restartTS) != 0 {
			this.Restart()
		}
		return err
	}
	this.conn = conn

	var ping = struct {
		Method string `json:"method"`
		ReqId  int64  `json:"req_id"`
	}{"ping", time.Now().UnixMilli()}

	err = this.conn.WriteJSON(ping)
	if err != nil {
		if len(this.restartTS) != 0 {
			this.Restart()
		}
		return err
	}

	for {
		var _, p, _ = conn.ReadMessage()
		var result = struct {
			Method  string `json:"method"`
			ReqId   int64  `json:"req_id"`
			TimeIn  string `json:"time_in"`
			TimeOut string `json:"time_out"`
		}{}

		_ = json.Unmarshal(p, &result)
		if result.Method != "pong" {
			continue
		} else {
			break
		}
	}

	go this.recvRoutine()
	go this.checkRoutine()
	go this.pingRoutine()

	return nil
}

func (this *WSSpotTradeKK) Stop() {

	if this.stopChecSign != nil {
		this.stopChecSign <- true
	}

	if this.conn != nil {
		_ = this.conn.Close()
		this.conn = nil
	}
	this.connId = ""
}

func (this *WSSpotTradeKK) Restart() {
	this.ErrorHandler(
		&WSRestartError{Msg: fmt.Sprintf("websocket will restart in next %d seconds...", this.restartSleepSec)},
	)
	this.restartTS[time.Now().Unix()] = this.connId
	this.Stop()

	time.Sleep(time.Duration(this.restartSleepSec) * time.Second)
	if err := this.Start(); err != nil {
		this.ErrorHandler(err)
		return
	}

	// subscribe unsubscribe the channel
	for _, channel := range this.subscribed {
		var err = this.conn.WriteJSON(map[string]string{
			"event":              "subscribe",
			"feed":               channel.(string),
			"api_key":            this.Config.ApiKey,
			"original_challenge": this.connId,
			"signed_challenge":   hashChallenge(this.Config.ApiSecretKey, this.connId),
		})
		if err != nil {
			this.ErrorHandler(err)
			var errMsg, _ = json.Marshal(channel.(string))
			this.ErrorHandler(fmt.Errorf("subscribe error: %s", string(errMsg)))
		}
	}

}

func (this *WSSpotTradeKK) startCheck() error {
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

func (this *WSSpotTradeKK) getConn(wss string) (*websocket.Conn, error) {
	this.initDefaultValue()
	var conn, _, err = websocket.DefaultDialer.Dial(
		wss,
		nil,
	)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func (this *WSSpotTradeKK) getLoginConn(wss string) (*websocket.Conn, error) {
	this.initDefaultValue()
	var kk = New(this.Config)
	if _, token, err := kk.GetToken(); err != nil {
		return nil, err
	} else {
		this.connId = token
	}

	if conn, _, err := websocket.DefaultDialer.Dial(
		wss,
		nil,
	); err != nil {
		return nil, err
	} else {
		return conn, nil
	}
}

func (this *WSSpotTradeKK) initDefaultValue() {
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

func (this *WSSpotTradeKK) checkRoutine() {
	var stopChecChn = make(chan bool, 1)
	this.stopChecSign = stopChecChn

	var ticker = time.NewTicker(DEFAULT_WEBSOCKET_PENDING_SEC * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// 超过x秒没有收到消息，重新连接，如果超出重连次数，ws将停止。
			if time.Now().Unix()-this.lastPingTS > DEFAULT_WEBSOCKET_PENDING_SEC {
				this.ErrorHandler(fmt.Errorf("ping timeout, last ping ts: %d", this.lastPingTS))
				this.Restart()
				continue
			}
		case _, opened := <-stopChecChn:
			if opened {
				close(stopChecChn)
			}
			this.stopChecSign = nil
			return
		}
	}
}

func (this *WSSpotTradeKK) recvRoutine() {
	for {
		var msgType, msg, readErr = this.conn.ReadMessage()
		if readErr != nil {
			// conn closed by user.
			if strings.Index(readErr.Error(), "use of closed network connection") > 0 {
				fmt.Println("conn closed by user. ")
				return
			}

			this.ErrorHandler(readErr)
			this.Restart()
			return
		}

		if msgType != websocket.TextMessage {
			continue
		}
		var event = struct {
			Feed string `json:"feed"`
		}{}
		_ = json.Unmarshal(msg, &event)
		this.lastPingTS = time.Now().Unix()
		if event.Feed != "heartbeat" {
			this.RecvHandler(string(msg))
		}
	}
}

func (this *WSSpotTradeKK) pingRoutine() {

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

			var ping = struct {
				Method string `json:"method"`
				ReqId  int64  `json:"req_id"`
			}{"ping", time.Now().UnixMilli()}

			var err = conn.WriteJSON(ping)
			if err != nil {
				fmt.Println(err)
			}
		case _, opened := <-stopPingChn:
			if opened {
				this.stopPingSign = nil
				close(stopPingChn)
			}
			return
		}
	}
}
