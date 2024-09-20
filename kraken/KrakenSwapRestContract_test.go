package kraken

import (
	"fmt"
	"net/http"
	"net/url"
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
		Endpoint: ENDPOINT,
		HttpClient: &http.Client{
			Transport: &http.Transport{
				Proxy: func(req *http.Request) (*url.URL, error) {
					return url.Parse(SWAP_PROXY_URL)
				},
			},
		},
		ApiKey:        SWAP_API_KEY,
		ApiSecretKey:  SWAP_API_SECRETKEY,
		ApiPassphrase: SWAP_API_PASSPHRASE,
		Location:      time.Now().Location(),
	}

	var kraken = New(config)
	if contracts, resp, err := kraken.Swap.GetContracts(); err != nil {
		t.Error(err)
		return
	} else {
		for _, contract := range contracts {
			fmt.Println(contract)
		}
		//go func() {
		t.Log(string(resp))
		//}()
	}
}
