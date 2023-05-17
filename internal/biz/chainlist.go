package biz

import (
	"context"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/go-kratos/kratos/v2/log"
	"gitlab.bixin.com/mili/node-proxy/api/chainlist/v1"
	"gitlab.bixin.com/mili/node-proxy/internal/data/models"
	"strings"
	"time"
)

type ChainListRepo interface {
	Create(ctx context.Context, chainNode *models.ChainNodeUrl) error
	Update(ctx context.Context, chainNode *models.ChainNodeUrl) error
	GetAllChainList(ctx context.Context) ([]*models.BlockChain, error)
	FindChainListByChainIds(ctx context.Context, chainIds []string) ([]*models.BlockChain, error)
	FindChainNodeUrlByChainIds(ctx context.Context, chainIds []string) ([]*models.ChainNodeUrl, error)
	FindChainNodeUrlListWithSource(ctx context.Context, chainId string, source uint8) ([]*models.ChainNodeUrl, error)
	GetByChainIdAndUrl(ctx context.Context, chainId string, url string) (*models.ChainNodeUrl, error)
	GetAllWithInUsed(ctx context.Context) ([]*models.ChainNodeUrl, error)
}

type ChainListUsecase struct {
	repo ChainListRepo
	log  *log.Helper
}

// NewChainListUsecase new a ChainList usecase.
func NewChainListUsecase(repo ChainListRepo, logger log.Logger) *ChainListUsecase {
	return &ChainListUsecase{repo: repo, log: log.NewHelper(logger)}
}

func (uc *ChainListUsecase) GetAllChainList(ctx context.Context) ([]*v1.GetAllChainListResp_Data, error) {
	chainList, err := uc.repo.GetAllChainList(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*v1.GetAllChainListResp_Data, len(chainList))
	for i, chain := range chainList {
		result[i] = &v1.GetAllChainListResp_Data{
			ChainId:        chain.ChainId,
			Name:           chain.Name,
			Title:          chain.Title,
			Chain:          fmt.Sprintf("%s%s", "evm", chain.ChainId),
			CurrencyName:   chain.CurrencyName,
			CurrencySymbol: chain.CurrencySymbol,
			Decimals:       uint32(chain.Decimals),
			Explorer:       chain.Explorer,
			ChainSlug:      chain.ChainSlug,
			Logo:           chain.Logo,
			Type:           "EVM",
			IsTest:         chain.IsTest,
		}
	}

	return result, nil
}

func (uc *ChainListUsecase) GetChainList(ctx context.Context, chainIds []string) ([]*v1.GetChainListResp_Data, error) {
	chainList, err := uc.repo.FindChainListByChainIds(ctx, chainIds)
	if err != nil {
		return nil, err
	}

	result := make([]*v1.GetChainListResp_Data, len(chainList))
	for i, chain := range chainList {
		result[i] = &v1.GetChainListResp_Data{
			ChainId:        chain.ChainId,
			Name:           chain.Name,
			Title:          chain.Title,
			Chain:          fmt.Sprintf("%s%s", "evm", chain.ChainId),
			CurrencyName:   chain.CurrencyName,
			CurrencySymbol: chain.CurrencySymbol,
			Decimals:       uint32(chain.Decimals),
			Explorer:       chain.Explorer,
			ChainSlug:      chain.ChainSlug,
			Logo:           chain.Logo,
			Type:           "EVM",
			IsTest:         chain.IsTest,
		}
	}

	return result, nil
}

func (uc *ChainListUsecase) GetChainNodeUrlList(ctx context.Context, chainId string) ([]*v1.GetChainNodeListResp_Data, error) {
	nodeUrlList, err := uc.repo.FindChainNodeUrlListWithSource(ctx, chainId, models.ChainNodeUrlSourcePublic)
	if err != nil {
		return nil, err
	}

	result := make([]*v1.GetChainNodeListResp_Data, len(nodeUrlList))
	for i, nodeUrl := range nodeUrlList {
		result[i] = &v1.GetChainNodeListResp_Data{
			Url:     nodeUrl.Url,
			Privacy: nodeUrl.Privacy,
		}
	}

	return result, nil
}

func (uc *ChainListUsecase) UseChainNode(ctx context.Context, chainId string, url string, source uint32) error {

	switch uint8(source) {
	case models.ChainNodeUrlSourcePublic:
		nodeUrl, err := uc.repo.GetByChainIdAndUrl(ctx, chainId, url)
		if err != nil {
			if strings.Contains(err.Error(), "record not found") {
				return errors.New("use chain node error: chain node not found")
			}
			uc.log.Error(err.Error())
			return err
		}

		nodeUrl.InUsed = true

		err = uc.repo.Update(ctx, nodeUrl)
		if err != nil {
			uc.log.Error(err.Error())
			return errors.New("use chain node error: update chain node failed")
		}
	case models.ChainNodeUrlSourceCustom:
		//连接节点
		client, err := ethclient.Dial(url)
		if err != nil {
			return errors.New("use chain node error: connect to  node failed")
		}

		//检查ChainId
		checkCtx, cancelFunc := context.WithTimeout(context.Background(), time.Second*10)
		defer cancelFunc()
		chainID, err := client.ChainID(checkCtx)
		if err != nil {
			return errors.New("use chain node error: get chain id failed")
		}

		if chainID.String() != chainId {
			return errors.New("use chain node error: chain id not match")
		}

		err = uc.repo.Create(context.Background(), &models.ChainNodeUrl{
			ChainId: chainId,
			Url:     url,
			Status:  models.ChainNodeUrlStatusAvailable, //可以获取ChainId,说明节点可用
			InUsed:  true,
			Source:  models.ChainNodeUrlSourceCustom,
		})

		if err != nil {
			//创建失败，说明自定义网络与已有网络重复，修改状态
			if strings.Contains(err.Error(), "idx_chain_node_urls_url") {
				nodeUrl, err := uc.repo.GetByChainIdAndUrl(context.Background(), chainId, url)
				if err != nil {
					return errors.New("use chain node error: get chain node failed")
				}
				nodeUrl.InUsed = true
				err = uc.repo.Update(context.Background(), nodeUrl)
				if err != nil {
					return errors.New("use chain node error: update chain node failed")
				}
				return nil
			}

			uc.log.Error(err.Error())
			return errors.New("use chain node error: create chain node failed")
		}

	default:
		return errors.New("use chain node error: source not support")
	}

	return nil
}

func (uc *ChainListUsecase) GetChainNodeInUsedList(ctx context.Context) ([]*v1.GetChainNodeInUsedListResp_Data, error) {
	nodeUrlInUsedList, err := uc.repo.GetAllWithInUsed(ctx)
	if err != nil {
		return nil, err
	}

	// 去重
	var chainIdSet = make(map[string]struct{})
	var chainIds []string

	for _, url := range nodeUrlInUsedList {
		if _, ok := chainIdSet[url.ChainId]; !ok {
			chainIds = append(chainIds, url.ChainId)
		}
	}

	blockChains, err := uc.repo.FindChainListByChainIds(ctx, chainIds)
	if err != nil {
		return nil, err
	}

	chainNodeUrls, err := uc.repo.FindChainNodeUrlByChainIds(ctx, chainIds)
	if err != nil {
		return nil, err
	}

	nodeUrlMap := make(map[string][]*models.ChainNodeUrl)
	for i, nodeUrl := range chainNodeUrls {
		nodeUrlMap[nodeUrl.ChainId] = append(nodeUrlMap[nodeUrl.ChainId], chainNodeUrls[i])
	}

	result := make([]*v1.GetChainNodeInUsedListResp_Data, len(blockChains))
	for i, chain := range blockChains {
		result[i] = &v1.GetChainNodeInUsedListResp_Data{
			ChainId:        chain.ChainId,
			Name:           chain.Name,
			Title:          chain.Title,
			Chain:          fmt.Sprintf("%s%s", "evm", chain.ChainId),
			CurrencyName:   chain.CurrencyName,
			CurrencySymbol: chain.CurrencySymbol,
			Decimals:       uint32(chain.Decimals),
			Explorer:       chain.Explorer,
			ChainSlug:      chain.ChainSlug,
			Logo:           chain.Logo,
			Type:           "EVM",
		}

		urls := make([]string, len(nodeUrlMap[chain.ChainId]))
		for j, node := range nodeUrlMap[chain.ChainId] {
			urls[j] = node.Url
		}

		result[i].Urls = urls
	}

	return result, nil
}
