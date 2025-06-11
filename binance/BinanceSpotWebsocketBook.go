package binance

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	. "github.com/deforceHK/goghostex"
)

type LocalSpotBooks struct {
	*WSMarketSpot
	BidData       map[string]map[int64]float64
	AskData       map[string]map[int64]float64
	SeqData       map[string]int64
	TsData        map[string]int64
	OrderBookMuxs map[string]*sync.Mutex
	Cache         map[string][]*DeltaSpotBook

	// if the channel is not nil, send the update message to the channel. User should read the channel in the loop.
	UpdateChan chan string
}

type DeltaSpotBook struct {
	EventType      string     `json:"e"`
	EventTimestamp int64      `json:"E"`
	Symbol         string     `json:"s"`
	StartSeq       int64      `json:"U"`
	EndSeq         int64      `json:"u"`
	Bids           [][]string `json:"b"`
	Asks           [][]string `json:"a"`
}

func (this *LocalSpotBooks) Init() error {
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
		this.Cache = make(map[string][]*DeltaSpotBook)
	}
	this.RecvHandler = func(s string) {
		this.ReceiveDelta(s)
	}
	//this.WSTradeUMBN.ErrorHandler = func(err error) {
	//
	//}
	return this.Start()
}

func (this *LocalSpotBooks) ReceiveDelta(msg string) {
	var delta = DeltaSpotBook{}

	_ = json.Unmarshal([]byte(msg), &delta)
	if delta.EventType != "depthUpdate" {
		// it's not for depth update, ignore it.
		log.Println(msg)
		return
	}

	var productId = delta.Symbol
	// 如果还没有锁，说明还没有申请过snapshot，或者snapshot重置了。
	if this.OrderBookMuxs[productId] == nil {
		this.BidData[productId] = make(map[int64]float64)
		this.AskData[productId] = make(map[int64]float64)
		if this.Cache[productId] == nil {
			this.Cache[productId] = []*DeltaSpotBook{&delta}
			go this.getSnapshot(productId, 0)
		} else {
			this.Cache[productId] = append(this.Cache[productId], &delta)
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
			this.TsData[productId] = cache.EventTimestamp
		}
		this.Cache[productId] = make([]*DeltaSpotBook, 0)
	}

	if !withCache && delta.StartSeq != (this.SeqData[productId]+1) {
		// 有丢包现象，需要重新申请snapshot
		fmt.Println(delta.StartSeq, this.SeqData[productId])
		go this.getSnapshot(productId, 0)
	} else {
		for _, bid := range delta.Bids {
			var price, _ = strconv.ParseFloat(bid[0], 64)
			var stdPrice = int64(price * 100000000)
			var volume, _ = strconv.ParseFloat(bid[1], 64)

			this.BidData[productId][stdPrice] = volume
		}

		for _, ask := range delta.Asks {
			var price, _ = strconv.ParseFloat(ask[0], 64)
			var stdPrice = int64(price * 100000000)
			var volume, _ = strconv.ParseFloat(ask[1], 64)

			this.AskData[productId][stdPrice] = volume
		}
		this.SeqData[productId] = delta.EndSeq
		this.TsData[productId] = delta.EventTimestamp

		if this.UpdateChan != nil {
			this.UpdateChan <- fmt.Sprintf("%s:%d", productId, delta.EventTimestamp)
		}
	}

}

func (this *LocalSpotBooks) getSnapshot(productId string, times int) {
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

	var depth, err = this.getDepthById(productId, 1000)
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

func (this *LocalSpotBooks) getDepthById(productId string, size int) (*Depth, error) {

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

	fmt.Println("https://api.binance.com/api/v3/depth?" + params.Encode())
	var resp, err = NewHttpRequest(
		this.Config.HttpClient,
		http.MethodGet,
		"https://api.binance.com/api/v3/depth?"+params.Encode(),
		"",
		map[string]string{
			"X-MBX-APIKEY": this.Config.ApiKey,
		},
	)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	err = json.Unmarshal(resp, &response)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	var now = time.Now()
	var depth = new(Depth)
	//depth.Pair = pair
	depth.Timestamp = now.UnixMilli()
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

	return depth, nil
}

func (this *LocalSpotBooks) Subscribe(pair Pair) {

	var symbol = pair.ToSymbol("", false)

	this.WSMarketSpot.Subscribe(fmt.Sprintf("%s@depth", symbol))
}

func (this *LocalSpotBooks) Unsubscribe(pair Pair) {

	var symbol = pair.ToSymbol("", false)

	this.WSMarketSpot.Unsubscribe(fmt.Sprintf("%s@depth", symbol))

}
