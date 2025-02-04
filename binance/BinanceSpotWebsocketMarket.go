package binance

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gorilla/websocket"

	. "github.com/deforceHK/goghostex"
)

type WSMarketSpot struct {
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

	stopChecSign chan bool
}

func (this *WSMarketSpot) Start() error {
	// it will return error if the restart limit is reached
	if err := this.startCheck(); err != nil {
		return err
	}

	var conn, err = this.noLoginConn("wss://stream.binance.com:9443/ws")
	if err != nil {
		// it means stopped at least once.
		if len(this.restartTS) != 0 {
			this.Restart()
		}
		return err
	}
	this.conn = conn

	go this.checRoutine()
	go this.recvRoutine()

	return nil
}

func (this *WSMarketSpot) Subscribe(v interface{}) {
	if item, ok := v.(string); ok {

		var req = struct {
			Id     string   `json:"id"`
			Method string   `json:"method"`
			Params []string `json:"params"`
		}{
			this.connId,
			"SUBSCRIBE",
			[]string{item},
		}

		if err := this.conn.WriteJSON(req); err != nil {
			this.ErrorHandler(err)
			return
		}
		this.subscribed = append(this.subscribed, item)
	}
}

func (this *WSMarketSpot) Unsubscribe(v interface{}) {
	if item, ok := v.(string); ok {

		var req = struct {
			Id     string   `json:"id"`
			Method string   `json:"method"`
			Params []string `json:"params"`
		}{
			this.connId,
			"UNSUBSCRIBE",
			[]string{item},
		}

		if err := this.conn.WriteJSON(req); err != nil {
			this.ErrorHandler(err)
			return
		}
	}
}

func (this *WSMarketSpot) Restart() {
	this.restartTS[time.Now().Unix()] = this.connId
	this.Stop()
	this.ErrorHandler(
		&WSRestartError{Msg: fmt.Sprintf("websocket will restart in next %d seconds......", this.restartSec)},
	)

	time.Sleep(time.Duration(this.restartSec) * time.Second)
	if err := this.Start(); err != nil {
		this.ErrorHandler(err)
		return
	}

	// subscribe unsubscribe the channel
	for _, v := range this.subscribed {
		if item, ok := v.(string); ok {
			var req = struct {
				Id     string   `json:"id"`
				Method string   `json:"method"`
				Params []string `json:"params"`
			}{
				this.connId,
				"SUBSCRIBE",
				[]string{item},
			}
			if err := this.conn.WriteJSON(req); err != nil {
				this.ErrorHandler(err)
			}
		}
	}
}

func (this *WSMarketSpot) noLoginConn(wss string) (*websocket.Conn, error) {
	this.initDefaultValue()
	var conn, _, err = websocket.DefaultDialer.Dial(
		wss,
		nil,
	)
	if err != nil {
		return nil, err
	}
	this.connId = UUID()
	return conn, nil
}

func (this *WSMarketSpot) startCheck() error {
	var restartNum, limitTS = 0, time.Now().Unix() - int64(this.restartLimitSec)
	for ts, _ := range this.restartTS {
		if ts > limitTS {
			restartNum++
		}
	}
	fmt.Println("Restarted time: ", restartNum, "Restart limit time: ", this.restartLimitNum)
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

func (this *WSMarketSpot) checRoutine() {
	var stopCheckChn = make(chan bool, 1)
	this.stopChecSign = stopCheckChn
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
		case _, opened := <-stopCheckChn:
			if opened {
				close(stopCheckChn)
			}
			this.stopChecSign = nil
			return
		}
	}
}

func (this *WSMarketSpot) recvRoutine() {
	var ticker = time.NewTicker(DEFAULT_WEBSOCKET_PENDING_SEC * time.Second)
	defer ticker.Stop()

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

		this.lastPingTS = time.Now().Unix()
		this.RecvHandler(string(msg))
	}

}

func (this *WSMarketSpot) Stop() {

	if this.stopChecSign != nil {
		this.stopChecSign <- true
	}

	if this.conn != nil {
		_ = this.conn.Close()
		this.conn = nil
	}
	this.connId = ""
}

func (this *WSMarketSpot) initDefaultValue() {
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
