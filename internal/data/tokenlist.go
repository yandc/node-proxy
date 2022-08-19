package data

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-redis/redis"
	v1 "gitlab.bixin.com/mili/node-proxy/api/tokenlist/v1"
	"gitlab.bixin.com/mili/node-proxy/internal/biz"
	"gitlab.bixin.com/mili/node-proxy/internal/conf"
	"gitlab.bixin.com/mili/node-proxy/pkg/token-list/tokenlist"
	"gorm.io/gorm"
)

type tokenListRepo struct {
	log *log.Helper
}

// NewTokenListRepo .
func NewTokenListRepo(conf *conf.TokenList, db *gorm.DB, client *redis.Client, logger log.Logger) biz.TokenListRepo {
	tokenlist.InitTokenList(conf, db, client, logger)
	return &tokenListRepo{
		log: log.NewHelper(logger),
	}
}

func (r *tokenListRepo) GetPrice(ctx context.Context, coinName, coinAddress, currency string) ([]byte, error) {
	r.log.WithContext(ctx).Infof("GetPrice", coinName, coinAddress, currency)
	var address, chainName []string
	if len(coinAddress) > 0 {
		address = strings.Split(coinAddress, ",")
	}
	if len(coinName) > 0 {
		chainName = strings.Split(coinName, ",")
	}
	price := tokenlist.GetTokenListPrice(chainName, address, currency)
	b, err := json.Marshal(price)
	if err != nil {
		r.log.WithContext(ctx).Error("marshal error", err)
	}
	return b, nil
}

//func (r *tokenListRepo) CreateTokenList(ctx context.Context) {
//	r.log.WithContext(ctx).Infof("CreateTokenList")
//	//tokenlist.CreateTokenList()
//}

func (r *tokenListRepo) GetTokenList(ctx context.Context, chain string) ([]*v1.GetTokenListResp_Data, error) {
	r.log.WithContext(ctx).Infof("GetTokenList", chain)
	return tokenlist.GetTokenList(chain)
}

func (r *tokenListRepo) AutoUpdateTokenList(ctx context.Context) {
	r.log.WithContext(ctx).Infof("AutoUpdateTokenList")
	tokenlist.AutoUpdateTokenList(true, true, true)
}

func (r *tokenListRepo) GetTokenInfo(ctx context.Context, addressInfo []*v1.GetTokenInfoReq_Data) (
	[]*v1.GetTokenInfoResp_Data, error) {
	r.log.WithContext(ctx).Infof("GetTokenInfo", addressInfo)
	return tokenlist.GetTokenInfo(addressInfo)
}
