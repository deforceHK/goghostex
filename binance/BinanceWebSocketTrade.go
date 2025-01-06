package binance

import (
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/websocket"

	. "github.com/deforceHK/goghostex"
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
	stopChecSign chan bool
}

type WSParamsBN struct {
	Id     string                 `json:"id"`
	Method string                 `json:"method"`
	Params map[string]interface{} `json:"params"`
}

func (this *WSTradeUMBN) Write(v interface{}) error {
	if this.conn == nil {
		return fmt.Errorf("conn is nil")
	}

	if req, ok := v.(WSParamsBN); !ok {
		return fmt.Errorf("v is not WSParamsBN")
	} else {
		req.Params["apiKey"] = this.Config.ApiKey
		req.Params["timestamp"] = time.Now().UnixMilli()
		var p = url.Values{}
		for key, value := range req.Params {
			p.Set(key, fmt.Sprintf("%v", value))
		}
		var sign, _ = GetParamHmacSHA256Sign(this.Config.ApiSecretKey, p.Encode())
		req.Params["signature"] = sign
		return this.conn.WriteJSON(req)
	}
}

func (this *WSTradeUMBN) Subscribe(v interface{}) {
	if err := this.Write(v); err != nil {
		this.ErrorHandler(err)
	}
}

func (this *WSTradeUMBN) Unsubscribe(v interface{}) {
	if err := this.Write(v); err != nil {
		this.ErrorHandler(err)
	}
}

func (this *WSTradeUMBN) Start() error {
	if err := this.startCheck(); err != nil {
		return err
	}

	var conn, err = this.noLoginConn("wss://ws-fapi.binance.com/ws-fapi/v1")
	if err != nil {
		// it means stopped at least once.
		if len(this.restartTS) != 0 {
			this.Restart()
		}
		return err
	}
	this.conn = conn

	//go this.pingRoutine() binance auto send pong to ping
	//go this.checRoutine() binance auto send pong to ping
	go this.recvRoutine()

	return nil
}

func (this *WSTradeUMBN) recvRoutine() {

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

func (this *WSTradeUMBN) Stop() {
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

func (this *WSTradeUMBN) Restart() {
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
	//for _, v := range this.subscribed {
	//	if item, ok := v.(string); ok {
	//		var req = WSMethodBN{
	//			this.connId,
	//			[]string{fmt.Sprintf("%s@%s", this.listenKey, item)},
	//			"REQUEST",
	//		}
	//		if err := this.conn.WriteJSON(req); err != nil {
	//			this.ErrorHandler(err)
	//		}
	//	}
	//}

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
		this.restartLimitNum = DERFAULT_WEBSOCKET_RESTART_LIMIT_NUM
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

func (this *WSTradeUMBN) noLoginConn(wss string) (*websocket.Conn, error) {
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
