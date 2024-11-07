package okex

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gorilla/websocket"

	. "github.com/deforceHK/goghostex"
)

type WSMarketOKEx struct {
	RecvHandler  func(string)
	ErrorHandler func(error)
	Config       *APIConfig

	conn   *websocket.Conn
	connId string

	restartSec      int
	restartLimitNum int // In X(restartLimitSec) seconds, the limit times of restart
	restartLimitSec int // In the seconds, the limit times(restartLimitNum) of restart

	restartTS map[int64]string

	subscribed []interface{}

	lastPingTS int64

	stopPingSign chan bool
	stopChecSign chan bool
}

func (this *WSMarketOKEx) Subscribe(v interface{}) {
	if err := this.conn.WriteJSON(v); err != nil {
		this.ErrorHandler(err)
		return
	}
	this.subscribed = append(this.subscribed, v)
}

func (this *WSMarketOKEx) Unsubscribe(v interface{}) {
	if err := this.conn.WriteJSON(v); err != nil {
		this.ErrorHandler(err)
		return
	}
	this.subscribed = append(this.subscribed, v)
}

func (this *WSMarketOKEx) Stop() {
	if this.stopPingSign != nil {
		this.stopPingSign <- true
	}

	if this.stopChecSign != nil {
		this.stopChecSign <- true
	}

	if this.conn != nil {
		_ = this.conn.Close()
		this.conn = nil
	}
	this.connId = ""
}

func (this *WSMarketOKEx) Restart() {
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

func (this *WSMarketOKEx) Start() error {
	// it will stop the ws if the restart limit is reached
	if err := this.startCheck(); err != nil {
		return err
	}

	var conn, err = this.getConn("wss://ws.okx.com:8443/ws/v5/public")
	if err != nil {
		if len(this.restartTS) != 0 {
			this.Restart()
		}
		return err
	}
	this.conn = conn
	this.connId = UUID()

	go this.pingRoutine()
	go this.checkRoutine()
	go this.recvRoutine()

	return nil
}

func (this *WSMarketOKEx) startCheck() error {
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

func (this *WSMarketOKEx) getConn(wss string) (*websocket.Conn, error) {
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

func (this *WSMarketOKEx) initDefaultValue() {
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
		this.restartLimitNum = DERFAULT_WEBSOCKET_RESTART_LIMIT_NUM
	}

	if this.restartLimitSec == 0 {
		this.restartLimitSec = DERFAULT_WEBSOCKET_RESTART_LIMIT_SEC
	}

	if this.restartTS == nil {
		this.restartTS = make(map[int64]string, 0)
	}

}

func (this *WSMarketOKEx) pingRoutine() {
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
			if err := conn.WriteMessage(websocket.TextMessage, []byte("ping")); err != nil {
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

func (this *WSMarketOKEx) checkRoutine() {
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
				this.stopChecSign = nil
				close(stopChecChn)
			}
			return
		}
	}
}

func (this *WSMarketOKEx) recvRoutine() {
	var conn = this.conn
	for {
		var msgType, msg, readErr = conn.ReadMessage()
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

		this.lastPingTS = time.Now().Unix()
		var msgStr = string(msg)
		if msgStr != "pong" {
			this.RecvHandler(msgStr)
		}
	}
}
