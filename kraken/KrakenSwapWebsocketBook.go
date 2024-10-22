package kraken

import (
	"encoding/json"
	"fmt"
	"sort"
	"sync"
	"time"

	. "github.com/deforceHK/goghostex"
)

type KKBook struct {
	ProductId string  `json:"product_id"`
	Side      string  `json:"side"`
	Seq       int64   `json:"seq"`
	Price     float64 `json:"price"`
	Qty       float64 `json:"qty"`
	Timestamp int64   `json:"timestamp"`
}

type KKSnapshot struct {
	ProductId string `json:"product_id"`
	Seq       int64  `json:"seq"`
	Bids      []*struct {
		Price float64 `json:"price"`
		Qty   float64 `json:"qty"`
	} `json:"bids"`
	Asks []*struct {
		Price float64 `json:"price"`
		Qty   float64 `json:"qty"`
	} `json:"asks"`
	Timestamp int64 `json:"timestamp"`
}

type LocalOrderBooks struct {
	*WSSwapMarketKK
	BidData       map[string]map[int64]float64
	AskData       map[string]map[int64]float64
	SeqData       map[string]int64
	OrderBookMuxs map[string]*sync.Mutex
}

func (this *LocalOrderBooks) Init() error {
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

	this.WSSwapMarketKK.RecvHandler = func(s string) {
		this.Receiver(s)
	}
	return this.Start()
}

//func (this *LocalOrderBooks) Restart() {
//	for productId := range this.OrderBookMuxs {
//		this.BidData[productId] = make(map[int64]float64)
//		this.AskData[productId] = make(map[int64]float64)
//		this.SeqData[productId] = 0
//	}
//
//	this.WSSwapMarketKK.Restart()
//}

func (this *LocalOrderBooks) Subscribe(pair Pair) {
	var symbol = pair.ToSymbol("", true)
	if symbol == "BTCUSD" {
		symbol = "XBTUSD"
	}

	var sub = struct {
		Event      string   `json:"event"`
		Feed       string   `json:"feed"`
		ProductIds []string `json:"product_ids"`
	}{
		"subscribe", "book", []string{fmt.Sprintf("PF_%s", symbol)},
	}

	this.WSSwapMarketKK.Subscribe(sub)
}

func (this *LocalOrderBooks) Unsubscribe(pair Pair) {
	var symbol = pair.ToSymbol("", true)
	if symbol == "BTCUSD" {
		symbol = "XBTUSD"
	}

	var sub = struct {
		Event      string   `json:"event"`
		Feed       string   `json:"feed"`
		ProductIds []string `json:"product_ids"`
	}{
		"unsubscribe", "book", []string{fmt.Sprintf("PF_%s", symbol)},
	}

	this.WSSwapMarketKK.Unsubscribe(sub)

}

func (this *LocalOrderBooks) Receiver(msg string) {
	var rawData = []byte(msg)
	var pre = struct {
		Feed string `json:"feed"`
	}{}
	_ = json.Unmarshal(rawData, &pre)

	if pre.Feed == "book" {
		var book = KKBook{}
		_ = json.Unmarshal(rawData, &book)
		this.recvBook(book)
	} else if pre.Feed == "book_snapshot" {
		var snapshot = KKSnapshot{}
		_ = json.Unmarshal(rawData, &snapshot)
		this.recvSnapshot(snapshot)
	} else {
		fmt.Println("The feed must in book_snapshot book")
	}
}

func (this *LocalOrderBooks) recvBook(book KKBook) {
	var mux, exist = this.OrderBookMuxs[book.ProductId]
	if !exist {
		return
	}

	mux.Lock()
	defer mux.Unlock()

	var stdPrice = int64(book.Price * 100000000)
	if book.Seq != this.SeqData[book.ProductId]+1 {
		//这样restart也可以，但是重新订阅是不是更轻量？
		this.Restart()
		return
	}

	if book.Side == "buy" {
		this.BidData[book.ProductId][stdPrice] = book.Qty
	} else {
		this.AskData[book.ProductId][stdPrice] = book.Qty
	}
	this.SeqData[book.ProductId] = book.Seq
}

func (this *LocalOrderBooks) recvSnapshot(snapshot KKSnapshot) {
	var _, exist = this.OrderBookMuxs[snapshot.ProductId]
	if !exist {
		this.OrderBookMuxs[snapshot.ProductId] = &sync.Mutex{}
	}

	var mux = this.OrderBookMuxs[snapshot.ProductId]
	mux.Lock()
	defer mux.Unlock()

	var bidData = make(map[int64]float64)
	var askData = make(map[int64]float64)
	for _, bid := range snapshot.Bids {
		var stdPrice = int64(bid.Price * 100000000)
		bidData[stdPrice] = bid.Qty
	}

	for _, ask := range snapshot.Asks {
		var stdPrice = int64(ask.Price * 100000000)
		askData[stdPrice] = ask.Qty
	}

	this.BidData[snapshot.ProductId] = bidData
	this.AskData[snapshot.ProductId] = askData
	this.SeqData[snapshot.ProductId] = snapshot.Seq

}

func (this *LocalOrderBooks) Snapshot(pair Pair) (*SwapDepth, error) {
	var symbol = pair.ToSymbol("", true)
	if symbol == "BTCUSD" {
		symbol = "XBTUSD"
	}
	var productId = fmt.Sprintf("PF_%s", symbol)

	if this.BidData[productId] == nil || this.AskData[productId] == nil || this.OrderBookMuxs[productId] == nil {
		return nil, fmt.Errorf("The order book data is not ready or you need subscribe the productid. ")
	}

	var mux = this.OrderBookMuxs[productId]
	mux.Lock()
	defer mux.Unlock()

	var now = time.Now()
	var depth = &SwapDepth{
		Pair:      pair,
		Timestamp: now.Unix(),
		Sequence:  this.SeqData[productId],
		Date:      now.Format(GO_BIRTHDAY),
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
