package utils

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"gitlab.bixin.com/mili/node-proxy/pkg/platform/types"
	"io/ioutil"
	"math/big"
	"net/http"
	"strings"
	"time"
)

func UpdateDecimals(amount string, decimals int) string {
	var result string
	//amount := balance.String()
	if len(amount) > decimals {
		result = fmt.Sprintf("%s.%s", amount[0:len(amount)-decimals], amount[len(amount)-decimals:])
	} else {
		sub := decimals - len(amount)
		var zero string
		for i := 0; i < sub; i++ {
			zero += "0"
		}
		result = "0." + zero + amount
	}
	return Clean(strings.TrimRight(result, "0"))
}

func BigIntString(balance *big.Int, decimals int) string {
	var result string
	amount := balance.String()
	if len(amount) > decimals {
		result = fmt.Sprintf("%s.%s", amount[0:len(amount)-decimals], amount[len(amount)-decimals:])
	} else {
		sub := decimals - len(amount)
		var zero string
		for i := 0; i < sub; i++ {
			zero += "0"
		}
		result = "0." + zero + amount
	}
	return Clean(strings.TrimRight(result, "0"))
}

func Clean(newNum string) string {
	stringBytes := bytes.TrimRight([]byte(newNum), "0")
	newNum = string(stringBytes)
	if stringBytes[len(stringBytes)-1] == 46 {
		newNum = newNum[:len(stringBytes)-1]
	}
	if stringBytes[0] == 46 {
		newNum = "0" + newNum
	}
	return newNum
}

func HttpsPost(url string, id int, method, jsonrpc string, out interface{}, params []interface{}, args ...interface{}) error {
	request := types.STCRequest{
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
	client := http.DefaultClient
	if len(args) > 0 {
		client.Timeout = time.Duration(args[0].(int)) * time.Millisecond
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	//fmt.Println(string(body))
	err = json.Unmarshal(body, out)
	return err
}

func HttpsForm(url, method string, params map[string]string, reqBody, out interface{}, args ...interface{}) error {
	var bodyReader string
	if value, ok := reqBody.(string); ok {
		bodyReader = value
	} else {
		bytes, _ := json.Marshal(reqBody)
		bodyReader = string(bytes)
	}
	req, err := http.NewRequest(method, url, strings.NewReader(bodyReader))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if params != nil {
		q := req.URL.Query()
		for k, v := range params {
			q.Add(k, v)
		}
		req.URL.RawQuery = q.Encode()
	}
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	if len(args) > 0 {
		client.Timeout = time.Duration(args[0].(int)) * time.Millisecond
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(body, &out); err != nil {
		return err
	}
	return nil
}

func HttpsGetForm(url string, params map[string]string, out interface{}) error {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if params != nil {
		q := req.URL.Query()
		for k, v := range params {
			q.Add(k, v)
		}
		req.URL.RawQuery = q.Encode()
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
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
	if err := json.Unmarshal(body, &out); err != nil {
		return err
	}
	return nil
}
