package binance

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"

	. "github.com/deforceHK/goghostex"
)

const (
	DEFAULT_WEBSOCKET_RESTART_SEC        = 30
	DEFAULT_WEBSOCKET_PING_SEC           = 60
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
	stopChecSign chan bool
}

type WSMethodBN struct {
	Id     string `json:"id"`
	Params struct {
		ApiKey    string `json:"apiKey"`
		Timestamp int64  `json:"timestamp"`
		Signature string `json:"signature"`
	} `json:"params"`
	Method string `json:"method"`
}

func (this *WSTradeUMBN) Subscribe(v interface{}) {
	if item, ok := v.(string); ok {
		var req = WSMethodBN{
			this.connId,
			struct {
				ApiKey    string `json:"apiKey"`
				Timestamp int64  `json:"timestamp"`
				Signature string `json:"signature"`
			}{
				this.Config.ApiKey,
				time.Now().UnixMilli(),
				"",
			},
			item,
		}

		if err := this.conn.WriteJSON(req); err != nil {
			this.ErrorHandler(err)
			return
		}
		this.subscribed = append(this.subscribed, item)
	}
}

func (this *WSTradeUMBN) Unsubscribe(v interface{}) {
	this.subscribed = append(this.subscribed, v)
	if err := this.conn.WriteJSON(v); err != nil {
		this.ErrorHandler(err)
	}
}

func (this *WSTradeUMBN) Start() error {
	if err := this.startCheck(); err != nil {
		return err
	}

	var conn, err = this.loginConn()
	if err != nil {
		// it means stopped at least once.
		if len(this.restartTS) != 0 {
			this.Restart()
		}
		return err
	}
	this.conn = conn

	//var req = WSMethodBN{
	//	this.connId, []string{"userDataStream.start"}, "REQUEST",
	//}
	//
	//err = conn.WriteJSON(req)
	//if err != nil {
	//	// it means stopped at least once.
	//	if len(this.restartTS) != 0 {
	//		this.Restart()
	//	}
	//	return err
	//}
	//
	//var _, p, readErr = conn.ReadMessage()
	//if readErr != nil {
	//	// it means stopped at least once.
	//	if len(this.restartTS) != 0 {
	//		this.Restart()
	//	}
	//	return readErr
	//}
	//
	//var result = struct {
	//	Id string `json:"id"`
	//}{}
	//
	//_ = json.Unmarshal(p, &result)

	go this.pingRoutine()
	go this.checRoutine()
	go this.recvRoutine()

	return nil
}

func (this *WSTradeUMBN) pingRoutine() {
	var stopPingChn = make(chan bool, 1)
	this.stopPingSign = stopPingChn
	var ticker = time.NewTicker(DEFAULT_WEBSOCKET_PING_SEC * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if this.conn == nil {
				continue
			}

			var bn = New(this.Config)
			var response = struct {
				ListenKey string `json:"listenKey"`
			}{}
			if resp, err := bn.Swap.DoRequest(
				http.MethodPut,
				"/fapi/v1/listenKey",
				"",
				&response,
				SETTLE_MODE_COUNTER,
			); err != nil {
				this.ErrorHandler(err)
			} else if response.ListenKey == "" {
				this.ErrorHandler(fmt.Errorf(string(resp)))
			} else {
				this.lastPingTS = time.Now().Unix()
			}
		case _, opened := <-stopPingChn:
			if opened {
				close(stopPingChn)
			}
			this.stopPingSign = nil
			return
		}
	}
}

func (this *WSTradeUMBN) checRoutine() {
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

func (this *WSTradeUMBN) recvRoutine() {
	var ticker = time.NewTicker(DEFAULT_WEBSOCKET_PENDING_SEC * time.Second)
	defer ticker.Stop()
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

func (this *WSTradeUMBN) loginConn() (*websocket.Conn, error) {
	this.initDefaultValue()

	var bn = New(this.Config)
	var response = struct {
		ListenKey string `json:"listenKey"`
	}{}
	if resp, err := bn.Swap.DoRequest(
		http.MethodPost,
		"/fapi/v1/listenKey",
		"",
		&response,
		SETTLE_MODE_COUNTER,
	); err != nil {
		return nil, err
	} else if response.ListenKey == "" {
		return nil, fmt.Errorf(string(resp))
	}

	var conn, _, err = websocket.DefaultDialer.Dial(
		fmt.Sprintf("wss://fstream.binance.com/ws/%s", response.ListenKey),
		nil,
	)
	if err != nil {
		return nil, err
	}

	this.connId = UUID()
	this.listenKey = response.ListenKey

	return conn, nil
}

type WSMarketUMBN struct {
	*WSTradeUMBN
}

func (this *WSMarketUMBN) Start() error {
	// it will return error if the restart limit is reached
	if err := this.startCheck(); err != nil {
		return err
	}

	var conn, err = this.noLoginConn("wss://fstream.binance.com/stream")
	if err != nil {
		// it means stopped at least once.
		if len(this.restartTS) != 0 {
			this.Restart()
		}
		return err
	}
	this.conn = conn

	//go this.pingRoutine()
	go this.checRoutine()
	go this.recvRoutine()

	return nil
}

func (this *WSMarketUMBN) Subscribe(v interface{}) {
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

func (this *WSMarketUMBN) Restart() {
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

func (this *WSMarketUMBN) noLoginConn(wss string) (*websocket.Conn, error) {
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
