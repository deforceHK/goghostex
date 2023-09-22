package okex

import (
	"fmt"
	"testing"
	"time"

	. "github.com/deforceHK/goghostex"
)

const (
	WEBSOCKET_API_KEY        = ""
	WEBSOCKET_API_SECRETKEY  = ""
	WEBSOCKET_API_PASSPHRASE = ""
)

func Test_OKExWebSocket(t *testing.T) {

	var ws = WSTradeOKEx{
		Config: &APIConfig{
			Endpoint:      ENDPOINT,
			HttpClient:    nil,
			ApiKey:        WEBSOCKET_API_KEY,
			ApiSecretKey:  WEBSOCKET_API_SECRETKEY,
			ApiPassphrase: WEBSOCKET_API_PASSPHRASE,
			Location:      time.Now().Location(),
		},
		RecvHandler: func(res string) {
			fmt.Println(res)
		},
	}

	ws.Start()

	var subParam = WSOpOKEx{
		Op: "subscribe",
		Args: []map[string]string{
			//{
			//	"channel": "funding-rate",
			//	"instId":  "BTC-USD-SWAP",
			//},
			//{
			//	"channel": "account",
			//	"ccy": "USDT",
			//	"extraParams": "0",
			//},
			//{
			//	"channel": "positions",
			//	"instType": "SWAP",
			//	//"extraParams": "0",
			//},
			{
				"channel": "balance_and_position",
			},
		},
	}
	ws.Subscribe(subParam)

	select {}

	//if err := ws.Login(); err != nil {
	//	t.Error(err)
	//	return
	//}
	//
	//var subArgs = []map[string]string{
	//	//{
	//	//	"channel": "funding-rate",
	//	//	"instId":  "BTC-USD-SWAP",
	//	//},
	//	{
	//		"channel": "account",
	//		"ccy":     "USDT",
	//	},
	//}
	//
	//var subParam = OKExWSOp{
	//	Op:   "subscribe",
	//	Args: subArgs,
	//}
	//// 将 subOp 转换为 json字符串
	//var subOpByte, _ = json.Marshal(subParam)
	//
	//ws.Start()
	//time.Sleep(60 * time.Second)
	//
	//if err := ws.Subscribe(
	//	string(subOpByte),
	//); err != nil {
	//	t.Error(err)
	//	return
	//}
	//
	//time.Sleep(60 * time.Second)
	//
	//ws.Close()

}

//func Test_OKExWebSocket1(t *testing.T) {
//
//	c, _, err := websocket.DefaultDialer.Dial("wss://ws.okx.com:8443/ws/v5/public", nil)
//	if err != nil {
//		log.Fatal("dial:", err)
//	}
//	defer c.Close()
//
//	var subParam = OKExWSOp{
//		Op: "subscribe",
//		Args: []map[string]string{
//			{
//				"channel": "funding-rate",
//				"instId":  "BTC-USD-SWAP",
//			},
//			//{
//			//	"channel": "account",
//			//	"ccy": "USDT",
//			//},
//		},
//	}
//
//	if err := c.WriteJSON(subParam); err != nil {
//		log.Fatal("subscribe:", err)
//	}
//
//	go func(conn *websocket.Conn) {
//		for {
//			time.Sleep(20 * time.Second)
//			conn.WriteMessage(websocket.TextMessage, []byte("ping"))
//		}
//	}(c)
//
//	for {
//		messageType, p, err := c.ReadMessage()
//		if err != nil {
//			log.Println("read:", err)
//			return
//		}
//		if messageType == websocket.TextMessage {
//			log.Printf("Received message: %s", p)
//		}
//	}
//}
