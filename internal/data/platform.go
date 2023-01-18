package data

import (
	"context"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-redis/redis"
	v1 "gitlab.bixin.com/mili/node-proxy/api/platform/v1"
	"gitlab.bixin.com/mili/node-proxy/internal/biz"
	"gitlab.bixin.com/mili/node-proxy/internal/conf"
	"gitlab.bixin.com/mili/node-proxy/pkg/nft/list"
	"gitlab.bixin.com/mili/node-proxy/pkg/platform"
	"gitlab.bixin.com/mili/node-proxy/pkg/platform/types"
)

type platformRepo struct {
	log *log.Helper
}

func NewPlatformRepo(conf []*conf.Platform, logger log.Logger, client *redis.Client) biz.PlatformRepo {
	platform.InitPlatform(conf, logger, client)
	return &platformRepo{
		log: log.NewHelper(logger),
	}

}

func (r *platformRepo) GetBalance(ctx context.Context, chain, address, tokenAddress, decimals string) (string, error) {
	r.log.WithContext(ctx).Infof("GetBalance", chain, address, tokenAddress, decimals)
	return platform.GetBalance(ctx, chain, address, tokenAddress, decimals)
}

func (r *platformRepo) BuildWasmRequest(ctx context.Context, chain, nodeRpc, functionName, params string) (
	*v1.BuildWasmRequestReply, error) {
	r.log.WithContext(ctx).Infof("BuildWasmRequest", chain, nodeRpc, functionName, params)
	if functionName == types.NFTINFO {
		return list.BuildNFTInfoFunc(chain, params)
	}

	result, err := platform.BuildWasmRequest(ctx, chain, nodeRpc, functionName, params)
	log.Infof("BuildWasmRequest result ", result)
	return result, err
}

func (r *platformRepo) AnalysisWasmResponse(ctx context.Context, chain, functionName, params,
	response string) (string, error) {
	log := r.log.WithContext(ctx)
	log.Infof("AnalysisWasmResponse", chain, functionName, params, response)
	var result string
	var err error
	if functionName == types.NFTINFO {
		result, err = list.AnalysisNFTResponse(chain, params, response)
	} else {
		result, err = platform.AnalysisWasmResponse(ctx, chain, functionName, params, response)
	}
	log.Infof("AnalysisWasmResponse result ", result)
	if err != nil {
		log.Error("AnalysisWasmResponse Error:", err)
	}
	return result, err
}

func (r *platformRepo) GetGasEstimate(ctx context.Context, chain, gasInfo string) (string, error) {
	log := r.log.WithContext(ctx)
	log.Infof("GetGasEstimate", chain, gasInfo)
	result, err := platform.GetGasEstimateTime(chain, gasInfo)
	log.Infof("GetGasEstimate result ", result)
	if err != nil {
		log.Error("GetGasEstimate Error:", err)
	}
	return result, err
}
