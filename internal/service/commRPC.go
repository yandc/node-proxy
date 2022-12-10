package service

import (
	"context"
	v1 "gitlab.bixin.com/mili/node-proxy/api/commRPC/v1"
)

type CommRPCService struct {
	v1.UnimplementedCommRPCServer
}

func NewCommRPCService() *CommRPCService {
	return &CommRPCService{}
}

type Test struct {
	Age  int
	Name string
}

func (s *CommRPCService) ExecNodeProxyRPC(context.Context, *v1.ExecNodeProxyRPCRequest) (*v1.ExecNodeProxyRPCReply, error) {
	return &v1.ExecNodeProxyRPCReply{
		Result: "",
		Ok:     true,
	}, nil
}
