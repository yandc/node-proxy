package data

import (
	"context"
	"fmt"
	"gitlab.bixin.com/mili/node-proxy/pkg/gasOracle"

	"github.com/go-kratos/kratos/v2/log"
	"gitlab.bixin.com/mili/node-proxy/internal/biz"
	"gitlab.bixin.com/mili/node-proxy/internal/conf"
	"gitlab.bixin.com/mili/node-proxy/pkg/platform"
	"gitlab.bixin.com/mili/node-proxy/pkg/platform/types"
	"gitlab.bixin.com/mili/node-proxy/pkg/token-list/tokenlist"
)

type commRPCRepo struct {
	log       *log.Helper
	platforms []*conf.Platform
	chainData map[string]string
}

func NewCommRPCRepo(logger log.Logger, platforms []*conf.Platform, chainData map[string]string) biz.CommRPCRepo {
	return &commRPCRepo{
		log:       log.NewHelper(logger),
		platforms: platforms,
		chainData: chainData,
	}
}

func (c *commRPCRepo) GetPriceV2(ctx context.Context, coinName, coinAddress []string, currency string) (map[string]map[string]string, error) {
	c.log.WithContext(ctx).Infof("GetPriceV2", coinName, coinAddress, currency)
	price := tokenlist.GetTokenListPrice(coinName, coinAddress, currency)
	fmt.Println("price==v2===", price)
	return price, nil
}

func (c *commRPCRepo) GetContractABI(ctx context.Context, chain, contract, methodId string) (interface{}, error) {
	c.log.WithContext(ctx).Infof("GetContractABI", chain, contract, methodId)
	ret, err := platform.GetContractABI(chain, contract, methodId)
	fmt.Println("GetContractABI==result===", ret)
	return ret, err
}

func (c *commRPCRepo) ParseDataByABI(ctx context.Context, chain, contract, data string) *types.ParseDataResponse {
	c.log.WithContext(ctx).Infof("ParseDataByABI", chain, contract, data)
	ret := platform.ParseDataByABI(chain, contract, data)
	fmt.Println("ParseDataByABI==result===", ret)
	return ret
}

func (c *commRPCRepo) GetPretreatmentAmount(ctx context.Context, chain, from, to, data, value string) map[string][]interface{} {
	c.log.WithContext(ctx).Infof("GetPretreatmentAmount", chain, from, to, data, value)
	ret := platform.GetPretreatmentAmount(chain, from, to, data, value)
	c.log.WithContext(ctx).Infof("GetPretreatmentAmount==result===", ret)
	return ret
}

func (c *commRPCRepo) IsContractAddress(ctx context.Context, chain, address string) (bool, error) {
	c.log.WithContext(ctx).Infof("IsContractAddress", chain, address)
	ret, err := platform.IsContractAddress(chain, address)
	c.log.WithContext(ctx).Infof("IsContractAddress==result===", ret)
	return ret, err
}

func (c *commRPCRepo) GetGasConstants(ctx context.Context) map[string]interface{} {
	c.log.WithContext(ctx).Infof("GetGasConstants")
	result := make(map[string]interface{})
	for _, p := range c.platforms {
		if p.GetGasDefaults() != nil {
			result[p.Chain] = p.GetGasDefaults()
		}
	}
	c.log.WithContext(ctx).Infof("GetGasConstants==result===", result)
	return result
}

func (c *commRPCRepo) GetChainDataConfig(ctx context.Context) map[string]interface{} {
	c.log.WithContext(ctx).Infof("GetChainDataConfig")
	result := make(map[string]interface{})
	gasDefaults := make(map[string]interface{})
	for _, p := range c.platforms {
		if p.GetGasDefaults() != nil {
			gasDefaults[p.Chain] = p.GetGasDefaults()
		}
	}
	result["gasDefaults"] = gasDefaults
	result["chainData"] = c.chainData
	c.log.WithContext(ctx).Infof("GetChainDataConfig==result===", result)
	return result
}

func (c *commRPCRepo) GetGasOracle(ctx context.Context, key string, cacheTime int64) string {
	c.log.WithContext(ctx).Infof("GetGasOracle")
	result := gasOracle.GetGasOracle(key, cacheTime)
	c.log.WithContext(ctx).Infof("GetGasOracle result=", result)
	return result
}
