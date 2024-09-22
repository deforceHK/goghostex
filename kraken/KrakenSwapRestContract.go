package kraken

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	. "github.com/deforceHK/goghostex"
)

const (
	SWAP_CONTRACT_URI = "/api/v3/instruments"
)

func (swap *Swap) getContract(pair Pair) *SwapContract {
	defer swap.Unlock()
	swap.Lock()

	var now = time.Now().In(swap.config.Location)
	if now.After(swap.nextUpdateContractTime) {
		_, err := swap.updateContracts()
		//重试三次
		for i := 0; err != nil && i < 3; i++ {
			time.Sleep(time.Second)
			_, err = swap.updateContracts()
		}

		// init fail at first time, get a default one.
		if swap.nextUpdateContractTime.IsZero() && err != nil {
			swap.initManual()
			swap.nextUpdateContractTime = now.Add(10 * time.Minute)
		}
	}

	var symbol = pair.ToSymbol("", true)
	var krSymbol = fmt.Sprintf("PF_%s", symbol)
	if symbol == "BTCUSD" {
		krSymbol = "PF_XBTUSD"
	}
	return swap.swapContracts.ContractNameKV[krSymbol]
}

func (swap *Swap) updateContracts() ([]byte, error) {
	var contracts, resp, err = swap.GetContracts()
	if err != nil {
		return resp, err
	}

	var swapContracts = SwapContracts{
		ContractNameKV: make(map[string]*SwapContract, 0),
	}

	for _, contract := range contracts {
		swapContracts.ContractNameKV[contract.ContractName] = contract
	}

	// setting next update time.
	var nowTime = time.Now().In(swap.config.Location)
	var nextUpdateTime = time.Date(
		nowTime.Year(), nowTime.Month(), nowTime.Day(),
		16, 0, 0, 0, swap.config.Location,
	)
	if nowTime.Hour() >= 16 {
		nextUpdateTime = nextUpdateTime.AddDate(0, 0, 1)
	}

	swap.nextUpdateContractTime = nextUpdateTime
	swap.swapContracts = swapContracts
	return resp, nil
}

func (swap *Swap) initManual() {
	swap.swapContracts = SwapContracts{
		ContractNameKV: map[string]*SwapContract{
			"PF_XBTUSD": {
				Pair:            Pair{BTC, USD},
				Symbol:          "btc_usd",
				Exchange:        KRAKEN,
				ContractName:    "PF_XBTUSD",
				SettleMode:      SETTLE_MODE_COUNTER,
				UnitAmount:      1,
				TickSize:        1,
				PricePrecision:  0,
				AmountPrecision: 0,
			},
			"PF_ETHUSD": {
				Pair:            Pair{ETH, USD},
				Symbol:          "eth_usd",
				Exchange:        KRAKEN,
				ContractName:    "PF_ETHUSD",
				SettleMode:      SETTLE_MODE_COUNTER,
				UnitAmount:      1,
				TickSize:        0.1,
				PricePrecision:  1,
				AmountPrecision: 0,
			},
			"PF_BNBUSD": {
				Pair:            Pair{BNB, USD},
				Symbol:          "bnb_usd",
				Exchange:        KRAKEN,
				ContractName:    "PF_BNBUSD",
				SettleMode:      SETTLE_MODE_COUNTER,
				UnitAmount:      1,
				TickSize:        0.01,
				PricePrecision:  2,
				AmountPrecision: 0,
			},
			"PF_SOLUSD": {
				Pair:            Pair{SOL, USD},
				Symbol:          "sol_usd",
				Exchange:        KRAKEN,
				ContractName:    "PF_SOLUSD",
				SettleMode:      SETTLE_MODE_COUNTER,
				UnitAmount:      1,
				TickSize:        0.01,
				PricePrecision:  2,
				AmountPrecision: 0,
			},
		},
	}
}

func (swap *Swap) GetContracts() ([]*SwapContract, []byte, error) {
	var results = struct {
		Result      string `json:"result"`
		Instruments []struct {
			Symbol       string  `json:"symbol"`
			Type         string  `json:"type"`
			Underlying   string  `json:"underlying"`
			TickSize     float64 `json:"tickSize"`
			ContractSize float64 `json:"contractSize"`
			OpeningDate  string  `json:"openingDate"`
		} `json:"instruments"`
	}{}

	if resp, err := swap.DoRequest(
		SWAP_KRAKEN_ENDPOINT,
		http.MethodGet,
		SWAP_CONTRACT_URI,
		"",
		&results,
	); err != nil {
		return nil, resp, err
	} else {
		if results.Result != "success" {
			return nil, resp, fmt.Errorf(string(resp))
		}

		var contracts = make([]*SwapContract, 0)
		for _, inst := range results.Instruments {
			// PI FF PF not swap
			if !strings.HasPrefix(inst.Symbol, "PF") {
				continue
			}

			if !strings.HasSuffix(inst.Symbol, "USD") {
				continue
			}
			//var openTime time.Time
			//if t, err := time.Parse(time.RFC3339, inst.OpeningDate);err!=nil{
			//	continue
			//}else{
			//	openTime = t
			//}
			var coin = inst.Symbol[3 : len(inst.Symbol)-3]
			if coin == "XBT" {
				coin = "BTC"
			}
			var pair = Pair{
				NewCurrency(coin, ""),
				USD,
			}

			contracts = append(contracts, &SwapContract{
				Pair:            pair,
				Symbol:          pair.ToSymbol("_", false),
				Exchange:        KRAKEN,
				ContractName:    inst.Symbol,
				SettleMode:      SETTLE_MODE_COUNTER,
				UnitAmount:      inst.ContractSize,
				TickSize:        inst.TickSize,
				PricePrecision:  GetPrecisionInt64(inst.TickSize),
				AmountPrecision: GetPrecisionInt64(inst.ContractSize),
			})
		}
		return contracts, resp, nil
	}

}
