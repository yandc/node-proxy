package biz

import (
	"context"
	"github.com/go-kratos/kratos/v2/log"
	v12 "gitlab.bixin.com/mili/node-proxy/api/nft-marketplace/v1"
	v2 "gitlab.bixin.com/mili/node-proxy/api/nft-marketplace/v2"
	v1 "gitlab.bixin.com/mili/node-proxy/api/nft/v1"
	"time"
)

type NFTRepo interface {
	GetNFTInfo(ctx context.Context, chain string, tokenInfos []*v1.GetNftInfoRequest_NftInfo) (*v1.GetNftReply, error)
	GetNftCollectionInfo(ctx context.Context, chain, address string) (*v1.GetNftCollectionInfoReply, error)
	AutoUpdateNFTInfo(ctx context.Context)
}

type NFTUsecase struct {
	repo                NFTRepo
	nftApiClient        v12.NFTApiClient
	collectionApiClient v2.CollectionApiClient
	log                 *log.Helper
}

// NewNFTUsecase new a TokenList usecase.
func NewNFTUsecase(repo NFTRepo, nftApiClient v12.NFTApiClient, collectionApiClient v2.CollectionApiClient, logger log.Logger) *NFTUsecase {
	return &NFTUsecase{repo: repo, nftApiClient: nftApiClient, collectionApiClient: collectionApiClient, log: log.NewHelper(logger)}
}

func (uc *NFTUsecase) GetNFTInfo(ctx context.Context, chain string, tokenInfos []*v1.GetNftInfoRequest_NftInfo) (*v1.GetNftReply, error) {
	return uc.repo.GetNFTInfo(ctx, chain, tokenInfos)
}

func (uc *NFTUsecase) GetNftCollectionInfo(ctx context.Context, chain, address string) (*v1.GetNftCollectionInfoReply, error) {
	return uc.repo.GetNftCollectionInfo(ctx, chain, address)
}

func (uc *NFTUsecase) AutoUpdateNFTInfo(ctx context.Context) {
	//uc.repo.AutoUpdateNFTInfo(ctx)
	transactionPlan := time.NewTicker(24 * time.Hour)
	for true {
		select {
		case <-transactionPlan.C:
			//uc.repo.AutoUpdateNFTInfo(ctx)
		}
	}
}
