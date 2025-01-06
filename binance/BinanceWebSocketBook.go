package binance

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
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
	TsData        map[string]int64
	OrderBookMuxs map[string]*sync.Mutex
	Cache         map[string][]*DeltaOrderBook

	// if the channel is not nil, send the update message to the channel. User should read the channel in the loop.
	UpdateChan chan string
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
	if this.TsData == nil {
		this.TsData = make(map[string]int64)
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
			this.TsData[productId] = cache.Timestamp
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
		this.TsData[productId] = delta.Data.Timestamp

		if this.UpdateChan != nil {
			this.UpdateChan <- fmt.Sprintf("%s:%d", productId, delta.Data.Timestamp)
		}
	}

}

func (this *LocalOrderBooks) getPairByProductId(productId string) Pair {
	var theId = productId
	if strings.Index(productId, "_") > 0 {
		theId = strings.Split(productId, "_")[0]
	}
	return Pair{
		Basis:   NewCurrency(theId[:len(theId)-4], ""),
		Counter: NewCurrency(theId[len(theId)-4:], ""),
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

	var depth, err = this.getDepthById(productId, 1000) //bn.Swap.GetDepth(pair, 1000)
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

func (this *LocalOrderBooks) Snapshot(pair Pair) (*Depth, error) {
	var productId = pair.ToSymbol("", false)
	var depth, err = this.SnapshotById(productId)
	//depth.Pair = pair
	return depth, err
}

func (this *LocalOrderBooks) SnapshotById(productId string) (*Depth, error) {
	if this.BidData[productId] == nil || this.AskData[productId] == nil || this.OrderBookMuxs[productId] == nil {
		return nil, fmt.Errorf("The order book data is not ready or you need subscribe the productid. ")
	}

	var mux = this.OrderBookMuxs[productId]
	mux.Lock()
	defer mux.Unlock()
	var pair = this.getPairByProductId(productId)
	var lastTime = time.UnixMilli(this.TsData[productId]).In(this.WSMarketUMBN.Config.Location)
	var depth = &Depth{
		Pair:      pair,
		Sequence:  this.SeqData[productId],
		Timestamp: lastTime.UnixMilli(),
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

func (this *LocalOrderBooks) SubscribeById(productId string) {
	this.WSMarketUMBN.Subscribe(fmt.Sprintf("%s@depth", productId))
}

func (this *LocalOrderBooks) Subscribe(pair Pair) {

	var symbol = pair.ToSymbol("", false)

	this.WSMarketUMBN.Subscribe(fmt.Sprintf("%s@depth", symbol))
}

func (this *LocalOrderBooks) Unsubscribe(pair Pair) {

	var symbol = pair.ToSymbol("", false)

	this.WSMarketUMBN.Unsubscribe(fmt.Sprintf("%s@depth", symbol))

}

func (this *LocalOrderBooks) getDepthById(productId string, size int) (*Depth, error) {

	var sizes = []int{5, 10, 20, 50, 100, 500, 1000}
	for _, s := range sizes {
		if size <= s {
			size = s
			break
		}
	}

	var response = struct {
		Asks         [][]interface{} `json:"asks"` // The binance return asks Ascending ordered
		Bids         [][]interface{} `json:"bids"` // The binance return bids Descending ordered
		LastUpdateId int64           `json:"lastUpdateId"`
	}{}

	var params = url.Values{}
	params.Set("symbol", productId)
	params.Set("limit", fmt.Sprintf("%d", size))
	var bn = New(this.Config)
	var _, err = bn.Swap.DoRequest(
		http.MethodGet,
		SWAP_COUNTER_DEPTH_URI+params.Encode(),

		"",
		&response,
		SETTLE_MODE_COUNTER,
	)

	//fmt.Println(string(resp))

	var now = time.Now()
	var depth = new(Depth)
	//depth.Pair = pair
	depth.Timestamp = now.UnixNano() / int64(time.Millisecond)
	depth.Date = now.In(this.Config.Location).Format(GO_BIRTHDAY)
	depth.Sequence = response.LastUpdateId

	for _, bid := range response.Bids {
		var price = ToFloat64(bid[0])
		var amount = ToFloat64(bid[1])
		depthItem := DepthRecord{Price: price, Amount: amount}
		depth.BidList = append(depth.BidList, depthItem)
	}

	for _, ask := range response.Asks {
		var price = ToFloat64(ask[0])
		var amount = ToFloat64(ask[1])
		depthItem := DepthRecord{Price: price, Amount: amount}
		depth.AskList = append(depth.AskList, depthItem)
	}

	return depth, err
}
