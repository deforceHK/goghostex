package kraken

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gorilla/websocket"

	. "github.com/deforceHK/goghostex"
)

const (
	DEFAULT_WEBSOCKET_RESTART_SLEEP_SEC  = 30
	DEFAULT_WEBSOCKET_PING_SEC           = 20
	DEFAULT_WEBSOCKET_PENDING_SEC        = 100
	DERFAULT_WEBSOCKET_RESTART_LIMIT_NUM = 10
	DERFAULT_WEBSOCKET_RESTART_LIMIT_SEC = 300
)

type WSSwapTradeKK struct {
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

func (this *WSSwapTradeKK) Subscribe(v interface{}) {
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

func (this *WSSwapTradeKK) Unsubscribe(v interface{}) {
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

func (this *WSSwapTradeKK) Write(v interface{}) error {
	if err := this.conn.WriteJSON(v); err != nil {
		return err
	}
	return nil
}

func (this *WSSwapTradeKK) Start() error {
	var stopErr = this.startCheck()
	if stopErr != nil {
		return stopErr
	}

	var conn, err = this.getConn("wss://futures.kraken.com/ws/v1")
	if err != nil {
		if len(this.restartTS) != 0 {
			this.Restart()
		}
		return err
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
		if len(this.restartTS) != 0 {
			this.Restart()
		}
		return err
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

	var heartBeat = struct {
		Event string `json:"event"`
		Feed  string `json:"feed"`
	}{
		Event: "subscribe",
		Feed:  "heartbeat",
	}

	err = this.conn.WriteJSON(heartBeat)
	if err != nil {
		if len(this.restartTS) != 0 {
			this.Restart()
		}
		return err
	}

	go this.recvRoutine()
	go this.checkRoutine()

	return nil
}

func (this *WSSwapTradeKK) Stop() {

	if this.stopChecSign != nil {
		this.stopChecSign <- true
	}

	if this.conn != nil {
		_ = this.conn.Close()
		this.conn = nil
	}
	this.connId = ""
}

func (this *WSSwapTradeKK) Restart() {
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

func (this *WSSwapTradeKK) startCheck() error {
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

func (this *WSSwapTradeKK) getConn(wss string) (*websocket.Conn, error) {
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

func (this *WSSwapTradeKK) initDefaultValue() {
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

func (this *WSSwapTradeKK) checkRoutine() {
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

func (this *WSSwapTradeKK) recvRoutine() {
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

func hashChallenge(apiSecret, challenge string) string {

	// 1. Hash the challenge with SHA-256
	var challengeHash = sha256.Sum256([]byte(challenge))

	// 2. Base64-decode your api_secret
	var decodedSecret, _ = base64.StdEncoding.DecodeString(apiSecret)

	// 3. Use the result of step 2 to hash the result of step 1 with HMAC-SHA-512
	var hmac = hmac.New(sha512.New, decodedSecret)
	hmac.Write(challengeHash[:])
	var signature = hmac.Sum(nil)

	// 4. Base64-encode the result of step 3
	return base64.StdEncoding.EncodeToString(signature)
}
