package okex

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"
)

func Test_OKExWebSocket(t *testing.T) {


	var subOp = OKExWSOp{
		 Op: "subscribe",
		 Args: []*OKExWSArg{{
			 Channel: "funding-rate",
			 InstId: "BTC-USD-SWAP",
		 }},
	}

	var ws = OKExTradeWebSocket{
		receive: func(res []byte) error {
			var response = string(res)
			if response == "pong" {
				return nil
			}
			fmt.Println(response)
			return nil
		},
	}

	ws.Init()

	//config := goghostex.APIConfig{
	//	ApiSecretKey: "",
	//	ApiKey: "",
	//	ApiPassphrase: "",
	//}
	//
	//if err := ws.Login(&config);err!=nil{
	//	t.Error(err)
	//	return
	//}


	// 将 subOp 转换为 json字符串
	var subOpByte, _ = json.Marshal(subOp)
	println(string(subOpByte))

	if err := ws.Subscribe(
		string(subOpByte),
	); err != nil {
		t.Error(err)
		return
	}

	ws.Start()

	time.Sleep(60 * time.Second)

	ws.Close()

}
