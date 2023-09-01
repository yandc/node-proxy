package starcoin

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-kratos/kratos/v2/log"
	v1 "gitlab.bixin.com/mili/node-proxy/api/platform/v1"
	v12 "gitlab.bixin.com/mili/node-proxy/api/tokenlist/v1"
	"gitlab.bixin.com/mili/node-proxy/pkg/platform/types"
	"gitlab.bixin.com/mili/node-proxy/pkg/platform/utils"
	"math/big"
	"strconv"
	"strings"
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
	chain  string
}

func NewSTCPlatform(chain string, rpcURL []string, logger log.Logger) types.Platform {
	log := log.NewHelper(log.With(logger, "module", "platform/bitcoin"))
	return &platform{rpcURL: rpcURL, log: log, chain: chain}
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

func (p *platform) BuildWasmRequest(ctx context.Context, nodeRpc, functionName, params string) (*v1.BuildWasmRequestReply, error) {
	return nil, nil
}
func (p *platform) AnalysisWasmResponse(ctx context.Context, functionName, params, response string) (string, error) {
	return "", nil
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

func listResource(url, address string) (*types.STCListResource, error) {
	method := "state.list_resource"
	d := map[string]bool{
		"decode": true,
	}
	params := []interface{}{address, d}
	resource := &types.STCListResource{}
	err := call(url, ID101, method, resource, params)
	return resource, err
}

func (p *platform) GetTokenType(token string) (*v12.GetTokenInfoResp_Data, error) {
	tokenInfo := strings.SplitN(token, "::", 3)
	var listResources *types.STCListResource
	var err error
	for i := 0; i < len(p.rpcURL); i++ {
		listResources, err = listResource(p.rpcURL[i], tokenInfo[0])
		if err != nil {
			p.log.Error("get stc resource error:", err)
			continue
		}
	}
	if err != nil {
		return nil, err
	}
	tokenKey := fmt.Sprintf("TokenInfo<%s>", token)
	decimals := 0
	for key, value := range listResources.Resources {
		if strings.Contains(key, tokenKey) {
			d := len(fmt.Sprintf("%d", value.Json.ScalingFactor))
			if d > 0 {
				decimals = d - 1
			}
		}
	}
	symbol := tokenInfo[2]
	if strings.Contains(tokenInfo[2], "LiquidityToken") {
		model := tokenInfo[2][15 : len(tokenInfo[2])-1]
		modelInfo := strings.Split(model, ",")
		if len(modelInfo) > 0 {
			name1, name2 := strings.SplitN(modelInfo[0], "::", 3)[2], strings.SplitN(modelInfo[1], "::", 3)[2]
			symbol = fmt.Sprintf("%v/%v", name1, name2)
		}
	}
	return &v12.GetTokenInfoResp_Data{
		Chain:    p.chain,
		Address:  token,
		Decimals: uint32(decimals),
		Symbol:   symbol,
		Name:     symbol,
	}, nil
}

func (p *platform) IsContractAddress(address string) (bool, error) {
	return false, nil
}

func (p *platform) GetERCType(token string) string {
	return ""
}
