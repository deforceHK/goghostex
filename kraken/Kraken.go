package kraken

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	. "github.com/deforceHK/goghostex"
)

const (
	ENDPOINT    = "https://api.kraken.com"
	API_V1      = "/0/public/"
	API_PRIVATE = "/0/private"

	KLINE_URI = "OHLC"
)

var _INERNAL_KLINE_PERIOD_CONVERTER = map[int]string{
	KLINE_PERIOD_1MIN:  "1",
	KLINE_PERIOD_5MIN:  "5",
	KLINE_PERIOD_15MIN: "15",
	KLINE_PERIOD_30MIN: "30",
	KLINE_PERIOD_1H:    "60",
	KLINE_PERIOD_4H:    "240",
	KLINE_PERIOD_1DAY:  "1440",
}

func New(config *APIConfig) *Kraken {
	var k = &Kraken{config: config}
	k.Spot = &Spot{k}
	k.Swap = &Swap{
		Kraken:        k,
		Locker:        new(sync.Mutex),
		swapContracts: SwapContracts{},
	}
	return k
}

type Kraken struct {
	config *APIConfig
	Spot   *Spot
	Swap   *Swap
	//Margin *Margin
	//Future *Future
}

func (k *Kraken) GetExchangeName() string {
	return KRAKEN
}

func (k *Kraken) DoRequest(httpMethod, uri, reqBody string, response interface{}) ([]byte, error) {
	resp, err := NewHttpRequest(
		k.config.HttpClient,
		httpMethod,
		k.config.Endpoint+uri,
		reqBody,
		map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		},
	)

	if err != nil {
		return nil, err
	} else {
		nowTimestamp := time.Now().Unix() * 1000
		if k.config.LastTimestamp < nowTimestamp {
			k.config.LastTimestamp = nowTimestamp
		}
		return resp, json.Unmarshal(resp, &response)
	}
}

func (k *Kraken) DoSignRequest(httpMethod, uri string, data interface{}, response interface{}) ([]byte, error) {
	var sign, signErr = k.GetKrakenSign(uri, data)
	if signErr != nil {
		return nil, signErr
	}

	var postData, _ = json.Marshal(data)
	resp, err := NewHttpRequest(
		k.config.HttpClient,
		httpMethod,
		k.config.Endpoint+uri,
		string(postData),
		map[string]string{
			"Content-Type": "application/json",
			"Accept":       "application/json",
			"API-Key":      k.config.ApiKey,
			"API-Sign":     sign,
		},
	)

	if err != nil {
		return nil, err
	} else {
		nowTimestamp := time.Now().Unix() * 1000
		if k.config.LastTimestamp < nowTimestamp {
			k.config.LastTimestamp = nowTimestamp
		}
		return resp, json.Unmarshal(resp, &response)
	}

}

func (k *Kraken) GetKrakenSign(urlPath string, data interface{}) (string, error) {
	var encodedData string

	switch v := data.(type) {
	case string:
		var jsonData map[string]interface{}
		if err := json.Unmarshal([]byte(v), &jsonData); err != nil {
			return "", err
		}
		encodedData = jsonData["nonce"].(string) + v
	case map[string]interface{}:
		//dataMap := url.Values{}
		//for key, value := range v {
		//	dataMap.Set(key, fmt.Sprintf("%v", value))
		//}
		//encodedData = v["nonce"].(string) + dataMap.Encode()
		body, _ := json.Marshal(v)
		encodedData = v["nonce"].(string) + string(body)
	default:
		return "", fmt.Errorf("invalid data type")
	}
	sha := sha256.New()
	sha.Write([]byte(encodedData))
	shaSum := sha.Sum(nil)

	message := append([]byte(urlPath), shaSum...)
	decodeSec, err := base64.StdEncoding.DecodeString(k.config.ApiSecretKey)
	if err != nil {
		return "", err
	}
	mac := hmac.New(sha512.New, decodeSec)
	mac.Write(message)
	macSum := mac.Sum(nil)
	sigDigest := base64.StdEncoding.EncodeToString(macSum)
	return sigDigest, nil
}

func (k *Kraken) GetToken() ([]byte, string, error) {
	var nonce = fmt.Sprintf("%d", time.Now().UnixNano())
	var data = map[string]interface{}{
		"nonce": nonce,
	}

	var response = struct {
		Error  []string `json:"error"`
		Result struct {
			Token   string `json:"token"`
			Expires int64  `json:"expires"`
		} `json:"result"`
	}{}
	var resp, err = k.DoSignRequest(
		http.MethodPost,
		API_PRIVATE+"/GetWebSocketsToken",
		data, &response,
	)

	if err != nil {
		return resp, "", err
	} else {
		if len(response.Error) != 0 {
			return resp, "", fmt.Errorf("%s", response.Error)
		}
		return resp, response.Result.Token, nil
	}
}
