package starcoin

import (
	"context"
	"encoding/json"
	"github.com/go-kratos/kratos/v2/log"
	"gitlab.bixin.com/mili/node-proxy/pkg/platform/types"
	"gitlab.bixin.com/mili/node-proxy/pkg/platform/utils"
	"math/big"
	"strconv"
)

const (
	ID101          = 101
	ID200          = 200
	JSONRPC        = "2.0"
	GAS_TOKEN_CODE = "0x1::STC::STC"
)

type platform struct {
	rpcURL []string
	log    *log.Helper
}

func NewSTCPlatform(rpcURL []string, logger log.Logger) types.Platform {
	log := log.NewHelper(log.With(logger, "module", "platform/bitcoin"))
	return &platform{rpcURL: rpcURL, log: log}
}

func (p *platform) GetBalance(ctx context.Context, address, tokenAddress, decimals string) (string, error) {

	//get native balance
	if tokenAddress == "" || address == tokenAddress {
		for i := 0; i < len(p.rpcURL); i++ {
			balance, err := getStcBalance(p.rpcURL[i], address, GAS_TOKEN_CODE, 9)
			if err != nil {
				p.log.Error("get stc balance error:", err)
				continue
			}
			return balance, nil
		}
	}
	if address != tokenAddress && tokenAddress != "" && decimals != "" {
		decimalsInt, _ := strconv.Atoi(decimals)
		for i := 0; i < len(p.rpcURL); i++ {
			balance, err := getStcBalance(p.rpcURL[i], address, tokenAddress, decimalsInt)
			if err != nil {
				p.log.Error("get stc balance error:", err)
				continue
			}
			return balance, nil
		}
	}
	return "0", nil
}

func getStcBalance(url, address, tokenAddress string, decimals int) (string, error) {
	method := "state.get_resource"
	d := map[string]bool{
		"decode": true,
	}
	params := []interface{}{address, "0x00000000000000000000000000000001::Account::Balance<" + tokenAddress + ">", d}
	balance := &types.STCBalance{}
	err := call(url, ID101, method, balance, params)
	if err != nil {
		return "", err
	}
	return utils.BigIntString(big.NewInt(balance.JSON.Token.Value), decimals), nil
}

func (p *platform) GetRpcURL() []string {
	return p.rpcURL
}

func call(url string, id int, method string, out interface{}, params []interface{}) error {
	var resp types.STCResponse
	err := utils.HttpsPost(url, id, method, JSONRPC, &resp, params)
	if err != nil {
		return err
	}
	if resp.Error != nil {
		return resp.Error
	}
	return json.Unmarshal(resp.Result, &out)
}