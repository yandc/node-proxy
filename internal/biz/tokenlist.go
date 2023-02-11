package biz

import (
	"context"
	"time"

	v1 "gitlab.bixin.com/mili/node-proxy/api/tokenlist/v1"

	"github.com/go-kratos/kratos/v2/log"
)

type TokenListRepo interface {
	GetPrice(ctx context.Context, coinName, coinAddress, currency string) ([]byte, error)
	//CreateTokenList(ctx context.Context)
	GetTokenList(ctx context.Context, chain string) ([]*v1.GetTokenListResp_Data, error)
	AutoUpdateTokenList(ctx context.Context)
	GetTokenInfo(ctx context.Context, addressInfo []*v1.GetTokenInfoReq_Data) ([]*v1.GetTokenInfoResp_Data, error)
	GetDBTokenInfo(ctx context.Context, addressInfo []*v1.GetTokenInfoReq_Data) ([]*v1.GetTokenInfoResp_Data, error)
	GetTokenTop20(ctx context.Context, chain string) ([]*v1.TokenInfoData, error)
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

//func (uc *TokenListUsecase) CreateTokenList(ctx context.Context) {
//	uc.repo.CreateTokenList(ctx)
//}

func (uc *TokenListUsecase) GetTokenList(ctx context.Context, chain string) ([]*v1.GetTokenListResp_Data, error) {
	return uc.repo.GetTokenList(ctx, chain)
}

func (uc *TokenListUsecase) AutoUpdateTokenList(ctx context.Context) {
	transactionPlan := time.NewTicker(24 * time.Hour)
	for true {
		select {
		case <-transactionPlan.C:
			uc.repo.AutoUpdateTokenList(ctx)
		}
	}
}

func (uc *TokenListUsecase) GetTokenInfo(ctx context.Context, addressInfo []*v1.GetTokenInfoReq_Data) (
	[]*v1.GetTokenInfoResp_Data, error) {
	return uc.repo.GetTokenInfo(ctx, addressInfo)
}

func (uc *TokenListUsecase) GetDBTokenInfo(ctx context.Context, addressInfo []*v1.GetTokenInfoReq_Data) (
	[]*v1.GetTokenInfoResp_Data, error) {
	return uc.repo.GetDBTokenInfo(ctx, addressInfo)
}

func (uc *TokenListUsecase) GetTokenTop20(ctx context.Context, chain string) ([]*v1.TokenInfoData, error) {
	return uc.repo.GetTokenTop20(ctx, chain)
}
