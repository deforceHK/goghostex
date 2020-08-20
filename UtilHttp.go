package goghostex

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

func NewHttpRequest(
	client *http.Client,
	reqType,
	reqUrl,
	postData string,
	requstHeaders map[string]string,
) ([]byte, error) {

	var req *http.Request
	if strings.ToUpper(reqType) == http.MethodGet {
		//todo if here change to the newRequestWithoutContext
		req, _ = http.NewRequest(strings.ToUpper(reqType), reqUrl, nil)
	} else {
		req, _ = http.NewRequest(strings.ToUpper(reqType), reqUrl, strings.NewReader(postData))
	}

	if req.Header.Get("User-Agent") == "" {
		req.Header.Set(
			"User-Agent",
			"Mozilla/5.0 (Windows NT 5.1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/31.0.1650.63 Safari/537.36",
		)
	}
	if requstHeaders != nil {
		for k, v := range requstHeaders {
			req.Header.Add(k, v)
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	bodyData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, errors.New(fmt.Sprintf(
			"HttpStatusCode: %d, HttpMethod: %s, Response: %s, Request: %s, Url: %s",
			resp.StatusCode,
			reqType,
			string(bodyData),
			postData,
			reqUrl,
		))
	}

	return bodyData, nil
}
