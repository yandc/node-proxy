package data

import (
	"context"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-redis/redis"
	"gitlab.bixin.com/mili/node-proxy/internal/biz"
	"gitlab.bixin.com/mili/node-proxy/internal/data/models"
	"gitlab.bixin.com/mili/node-proxy/pkg/chainlist"
	"gorm.io/gorm"
)

type chainListRepo struct {
	log *log.Helper
}

// NewChainListRepo .
func NewChainListRepo(db *gorm.DB, client *redis.Client, logger log.Logger) biz.ChainListRepo {
	chainlist.InitChainList(db, client, logger)
	return &chainListRepo{
		log: log.NewHelper(logger),
	}
}

func (r chainListRepo) Create(ctx context.Context, chainNode *models.ChainNodeUrl) error {
	return chainlist.Create(chainNode)
}

func (r chainListRepo) Update(ctx context.Context, chainNode *models.ChainNodeUrl) error {
	return chainlist.Update(chainNode)
}

func (r chainListRepo) GetAllChainList(ctx context.Context) ([]*models.BlockChain, error) {
	return chainlist.GetAllBlockChain()
}

func (r chainListRepo) GetBlockChainByChainId(ctx context.Context, chainId string) (*models.BlockChain, error) {
	return chainlist.GetBlockChainByChainId(chainId)
}

func (r chainListRepo) FindBlockChainsByChainIds(ctx context.Context, chainIds []string) ([]*models.BlockChain, error) {
	return chainlist.FindBlockChainByChainIds(chainIds)
}

func (r chainListRepo) FindChainNodeUrlByChainIds(ctx context.Context, chainIds []string) ([]*models.ChainNodeUrl, error) {
	return chainlist.FindChainNodeUrlByChainIds(chainIds)
}

func (r chainListRepo) FindChainNodeUrlListWithSource(ctx context.Context, chainId string, source uint8) ([]*models.ChainNodeUrl, error) {
	return chainlist.FindChainNodeUrlListWithSource(chainId, source)
}

func (r chainListRepo) GetByChainIdAndUrl(ctx context.Context, chainId string, url string) (*models.ChainNodeUrl, error) {
	return chainlist.GetByChainIdAndUrl(chainId, url)
}

func (r chainListRepo) GetAllWithInUsed(ctx context.Context) ([]*models.ChainNodeUrl, error) {
	return chainlist.GetAllWithInUsed()
}

func (r chainListRepo) GetChainListByType(ctx context.Context, chainType string) ([]*models.BlockChain, error) {
	return chainlist.FindBlockChainByChainType(chainType)
}

func (r chainListRepo) CheckChainIdByType(ctx context.Context, chainType, chainId, rpc string) error {
	return chainlist.CheckTypeChainId(chainType, chainId, rpc)
}
