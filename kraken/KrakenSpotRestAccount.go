package kraken

import (
	"encoding/json"
	"fmt"
	"time"

	. "github.com/deforceHK/goghostex"
)

func (s *Spot) GetAccount() (*Account, []byte, error) {
	var nowTS = fmt.Sprintf("%d", time.Now().UnixNano())
	var data = map[string]interface{}{
		"nonce": nowTS,
	}

	var sign, err = s.GetKrakenSign("/0/private/Balance", data)
	if err != nil {
		return nil, nil, err
	}

	body, _ := json.Marshal(data)

	resp, err := s.DoSignRequest("POST", "/0/private/Balance", string(body), sign, nil)
	if err != nil {
		return nil, nil, err
	} else {
		return nil, resp, nil
	}

}
