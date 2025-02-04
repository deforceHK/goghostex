package kraken

import (
	"fmt"
	"time"

	. "github.com/deforceHK/goghostex"
)

func (s *Spot) GetAccount() (*Account, []byte, error) {
	var nowTS = fmt.Sprintf("%d", time.Now().UnixNano())
	var data = map[string]interface{}{
		"nonce": nowTS,
	}

	resp, err := s.DoSignRequest("POST", "/0/private/Balance", data, nil)
	if err != nil {
		return nil, nil, err
	} else {
		return nil, resp, nil
	}
}
