package service

import (
	"context"
	"gitlab.bixin.com/mili/node-proxy/internal/biz"
	"time"

	pb "gitlab.bixin.com/mili/node-proxy/api/platform/v1"
)

type PlatformService struct {
	pb.UnimplementedPlatformServer
	pc *biz.PlatformUseCase
}

func NewPlatformService(pc *biz.PlatformUseCase) *PlatformService {
	return &PlatformService{pc: pc}
}

func (s *PlatformService) GetBalance(ctx context.Context, req *pb.GetBalanceRequest) (*pb.GetBalanceReply, error) {
	// 设置接口 3s 超时
	subctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	balance, err := s.pc.GetBalance(subctx, req.Chain, req.Address,req.TokenAddress,req.Decimals)
	return &pb.GetBalanceReply{Balance: balance}, err
}
