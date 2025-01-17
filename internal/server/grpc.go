package server

import (
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/robfig/cron/v3"
	v15 "gitlab.bixin.com/mili/node-proxy/api/chainlist/v1"
	v14 "gitlab.bixin.com/mili/node-proxy/api/commRPC/v1"
	v13 "gitlab.bixin.com/mili/node-proxy/api/nft/v1"
	v12 "gitlab.bixin.com/mili/node-proxy/api/platform/v1"

	v1 "gitlab.bixin.com/mili/node-proxy/api/tokenlist/v1"
	"gitlab.bixin.com/mili/node-proxy/internal/conf"
	"gitlab.bixin.com/mili/node-proxy/internal/service"
)

// NewGRPCServer new a gRPC server.
func NewGRPCServer(c *conf.Server, tokenList *service.TokenlistService, chainList *service.ChainListService, platform *service.PlatformService,
	nft *service.NFTService, commService *service.CommRPCService, jobManager *cron.Cron, logger log.Logger) *grpc.Server {
	var opts = []grpc.ServerOption{
		grpc.Middleware(
			recovery.Recovery(),
		),
	}
	if c.Grpc.Network != "" {
		opts = append(opts, grpc.Network(c.Grpc.Network))
	}
	if c.Grpc.Addr != "" {
		opts = append(opts, grpc.Address(c.Grpc.Addr))
	}
	if c.Grpc.Timeout != nil {
		opts = append(opts, grpc.Timeout(c.Grpc.Timeout.AsDuration()))
	}
	srv := grpc.NewServer(opts...)
	v1.RegisterTokenlistServer(srv, tokenList)
	v12.RegisterPlatformServer(srv, platform)
	v13.RegisterNftServer(srv, nft)
	v14.RegisterCommRPCServer(srv, commService)
	v15.RegisterChainListServer(srv, chainList)

	jobManager.Start()

	return srv
}
