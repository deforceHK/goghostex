package goghostex

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type WsConfig struct {
	WsUrl                 string              // websocket server url, necessary
	ProxyUrl              string              // proxy url, not necessary
	ReqHeaders            map[string][]string // set the head info ,when connecting, not necessary
	HeartbeatIntervalTime time.Duration       // the heartbeat interval, necessary
	HeartbeatData         []byte              // the raw text of heartbeat data for example: ping, necessary if heartbeatfunc is nil
	HeartbeatDataType     int

	//HeartbeatFunc         func() interface{}           // the json text of heartbeat data , necessary if heartbeatdata is nil
	ReconnectIntervalTime time.Duration                // force reconnect on XXX time duration, not necessary
	ProtoHandleFunc       func([]byte) error           // the message handle func, necessary
	UnCompressFunc        func([]byte) ([]byte, error) // the uncompress func, not necessary
	ErrorHandleFunc       func(err error)              // the error handle func, not necessary
	IsDump                bool                         // is print the connect info, not necessary
}

type WsConn struct {
	*websocket.Conn
	sync.Mutex
	WsConfig

	activeTime  time.Time
	activeTimeL sync.Mutex

	mu             chan struct{} // lock write data
	closeHeartbeat chan struct{}
	closeReconnect chan struct{}
	closeRecv      chan struct{}
	closeCheck     chan struct{}
	subs           []interface{}
}

// websocket build config
type WsBuilder struct {
	wsConfig *WsConfig
}

func NewWsBuilder() *WsBuilder {
	return &WsBuilder{&WsConfig{}}
}

func (b *WsBuilder) WsUrl(wsUrl string) *WsBuilder {
	b.wsConfig.WsUrl = wsUrl
	return b
}

func (b *WsBuilder) ProxyUrl(proxyUrl string) *WsBuilder {
	b.wsConfig.ProxyUrl = proxyUrl
	return b
}

func (b *WsBuilder) ReqHeader(key, value string) *WsBuilder {
	b.wsConfig.ReqHeaders[key] = append(b.wsConfig.ReqHeaders[key], value)
	return b
}

func (b *WsBuilder) Dump() *WsBuilder {
	b.wsConfig.IsDump = true
	return b
}

func (b *WsBuilder) Heartbeat(data []byte, dataType int, t time.Duration) *WsBuilder {
	b.wsConfig.HeartbeatIntervalTime = t
	b.wsConfig.HeartbeatData = data
	b.wsConfig.HeartbeatDataType = dataType
	return b
}

//func (b *WsBuilder) Heartbeat2(heartbeat func() interface{}, t time.Duration) *WsBuilder {
//	b.wsConfig.HeartbeatIntervalTime = t
//	b.wsConfig.HeartbeatFunc = heartbeat
//	return b
//}

func (b *WsBuilder) ReconnectIntervalTime(t time.Duration) *WsBuilder {
	b.wsConfig.ReconnectIntervalTime = t
	return b
}

func (b *WsBuilder) ProtoHandleFunc(f func([]byte) error) *WsBuilder {
	b.wsConfig.ProtoHandleFunc = f
	return b
}

func (b *WsBuilder) UnCompressFunc(f func([]byte) ([]byte, error)) *WsBuilder {
	b.wsConfig.UnCompressFunc = f
	return b
}

func (b *WsBuilder) ErrorHandleFunc(f func(err error)) *WsBuilder {
	b.wsConfig.ErrorHandleFunc = f
	return b
}

func (b *WsBuilder) Build() *WsConn {
	if b.wsConfig.ErrorHandleFunc == nil {
		b.wsConfig.ErrorHandleFunc = func(err error) {
			log.Println(err)
		}
	}
	wsConn := &WsConn{WsConfig: *b.wsConfig}
	return wsConn.NewWs()
}

func (ws *WsConn) NewWs() *WsConn {
	ws.Lock()
	defer ws.Unlock()

	ws.connect()

	ws.mu = make(chan struct{}, 1)
	ws.closeHeartbeat = make(chan struct{}, 1)
	ws.closeReconnect = make(chan struct{}, 1)
	ws.closeRecv = make(chan struct{}, 1)
	ws.closeCheck = make(chan struct{}, 1)

	ws.HeartbeatTimer()
	ws.ReConnectTimer()
	ws.checkStatusTimer()

	return ws
}

func (ws *WsConn) connect() {
	dialer := websocket.DefaultDialer

	if ws.ProxyUrl != "" {
		proxy, err := url.Parse(ws.ProxyUrl)
		if err == nil {
			log.Println("proxy url :", proxy)
			dialer.Proxy = http.ProxyURL(proxy)
		} else {
			log.Println("proxy url error ? ", err)
		}
	}

	wsConn, resp, err := dialer.Dial(ws.WsUrl, http.Header(ws.ReqHeaders))
	if err != nil {
		panic(err)
	}

	ws.Conn = wsConn

	if ws.IsDump {
		dumpData, _ := httputil.DumpResponse(resp, true)
		log.Println(string(dumpData))
	}

	ws.UpdateActiveTime()
}

func (ws *WsConn) SendJsonMessage(v interface{}) error {
	ws.mu <- struct{}{}
	defer func() {
		<-ws.mu
	}()
	return ws.WriteJSON(v)
}

func (ws *WsConn) SendTextMessage(data []byte) error {
	ws.mu <- struct{}{}
	defer func() {
		<-ws.mu
	}()
	return ws.WriteMessage(websocket.TextMessage, data)
}

func (ws *WsConn) ReConnect() {
	ws.Lock()
	defer ws.Unlock()

	log.Println("close ws  error :", ws.Close())
	time.Sleep(time.Second)

	ws.connect()

	//re subscribe
	for _, sub := range ws.subs {
		log.Println("subscribe:", sub)
		_ = ws.SendJsonMessage(sub)
	}
}

func (ws *WsConn) ReConnectTimer() {
	if ws.ReconnectIntervalTime == 0 {
		return
	}
	timer := time.NewTimer(ws.ReconnectIntervalTime)

	go func() {
		ws.clearChannel(ws.closeReconnect)

		for {
			select {
			case <-timer.C:
				log.Println("reconnect websocket")
				ws.ReConnect()
				timer.Reset(ws.ReconnectIntervalTime)
			case <-ws.closeReconnect:
				timer.Stop()
				log.Println("close websocket connect ,  exiting reconnect timer goroutine.")
				return
			}
		}
	}()
}

func (ws *WsConn) checkStatusTimer() {
	if ws.HeartbeatIntervalTime == 0 {
		return
	}

	timer := time.NewTimer(ws.HeartbeatIntervalTime)

	go func() {
		ws.clearChannel(ws.closeCheck)

		for {
			select {
			case <-timer.C:
				now := time.Now()
				if now.Sub(ws.activeTime) >= 2*ws.HeartbeatIntervalTime {
					log.Println("active time [ ", ws.activeTime, " ] has expired , begin reconnect ws.")
					ws.ReConnect()
				}
				timer.Reset(ws.HeartbeatIntervalTime)
			case <-ws.closeCheck:
				log.Println("check status timer exiting")
				return
			}
		}
	}()
}

func (ws *WsConn) HeartbeatTimer() {
	log.Println("heartbeat interval time = ", ws.HeartbeatIntervalTime)
	if ws.HeartbeatIntervalTime == 0 || (ws.HeartbeatDataType == 0 && ws.HeartbeatData == nil) {
		return
	}

	timer := time.NewTicker(ws.HeartbeatIntervalTime)
	go func() {
		ws.clearChannel(ws.closeHeartbeat)

		for {
			select {
			case <-timer.C:
				err := ws.WriteMessage(ws.HeartbeatDataType, ws.HeartbeatData)
				if err != nil {
					log.Println("heartbeat error , ", err)
					time.Sleep(time.Second)
				}
			case <-ws.closeHeartbeat:
				timer.Stop()
				log.Println("close websocket connect , exiting heartbeat goroutine.")
				return
			}
		}
	}()
}

func (ws *WsConn) Subscribe(subEvent interface{}) error {
	log.Println("Subscribe:", subEvent)
	err := ws.SendJsonMessage(subEvent)
	if err != nil {
		return err
	}
	ws.subs = append(ws.subs, subEvent)
	return nil
}

func (ws *WsConn) ReceiveMessage() {
	ws.clearChannel(ws.closeRecv)

	go func() {
		for {

			if len(ws.closeRecv) > 0 {
				<-ws.closeRecv
				log.Println("close websocket , exiting receive message goroutine.")
				return
			}

			t, msg, err := ws.ReadMessage()
			if err != nil {
				ws.ErrorHandleFunc(err)
				time.Sleep(time.Second)
				continue
			}

			switch t {
			case websocket.TextMessage:
				ws.ProtoHandleFunc(msg)
			case websocket.BinaryMessage:
				if ws.UnCompressFunc == nil {
					ws.ProtoHandleFunc(msg)
				} else {
					msg2, err := ws.UnCompressFunc(msg)
					if err != nil {
						ws.ErrorHandleFunc(fmt.Errorf("%s,%s", "un compress error", err.Error()))
					} else {
						err := ws.ProtoHandleFunc(msg2)
						if err != nil {
							ws.ErrorHandleFunc(err)
						}
					}
				}
			case websocket.CloseMessage:
				ws.CloseWs()
				return
			default:
				log.Println("error websocket message type , content is :\n", string(msg))
			}
		}
	}()
}

func (ws *WsConn) UpdateActiveTime() {
	ws.activeTimeL.Lock()
	defer ws.activeTimeL.Unlock()

	ws.activeTime = time.Now()
}

func (ws *WsConn) CloseWs() {
	ws.clearChannel(ws.closeCheck)
	ws.clearChannel(ws.closeReconnect)
	ws.clearChannel(ws.closeHeartbeat)
	ws.clearChannel(ws.closeRecv)

	ws.closeReconnect <- struct{}{}
	ws.closeHeartbeat <- struct{}{}
	ws.closeRecv <- struct{}{}
	ws.closeCheck <- struct{}{}

	err := ws.Close()
	if err != nil {
		log.Println("close websocket error , ", err)
	}
}

func (ws *WsConn) clearChannel(c chan struct{}) {
	for {
		if len(c) > 0 {
			<-c
		} else {
			break
		}
	}
}
