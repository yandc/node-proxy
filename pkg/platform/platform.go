package platform

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-redis/redis"
	v1 "gitlab.bixin.com/mili/node-proxy/api/platform/v1"
	v12 "gitlab.bixin.com/mili/node-proxy/api/tokenlist/v1"
	"gitlab.bixin.com/mili/node-proxy/internal/conf"
	"gitlab.bixin.com/mili/node-proxy/pkg/chainlist"
	"gitlab.bixin.com/mili/node-proxy/pkg/platform/aptos"
	"gitlab.bixin.com/mili/node-proxy/pkg/platform/bitcoin"
	"gitlab.bixin.com/mili/node-proxy/pkg/platform/casper"
	"gitlab.bixin.com/mili/node-proxy/pkg/platform/ethereum"
	"gitlab.bixin.com/mili/node-proxy/pkg/platform/solana"
	"gitlab.bixin.com/mili/node-proxy/pkg/platform/starcoin"
	"gitlab.bixin.com/mili/node-proxy/pkg/platform/sui"
	"gitlab.bixin.com/mili/node-proxy/pkg/platform/tron"
	"gitlab.bixin.com/mili/node-proxy/pkg/platform/types"
	"gitlab.bixin.com/mili/node-proxy/pkg/platform/utils"
	"math"
	"math/big"
	"strings"
)

const (
	STC              = "STC"
	BTC              = "BTC"
	EVM              = "EVM"
	TVM              = "TVM"
	SUI              = "SUI"
	APTOS            = "APTOS"
	SOL              = "SOL"
	CSPR             = "CSPR"
	REDIS_ESTIME_KEY = "platform:estime:"
)

type TypeAndRpc struct {
	Type   string
	RpcURL []string
}

type config struct {
	log         *log.Helper
	logger      log.Logger
	chainInfo   map[string]TypeAndRpc
	redisClient *redis.Client
}

var c config

func InitPlatform(conf []*conf.Platform, logger log.Logger, client *redis.Client) {
	log := log.NewHelper(log.With(logger, "module", "platform/InitPlatform"))
	tempMap := make(map[string]TypeAndRpc, len(conf))
	for _, chainInfo := range conf {
		tempMap[chainInfo.Chain] = TypeAndRpc{
			Type:   chainInfo.Type,
			RpcURL: chainInfo.RpcURL,
		}
	}
	c = config{
		log:         log,
		logger:      logger,
		chainInfo:   tempMap,
		redisClient: client,
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
	result, err := platform.BuildWasmRequest(ctx, nodeRpc, functionName, params)
	if err != nil {
		c.log.Error("BuildWasmRequest Error:", err)
	}
	return result, err
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
		case CSPR:
			return casper.NewCasperPlatform(chain, value.RpcURL, c.logger)
		}
	} else if strings.HasPrefix(strings.ToLower(chain), "evm") { //支持自定义EVM
		url := getRpcUrl(chain)
		return ethereum.NewEVMPlatform(chain, url, c.logger)
	}

	return nil
}

func getRpcUrl(chain string) []string {
	_, chainId, found := strings.Cut(strings.ToLower(chain), "evm")
	if !found {
		c.log.Error("unsupported evm chain")
		return nil
	}
	nodeUrls, err := chainlist.FindChainNodeUrlList(chainId)
	if err != nil {
		c.log.Error("get chain node url list error", "err", err)
		return nil
	}

	rpcUrls := make([]string, len(nodeUrls))
	for i, nodeUrl := range nodeUrls {
		rpcUrls[i] = nodeUrl.Url
	}

	return rpcUrls
}

func GetSUINFTInfo(chain, objectId string) (*types.SuiNFTObjectResponse, error) {
	platform := newPlatform(chain)
	if platform == nil {
		c.log.Error("get platform is nil")
		return nil, errors.New("platform is nil")
	}
	suiPlatform := sui.Platform2SUIPlatform(platform)
	if platform == nil {
		c.log.Error("sui platform is nil")
		return nil, errors.New("sui platform is nil")
	}
	return suiPlatform.GetNFTObject(objectId)
}

// GetGasEstimateTime returns the estimated time, in seconds, for a transaction to be confirmed on the blockchain.
func GetGasEstimateTime(chain string, gasInfo string) (string, error) {
	switch chain {
	case "ETH":
		return getETHGasEstimate(gasInfo)
	case "BSC", "Fantom", "Polygon", "Avalanche":
		return getEVMGasEstimate(chain, gasInfo)
	}
	return "", nil
}

func getEVMGasEstimate(chain string, gasInfo string) (string, error) {
	var url string
	var blockInterval int
	switch chain {
	case "BSC":
		url = "https://gbsc.blockscan.com/gasapi.ashx?apikey=key&method=pendingpooltxgweidata"
		blockInterval = 3
	case "Fantom":
		url = "https://gftm.blockscan.com/gasapi.ashx?apikey=key&method=pendingpooltxgweidata"
		blockInterval = 2
	case "Polygon":
		url = "https://gpoly.blockscan.com/gasapi.ashx?apikey=key&method=pendingpooltxgweidata"
		blockInterval = 2
	case "Avalanche":
		url = "https://gavax.blockscan.com/gasapi.ashx?apikey=key&method=pendingpooltxgweidata"
		blockInterval = 2
	default:
		return "", nil
	}
	var gasMap map[string]string
	if err := json.Unmarshal([]byte(gasInfo), &gasMap); err != nil {
		return "", err
	}
	gasPrice := gasMap["gas_price"]
	tempGasPrice, flag := new(big.Float).SetString(gasPrice)
	if !flag {
		return "", errors.New("float set string error:" + gasPrice)
	}
	gasPriceGWei, _ := new(big.Float).Quo(tempGasPrice, big.NewFloat(1000000000)).Float64()
	out := &types.EVMGasEstimate{}
	redisKey := REDIS_ESTIME_KEY + chain
	esTimeData, updateFlag, _ := utils.GetESTimeRedisValueByKey(c.redisClient, redisKey)
	if esTimeData != "" {
		if err := json.Unmarshal([]byte(esTimeData), &out.Result); err != nil {
			return "", err
		}
	}
	if esTimeData == "" || updateFlag {
		err := utils.HttpsGetForm(url, nil, out)
		if err != nil {
			return "", err
		}
		if out.Status != "1" {
			return "", errors.New(out.Message)
		}
		resultByte, _ := json.Marshal(out.Result)
		err = utils.SetESTimeRedisKey(c.redisClient, redisKey, string(resultByte))
		if err != nil {
			c.log.Error("set estime redis error:", err)
		}
	}

	var data [][]float64
	if err := json.Unmarshal([]byte(out.Result.Data), &data); err != nil {
		return "", err
	}
	txSum := 0
	for i := 0; i < len(data); i++ {
		if data[i][0] > gasPriceGWei {
			txSum += int(data[i][1])
		} else {
			break
		}
	}
	block := int(math.Ceil(float64(txSum) / float64(out.Result.Avgtxnsperblock)))
	t := block * blockInterval
	if t == 0 {
		t = blockInterval
	}
	return fmt.Sprintf("%v", t), nil
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
