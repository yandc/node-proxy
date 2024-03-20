package chainlist

import (
	"errors"
	"fmt"
	"github.com/go-kratos/kratos/v2/log"
	"gitlab.bixin.com/mili/node-proxy/internal/data/models"
	"gitlab.bixin.com/mili/node-proxy/pkg/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"strconv"
	"strings"
	"sync"
)

type cosmosNativeInfo struct {
	Chains []struct {
		Name         string `json:"name"`
		Path         string `json:"path"`
		ChainName    string `json:"chain_name"`
		NetworkType  string `json:"network_type"`
		PrettyName   string `json:"pretty_name"`
		ChainID      string `json:"chain_id"`
		Status       string `json:"status"`
		Bech32Prefix string `json:"bech32_prefix"`
		Slip44       int    `json:"slip44,omitempty"`
		Symbol       string `json:"symbol,omitempty"`
		Display      string `json:"display,omitempty"`
		Denom        string `json:"denom,omitempty"`
		Decimals     int    `json:"decimals,omitempty"`
		CoingeckoID  string `json:"coingecko_id,omitempty"`
		Image        string `json:"image,omitempty"`
		Website      string `json:"website,omitempty"`
		Height       uint64 `json:"height"`
		BestApis     struct {
			Rest []struct {
				Address  string `json:"address"`
				Provider string `json:"provider"`
			} `json:"rest"`
		} `json:"best_apis"`
		ProxyStatus struct {
			Rest bool `json:"rest"`
			RPC  bool `json:"rpc"`
		} `json:"proxy_status"`
		Versions struct {
			CosmosSdkVersion   string `json:"cosmos_sdk_version"`
			TendermintVersion  string `json:"tendermint_version"`
			ApplicationVersion string `json:"application_version"`
		} `json:"versions"`
		CosmwasmEnabled bool `json:"cosmwasm_enabled,omitempty"`
		Explorers       []struct {
			URL         string `json:"url"`
			TxPage      string `json:"tx_page"`
			Kind        string `json:"kind,omitempty"`
			AccountPage string `json:"account_page"`
		} `json:"explorers,omitempty"`
		Services struct {
			StakingRewards struct {
				Name   string `json:"name"`
				Slug   string `json:"slug"`
				Symbol string `json:"symbol"`
			} `json:"staking_rewards"`
		} `json:"services,omitempty"`
		Keywords []string `json:"keywords,omitempty"`
	} `json:"chains"`
}

type cosmosChainId struct {
	Block struct {
		Header struct {
			ChainID string `json:"chain_id"`
			Height  string `json:"height"`
		} `json:"header"`
	} `json:"block"`
}

func GetNativeCosmosChainList(log *log.Helper, db *gorm.DB) {
	subJobName := "GetNativeCosmosChainList"
	log.Infof("子任务执行开始：%s", subJobName)
	url := "https://chains.cosmos.directory/"
	chainListData := &cosmosNativeInfo{}
	if err := utils.HttpsGet(url, nil, nil, chainListData); err != nil {
		log.Errorf("任务执行失败：%s,err：%s", subJobName, err.Error())
		return
	}
	fmt.Println("length=1=", len(chainListData.Chains))
	for _, chain := range chainListData.Chains {
		if chain.Status != "live" {
			continue
		}
		blockChain := &models.BlockChain{
			Name:           chain.PrettyName,
			Chain:          strings.ToUpper(chain.ChainName[:1]) + chain.ChainName[1:],
			ChainId:        chain.ChainID,
			CurrencyName:   chain.Name,
			CurrencySymbol: chain.Symbol,
			Decimals:       uint8(chain.Decimals),
			IsTest:         false,
			GetPriceKey:    chain.CoingeckoID,
			ChainType:      models.ChainTypeCOSMOS,
			Logo:           chain.Image,
			ChainSlug:      chain.Services.StakingRewards.Slug,
			Prefix:         chain.Bech32Prefix,
			Denom:          chain.Denom,
		}
		if strings.Contains(strings.ToLower(chain.NetworkType), "testnet") {
			blockChain.IsTest = true
			blockChain.Chain = strings.TrimRight(blockChain.Chain, "testnet") + "TEST"
		}
		if chain.Explorers != nil && len(chain.Explorers) > 0 {
			blockChain.Explorer = chain.Explorers[0].URL
			if chain.Explorers[0].AccountPage != "" {
				blockChain.ExplorerAddress = strings.Split(chain.Explorers[0].AccountPage, "${accountAddress}")[0]
				//blockChain.ExplorerAddress = chain.Explorers[0].AccountPage
			} else {
				blockChain.ExplorerAddress = GetExplorerURL(chain.Explorers[0].URL) + "address/"
			}
			if chain.Explorers[0].TxPage != "" {
				blockChain.ExplorerTx = strings.Split(chain.Explorers[0].TxPage, "${txHash}")[0]
			} else {
				blockChain.ExplorerTx = GetExplorerURL(chain.Explorers[0].URL) + "tx/"
			}
		}
		if err := db.Clauses(clause.OnConflict{DoNothing: true}).Create(blockChain).Error; err != nil {
			log.Errorf("任务执行异常：%s,err：%s", subJobName, err.Error())
		}
		var nodeUrls []*models.ChainNodeUrl
		for _, rpc := range chain.BestApis.Rest {
			if !strings.HasPrefix(rpc.Address, "https://") {
				continue
			}
			rpc.Address = strings.TrimRight(rpc.Address, "/")
			nodeUrl := &models.ChainNodeUrl{
				ChainId: chain.ChainID,
				Url:     rpc.Address,
				Height:  chain.Height,
				Status:  models.ChainNodeUrlStatusAvailable, //默认可用
				Source:  models.ChainNodeUrlSourcePublic,
			}
			nodeUrls = append(nodeUrls, nodeUrl)
		}
		group := sync.WaitGroup{}
		for _, nodeUrl := range nodeUrls {
			group.Add(1)
			go checkCOSMOSChainId(nodeUrl, &group)
		}
		group.Wait()

		if len(nodeUrls) != 0 {
			if err := db.Clauses(clause.OnConflict{DoNothing: true}).Create(nodeUrls).Error; err != nil {
				log.Errorf("任务执行异常：%s,err：%s", subJobName, err.Error())
			}
		}
	}
	log.Infof("子任务执行完成：%s", subJobName)

}

func checkCOSMOSChainId(nodeUrl *models.ChainNodeUrl, group *sync.WaitGroup) error {
	defer group.Done()
	//url := fmt.Sprintf("%s/cosmos/base/tendermint/v1beta1/blocks/latest", nodeUrl.Url)
	//out := &cosmosChainId{}
	//err := utils.HttpsGet(url, nil, nil, out)
	//if err != nil {
	//	nodeUrl.Status = 2
	//	return errors.New(fmt.Sprintf("check chainId error , err : %s", err.Error()))
	//}
	//if out.Block.Header.ChainID != nodeUrl.ChainId {
	//	nodeUrl.Status = 2
	//	return errors.New(fmt.Sprintf("chain id not match,expect:%d,get:%d", out.Block.Header.ChainID, nodeUrl.ChainId))
	//}
	heightStr, err := CheckCosmosChainId(nodeUrl.ChainId, nodeUrl.Url)
	if err != nil {
		nodeUrl.Status = 2
		return err
	}
	if height, err := strconv.ParseUint(heightStr, 10, 64); err != nil {
		nodeUrl.Height = height
	}
	return nil
}

func CheckCosmosChainId(chainId, rpc string) (string, error) {
	url := fmt.Sprintf("%s/cosmos/base/tendermint/v1beta1/blocks/latest", rpc)
	out := &cosmosChainId{}
	err := utils.HttpsGet(url, nil, nil, out)
	if err != nil {
		return "", errors.New(fmt.Sprintf("check chainId error , err : %s", err.Error()))
	}
	if out.Block.Header.ChainID != chainId {
		return "", errors.New(fmt.Sprintf("chain id not match,expect:%v,get:%v", out.Block.Header.ChainID, chainId))
	}
	return out.Block.Header.Height, nil
}
