package kraken

import (
	"fmt"
	. "github.com/deforceHK/goghostex"
	"net/http"
	"strings"
)

const (
	SWAP_CONTRACT_URI = "/derivatives/api/v3/instruments"
)

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

	if resp, err := swap.DoRequest(http.MethodGet, SWAP_CONTRACT_URI, "", &results); err != nil {
		return nil, resp, err
	} else {
		if results.Result != "success" {
			return nil, resp, fmt.Errorf(string(resp))
		}

		var contracts = make([]*SwapContract, 0)
		for _, inst := range results.Instruments {
			// PI FF FI not swap
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
