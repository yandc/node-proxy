package chainlist

import (
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-redis/redis"
	"gitlab.bixin.com/mili/node-proxy/internal/data/models"
	"gorm.io/gorm"
)

type config struct {
	db          *gorm.DB
	log         *log.Helper
	redisClient *redis.Client
}

var c config

func InitChainList(db *gorm.DB, client *redis.Client, logger log.Logger) {
	log := log.NewHelper(log.With(logger, "module", "chainlist/InitChainList"))
	c = config{
		db:          db,
		log:         log,
		redisClient: client,
	}
}

func Create(chainNode *models.ChainNodeUrl) error {
	return c.db.Create(chainNode).Error
}

func Update(chainNode *models.ChainNodeUrl) error {
	return c.db.Save(chainNode).Error
}

func GetAllBlockChain() ([]*models.BlockChain, error) {
	var blockChains []*models.BlockChain
	if err := c.db.Order("id asc").Find(&blockChains).Error; err != nil {
		return nil, err
	}
	return blockChains, nil
}

func FindBlockChainByChainIds(chainIds []string) ([]*models.BlockChain, error) {
	var blockChains []*models.BlockChain
	if err := c.db.Where("chain_id in ?", chainIds).Order("id asc").Find(&blockChains).Error; err != nil {
		return nil, err
	}
	return blockChains, nil
}

func FindChainNodeUrlByChainIds(chainIds []string) ([]*models.ChainNodeUrl, error) {
	var nodeUrls []*models.ChainNodeUrl
	if err := c.db.Where("chain_id in ? and status = 1", chainIds).Order("id asc").Find(&nodeUrls).Error; err != nil {
		return nil, err
	}
	return nodeUrls, nil
}

func FindChainNodeUrlList(chainId string) ([]*models.ChainNodeUrl, error) {
	var nodeUrls []*models.ChainNodeUrl
	if err := c.db.Where("chain_id = ? and status = 1", chainId).Order("height desc,latency asc ").Find(&nodeUrls).Error; err != nil {
		return nil, err
	}
	return nodeUrls, nil
}

func FindChainNodeUrlListWithSource(chainId string, source uint8) ([]*models.ChainNodeUrl, error) {
	var nodeUrls []*models.ChainNodeUrl
	if err := c.db.Where("chain_id = ? and source = ? and status = 1", chainId, source).Order("height desc,latency asc ").Find(&nodeUrls).Error; err != nil {
		return nil, err
	}
	return nodeUrls, nil
}

func GetByChainIdAndUrl(chainId string, url string) (*models.ChainNodeUrl, error) {
	var nodeUrl models.ChainNodeUrl
	if err := c.db.Where("chain_id = ? and url = ?", chainId, url).First(&nodeUrl).Error; err != nil {
		return nil, err
	}

	return &nodeUrl, nil
}

func GetAllWithInUsed() ([]*models.ChainNodeUrl, error) {
	var nodeUrls []*models.ChainNodeUrl
	if err := c.db.Where("in_used = ?", true).Find(&nodeUrls).Error; err != nil {
		return nil, err
	}

	return nodeUrls, nil
}
