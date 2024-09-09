package kraken

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	. "github.com/deforceHK/goghostex"
)

/**
* unit test cmd
* go test -v ./kraken/... -count=1 -run=TestSwap_Kraken_GetContracts
*
**/
func TestSwap_Kraken_GetContracts(t *testing.T) {
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

	var kraken = New(config)
	if contracts, _, err := kraken.Swap.GetContracts(); err != nil {
		t.Error(err)
		return
	} else {
		for _, contract := range contracts {
			fmt.Println(contract)
		}
	}
}
