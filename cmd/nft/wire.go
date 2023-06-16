//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package main

import (
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"gitlab.bixin.com/mili/node-proxy/internal/biz"
	"gitlab.bixin.com/mili/node-proxy/internal/conf"
	"gitlab.bixin.com/mili/node-proxy/internal/data"
)

// initNFTcase init kratos application tokenlistcase.
func initNFTcase(confData *conf.Data, nftList *conf.NFTList, logger log.Logger) (*biz.NFTUsecase, func(), error) {
	panic(wire.Build(data.ProviderSet, biz.ProviderSet))
}
