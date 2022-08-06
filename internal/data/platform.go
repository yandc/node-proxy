package data

import (
	"context"
	"github.com/go-kratos/kratos/v2/log"
	"gitlab.bixin.com/mili/node-proxy/internal/biz"
	"gitlab.bixin.com/mili/node-proxy/internal/conf"
	"gitlab.bixin.com/mili/node-proxy/pkg/platform"
)

type platformRepo struct {
	log *log.Helper
}

func NewPlatformRepo(conf []*conf.Platform, logger log.Logger) biz.PlatformRepo {
	platform.InitPlatform(conf, logger)
	return &platformRepo{
		log: log.NewHelper(logger),
	}

}

func (r *platformRepo) GetBalance(ctx context.Context, chain, address, tokenAddress, decimals string) (string, error) {
	r.log.WithContext(ctx).Infof("GetBalance", chain, address, tokenAddress, decimals)
	return platform.GetBalance(ctx, chain, address, tokenAddress, decimals)
}
