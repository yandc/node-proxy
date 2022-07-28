package service

import (
	"context"
	"time"

	pb "gitlab.bixin.com/mili/node-proxy/api/tokenlist/v1"
	"gitlab.bixin.com/mili/node-proxy/internal/biz"
)

type TokenlistService struct {
	pb.UnimplementedTokenlistServer
	uc *biz.TokenListUsecase
}

func NewTokenlistService(uc *biz.TokenListUsecase) *TokenlistService {
	return &TokenlistService{uc: uc}
}

func (s *TokenlistService) GetPrice(ctx context.Context, req *pb.PriceReq) (*pb.PriceResp, error) {
	// 设置接口 3s 超时
	subctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	price, err := s.uc.GetPrice(subctx, req.CoinNames, req.CoinAddresses, req.Currency)
	return &pb.PriceResp{
		Data: price,
	}, err
}

func (s *TokenlistService) GetTokenList(ctx context.Context, req *pb.GetTokenListReq) (*pb.GetTokenListResp, error) {
	// 设置接口 3s 超时
	subctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	data, err := s.uc.GetTokenList(subctx, req.Chain)
	return &pb.GetTokenListResp{
		Data: data,
	}, err
}
