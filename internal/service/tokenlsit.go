package service

import (
	"context"
	pb "node-proxy/api/tokenlist/v1"
	"node-proxy/internal/biz"
)

type TokenlsitService struct {
	pb.UnimplementedTokenlsitServer
	uc *biz.TokenListUsecase
}

func NewTokenlsitService(uc *biz.TokenListUsecase) *TokenlsitService {
	return &TokenlsitService{uc: uc}
}

func (s *TokenlsitService) GetPrice(ctx context.Context, req *pb.PriceReq) (*pb.PriceResp, error) {
	price, err := s.uc.GetPrice(ctx, req.CoinNames, req.CoinAddresses, req.Currency)
	return &pb.PriceResp{Data: price}, err
}

func (s *TokenlsitService) CreateTokenList(ctx context.Context, req *pb.CreateTokenListReq) (*pb.CreateTokenListResp, error) {
	s.uc.CreateTokenList(ctx)
	return &pb.CreateTokenListResp{Message: "success"}, nil
}

func (s *TokenlsitService) GetTokenList(ctx context.Context, req *pb.GetTokenListReq) (*pb.GetTokenListResp, error) {
	data, err := s.uc.GetTokenList(ctx, req.Chain)
	return &pb.GetTokenListResp{Data: data}, err
}
