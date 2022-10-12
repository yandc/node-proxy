package biz

import (
	"context"
	"github.com/go-kratos/kratos/v2/log"
	v1 "gitlab.bixin.com/mili/node-proxy/api/platform/v1"
)

type PlatformRepo interface {
	GetBalance(ctx context.Context, chain, address, tokenAddress, decimals string) (string, error)
	BuildWasmRequest(ctx context.Context, chain, nodeRpc, functionName, params string) (*v1.BuildWasmRequestReply, error)
	AnalysisWasmResponse(ctx context.Context, chain, functionName, params, response string) (string, error)
	GetGasEstimate(ctx context.Context, chain, gasInfo string) (string, error)
}

type PlatformUseCase struct {
	repo PlatformRepo
	log  *log.Helper
}

func NewPlatformUseCase(repo PlatformRepo, logger log.Logger) *PlatformUseCase {
	return &PlatformUseCase{repo: repo, log: log.NewHelper(logger)}
}

func (uc *PlatformUseCase) GetBalance(ctx context.Context, chain, address, tokenAddress, decimals string) (string, error) {
	return uc.repo.GetBalance(ctx, chain, address, tokenAddress, decimals)
}

func (uc *PlatformUseCase) BuildWasmRequest(ctx context.Context, chain, nodeRpc, functionName,
	params string) (*v1.BuildWasmRequestReply, error) {
	return uc.repo.BuildWasmRequest(ctx, chain, nodeRpc, functionName, params)
}

func (uc *PlatformUseCase) AnalysisWasmResponse(ctx context.Context, chain, functionName, params,
	response string) (string, error) {
	return uc.repo.AnalysisWasmResponse(ctx, chain, functionName, params, response)
}

func (uc *PlatformUseCase) GetGasEstimate(ctx context.Context, chain, gasInfo string) (string, error) {
	return uc.repo.GetGasEstimate(ctx, chain, gasInfo)
}
