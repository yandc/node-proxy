package chainlist

import (
	"errors"
	"fmt"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-redis/redis"
	v1 "gitlab.bixin.com/mili/node-proxy/api/chainlist/v1"
	"gitlab.bixin.com/mili/node-proxy/internal/data/models"
	"gitlab.bixin.com/mili/node-proxy/pkg/utils"
	"gorm.io/gorm"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
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

func GetBlockChainByChainId(chainId string) (*models.BlockChain, error) {
	var blockChain *models.BlockChain
	if err := c.db.Where("chain_id = ?", chainId).First(&blockChain).Error; err != nil {
		return nil, err
	}
	return blockChain, nil
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
	if err := c.db.Where("chain_id in ?", chainIds).Order("id asc").Find(&nodeUrls).Error; err != nil {
		return nil, err
	}
	return nodeUrls, nil
}

func FindChainNodeUrlList(chainId string) ([]*models.ChainNodeUrl, error) {
	var nodeUrls []*models.ChainNodeUrl
	if err := c.db.Where("chain_id = ?", chainId).Order("height desc,latency asc ").Find(&nodeUrls).Error; err != nil {
		return nil, err
	}
	return nodeUrls, nil
}

func FindChainNodeUrlListWithSource(chainId string, source uint8) ([]*models.ChainNodeUrl, error) {
	var nodeUrls []*models.ChainNodeUrl
	if err := c.db.Where("chain_id = ? and source = ?", chainId, source).Order("height desc,latency asc ").Find(&nodeUrls).Error; err != nil {
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

func GetPriceKeyBySymbol(symbol string) string {
	var tempCoinGeckoList models.CoinGeckoList
	c.db.Where("symbol = ?", strings.ToLower(symbol)).First(&tempCoinGeckoList)
	if tempCoinGeckoList.CgId != "" {
		return tempCoinGeckoList.CgId
	}
	return ""
}

func FindBlockChainByChainType(chainType string) ([]*models.BlockChain, error) {
	var blockChains []*models.BlockChain
	if err := c.db.Where("chain_type = ?", chainType).Order("id asc").Find(&blockChains).Error; err != nil {
		return nil, err
	}
	return blockChains, nil
}

func CheckTypeChainId(chainType, chainId, rpc string) error {
	switch chainType {
	case models.ChainTypeEVM:
		return CheckEVMChainIdByURL(chainId, rpc)
	case models.ChainTypeCOSMOS:
		_, err := CheckCosmosChainId(chainId, rpc)
		return err
	}
	return errors.New(chainType + " is not support")
}

func GetExplorerURL(explorer string) string {
	if strings.HasSuffix(explorer, "/") {
		return explorer
	}
	return explorer + "/"
}

type BlockChainInfo struct {
	*v1.BlockChainData
	Currency string     `json:"currency"`
	Hrp      string     `json:"hrp"`
	Nodes    []NodeInfo `json:"nodes"`
}

type NodeInfo struct {
	URL     string `json:"url"`
	Privacy string `json:"privacy"`
}

func UpLoadChainList2CDN() {
	chainTypes := []string{"EVM", "COSMOS"}
	localPath := "chainList"
	exist, _ := utils.PathExists(localPath)
	if exist {
		err := os.RemoveAll(localPath)
		c.log.Error("remove path error:", err.Error())
	}
	os.MkdirAll(localPath, 0777)
	for _, chainType := range chainTypes {
		blockChains, err := FindBlockChainByChainType(chainType)
		if err != nil {
			continue
		}
		blockChainInfo := make([]BlockChainInfo, 0, len(blockChains))
		for _, blockChain := range blockChains {
			nodeURLS, err := FindChainNodeUrlListWithSource(blockChain.ChainId, models.ChainNodeUrlSourcePublic)
			if err != nil || nodeURLS == nil || len(nodeURLS) == 0 {
				continue
			}
			urls := make([]NodeInfo, 0, len(nodeURLS))
			for _, nodeUrl := range nodeURLS {
				urls = append(urls, NodeInfo{
					URL:     nodeUrl.Url,
					Privacy: nodeUrl.Privacy,
				})
			}
			blockChainData := DBBlockChain2Data(blockChain)
			blockChainInfo = append(blockChainInfo, BlockChainInfo{
				BlockChainData: blockChainData, Currency: blockChain.CurrencySymbol, Hrp: blockChain.Prefix, Nodes: urls,
			})
		}
		//写数据到本地本地
		fileName := fmt.Sprintf("%s/%s.json", localPath, chainType)
		err = utils.WriteInfoToFile(fileName, blockChainInfo)
		if err != nil {
			c.log.Error("编码错误", err.Error())
		}
	}
	var paths []string
	filepath.Walk(localPath, func(path string, info fs.FileInfo, err error) error {
		if !info.IsDir() {
			paths = append(paths, path)
		}
		return nil
	})
	//上传到cdn
	utils.UploadFileToS3(paths)
	utils.UpLoadFile2QiNiu(paths)
	//删除目录
	err := os.RemoveAll(localPath)
	if err != nil {
		c.log.Error("remove path error", err.Error())
	}
}

func DBBlockChain2Data(chain *models.BlockChain) *v1.BlockChainData {
	if chain.GetPriceKey == "" {
		chain.GetPriceKey = GetPriceKeyBySymbol(chain.CurrencySymbol)
	}
	chainName := chain.Chain
	if chain.ChainType == models.ChainTypeEVM {
		chainName = fmt.Sprintf("%s%s", "evm", chain.ChainId)
	} else if chain.ChainType == models.ChainTypeCOSMOS {
		chainName = fmt.Sprintf("%s%s", "cosmos", chain.ChainId)
	}

	data := &v1.BlockChainData{
		ChainId:        chain.ChainId,
		Name:           chain.Name,
		Title:          chain.Title,
		Chain:          chainName,
		CurrencyName:   chain.CurrencyName,
		CurrencySymbol: chain.CurrencySymbol,
		Decimals:       uint32(chain.Decimals),
		Explorer:       chain.Explorer,
		ChainSlug:      chain.ChainSlug,
		Logo:           chain.Logo,
		Type:           chain.ChainType,
		IsTest:         chain.IsTest,
		GetPriceKey:    chain.GetPriceKey,
		Prefix:         chain.Prefix,
		Denom:          chain.Denom,
		ExplorerTx:     chain.ExplorerTx,
		ExplorerAddr:   chain.ExplorerAddress,
	}
	if data.ExplorerAddr == "" {
		data.ExplorerAddr = GetExplorerURL(data.Explorer) + "address/"
	}
	if data.ExplorerTx == "" {
		data.ExplorerTx = GetExplorerURL(data.Explorer) + "tx/"
	}
	return data
}
