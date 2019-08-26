package goghostex

import (
	"bytes"
	"compress/flate"
	"io/ioutil"
	"testing"
	"time"
)

func Test_time(t *testing.T) {
	t.Log(time.Now().Unix())
}

func ProtoHandle(data []byte) error {
	println(string(data))
	return nil
}

func GzipDecode(in []byte) ([]byte, error) {
	reader := flate.NewReader(bytes.NewReader(in))
	defer func() { _ = reader.Close() }()

	return ioutil.ReadAll(reader)

}

func TestNewWsConn(t *testing.T) {
	clientId := "a"
	//ts := time.Now().Unix()*1000 + 42029
	//args := make([]interface{}, 0)
	//args = append(args, ts)
	//
	//ping := fmt.Sprintf("{\"cmd\":\"ping\",\"args\":[%d],\"id\":\"%s\"}", ts, clientId)
	//ping2 := map[string]interface{}{
	//	"cmd":  "ping",
	//	"id":   clientId,
	//	"args": args}
	//ping3, err := json.Marshal(ping2)
	//
	//fmt.Println(ping)
	//fmt.Println(ping2)
	//fmt.Println(err, string(ping3))

	ws := NewWsBuilder().Dump().WsUrl(
		"wss://real.okex.com:8443/ws/v3",
	).ProxyUrl(
		"socks5://127.0.0.1:1090",
	).Heartbeat(
		[]byte("ping"),
		5*time.Second,
	).UnCompressFunc(GzipDecode).ProtoHandleFunc(ProtoHandle).Build()
	t.Log(ws.Subscribe(map[string]string{
		//"cmd":"sub", "args":"[\"ticker.btcusdt\"]", "id": clientId}))
		"cmd": "sub", "args": "ticker.btcusdt", "id": clientId}))

	//_ = ws.WriteMessage(
	//	websocket.TextMessage,
	//	[]byte(`{"channel":"ok_sub_futureusd_btc_depth_quarter","event":"addChannel"}`),
	//)
	ws.ReceiveMessage()
	time.Sleep(time.Second * 20)
	ws.CloseWs()
}
