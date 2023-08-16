package utils

import (
	"bytes"
	"encoding/json"
	"github.com/Danny-Dasilva/CycleTLS/cycletls"
	"github.com/RomainMichau/cloudscraper_go/cloudscraper"
	"strings"
)

func ListToString(list interface{}) string {
	tsjsons, err := JsonEncode(list)
	if err != nil {
		return ""
	}
	tsjsons = tsjsons[1 : len(tsjsons)-1]
	return tsjsons
}

func JsonEncode(source interface{}) (string, error) {
	bytesBuffer := &bytes.Buffer{}
	encoder := json.NewEncoder(bytesBuffer)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(source)
	if err != nil {
		return "", err
	}

	jsons := string(bytesBuffer.Bytes())
	tsjsons := strings.TrimSuffix(jsons, "\n")
	return tsjsons, nil
}

func CommDoWebRequest(url string) (string, error) {
	cli, err := cloudscraper.Init(false, false)
	options := cycletls.Options{
		Headers: map[string]string{"User-Agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 12_2_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/98.0.4758.102 Safari/537.36",
			"Accept":       "application/json",
			"Content-Type": "application/json"},

		//Proxy:           "http://127.0.0.1:1087",
		Timeout:         10,
		DisableRedirect: true,
	}
	resp, err := cli.Do(url, options, "GET")
	return resp.Body, err
}
