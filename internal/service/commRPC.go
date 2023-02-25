package service

import (
	"context"
	v1 "gitlab.bixin.com/mili/node-proxy/api/commRPC/v1"
	"gitlab.bixin.com/mili/node-proxy/internal/biz"
	"time"
)

type CommRPCService struct {
	v1.UnimplementedCommRPCServer
	uc *biz.CommRPCUsecase
}

func NewCommRPCService(uc *biz.CommRPCUsecase) *CommRPCService {
	return &CommRPCService{uc: uc}
}

type Test struct {
	Age  int
	Name string
}

func (s *CommRPCService) ExecNodeProxyRPC(ctx context.Context, req *v1.ExecNodeProxyRPCRequest) (*v1.ExecNodeProxyRPCReply, error) {
	// 设置接口 3s 超时
	subctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	return s.uc.ExecNodeProxyRPC(subctx, req)
}
