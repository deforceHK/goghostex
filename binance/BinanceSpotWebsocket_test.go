package binance

import "testing"

func TestBinanceFutureWebsocket_Start(t *testing.T) {
	ws := BinanceWebsocket{
		proxyUrl: "socks5://127.0.0.1:1090",
	}

	ws.Init()

	ws.Subscribe("bnbbtc@kline_1m")
	ws.Subscribe("bnbbtc@kline_1h")
	ws.Start()
}
