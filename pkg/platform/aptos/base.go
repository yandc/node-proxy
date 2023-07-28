package aptos

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-kratos/kratos/v2/log"
	v1 "gitlab.bixin.com/mili/node-proxy/api/platform/v1"
	v12 "gitlab.bixin.com/mili/node-proxy/api/tokenlist/v1"
	"gitlab.bixin.com/mili/node-proxy/pkg/platform/types"
	"gitlab.bixin.com/mili/node-proxy/pkg/platform/utils"
	"strconv"
	"strings"
)

const TYPE_PREFIX = "0x1::coin::CoinStore"
const TOKEN_INFO_PREFIX = "0x1::coin::CoinInfo"
const APTOS_DECIMALS = 8

type platform struct {
	rpcURL []string
	log    *log.Helper
	chain  string
}

func NewAptosPlatform(chain string, rpcURL []string, logger log.Logger) types.Platform {
	log := log.NewHelper(log.With(logger, "module", "platform/aptos"))
	return &platform{rpcURL: rpcURL, log: log, chain: chain}
}

func (p *platform) GetRpcURL() []string {
	return p.rpcURL
}

func (p *platform) GetBalance(ctx context.Context, address, tokenAddress, decimals string) (string, error) {
	resourceType := fmt.Sprintf("%s<%s>", TYPE_PREFIX, "0x1::aptos_coin::AptosCoin")
	decimalsInt := APTOS_DECIMALS
	if address != tokenAddress && tokenAddress != "" && decimals != "" {
		resourceType = fmt.Sprintf("%s<%s>", TYPE_PREFIX, tokenAddress)
		decimalsInt, _ = strconv.Atoi(decimals)
	}
	for i := 0; i < len(p.rpcURL); i++ {
		balance, err := getBalance(p.rpcURL[i], address, resourceType, decimalsInt)
		if err != nil {
			p.log.Error("get token balance error:", err)
			continue
		}
		return balance, nil
	}
	return "", nil
}

func (p *platform) BuildWasmRequest(ctx context.Context, nodeRpc, functionName, params string) (*v1.BuildWasmRequestReply, error) {
	return nil, nil
}

func (p *platform) AnalysisWasmResponse(ctx context.Context, functionName, params, response string) (string, error) {
	return "", nil
}

func (p *platform) GetTokenType(token string) (*v12.GetTokenInfoResp_Data, error) {
	sourceToken := token
	if strings.Contains(token, "::") {
		sourceToken = strings.Split(token, "::")[0]
	}
	tokenResource := fmt.Sprintf("%s<%s>", TOKEN_INFO_PREFIX, token)
	for i := 0; i < len(p.rpcURL); i++ {
		resources := getResourceByAddress(p.rpcURL[i], sourceToken, tokenResource)
		if resources != nil {
			var tokenInfo types.AptosTokenInfo
			b, _ := json.Marshal(resources.Data)
			json.Unmarshal(b, &tokenInfo)
			return &v12.GetTokenInfoResp_Data{
				Chain:    p.chain,
				Address:  token,
				Decimals: uint32(tokenInfo.Decimals),
				Symbol:   strings.ToUpper(tokenInfo.Symbol),
				Name:     tokenInfo.Name,
			}, nil
		}
	}
	return nil, nil
}

func getBalance(noderpc, address string, resourceType string, decimals int) (string, error) {
	//url := c.url
	url := fmt.Sprintf("%s/accounts/%s/resource/%s", noderpc, address, resourceType)
	out := &types.AptosBalanceResp{}
	err := utils.HttpsGetForm(url, nil, out)
	if err != nil {
		return "", err
	}
	if out.Message != "" {
		return "", errors.New(out.Message)
	}
	balance := utils.UpdateDecimals(out.Data.Coin.Value, decimals)
	return balance, nil
}

func getResourceByAddress(noderpc, address, resourceType string) *types.AptosResourceResp {
	url := fmt.Sprintf("%s/accounts/%s/resource/%s", noderpc, address, resourceType)
	out := &types.AptosResourceResp{}
	err := utils.HttpsGetForm(url, nil, out)
	if err != nil {
		return nil
	}
	return out
}

func (p *platform) IsContractAddress(address string) (bool, error) {
	return false, nil
}
