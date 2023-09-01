package tron

import (
	"context"
	"encoding/hex"
	"github.com/btcsuite/btcutil/base58"
	"github.com/go-kratos/kratos/v2/log"
	v1 "gitlab.bixin.com/mili/node-proxy/api/platform/v1"
	v12 "gitlab.bixin.com/mili/node-proxy/api/tokenlist/v1"
	"gitlab.bixin.com/mili/node-proxy/pkg/platform/types"
	"gitlab.bixin.com/mili/node-proxy/pkg/platform/utils"
	"strconv"
	"strings"
)

const (
	TronscanURL  = "tronscan.org"
	TrongridURL  = "trongrid.io"
	TronstackURL = "tronstack.io"
	TRC20        = "trc20"
)

type platform struct {
	rpcURL []string
	log    *log.Helper
	chain  string
}

func NewTronPlatform(chain string, rpcURL []string, logger log.Logger) types.Platform {
	log := log.NewHelper(log.With(logger, "module", "platform/tron"))
	return &platform{rpcURL: rpcURL, log: log, chain: chain}
}

func (p *platform) GetBalance(ctx context.Context, address, tokenAddress, decimals string) (string, error) {
	if tokenAddress == "" || address == tokenAddress {
		for i := 0; i < len(p.rpcURL); i++ {
			client := getClient(p.rpcURL[i])
			balance, err := client.GetBalance(address)
			if err != nil {
				p.log.Error("get balance error:", err)
				continue
			}
			return balance, nil
		}
	}

	if address != tokenAddress && tokenAddress != "" && decimals != "" {
		decimalsInt, _ := strconv.Atoi(decimals)
		for i := 0; i < len(p.rpcURL); i++ {
			client := getClient(p.rpcURL[i])
			balance, err := client.GetTokenBalance(address, tokenAddress, decimalsInt)
			if err != nil {
				p.log.Error("get token balance error:", err)
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

func (p *platform) GetTokenType(token string) (*v12.GetTokenInfoResp_Data, error) {
	var url string
	if strings.Contains(p.chain, "TEST") {
		url = "https://shastapi.tronscan.org/api/contract"
	} else {
		url = "https://apilist.tronscan.org/api/contract"
	}
	params := map[string]string{
		"contract": token,
	}
	out := &types.TronTokenInfo{}
	err := utils.HttpsGetForm(url, params, out)
	if err != nil {
		return nil, err
	}
	if len(out.Data) == 0 {
		return nil, nil
	}
	return &v12.GetTokenInfoResp_Data{
		Chain:    p.chain,
		Address:  token,
		Decimals: uint32(out.Data[0].TokenInfo.TokenDecimal),
		Symbol:   strings.ToUpper(out.Data[0].TokenInfo.TokenName),
	}, nil
}

func (p *platform) GetERCType(token string) string {
	var url string
	if strings.Contains(p.chain, "TEST") {
		url = "https://shastapi.tronscan.org/api/contract"
	} else {
		url = "https://apilist.tronscan.org/api/contract"
	}
	params := map[string]string{
		"contract": token,
	}
	out := &types.TronTokenInfo{}
	err := utils.HttpsGetForm(url, params, out)
	if err != nil {
		return ""
	}
	if len(out.Data) == 0 {
		return ""
	}
	return out.Data[0].TokenInfo.TokenType
}

func (p *platform) GetRpcURL() []string {
	return p.rpcURL
}

func getClient(url string) types.TronClient {
	if strings.Contains(url, TronscanURL) {
		return TronScanClient{url: url}
	} else if strings.Contains(url, TrongridURL) || strings.Contains(url, TronstackURL) {
		return TronGridClient{url: url}
	}
	return nil
}

func Base58ToHex(address string) string {
	decodeCheck := base58.Decode(address)
	if len(decodeCheck) < 4 {
		return "0"
	}
	decodeData := decodeCheck[:len(decodeCheck)-4]
	result := hex.EncodeToString(decodeData)
	return result
}

func (p *platform) IsContractAddress(address string) (bool, error) {
	for i := 0; i < len(p.rpcURL); i++ {
		client := getClient(p.rpcURL[i])
		isContract, err := client.IsContractAddress(address)
		if err != nil {
			p.log.Error("IsContractAddress error:", err)
			continue
		}
		return isContract, nil
	}
	return false, nil
}
