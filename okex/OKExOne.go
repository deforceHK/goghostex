package okex

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	. "github.com/deforceHK/goghostex"
)

const OKEXONE_ENDPOINT = "https://www.okx.com"

type OKExOne struct {
	HttpClient    *http.Client
	Endpoint      string
	ApiKey        string
	ApiSecretKey  string
	ApiPassphrase string

	LastTimestamp int64
	Location      *time.Location
}

type Instrument struct {
	InstType       string  `json:"instType"`
	InstId         string  `json:"instId"`
	BaseCcy        string  `json:"baseCcy"`
	QuoteCcy       string  `json:"quoteCcy"`
	SettleCcy      string  `json:"settleCcy"`
	CtVal          string  `json:"ctVal"`
	CtMult         string  `json:"ctMult"`
	CtValCcy       string  `json:"ctValCcy"`
	OptType        string  `json:"optType"`
	Stk            string  `json:"stk"`
	ListTime       string   `json:"listTime"`
	AuctionEndTime string   `json:"auctionEndTime"`
	ExpTime        string   `json:"expTime"`
	Lever          string   `json:"lever"`
	TickSz         string `json:"tickSz"`
	LotSz          string `json:"lotSz"`
	MinSz          string `json:"minSz"`
	CtType         string  `json:"ctType"`
	State          string  `json:"state"`
	RuleType       string  `json:"ruleType"`
	MaxLmtSz       string `json:"maxLmtSz"`
	MaxMktSz       string `json:"maxMktSz"`
	MaxLmtAmt      string `json:"maxLmtAmt"`
	MaxMktAmt      string `json:"maxMktAmt"`
}

func (ok *OKExOne) Init() error {
	if ok.HttpClient == nil {
		ok.HttpClient = &http.Client{}
	}

	if ok.Endpoint == "" {
		ok.Endpoint = OKEXONE_ENDPOINT
	}

	if ok.Location == nil {
		var loc, _ = time.LoadLocation("Asia/Shanghai")
		ok.Location = loc
	}

	return nil
}

func (ok *OKExOne) RequestDirect(
	httpMethod,
	uri,
	reqBody string,
	response interface{},
) ([]byte, error) {

	var reqUrl = ok.Endpoint + uri
	var resp, err = NewHttpRequest(
		ok.HttpClient,
		httpMethod, reqUrl,
		reqBody,
		map[string]string{
			CONTENT_TYPE: APPLICATION_JSON_UTF8,
			ACCEPT:       APPLICATION_JSON,
		},
	)
	if err != nil {
		return nil, err
	} else {
		var nowTimestamp = time.Now().Unix() * 1000
		if nowTimestamp > ok.LastTimestamp {
			ok.LastTimestamp = nowTimestamp
		}
		return resp, json.Unmarshal(resp, &response)
	}
}

func (ok *OKExOne) RequestAuth(
	httpMethod,
	uri string,
	request,
	response interface{},
) ([]byte, error) {
	if request == nil {
		return nil, errors.New("illegal parameter")
	}
	var data, err = json.Marshal(request)
	if err != nil {
		return nil, err
	}

	var requestBody = string(data)

	var url = ok.Endpoint + uri
	sign, timestamp := ok.doParamSign(httpMethod, uri, requestBody)
	var resp, respErr = NewHttpRequest(
		ok.HttpClient,
		httpMethod,
		url, requestBody,
		map[string]string{
			CONTENT_TYPE:         APPLICATION_JSON_UTF8,
			ACCEPT:               APPLICATION_JSON,
			OK_ACCESS_KEY:        ok.ApiKey,
			OK_ACCESS_PASSPHRASE: ok.ApiPassphrase,
			OK_ACCESS_SIGN:       sign,
			OK_ACCESS_TIMESTAMP:  fmt.Sprint(timestamp),
		},
	)
	if respErr != nil {
		return nil, respErr
	} else {
		nowTimestamp := time.Now().Unix() * 1000
		if nowTimestamp > ok.LastTimestamp {
			ok.LastTimestamp = nowTimestamp
		}
		return resp, json.Unmarshal(resp, &response)
	}
}

func (ok *OKExOne) doParamSign(httpMethod, uri, requestBody string) (string, string) {
	var utcTime = time.Now().UTC()
	var isoTime = utcTime.String()
	isoBytes := []byte(isoTime)
	isoTime = string(isoBytes[:10]) + "T" + string(isoBytes[11:23]) + "Z"

	preText := fmt.Sprintf("%s%s%s%s", isoTime, strings.ToUpper(httpMethod), uri, requestBody)
	sign, _ := GetParamHmacSHA256Base64Sign(ok.ApiSecretKey, preText)
	return sign, isoTime
}

func (ok *OKExOne) GetProducts(tradeType string) ([]byte, []*Instrument, error) {
	var uri = "/api/v5/public/instruments?"
	var params = url.Values{}
	params.Set("instType", tradeType)
	uri += params.Encode()

	var response struct {
		Code string        `json:"code"`
		Msg  string        `json:"msg"`
		Data []*Instrument `json:"data"`
	}
	var rawResp, err = ok.RequestDirect(http.MethodGet, uri, "", &response)
	if err != nil {
		return rawResp, nil, err
	}
	if response.Code != "0" {
		return rawResp, nil, errors.New(response.Msg)
	}

	return rawResp, response.Data, nil
}
