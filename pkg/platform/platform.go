package platform

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/go-kratos/kratos/v2/log"
	v1 "gitlab.bixin.com/mili/node-proxy/api/platform/v1"
	v12 "gitlab.bixin.com/mili/node-proxy/api/tokenlist/v1"
	"gitlab.bixin.com/mili/node-proxy/internal/conf"
	"gitlab.bixin.com/mili/node-proxy/pkg/platform/aptos"
	"gitlab.bixin.com/mili/node-proxy/pkg/platform/bitcoin"
	"gitlab.bixin.com/mili/node-proxy/pkg/platform/ethereum"
	"gitlab.bixin.com/mili/node-proxy/pkg/platform/solana"
	"gitlab.bixin.com/mili/node-proxy/pkg/platform/starcoin"
	"gitlab.bixin.com/mili/node-proxy/pkg/platform/sui"
	"gitlab.bixin.com/mili/node-proxy/pkg/platform/tron"
	"gitlab.bixin.com/mili/node-proxy/pkg/platform/types"
	"gitlab.bixin.com/mili/node-proxy/pkg/platform/utils"
)

const (
	STC   = "STC"
	BTC   = "BTC"
	EVM   = "EVM"
	TVM   = "TVM"
	SUI   = "SUI"
	APTOS = "APTOS"
	SOL   = "SOL"
)

type TypeAndRpc struct {
	Type   string
	RpcURL []string
}

type config struct {
	log       *log.Helper
	logger    log.Logger
	chainInfo map[string]TypeAndRpc
}

var c config

func InitPlatform(conf []*conf.Platform, logger log.Logger) {
	log := log.NewHelper(log.With(logger, "module", "platform/InitPlatform"))
	tempMap := make(map[string]TypeAndRpc, len(conf))
	for _, chainInfo := range conf {
		tempMap[chainInfo.Chain] = TypeAndRpc{
			Type:   chainInfo.Type,
			RpcURL: chainInfo.RpcURL,
		}
	}
	c = config{
		log:       log,
		logger:    logger,
		chainInfo: tempMap,
	}
}

func GetBalance(ctx context.Context, chain, address, tokenAddress, decimals string) (string, error) {
	platform := newPlatform(chain)
	if platform == nil {
		c.log.Error("get platform is nil")
		return "", errors.New("platform is nil")
	}

	return platform.GetBalance(ctx, address, tokenAddress, decimals)
}

func BuildWasmRequest(ctx context.Context, chain, nodeRpc, functionName, params string) (*v1.BuildWasmRequestReply, error) {
	platform := newPlatform(chain)
	if platform == nil {
		c.log.Error("get platform is nil")
		return nil, errors.New("platform is nil")
	}
	return platform.BuildWasmRequest(ctx, nodeRpc, functionName, params)
}

func AnalysisWasmResponse(ctx context.Context, chain, functionName, params, response string) (string, error) {
	platform := newPlatform(chain)
	if platform == nil {
		c.log.Error("get platform is nil")
		return "", errors.New("platform is nil")
	}
	return platform.AnalysisWasmResponse(ctx, functionName, params, response)
}

func GetPlatformTokenInfo(chain, token string) (*v12.GetTokenInfoResp_Data, error) {
	platform := newPlatform(chain)
	if platform == nil {
		c.log.Error("get platform is nil")
		return nil, errors.New("platform is nil")
	}
	return platform.GetTokenType(token)
}

func newPlatform(chain string) types.Platform {
	if value, ok := c.chainInfo[chain]; ok {
		switch value.Type {
		case EVM:
			return ethereum.NewEVMPlatform(chain, value.RpcURL, c.logger)
		case STC:
			return starcoin.NewSTCPlatform(chain, value.RpcURL, c.logger)
		case BTC:
			return bitcoin.NewBTCPlatform(chain, value.RpcURL, c.logger)
		case TVM:
			return tron.NewTronPlatform(chain, value.RpcURL, c.logger)
		case SUI:
			return sui.NewSuiPlatform(chain, value.RpcURL, c.logger)
		case APTOS:
			return aptos.NewAptosPlatform(chain, value.RpcURL, c.logger)
		case SOL:
			return solana.NewSolanaPlatform(chain, value.RpcURL, c.logger)
		}
	}
	return nil
}

// GetGasEstimateTime returns the estimated time, in seconds, for a transaction to be confirmed on the blockchain.
func GetGasEstimateTime(chain string, gasInfo string) (string, error) {
	switch chain {
	case "ETH":
		return getETHGasEstimate(gasInfo)
	}
	return "", nil
}

func getETHGasEstimate(gasInfo string) (string, error) {
	var gasMap map[string]string
	if err := json.Unmarshal([]byte(gasInfo), &gasMap); err != nil {
		return "", err
	}
	url := "https://api.etherscan.io/api"
	gasPrice := gasMap["gas_price"]
	params := map[string]string{
		"module":   "gastracker",
		"action":   "gasestimate",
		"gasprice": gasPrice,
		"apikey":   "CT5GUMRVZMMB94IZ34SNWSI5MEBPBXPPIK",
	}
	out := &types.ETHGasEstimate{}
	err := utils.HttpsGetForm(url, params, out)
	if err != nil {
		return "", err
	}
	if out.Status != "1" {
		return "", errors.New(out.Result)
	}
	return out.Result, nil
}
