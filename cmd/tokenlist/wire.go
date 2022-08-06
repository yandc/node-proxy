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

// initTokenListcase init kratos application tokenlistcase.
func initTokenListcase(confData *conf.Data, tokenList *conf.TokenList, logger log.Logger) (*biz.TokenListUsecase, func(), error) {
	panic(wire.Build(data.ProviderSet, biz.ProviderSet))
}
