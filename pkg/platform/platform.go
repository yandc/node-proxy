package platform

import (
	"context"
	"errors"
	"github.com/go-kratos/kratos/v2/log"
	"gitlab.bixin.com/mili/node-proxy/internal/conf"
	"gitlab.bixin.com/mili/node-proxy/pkg/platform/bitcoin"
	"gitlab.bixin.com/mili/node-proxy/pkg/platform/ethereum"
	"gitlab.bixin.com/mili/node-proxy/pkg/platform/starcoin"
	"gitlab.bixin.com/mili/node-proxy/pkg/platform/tron"
	"gitlab.bixin.com/mili/node-proxy/pkg/platform/types"
)

const (
	STC = "STC"
	BTC = "BTC"
	EVM = "EVM"
	TVM = "TVM"
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

func newPlatform(chain string) types.Platform {
	if value, ok := c.chainInfo[chain]; ok {
		switch value.Type {
		case EVM:
			return ethereum.NewEVMPlatform(value.RpcURL, c.logger)
		case STC:
			return starcoin.NewSTCPlatform(value.RpcURL, c.logger)
		case BTC:
			return bitcoin.NewBTCPlatform(value.RpcURL, c.logger)
		case TVM:
			return tron.NewTronPlatform(value.RpcURL, c.logger)
		}
	}
	return nil
}
