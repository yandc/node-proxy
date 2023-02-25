package utils

import (
	"bytes"
	"encoding/json"
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
