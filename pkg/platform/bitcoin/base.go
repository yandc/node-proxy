package bitcoin

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/go-kratos/kratos/v2/log"
	"gitlab.bixin.com/mili/node-proxy/pkg/platform/types"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const (
	ChainURL       = "chain.so"
	BlockstreamURL = "blockstream.info"
	BlockcypherURL = "api.blockcypher.com"
)

type platform struct {
	rpcURL []string
	log    *log.Helper
}

func NewBTCPlatform(rpcURL []string, logger log.Logger) types.Platform {
	log := log.NewHelper(log.With(logger, "module", "platform/bitcoin"))
	return &platform{rpcURL: rpcURL, log: log}
}

func (p *platform) GetBalance(ctx context.Context, address, tokenAddress, decimals string) (string, error) {
	for i := 0; i < len(p.rpcURL); i++ {
		client := getClient(p.rpcURL[i])
		if client != nil {
			balance, err := client.GetBalance(address)
			if err != nil {
				p.log.Error("get balance error:", err)
				continue
			}
			return balance,nil
		}
	}
	return "0", nil
}

func getClient(url string) types.BtcClient {
	if strings.Contains(url, ChainURL) {
		return ChainClient{url: url}
	} else if strings.Contains(url, BlockstreamURL) {
		return BlockStreamClient{url: url}
	} else if strings.Contains(url, BlockcypherURL) {
		return BlockCypherClient{url: url}
	}
	return nil
}

func (p *platform) GetRpcURL() []string {
	return p.rpcURL
}

func buildURL(u string, params map[string]string) (target *url.URL, err error) {
	target, err = url.Parse(u)
	if err != nil {
		return
	}
	values := target.Query()
	//Set parameters
	for k, v := range params {
		values.Set(k, v)
	}
	//add token to url, if present

	target.RawQuery = values.Encode()
	return
}

//getResponse is a boilerplate for HTTP GET responses.
func getResponse(target *url.URL, decTarget interface{}) (err error) {
	resp, err := http.Get(target.String())
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		err = respErrorMaker(resp.StatusCode, resp.Body)
		return
	}
	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(decTarget)
	return
}

//respErrorMaker checks error messages/if they are multiple errors
//serializes them into a single error message
func respErrorMaker(statusCode int, body io.Reader) (err error) {
	status := "HTTP " + strconv.Itoa(statusCode) + " " + http.StatusText(statusCode)
	if statusCode == 429 {
		err = errors.New(status)
		return
	}
	type errorJSON struct {
		Err    string `json:"error"`
		Errors []struct {
			Err string `json:"error"`
		} `json:"errors"`
	}
	var msg errorJSON
	dec := json.NewDecoder(body)
	err = dec.Decode(&msg)
	if err != nil {
		return err
	}
	var errtxt string
	errtxt += msg.Err
	for i, v := range msg.Errors {
		if i == len(msg.Errors)-1 {
			errtxt += v.Err
		} else {
			errtxt += v.Err + ", "
		}
	}
	if errtxt == "" {
		err = errors.New(status)
	} else {
		err = errors.New(status + ", Message(s): " + errtxt)
	}
	return
}
