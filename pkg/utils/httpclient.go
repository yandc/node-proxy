package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

//jsonRequest is a jsonrpc request
type jsonRequest struct {
	ID      int           `json:"id"`
	Jsonrpc string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
}

// jsonResponse is a jsonrpc response
type jsonResponse struct {
	ID     uint64          `json:"id"`
	Result json.RawMessage `json:"result"`
	Error  *ErrorObject    `json:"error,omitempty"`
}

// ErrorObject is a jsonrpc error
type ErrorObject struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Error implements error interface
func (e *ErrorObject) Error() string {
	data, err := json.Marshal(e)
	if err != nil {
		return fmt.Sprintf("jsonrpc.internal marshal error: %v", err)
	}
	return string(data)
}

func JsonHttpsPost(url string, id int, method, jsonrpc string, out interface{}, params []interface{}) error {
	request := jsonRequest{
		ID:      id,
		Jsonrpc: jsonrpc,
		Method:  method,
		Params:  params,
	}
	str, err := json.Marshal(request)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(string(str)))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	var jsonResp jsonResponse
	if err = json.Unmarshal(body, &jsonResp); err != nil {
		return err
	}
	if jsonResp.Error != nil {
		return errors.New(jsonResp.Error.Message)
	}
	return json.Unmarshal(jsonResp.Result, &out)
}
