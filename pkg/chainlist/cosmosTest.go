package chainlist

import (
	"fmt"
	"github.com/go-kratos/kratos/v2/log"
	"gitlab.bixin.com/mili/node-proxy/internal/data/models"
	"gitlab.bixin.com/mili/node-proxy/pkg/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"strings"
	"sync"
)

type cosmosTestPath struct {
	Payload struct {
		AllShortcutsEnabled bool `json:"allShortcutsEnabled"`
		Tree                struct {
			Items []struct {
				Name        string `json:"name"`
				Path        string `json:"path"`
				ContentType string `json:"contentType"`
			} `json:"items"`
			TemplateDirectorySuggestionURL interface{} `json:"templateDirectorySuggestionUrl"`
			Readme                         interface{} `json:"readme"`
			TotalCount                     int         `json:"totalCount"`
			ShowBranchInfobar              bool        `json:"showBranchInfobar"`
		} `json:"tree"`
	} `json:"payload"`
}

type cosmosTestChainInfo struct {
	Schema       string   `json:"$schema"`
	ChainName    string   `json:"chain_name"`
	Status       string   `json:"status"`
	NetworkType  string   `json:"network_type"`
	PrettyName   string   `json:"pretty_name"`
	ChainID      string   `json:"chain_id"`
	Bech32Prefix string   `json:"bech32_prefix"`
	DaemonName   string   `json:"daemon_name"`
	NodeHome     string   `json:"node_home"`
	KeyAlgos     []string `json:"key_algos"`
	Slip44       int      `json:"slip44"`
	Fees         struct {
		FeeTokens []struct {
			Denom string `json:"denom"`
		} `json:"fee_tokens"`
	} `json:"fees"`
	Apis struct {
		Rest []struct {
			Address  string `json:"address"`
			Provider string `json:"provider"`
		} `json:"rest"`
	} `json:"apis"`
	Explorers []struct {
		Kind        string `json:"kind"`
		URL         string `json:"url"`
		TxPage      string `json:"tx_page"`
		AccountPage string `json:"account_page"`
	} `json:"explorers"`
}

func GetTestCosmosChainList(log *log.Helper, db *gorm.DB) {
	subJobName := "GetTestCosmosChainList"
	log.Infof("子任务执行开始：%s", subJobName)
	dirPath := "https://raw.githubusercontent.com/cosmos/chain-registry/master"
	url := "https://github.com/cosmos/chain-registry/blob/master/testnets"
	//	req.Header.Set("Accept", "application/json")
	header := map[string]string{
		"Accept": "application/json",
	}
	testPaths := &cosmosTestPath{}
	if err := utils.HttpsGet(url, nil, header, testPaths); err != nil {
		log.Errorf("任务执行失败：%s,err：%s", subJobName, err.Error())
		return
	}
	fmt.Println("length=", len(testPaths.Payload.Tree.Items))
	for _, item := range testPaths.Payload.Tree.Items {
		if !strings.HasPrefix(item.Name, "_") && item.ContentType == "directory" {
			dirUrl := fmt.Sprintf("%s/%s/chain.json", dirPath, item.Path)
			chainInfo := &cosmosTestChainInfo{}
			if err := utils.HttpsGet(dirUrl, nil, nil, chainInfo); err != nil {
				continue
			}
			if chainInfo.Status != "live" {
				continue
			}
			chainName := strings.Replace(chainInfo.ChainName, "testnet", "", -1) + "TEST"
			blockChain := &models.BlockChain{
				Name:      chainInfo.PrettyName,
				Chain:     strings.ToUpper(chainName[:1]) + chainName[1:],
				ChainId:   chainInfo.ChainID,
				IsTest:    true,
				ChainType: models.ChainTypeCOSMOS,
				Prefix:    chainInfo.Bech32Prefix,
			}
			if chainInfo.Fees.FeeTokens != nil && len(chainInfo.Fees.FeeTokens) > 0 {
				blockChain.Denom = chainInfo.Fees.FeeTokens[0].Denom
			} else {
				continue
			}
			if strings.HasPrefix(blockChain.Denom, "u") {
				blockChain.Decimals = 6
			}
			if chainInfo.Explorers != nil && len(chainInfo.Explorers) > 0 {
				blockChain.Explorer = chainInfo.Explorers[0].URL
				if chainInfo.Explorers[0].AccountPage != "" {
					blockChain.ExplorerAddress = strings.Split(chainInfo.Explorers[0].AccountPage, "${accountAddress}")[0]
					//blockChain.ExplorerAddress = chain.Explorers[0].AccountPage
				} else {
					blockChain.ExplorerAddress = GetExplorerURL(chainInfo.Explorers[0].URL) + "address/"
				}
				if chainInfo.Explorers[0].TxPage != "" {
					blockChain.ExplorerTx = strings.Split(chainInfo.Explorers[0].TxPage, "${txHash}")[0]
				} else {
					blockChain.ExplorerTx = GetExplorerURL(chainInfo.Explorers[0].URL) + "tx/"
				}
			}
			if strings.Contains(chainInfo.ChainName, "testnet") {
				tempChainName := strings.Split(chainInfo.ChainName, "testnet")[0]
				nativeChainName := strings.ToUpper(tempChainName[:1]) + tempChainName[1:]
				dbBlockChain := &models.BlockChain{}
				if err := db.Where("chain = ? and is_test = ?", nativeChainName, false).First(&dbBlockChain).Error; err == nil {
					blockChain.Decimals = dbBlockChain.Decimals
					blockChain.Logo = dbBlockChain.Logo
					blockChain.CurrencySymbol = dbBlockChain.CurrencySymbol
				}
			}
			if err := db.Clauses(clause.OnConflict{DoNothing: true}).Create(blockChain).Error; err != nil {
				log.Errorf("任务执行异常：%s,err：%s", subJobName, err.Error())
			}
			var nodeUrls []*models.ChainNodeUrl
			for _, rpc := range chainInfo.Apis.Rest {
				if !strings.HasPrefix(rpc.Address, "https://") {
					continue
				}
				rpc.Address = strings.TrimRight(rpc.Address, "/")
				nodeUrl := &models.ChainNodeUrl{
					ChainId: chainInfo.ChainID,
					Url:     rpc.Address,
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
	}
	log.Infof("子任务执行完成：%s", subJobName)
}
