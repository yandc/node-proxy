package biz

import (
	"context"
	v1 "node-proxy/api/tokenlist/v1"

	"github.com/go-kratos/kratos/v2/log"
)

type TokenListRepo interface {
	GetPrice(ctx context.Context, coinName, coinAddress, currency string) ([]byte, error)
	CreateTokenList(ctx context.Context)
	GetTokenList(ctx context.Context, chain string) ([]*v1.GetTokenListResp_Data, error)
}

type TokenListUsecase struct {
	repo TokenListRepo
	log  *log.Helper
}

// NewTokenListUsecase new a TokenList usecase.
func NewTokenListUsecase(repo TokenListRepo, logger log.Logger) *TokenListUsecase {
	return &TokenListUsecase{repo: repo, log: log.NewHelper(logger)}
}

func (uc *TokenListUsecase) GetPrice(ctx context.Context, coinName, coinAddress, currency string) ([]byte, error) {
	return uc.repo.GetPrice(ctx, coinName, coinAddress, currency)
}

func (uc *TokenListUsecase) CreateTokenList(ctx context.Context) {
	uc.repo.CreateTokenList(ctx)
}

func (uc *TokenListUsecase) GetTokenList(ctx context.Context, chain string) ([]*v1.GetTokenListResp_Data, error) {
	return uc.repo.GetTokenList(ctx, chain)
}
