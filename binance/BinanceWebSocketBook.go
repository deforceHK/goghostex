package binance

import (
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	. "github.com/deforceHK/goghostex"
)

type LocalOrderBooks struct {
	*WSMarketUMBN
	BidData       map[string]map[int64]float64
	AskData       map[string]map[int64]float64
	SeqData       map[string]int64
	OrderBookMuxs map[string]*sync.Mutex
	Cache         map[string][]*DeltaOrderBook
}

type DeltaOrderBook struct {
	EventType      string     `json:"e"`
	EventTimestamp int64      `json:"E"`
	Timestamp      int64      `json:"T"`
	Symbol         string     `json:"s"`
	StartSeq       int64      `json:"U"`
	EndSeq         int64      `json:"u"`
	PrevSeq        int64      `json:"pu"`
	Bids           [][]string `json:"b"`
	Asks           [][]string `json:"a"`
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
	if this.Cache == nil {
		this.Cache = make(map[string][]*DeltaOrderBook)
	}
	this.RecvHandler = func(s string) {
		this.ReceiveDelta(s)
	}
	//this.WSTradeUMBN.ErrorHandler = func(err error) {
	//
	//}
	return this.Start()
}

func (this *LocalOrderBooks) Restart() {
	for productId, _ := range this.OrderBookMuxs {
		var mux = this.OrderBookMuxs[productId]
		mux.Lock()
		this.OrderBookMuxs[productId] = nil
		this.Cache[productId] = nil
		mux.Unlock()
	}

	this.WSMarketUMBN.Restart()
}

func (this *LocalOrderBooks) ReceiveDelta(msg string) {
	var delta = struct {
		Stream string          `json:"stream"`
		Data   *DeltaOrderBook `json:"data"`
	}{}

	_ = json.Unmarshal([]byte(msg), &delta)
	if delta.Stream == "" {
		log.Println(msg)
		return
	}

	var productId = strings.Split(delta.Stream, "@")[0]
	// 如果还没有锁，说明还没有申请过snapshot，或者snapshot重置了。
	if this.OrderBookMuxs[productId] == nil {
		this.BidData[productId] = make(map[int64]float64)
		this.AskData[productId] = make(map[int64]float64)
		if this.Cache[productId] == nil {
			this.Cache[productId] = []*DeltaOrderBook{delta.Data}
			go this.getSnapshot(productId, 0)
		} else {
			this.Cache[productId] = append(this.Cache[productId], delta.Data)
		}
		return
	}

	//	已经有了snapshot，则直接处理delta
	var mux = this.OrderBookMuxs[productId]
	mux.Lock()
	defer mux.Unlock()

	var withCache = false
	if len(this.Cache[productId]) > 0 {
		for _, cache := range this.Cache[productId] {
			if cache.EndSeq < this.SeqData[productId] {
				withCache = true
				continue
			}

			withCache = false
			for _, bid := range cache.Bids {
				var price, _ = strconv.ParseFloat(bid[0], 64)
				var stdPrice = int64(price * 100000000)
				var volume, _ = strconv.ParseFloat(bid[1], 64)

				this.BidData[productId][stdPrice] = volume
			}

			for _, ask := range cache.Asks {
				var price, _ = strconv.ParseFloat(ask[0], 64)
				var stdPrice = int64(price * 100000000)
				var volume, _ = strconv.ParseFloat(ask[1], 64)

				this.AskData[productId][stdPrice] = volume
			}
			this.SeqData[productId] = cache.EndSeq
		}
		this.Cache[productId] = make([]*DeltaOrderBook, 0)
	}

	if !withCache && delta.Data.PrevSeq != this.SeqData[productId] {
		// 有丢包现象，需要重新申请snapshot
		go this.getSnapshot(productId, 0)
	} else {
		for _, bid := range delta.Data.Bids {
			var price, _ = strconv.ParseFloat(bid[0], 64)
			var stdPrice = int64(price * 100000000)
			var volume, _ = strconv.ParseFloat(bid[1], 64)

			this.BidData[productId][stdPrice] = volume
		}

		for _, ask := range delta.Data.Asks {
			var price, _ = strconv.ParseFloat(ask[0], 64)
			var stdPrice = int64(price * 100000000)
			var volume, _ = strconv.ParseFloat(ask[1], 64)

			this.AskData[productId][stdPrice] = volume
		}
		this.SeqData[productId] = delta.Data.EndSeq
	}

}
func (this *LocalOrderBooks) getPairByProductId(productId string) Pair {
	return Pair{
		Basis:   NewCurrency(productId[:len(productId)-4], ""),
		Counter: NewCurrency(productId[len(productId)-4:], ""),
	}

}

func (this *LocalOrderBooks) getSnapshot(productId string, times int) {
	if this.OrderBookMuxs[productId] != nil {
		var mux = this.OrderBookMuxs[productId]
		mux.Lock()
		this.OrderBookMuxs[productId] = nil
		this.Cache[productId] = nil
		mux.Unlock()
		return
	}
	if times > 5 {
		this.Stop()
		this.ErrorHandler(&WSStopError{
			Msg: "get snapshot failed, and retry 5 times, stop the websocket. ",
		})
		return
	}

	var pair = this.getPairByProductId(productId)
	var bn = New(this.Config)
	var depth, _, err = bn.Swap.GetDepth(pair, 1000)
	if err != nil {
		this.ErrorHandler(err)
		time.Sleep(5 * time.Second)
		this.getSnapshot(productId, times+1)
		return
	}

	this.OrderBookMuxs[productId] = &sync.Mutex{}
	var mux = this.OrderBookMuxs[productId]
	mux.Lock()
	this.SeqData[productId] = depth.Sequence

	for _, bid := range depth.BidList {
		var stdPrice = int64(bid.Price * 100000000)
		this.BidData[productId][stdPrice] = bid.Amount
	}

	for _, ask := range depth.AskList {
		var stdPrice = int64(ask.Price * 100000000)
		this.AskData[productId][stdPrice] = ask.Amount
	}
	mux.Unlock()
}

func (this *LocalOrderBooks) Snapshot(pair Pair) (*SwapDepth, error) {
	var productId = pair.ToSymbol("", false)
	if this.BidData[productId] == nil || this.AskData[productId] == nil || this.OrderBookMuxs[productId] == nil {
		return nil, fmt.Errorf("The order book data is not ready or you need subscribe the productid. ")
	}

	var mux = this.OrderBookMuxs[productId]
	mux.Lock()
	defer mux.Unlock()

	var now = time.Now()
	var depth = &SwapDepth{
		Pair:      pair,
		Timestamp: now.UnixMilli(),
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
		for priceKey, amountValue := range this.BidData[productId] {
			if amountValue > 0 {
				continue
			}
			delete(this.BidData[productId], priceKey)
		}
		for priceKey, amountValue := range this.AskData[productId] {
			if amountValue > 0 {
				continue
			}
			delete(this.AskData[productId], priceKey)
		}
	}

	return depth, nil
}

func (this *LocalOrderBooks) Subscribe(pair Pair) {

	var symbol = pair.ToSymbol("", false)

	this.WSMarketUMBN.Subscribe(fmt.Sprintf("%s@depth", symbol))
}

func (this *LocalOrderBooks) Unsubscribe(pair Pair) {

	var symbol = pair.ToSymbol("", false)

	this.WSMarketUMBN.Unsubscribe(fmt.Sprintf("%s@depth", symbol))

}
