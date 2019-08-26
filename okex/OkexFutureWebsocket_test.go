package okex

import (
	"testing"
	"time"
)

func TestStart(t *testing.T) {

	ws := OKexFutureWebsocket{
		proxyUrl: "socks5://127.0.0.1:1090",
	}

	ws.Init()

	//config := goghostex.APIConfig{
	//	ApiSecretKey: "B7318B036B1C5C37BEA45DC3B12AD804",
	//	ApiKey: "a127cc13-2c21-4b19-9a3b-7be62ca8a6f1",
	//	ApiPassphrase: "strengthening",
	//}

	//if err := ws.Login(&config);err!=nil{
	//	t.Error(err)
	//	return
	//}

	if err := ws.Subscribe(
		`{"op": "subscribe", "args": ["futures/ticker:BTC-USD-190927", "futures/candle60s:BTC-USD-190927"]}`,
	); err != nil {
		t.Error(err)
		return
	}

	ws.Start()

	time.Sleep(60 * time.Second)

	ws.Close()

}
