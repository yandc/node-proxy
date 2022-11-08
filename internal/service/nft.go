package service

import (
	"context"
	v1 "gitlab.bixin.com/mili/node-proxy/api/nft/v1"
	"gitlab.bixin.com/mili/node-proxy/internal/biz"
	"time"
)

type NFTService struct {
	v1.UnimplementedNftServer
	uc *biz.NFTUsecase
}

func NewNFTService(uc *biz.NFTUsecase) *NFTService {
	return &NFTService{uc: uc}
}

func (s *NFTService) GetNftInfo(ctx context.Context, req *v1.GetNftInfoRequest) (*v1.GetNftReply, error) {
	// 设置接口 3s 超时
	subctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	return s.uc.GetNFTInfo(subctx, req.Chain, req.NftInfo)
}

func (s *NFTService) NetNftCollectionInfo(ctx context.Context, req *v1.GetNftCollectionInfoReq) (
	*v1.GetNftCollectionInfoReply, error) {
	// 设置接口 3s 超时
	subctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	return s.uc.NetNftCollectionInfo(subctx, req.Chain, req.Address)
}
