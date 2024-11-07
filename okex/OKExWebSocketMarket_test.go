package okex

import (
	"fmt"
	"testing"
	"time"

	. "github.com/deforceHK/goghostex"
)

// go test -v ./okex/... -count=1 -run=Test_OKExWebSocketMarket
func Test_OKExWebSocketMarket(t *testing.T) {

	var ws = WSMarketOKEx{
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
			{
				"channel": "books",
				"instId":  "BTC-USDT-SWAP",
			},
		},
	}
	ws.Subscribe(subParam)

	select {}

}
