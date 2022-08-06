package biz

import (
	"context"
	"github.com/go-kratos/kratos/v2/log"
)

type PlatformRepo interface {
	GetBalance(ctx context.Context, chain, address, tokenAddress, decimals string) (string, error)
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
