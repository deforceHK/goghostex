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

type WSSpotMarketKK struct {
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
}

func (this *WSSpotMarketKK) Start() error {
	var stopErr = this.startCheck()
	if stopErr != nil {
		return stopErr
	}

	var conn, err = this.getConn("wss://ws.kraken.com/v2")
	if err != nil {
		if len(this.restartTS) != 0 {
			this.Restart()
		}
		return err
	}
	this.conn = conn
	this.connId = UUID()

	//var heartBeat = struct {
	//	Method string `json:"method"`
	//	ReqId  int64  `json:"req_id"`
	//}{
	//	"ping",
	//	time.Now().UnixMilli(),
	//}
	//
	//err = this.conn.WriteJSON(heartBeat)
	//if err != nil {
	//	if len(this.restartTS) != 0 {
	//		this.Restart()
	//	}
	//	return err
	//}

	go this.recvRoutine()
	go this.checkRoutine()
	return nil

}

func (this *WSSpotMarketKK) Subscribe(v interface{}) {
	var err = this.conn.WriteJSON(v)
	if err != nil {
		this.ErrorHandler(err)
	}
	this.subscribed = append(this.subscribed, v)
}

func (this *WSSpotMarketKK) Unsubscribe(v interface{}) {
	var err = this.conn.WriteJSON(v)
	if err != nil {
		this.ErrorHandler(err)
	}
}

func (this *WSSpotMarketKK) Restart() {
	this.ErrorHandler(
		&WSRestartError{
			Msg: fmt.Sprintf("kk market websocket will restart in next %d seconds...", this.restartSleepSec),
		},
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
		var err = this.conn.WriteJSON(channel)
		if err != nil {
			this.ErrorHandler(err)
			var errMsg, _ = json.Marshal(channel)
			this.ErrorHandler(fmt.Errorf("subscribe error: %s", string(errMsg)))
		}
	}

}

func (this *WSSpotMarketKK) checkRoutine() {
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

func (this *WSSpotMarketKK) recvRoutine() {
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
			Channel string `json:"channel"`
		}{}
		_ = json.Unmarshal(msg, &event)
		this.lastPingTS = time.Now().Unix()
		if event.Channel != "heartbeat" {
			this.RecvHandler(string(msg))
		}
	}
}

func (this *WSSpotMarketKK) startCheck() error {
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

func (this *WSSpotMarketKK) getConn(wss string) (*websocket.Conn, error) {
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

func (this *WSSpotMarketKK) initDefaultValue() {
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

func (this *WSSpotMarketKK) Stop() {

	if this.stopChecSign != nil {
		this.stopChecSign <- true
	}

	if this.conn != nil {
		_ = this.conn.Close()
		this.conn = nil
	}
	this.connId = ""
}
