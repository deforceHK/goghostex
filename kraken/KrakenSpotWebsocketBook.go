package kraken

import (
	"encoding/json"
	"fmt"
	"sort"
	"sync"
	"time"

	. "github.com/deforceHK/goghostex"
)

type SpotOrderBooks struct {
	*WSSpotMarketKK
	BidData       map[string]map[int64]float64
	AskData       map[string]map[int64]float64
	SeqData       map[string]int64
	TsData        map[string]int64
	OrderBookMuxs map[string]*sync.Mutex

	// if the channel is not nil, send the update message to the channel. User should read the channel in the loop.
	UpdateChan chan string
}

type KKBookSnapshot struct {
	Channel string `json:"channel"`
	Type    string `json:"type"`
	Data    []struct {
		Symbol string `json:"symbol"`
		Bids   []struct {
			Price float64 `json:"price"`
			Qty   float64 `json:"qty"`
		} `json:"bids"`
		Asks []struct {
			Price float64 `json:"price"`
			Qty   float64 `json:"qty"`
		} `json:"asks"`
		Checksum int64 `json:"checksum"`
	} `json:"data"`
}

type KKBookUpdate struct {
	Channel string `json:"channel"`
	Type    string `json:"type"`
	Data    []struct {
		Symbol string `json:"symbol"`
		Bids   []struct {
			Price float64 `json:"price"`
			Qty   float64 `json:"qty"`
		} `json:"bids"`
		Asks []struct {
			Price float64 `json:"price"`
			Qty   float64 `json:"qty"`
		} `json:"asks"`
		Checksum  int64  `json:"checksum"`
		Timestamp string `json:"timestamp"`
	} `json:"data"`
}

func (this *SpotOrderBooks) Init() error {
	if this.OrderBookMuxs == nil {
		this.OrderBookMuxs = make(map[string]*sync.Mutex)
	}
	if this.BidData == nil {
		this.BidData = make(map[string]map[int64]float64)
	}
	if this.AskData == nil {
		this.AskData = make(map[string]map[int64]float64)
	}
	if this.SeqData == nil {
		this.SeqData = make(map[string]int64)
	}
	if this.TsData == nil {
		this.TsData = make(map[string]int64)
	}

	this.WSSpotMarketKK.RecvHandler = func(s string) {
		this.Receiver(s)
	}
	return this.Start()
}

func (this *SpotOrderBooks) Subscribe(pair Pair) {
	var symbol = pair.ToSymbol("/", true)
	var sub = struct {
		Method string `json:"method"`
		Params struct {
			Channel string   `json:"channel"`
			Symbol  []string `json:"symbol"`
			Depth   int64    `json:"depth"`
		} `json:"params"`
	}{
		Method: "subscribe",
		Params: struct {
			Channel string   `json:"channel"`
			Symbol  []string `json:"symbol"`
			Depth   int64    `json:"depth"`
		}{
			"book",
			[]string{symbol},
			500,
		},
	}

	this.WSSpotMarketKK.Subscribe(sub)
}

func (this *SpotOrderBooks) Unsubscribe(pair Pair) {
	var symbol = pair.ToSymbol("/", true)
	var sub = struct {
		Method string `json:"method"`
		Params struct {
			Channel string   `json:"channel"`
			Symbol  []string `json:"symbol"`
			Depth   int64    `json:"depth"`
		} `json:"params"`
	}{
		Method: "unsubscribe",
		Params: struct {
			Channel string   `json:"channel"`
			Symbol  []string `json:"symbol"`
			Depth   int64    `json:"depth"`
		}{
			"book",
			[]string{symbol},
			500,
		},
	}
	this.WSSpotMarketKK.Unsubscribe(sub)

}

func (this *SpotOrderBooks) Receiver(msg string) {
	var rawData = []byte(msg)
	var pre = struct {
		Channel string `json:"channel"`
		Type    string `json:"type"`
	}{}

	_ = json.Unmarshal(rawData, &pre)
	if pre.Channel != "book" {
		fmt.Println("The feed must in book_snapshot book")
		return
	}

	if pre.Type == "update" {
		var book = KKBookUpdate{}
		_ = json.Unmarshal(rawData, &book)
		this.recvBook(book)
	} else if pre.Type == "snapshot" {
		var snapshot = KKBookSnapshot{}
		_ = json.Unmarshal(rawData, &snapshot)
		this.recvSnapshot(snapshot)
	}
}

func (this *SpotOrderBooks) recvBook(book KKBookUpdate) {
	for _, data := range book.Data {
		var mux, exist = this.OrderBookMuxs[data.Symbol]
		if !exist {
			return
		}

		mux.Lock()
		//var stdPrice = int64(book.Price * 100000000)
		//if book.Seq != this.SeqData[book.ProductId]+1 {
		//	//这样restart也可以，但是重新订阅是不是更轻量？
		//	this.Resubscribe(book.ProductId)
		//	return
		//}
		for _, bid := range data.Bids {
			var stdPrice = int64(bid.Price * 100000000)
			this.BidData[data.Symbol][stdPrice] = bid.Qty
		}

		for _, ask := range data.Asks {
			var stdPrice = int64(ask.Price * 100000000)
			this.AskData[data.Symbol][stdPrice] = ask.Qty
		}
		var updateTime, _ = time.ParseInLocation(time.RFC3339, data.Timestamp, this.Config.Location)
		this.SeqData[data.Symbol] = data.Checksum
		this.TsData[data.Symbol] = updateTime.UnixMilli()
		mux.Unlock()

		if this.UpdateChan != nil {
			this.UpdateChan <- fmt.Sprintf("%s:%d", data.Symbol, updateTime.UnixMilli())
		}
	}
}

func (this *SpotOrderBooks) recvSnapshot(snapshot KKBookSnapshot) {
	for _, data := range snapshot.Data {

		var _, exist = this.OrderBookMuxs[data.Symbol]
		if !exist {
			this.OrderBookMuxs[data.Symbol] = &sync.Mutex{}
		}

		var mux = this.OrderBookMuxs[data.Symbol]
		mux.Lock()

		var bidData = make(map[int64]float64)
		var askData = make(map[int64]float64)
		for _, bid := range data.Bids {
			var stdPrice = int64(bid.Price * 100000000)
			bidData[stdPrice] = bid.Qty
		}

		for _, ask := range data.Asks {
			var stdPrice = int64(ask.Price * 100000000)
			askData[stdPrice] = ask.Qty
		}

		this.BidData[data.Symbol] = bidData
		this.AskData[data.Symbol] = askData
		this.SeqData[data.Symbol] = data.Checksum
		this.TsData[data.Symbol] = 0
		mux.Unlock()
	}
}

func (this *SpotOrderBooks) Snapshot(pair Pair) (*Depth, error) {
	var productId = pair.ToSymbol("/", true)

	if this.BidData[productId] == nil || this.AskData[productId] == nil || this.OrderBookMuxs[productId] == nil {
		return nil, fmt.Errorf("The order book data is not ready or you need subscribe the productid. ")
	}

	var mux = this.OrderBookMuxs[productId]
	mux.Lock()
	defer mux.Unlock()

	var lastTime = time.UnixMilli(this.TsData[productId]).In(this.WSSpotMarketKK.Config.Location)
	var lastTS = lastTime.UnixMilli()
	var depth = &Depth{
		Pair:      pair,
		Timestamp: lastTS,
		Sequence:  this.SeqData[productId],
		Date:      lastTime.Format(GO_BIRTHDAY),
		AskList:   make(DepthRecords, 0),
		BidList:   make(DepthRecords, 0),
	}
	var zeroCount, sumCount = 0.0, 0.0
	for stdPrice, amount := range this.BidData[productId] {
		if amount > 0 {
			depth.BidList = append(depth.BidList, DepthRecord{
				Price:  float64(stdPrice) / 100000000,
				Amount: amount,
			})
		} else {
			zeroCount++
		}
		sumCount++
	}

	for stdPrice, amount := range this.AskData[productId] {
		if amount > 0 {
			depth.AskList = append(depth.AskList, DepthRecord{
				Price:  float64(stdPrice) / 100000000,
				Amount: amount,
			})
		} else {
			zeroCount++
		}
		sumCount++
	}
	sort.Sort(sort.Reverse(depth.BidList))
	sort.Sort(depth.AskList)

	// collect the zero amount data
	if zeroCount/sumCount > 0.3 {
		for stdPrice, amount := range this.BidData[productId] {
			if amount > 0 {
				continue
			}
			delete(this.BidData[productId], stdPrice)
		}
		for stdPrice, amount := range this.AskData[productId] {
			if amount > 0 {
				continue
			}
			delete(this.AskData[productId], stdPrice)
		}
	}
	return depth, nil
}

func (this *SpotOrderBooks) Resubscribe(productId string) {
	var unSub = struct {
		Method string `json:"method"`
		Params struct {
			Channel string   `json:"channel"`
			Symbol  []string `json:"symbol"`
			Depth   int64    `json:"depth"`
		} `json:"params"`
	}{
		Method: "unsubscribe",
		Params: struct {
			Channel string   `json:"channel"`
			Symbol  []string `json:"symbol"`
			Depth   int64    `json:"depth"`
		}{
			"book",
			[]string{productId},
			500,
		},
	}

	var err = this.conn.WriteJSON(unSub)
	if err != nil {
		this.ErrorHandler(err)
	}
	time.Sleep(10 * time.Second)

	var sub = struct {
		Method string `json:"method"`
		Params struct {
			Channel string   `json:"channel"`
			Symbol  []string `json:"symbol"`
			Depth   int64    `json:"depth"`
		} `json:"params"`
	}{
		Method: "subscribe",
		Params: struct {
			Channel string   `json:"channel"`
			Symbol  []string `json:"symbol"`
			Depth   int64    `json:"depth"`
		}{
			"book",
			[]string{productId},
			500,
		},
	}
	err = this.conn.WriteJSON(sub)
	if err != nil {
		this.ErrorHandler(err)
	}
}
