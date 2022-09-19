package tron

import (
	"context"
	"encoding/hex"
	"github.com/btcsuite/btcutil/base58"
	"github.com/go-kratos/kratos/v2/log"
	v1 "gitlab.bixin.com/mili/node-proxy/api/platform/v1"
	"gitlab.bixin.com/mili/node-proxy/pkg/platform/types"
	"strconv"
	"strings"
)

const (
	TronscanURL  = "tronscan.org"
	TrongridURL  = "trongrid.io"
	TronstackURL = "tronstack.io"
)

type platform struct {
	rpcURL []string
	log    *log.Helper
}

func NewTronPlatform(rpcURL []string, logger log.Logger) types.Platform {
	log := log.NewHelper(log.With(logger, "module", "platform/tron"))
	return &platform{rpcURL: rpcURL, log: log}
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
