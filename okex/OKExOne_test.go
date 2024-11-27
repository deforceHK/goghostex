package okex

import (
	"net/http"
	"net/url"
	"testing"
)

/**
* unit test cmd
* go test -v ./okex/... -count=1 -run=TestOKExOne_GetProducts
*
**/

func TestOKExOne_GetProducts(t *testing.T) {

	var okOne = OKExOne{
		HttpClient: &http.Client{
			Transport: &http.Transport{
				Proxy: func(req *http.Request) (*url.URL, error) {
					return url.Parse(PROXY_URL)
				},
			},
		},
		//Location:loc,
	}

	var err = okOne.Init()
	if err != nil {
		t.Error(err)
		return
	}

	var resp, products, productErr = okOne.GetProducts("FUTURES")
	if productErr != nil {
		t.Log(string(resp))
		t.Error(productErr)
		return
	}

	t.Log(string(resp))

	for _, p := range products {
		t.Log(p.InstId)
	}
}
