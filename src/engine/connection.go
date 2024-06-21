package engine

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	base "src/utils"
	"sync/atomic"
)

type Connection struct {
	Chain                string
	LatestBlockTimestamp base.Timestamp
	LatestBlock          string
}

type rpcResponse[T any] struct {
	Result T             `json:"result"`
	Error  *eip1474Error `json:"error"`
}

type eip1474Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Params are used during calls to the RPC.
type Params []interface{}

var rpcCounter uint32

type rpcPayload struct {
	Jsonrpc string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  `json:"params"`
	ID      int `json:"id"`
}

func rpcCall[T any](url string, headers map[string]string, method string, params Params) (*T, error) {
	payloadToSend := rpcPayload{
		Jsonrpc: "2.0",
		Method:  method,
		Params:  params,
		ID:      int(uint32(atomic.AddUint32(&rpcCounter, 1))),
	}

	if plBytes, err := json.Marshal(payloadToSend); err != nil {
		return nil, err
	} else {
		body := bytes.NewReader(plBytes)
		if request, err := http.NewRequest("POST", url, body); err != nil {
			return nil, err
		} else {
			request.Header.Set("Content-Type", "application/json")
			for key, value := range headers {
				request.Header.Set(key, value)
			}

			client := &http.Client{}
			if response, err := client.Do(request); err != nil {
				return nil, err
			} else if response.StatusCode != 200 {
				return nil, fmt.Errorf("%s: %d", response.Status, response.StatusCode)
			} else {
				defer response.Body.Close()

				if theBytes, err := io.ReadAll(response.Body); err != nil {
					return nil, err
				} else {
					var result rpcResponse[T]
					if err = json.Unmarshal(theBytes, &result); err != nil {
						return nil, err
					} else {
						if result.Error != nil {
							return nil, fmt.Errorf("%d: %s", result.Error.Code, result.Error.Message)
						}
						return &result.Result, nil
					}
				}
			}
		}
	}
}
