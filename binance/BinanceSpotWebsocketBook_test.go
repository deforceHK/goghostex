package binance

import (
	"net/http"
	"testing"
	"time"

	. "github.com/deforceHK/goghostex"
)

// go test -v ./binance/... -count=1 -run=TestBinanceWebsocketSpotBook
func TestBinanceWebsocketSpotBook(t *testing.T) {

	var config = &APIConfig{
		Endpoint:   ENDPOINT,
		HttpClient: &http.Client{
			//Transport: &http.Transport{
			//	Proxy: func(req *http.Request) (*url.URL, error) {
			//		return url.Parse(PROXY_URL)
			//	},
			//},
		},
		ApiKey:        SWAP_API_KEY,
		ApiSecretKey:  SWAP_API_SECRETKEY,
		ApiPassphrase: SWAP_API_PASSPHRASE,
		Location:      time.Now().Location(),
	}

	var book = &LocalSpotBooks{
		WSMarketSpot: &WSMarketSpot{
			Config: config,
		},
	}

	var err = book.Init()
	if err != nil {
		t.Error(err)
		return
	}

	book.Subscribe(Pair{BTC, USDT})
	book.Subscribe(Pair{ETH, BTC})

	for i := 0; ; i++ {
		time.Sleep(5 * time.Second)
		if i%2 == 0 {
			depth, depthErr := book.Snapshot(Pair{BTC, USDT})
			if depthErr != nil {
				t.Error(depthErr)
				return
			}
			t.Log(
				depth.BidList[0].Price,
				depth.BidList[0].Amount,
				depth.AskList[0].Price,
				depth.AskList[0].Amount,
				depth.Pair.ToSymbol("", true),
			)
		} else {
			depth, depthErr := book.Snapshot(Pair{ETH, BTC})
			if depthErr != nil {
				t.Error(depthErr)
				return
			}
			t.Log(
				depth.BidList[0].Price,
				depth.BidList[0].Amount,
				depth.AskList[0].Price,
				depth.AskList[0].Amount,
				depth.Pair.ToSymbol("", true),
			)
		}
	}

	//select {}

}
