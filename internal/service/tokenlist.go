package service

import (
	"context"
	"fmt"
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

func (s *TokenlistService) GetTokenInfo(ctx context.Context, req *pb.GetTokenInfoReq) (*pb.GetTokenInfoResp, error) {
	// 设置接口 3s 超时
	subctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	fmt.Println(subctx, len(req.Data))
	data, err := s.uc.GetTokenInfo(ctx, req.Data)
	return &pb.GetTokenInfoResp{
		Data: data,
	}, err
}

func (s *TokenlistService) GetDBTokenInfo(ctx context.Context, req *pb.GetTokenInfoReq) (*pb.GetTokenInfoResp, error) {
	// 设置接口 3s 超时
	subctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	fmt.Println(subctx, len(req.Data))
	data, err := s.uc.GetDBTokenInfo(ctx, req.Data)
	return &pb.GetTokenInfoResp{
		Data: data,
	}, err
}

func (s *TokenlistService) GetTokenTop20(ctx context.Context, req *pb.GetTokenTop20Req) (*pb.GetTokenTop20Resp, error) {
	// 设置接口 3s 超时
	subctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	data, err := s.uc.GetTokenTop20(subctx, req.Chain)
	return &pb.GetTokenTop20Resp{
		Data: data,
	}, err
}

func (s *TokenlistService) IsFake(ctx context.Context, req *pb.IsFakeReq) (*pb.IsFakeResp, error) {
	// 设置接口 3s 超时
	subctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	data := s.uc.IsFakeResp(subctx, req.Chain, req.Symbol, req.Address)
	return &pb.IsFakeResp{
		Data: data,
	}, nil
}
