package utils

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
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

func HttpsGet(url string, params, headers map[string]string, out interface{}) error {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	for key, value := range headers {
		req.Header.Add(key, value)
	}
	q := req.URL.Query()
	for k, v := range params {
		q.Add(k, v)
	}
	req.URL.RawQuery = q.Encode()
	//client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if 200 != resp.StatusCode {
		return fmt.Errorf("%s", body)
	}
	if err := json.Unmarshal(body, &out); err != nil {
		return err
	}
	return nil
}

func CommHttpsForm(url, method string, params, headers map[string]string, reqBody, out interface{}) error {
	var bodyReader io.Reader
	if reqBody != nil {
		if value, ok := reqBody.(string); ok {
			bodyReader = strings.NewReader(value)
		} else {
			byts, err := json.Marshal(reqBody)
			if err != nil {
				return err
			}
			bodyReader = bytes.NewReader(byts)
		}
	}
	req, err := http.NewRequest(method, url, bodyReader)

	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	if params != nil {
		q := req.URL.Query()
		for k, v := range params {
			q.Add(k, v)
		}
		req.URL.RawQuery = q.Encode()
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return errors.New(string(body))
	}
	if err = json.Unmarshal(body, &out); err != nil {

		return err
	}
	return nil
}

func HttpsParamsPost(url string, params interface{}) (string, error) {
	var bodyReader string
	if value, ok := params.(string); ok {
		bodyReader = value
	} else {
		bytes, err := json.Marshal(params)
		if err != nil {
			return "", err
		}
		bodyReader = string(bytes)
	}
	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(bodyReader))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	client := http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", errors.New(string(body))
	}
	return string(body), nil
}

var GlobalTransport *http.Transport

var client *http.Client

func init() {
	//uu, _ := url.Parse("http://127.0.0.1:1087")
	GlobalTransport = &http.Transport{
		//TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		//Proxy:           http.ProxyURL(uu),
		MaxConnsPerHost: 50,
	}
	client = &http.Client{
		Transport: GlobalTransport,
	}
}

func GetGlobalClient() *http.Client {
	return client
}
