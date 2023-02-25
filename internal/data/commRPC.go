package data

import (
	"context"
	"fmt"
	"github.com/go-kratos/kratos/v2/log"
	"gitlab.bixin.com/mili/node-proxy/internal/biz"
	"gitlab.bixin.com/mili/node-proxy/pkg/token-list/tokenlist"
)

type commRPCRepo struct {
	log *log.Helper
}

func NewCommRPCRepo(logger log.Logger) biz.CommRPCRepo {
	return &commRPCRepo{
		log: log.NewHelper(logger),
	}
}

func (c *commRPCRepo) GetPriceV2(ctx context.Context, coinName, coinAddress []string, currency string) (map[string]map[string]string, error) {
	c.log.WithContext(ctx).Infof("GetPriceV2", coinName, coinAddress, currency)
	price := tokenlist.GetTokenListPrice(coinName, coinAddress, currency)
	fmt.Println("price==v2===", price)
	return price, nil
}
