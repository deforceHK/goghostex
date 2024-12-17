package okex

import (
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strconv"
	"sync"
	"time"

	. "github.com/deforceHK/goghostex"
)

type LocalOrderBooks struct {
	*WSMarketOKEx
	BidData      map[string]map[int64]float64
	AskData      map[string]map[int64]float64
	SeqData      map[string]int64
	TsData       map[string]int64
	OrderBookMux *sync.RWMutex

	// if the channel is not nil, send the update message to the channel. User should read the channel in the loop.
	UpdateChan chan string
}

type OKBook struct {
	Action string `json:"action"`
	Arg    struct {
		Channel string `json:"channel"`
		InstId  string `json:"instId"`
	} `json:"arg"`
	Data []struct {
		Asks      [][]string `json:"asks"`
		Bids      [][]string `json:"bids"`
		Timestamp int64      `json:"ts,string"`
		Checksum  int64      `json:"checksum"`
		SeqId     int64      `json:"seqId"`
		PrevSeqId int64      `json:"prevSeqId"`
	} `json:"data"`
}

func (this *LocalOrderBooks) Init() error {
	if this.OrderBookMux == nil {
		this.OrderBookMux = &sync.RWMutex{}
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

	this.WSMarketOKEx.RecvHandler = func(s string) {
		this.Receiver(s)
	}
	return this.Start()
}

func (this *LocalOrderBooks) Receiver(msg string) {
	this.OrderBookMux.Lock()
	defer this.OrderBookMux.Unlock()
	var rawData = []byte(msg)
	var delta = OKBook{}
	_ = json.Unmarshal(rawData, &delta)

	if delta.Action == "snapshot" || delta.Action == "update" {
		var instId = delta.Arg.InstId
		var seqId = delta.Data[0].SeqId
		var prevSeqId = delta.Data[0].PrevSeqId
		var timestamp = delta.Data[0].Timestamp

		if delta.Action == "snapshot" {
			var bidData = make(map[int64]float64)
			var askData = make(map[int64]float64)

			for _, bid := range delta.Data[0].Bids {
				var price, _ = strconv.ParseFloat(bid[0], 64)
				var stdPrice = int64(price * 100000000)
				var amount, _ = strconv.ParseFloat(bid[1], 64)
				bidData[stdPrice] = amount
			}

			for _, ask := range delta.Data[0].Asks {
				var price, _ = strconv.ParseFloat(ask[0], 64)
				var stdPrice = int64(price * 100000000)
				var amount, _ = strconv.ParseFloat(ask[1], 64)
				askData[stdPrice] = amount
			}

			this.BidData[instId] = bidData
			this.AskData[instId] = askData
			this.SeqData[instId] = seqId
			this.TsData[instId] = timestamp
		} else {
			if prevSeqId != this.SeqData[instId] {
				log.Println(fmt.Sprintf(
					"The prevSeqId %d is not equal to the last seqId %d, in product %s. ",
					prevSeqId, this.SeqData[instId], instId,
				))

				this.Resubscribe(instId)
				return
			}

			for _, bid := range delta.Data[0].Bids {
				var price, _ = strconv.ParseFloat(bid[0], 64)
				var stdPrice = int64(price * 100000000)
				var amount, _ = strconv.ParseFloat(bid[1], 64)
				this.BidData[instId][stdPrice] = amount
				if amount == 0 {
					delete(this.BidData[instId], stdPrice)
				}
			}

			for _, ask := range delta.Data[0].Asks {
				var price, _ = strconv.ParseFloat(ask[0], 64)
				var stdPrice = int64(price * 100000000)
				var amount, _ = strconv.ParseFloat(ask[1], 64)
				this.AskData[instId][stdPrice] = amount
				if amount == 0 {
					delete(this.AskData[instId], stdPrice)
				}
			}
			this.SeqData[instId] = seqId
			this.TsData[instId] = timestamp

			if this.UpdateChan != nil {
				this.UpdateChan <- fmt.Sprintf("%s:%d", instId, timestamp)
			}
		}
	} else {
		fmt.Println(msg)
		fmt.Println("The action must in snapshot/update. ")
	}
}

func (this *LocalOrderBooks) Resubscribe(productId string) {
	var unSub = WSOpOKEx{
		Op: "unsubscribe",
		Args: []map[string]string{
			{
				"channel": "books",
				"instId":  productId,
			},
		},
	}

	var err = this.conn.WriteJSON(unSub)
	if err != nil {
		this.ErrorHandler(err)
	}
	time.Sleep(10 * time.Second)

	var sub = WSOpOKEx{
		Op: "subscribe",
		Args: []map[string]string{
			{
				"channel": "books",
				"instId":  productId,
			},
		},
	}

	err = this.conn.WriteJSON(sub)
	if err != nil {
		this.ErrorHandler(err)
	}
}

func (this *LocalOrderBooks) Snapshot(pair Pair) (*SwapDepth, error) {
	this.OrderBookMux.RLock()
	defer this.OrderBookMux.RUnlock()
	var symbol = pair.ToSymbol("-", true)
	var productId = fmt.Sprintf("%s-SWAP", symbol)

	var lastTime = time.UnixMilli(this.TsData[productId]).In(this.WSMarketOKEx.Config.Location)
	var depth = &SwapDepth{
		Pair:      pair,
		Timestamp: this.TsData[productId],
		Sequence:  this.SeqData[productId],
		Date:      lastTime.Format(GO_BIRTHDAY),
		AskList:   make(DepthRecords, 0),
		BidList:   make(DepthRecords, 0),
	}

	for stdPrice, amount := range this.BidData[productId] {
		depth.BidList = append(depth.BidList, DepthRecord{
			Price:  float64(stdPrice) / 100000000,
			Amount: amount,
		})
	}

	for stdPrice, amount := range this.AskData[productId] {
		depth.AskList = append(depth.AskList, DepthRecord{
			Price:  float64(stdPrice) / 100000000,
			Amount: amount,
		})
	}
	sort.Sort(sort.Reverse(depth.BidList))
	sort.Sort(depth.AskList)
	return depth, nil
}

func (this *LocalOrderBooks) Subscribe(pair Pair) {
	var symbol = pair.ToSymbol("-", true)
	var productId = fmt.Sprintf("%s-SWAP", symbol)

	var sub = WSOpOKEx{
		Op: "subscribe",
		Args: []map[string]string{
			{
				"channel": "books",
				"instId":  productId,
			},
		},
	}
	this.WSMarketOKEx.Subscribe(sub)
}

func (this *LocalOrderBooks) Unsubscribe(pair Pair) {
	var symbol = pair.ToSymbol("-", true)
	var productId = fmt.Sprintf("%s-SWAP", symbol)

	var unSub = WSOpOKEx{
		Op: "unsubscribe",
		Args: []map[string]string{
			{
				"channel": "books",
				"instId":  productId,
			},
		},
	}
	this.WSMarketOKEx.Unsubscribe(unSub)
}

func (this *LocalOrderBooks) SubSpotSwapPair(pair Pair) {
	this.WSMarketOKEx.Subscribe(WSOpOKEx{
		Op: "subscribe",
		Args: []map[string]string{
			{
				"channel": "books",
				"instId":  pair.ToSymbol("-", true),
			},
		},
	})
	this.WSMarketOKEx.Subscribe(WSOpOKEx{
		Op: "subscribe",
		Args: []map[string]string{
			{
				"channel": "books",
				"instId":  fmt.Sprintf("%s-SWAP", pair.ToSymbol("-", true)),
			},
		},
	})
}

func (this *LocalOrderBooks) SubscribeById(productId string) {
	var sub = WSOpOKEx{
		Op: "subscribe",
		Args: []map[string]string{
			{
				"channel": "books",
				"instId":  productId,
			},
		},
	}
	this.WSMarketOKEx.Subscribe(sub)
}

func (this *LocalOrderBooks) UnsubscribeById(productId string) {
	var unSub = WSOpOKEx{
		Op: "unsubscribe",
		Args: []map[string]string{
			{
				"channel": "books",
				"instId":  productId,
			},
		},
	}
	this.WSMarketOKEx.Unsubscribe(unSub)
}

func (this *LocalOrderBooks) SnapshotById(productId string) (*Depth, error) {
	this.OrderBookMux.RLock()
	defer this.OrderBookMux.RUnlock()

	var lastTime = time.UnixMilli(this.TsData[productId]).In(this.WSMarketOKEx.Config.Location)
	var depth = &Depth{
		Timestamp: this.TsData[productId],
		Sequence:  this.SeqData[productId],
		Date:      lastTime.Format(GO_BIRTHDAY),
		AskList:   make(DepthRecords, 0),
		BidList:   make(DepthRecords, 0),
	}

	for stdPrice, amount := range this.BidData[productId] {
		depth.BidList = append(depth.BidList, DepthRecord{
			Price:  float64(stdPrice) / 100000000,
			Amount: amount,
		})
	}

	for stdPrice, amount := range this.AskData[productId] {
		depth.AskList = append(depth.AskList, DepthRecord{
			Price:  float64(stdPrice) / 100000000,
			Amount: amount,
		})
	}
	sort.Sort(sort.Reverse(depth.BidList))
	sort.Sort(depth.AskList)
	return depth, nil
}
