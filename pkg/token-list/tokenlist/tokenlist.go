package tokenlist

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-redis/redis"
	"github.com/qiniu/go-sdk/v7/auth/qbox"
	"github.com/qiniu/go-sdk/v7/cdn"
	"github.com/qiniu/go-sdk/v7/storage"
	"github.com/shopspring/decimal"
	v12 "gitlab.bixin.com/mili/node-proxy/api/market/v1"
	v1 "gitlab.bixin.com/mili/node-proxy/api/tokenlist/v1"
	"gitlab.bixin.com/mili/node-proxy/internal/conf"
	"gitlab.bixin.com/mili/node-proxy/internal/data/models"
	"gitlab.bixin.com/mili/node-proxy/pkg/lark"
	"gitlab.bixin.com/mili/node-proxy/pkg/platform"
	"gitlab.bixin.com/mili/node-proxy/pkg/token-list/types"
	"gitlab.bixin.com/mili/node-proxy/pkg/token-list/utils"
	utils3 "gitlab.bixin.com/mili/node-proxy/pkg/utils"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type config struct {
	db            *gorm.DB
	log           *log.Helper
	redisClient   *redis.Client
	logoPrefix    string
	awsLogoPrefie string
	qiniu         types.QiNiuConf
	aws           []*conf.TokenList_AWS
	chains        []string
}

const (
	REDIS_PRICE_KEY    = "tokenlist:price:"
	REDIS_LIST_KEY     = "tokenlist:list:"
	REDIS_TOKEN_KEY    = "tokenlist:tokeninfo:"
	REDIS_TOP20_KEY    = "tokenlist:top20:"
	REDIS_TOP50_KEY    = "tokenlist:top50:"
	REDIS_FAKECOIN_KEY = "tokenlist:fakecoin:"
)

var c config

var updatePriceChannel chan []string
var createPriceChannel chan []string

func InitTokenList(conf *conf.TokenList, db *gorm.DB, client *redis.Client, logger log.Logger) {
	log := log.NewHelper(log.With(logger, "module", "tokenlist/InittokenList"))
	updatePriceChannel = make(chan []string, 100)
	createPriceChannel = make(chan []string, 200)
	c = config{
		db:            db,
		log:           log,
		redisClient:   client,
		logoPrefix:    conf.LogoPrefix,
		awsLogoPrefie: conf.AwsLogoPrefix,
		qiniu: types.QiNiuConf{
			AccessKey: conf.Qiniu.AccessKey,
			SecretKey: conf.Qiniu.SecretKey,
			Bucket:    conf.Qiniu.Bucket,
			KeyPrefix: conf.Qiniu.KeyPrefix,
		},
		aws:    conf.Aws,
		chains: conf.Chains,
	}

	InitCG(conf.Coingecko.BaseURL, db, logger)
	InitCMC(conf.Coinmarketcap.BaseURL, conf.Coinmarketcap.Key, db, logger)
}

func HandlerPrice() {
	for {
		select {
		case priceChains := <-updatePriceChannel:
			go updateCreatePrice(priceChains)
		case priceChains := <-createPriceChannel:
			go updateCreatePrice(priceChains)
		}
	}
}

func updateCreatePrice(priceChains []string) {
	var cgIds, cmcIds []string
	addressIdMap := make(map[string]string, len(priceChains))
	tokenAddresses := make([]*v12.Tokens, 0, len(priceChains))

	//处理地址
	for _, priceChain := range priceChains {
		if !strings.Contains(priceChain, "_") {
			cgIds = append(cgIds, priceChain)
			continue
		}
		addressInfo := strings.Split(priceChain, "_")
		chain := utils.GetChainNameByPlatform(addressInfo[0])
		address := utils.GetUnificationAddress(chain, addressInfo[1])
		//if strings.HasPrefix(address, "0x") && chain != utils.STARCOIN_CHAIN && chain != utils.APTOS_CHAIN && !strings.Contains(chain, "SUI") {
		//	address = strings.ToLower(address)
		//}
		var tempTokenList models.TokenList
		c.db.Where("chain = ? AND address = ?", chain, address).First(&tempTokenList)

		var id string
		if tempTokenList.CgId != "" {
			id = tempTokenList.CgId
			cgIds = append(cgIds, tempTokenList.CgId)
		} else if tempTokenList.CmcId > 0 {
			id = fmt.Sprintf("%d", tempTokenList.CmcId)
			cmcIds = append(cmcIds, fmt.Sprintf("%v", id))
		} else if len(address) > 0 {
			//tokenAddress = append(tokenAddress, address)
			tokenAddresses = append(tokenAddresses, &v12.Tokens{
				Chain:   utils3.GetChainByHandler(addressInfo[0]),
				Address: address,
			})
			id = fmt.Sprintf("%s_%s", utils3.GetChainByHandler(addressInfo[0]), address)
		}
		if id != "" {
			addressIdMap[id] = fmt.Sprintf("%s_%s", addressInfo[0], address)
		}
	}
	if len(cgIds) > 0 {
		go handlerCgPrice(cgIds, addressIdMap)
	}
	if len(cmcIds) > 0 {
		go handlerCMCPrice(cmcIds, addressIdMap)
	}
	if len(tokenAddresses) > 0 {
		go handlerPriceByToken(tokenAddresses, addressIdMap)
	}

}

func CreateSource() {
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		CreateCoinMarketCap()
	}()
	go func() {
		defer wg.Done()
		CreateCoinGecko()
	}()
	wg.Wait()
}

func CreateTokenList() {
	//CreateSource()
	coinMarketCaps, err := GetAllCoinMarketCap()
	if err != nil {
		c.log.Error("GetAllCoinMarketCap Error:", err)
	}
	coinGeckos, err := GetAllCoinGecko()
	if err != nil {
		c.log.Error("GetAllCoinGecko Error:", err)
	}
	decimalsInfo := GetDecimalsInfo()
	tokenListMap := make(map[string]*models.TokenList)
	noDecimals := make(map[string][]string)
	for _, cg := range coinGeckos {
		var platform map[string]interface{}
		json.Unmarshal([]byte(cg.Platform), &platform)
		for key, value := range platform {
			if key != "" && value != "" {
				if cgcValue, cgcOk := utils.CGCNameChainMap[key]; cgcOk {
					key = cgcValue
				}
				chain := utils.GetPlatform(key)
				var address string
				if strings.HasPrefix(value.(string), "0x") {
					address = strings.ToLower(value.(string))
				} else {
					address = value.(string)
				}
				var decimals int
				if tokenInfo, ok := decimalsInfo[chain+":"+address]; ok {
					decimals = tokenInfo.Decimals
				}

				if decimals == 0 {
					noDecimals[chain] = append(noDecimals[chain], address)
				}
				tokenListMap[chain+":"+address] = &models.TokenList{
					CgId:        cg.Id,
					Name:        cg.Name,
					Symbol:      strings.ToUpper(cg.Symbol),
					WebSite:     cg.Homepage,
					Description: cg.Description,
					Chain:       chain,
					Address:     address,
					Logo:        cg.Image,
					Decimals:    decimals,
				}
			}
		}
	}
	count := 0
	//handler coinMarketCap
	for _, cmc := range coinMarketCaps {
		var platform []types.Platform
		json.Unmarshal([]byte(cmc.Platform), &platform)
		for _, p := range platform {
			if p.ContractAddress != "" {
				var address, chain string
				if strings.HasPrefix(p.ContractAddress, "0x") {
					address = strings.ToLower(p.ContractAddress)
				} else {
					address = p.ContractAddress
				}

				if value, ok := utils.CMCNameChainMap[p.Platform.Name]; ok {
					chain = value
				} else {
					chain = utils.GetPlatform(p.Platform.Coin.Slug)
				}
				key := chain + ":" + address
				if value, ok := tokenListMap[key]; ok {
					value.CmcId = cmc.Id
					value.Logo = cmc.Logo
					value.WebSite = cmc.WebSite
					count++
				} else {
					var decimals int
					if tokenInfo, tok := decimalsInfo[chain+":"+address]; tok {
						decimals = tokenInfo.Decimals
					}
					if decimals == 0 {
						noDecimals[chain] = append(noDecimals[chain], address)
					}
					tokenListMap[key] = &models.TokenList{
						CmcId:       cmc.Id,
						Name:        cmc.Name,
						Symbol:      strings.ToUpper(cmc.Symbol),
						Logo:        cmc.Logo,
						Description: cmc.Description,
						WebSite:     cmc.WebSite,
						Chain:       chain,
						Address:     address,
						Decimals:    decimals,
					}
				}
			}
		}
	}
	decimalsDBCount := 0
	for key, token := range decimalsInfo {
		if _, ok := tokenListMap[key]; !ok {
			decimalsDBCount++
			chain := strings.SplitN(key, ":", 2)[0]
			tokenListMap[key] = &models.TokenList{
				Name:     token.Name,
				Address:  token.Address,
				Chain:    chain,
				Symbol:   strings.ToUpper(token.Symbol),
				Logo:     token.LogoURI,
				Decimals: token.Decimals,
			}
		}
	}
	decimalsMap := utils.GetDecimalsByMap(noDecimals)
	tokenLists := make([]models.TokenList, 0, len(tokenListMap))
	decimalsCount := 0
	for _, value := range tokenListMap {
		if decimal, ok := decimalsMap[value.Chain+":"+value.Address]; ok && decimal > 0 {
			value.Decimals = decimal
			decimalsCount++
		}
		tokenLists = append(tokenLists, *value)
	}
	patchSize := 1000
	end := 0
	for i := 0; i < len(tokenLists); i += patchSize {
		if i+patchSize > len(tokenLists) {
			end = len(tokenLists)
		} else {
			end = i + patchSize
		}
		temp := tokenLists[i:end]
		result := c.db.Clauses(clause.OnConflict{
			UpdateAll: true,
		}).Create(&temp)
		if result.Error != nil {
			c.log.Error("create db error:", result.Error)
		}
	}

}

func GetTokenListPrice(chains, addresses []string, currency string) map[string]map[string]string {
	result := make(map[string]map[string]string)
	updateChains := make([]string, 0, len(addresses))
	createChains := make([]string, 0, len(addresses))

	//native price
	if len(chains) > 0 {
		var cmLists []models.CoinGecko
		chainMap := make(map[string][]string)
		for i := 0; i < len(chains); i++ {
			if chains[i] == "Huobi-Token" {
				chains[i] = "Huobi"
			}
		}
		err := c.db.Where("name in ?", chains).Find(&cmLists).Error
		if err != nil {
			c.log.Error("find cmList error:", err)
		}
		tempChain := make([]string, 0, len(chains))
		if len(cmLists) > 0 {
			tempChainMap := make(map[string]struct{}, len(cmLists))
			for _, cm := range cmLists {
				tempChain = append(tempChain, cm.Id)
				tempChainMap[cm.Name] = struct{}{}
				if cm.Name == "Huobi" {
					chainMap[cm.Id] = append(chainMap[cm.Id], "Huobi-Token")
				} else {
					chainMap[cm.Id] = append(chainMap[cm.Id], cm.Name)
				}
			}
			if len(chains) != len(cmLists) {
				for _, chain := range chains {
					if _, ok := tempChainMap[chain]; !ok {
						tempChain = append(tempChain, chain)
					}
				}
			}
		} else {
			tempChain = append(tempChain, chains...)
		}
		for _, chain := range tempChain {
			//get id by chain
			key := REDIS_PRICE_KEY + fmt.Sprintf("%s:%s", chain, strings.ToLower(currency))
			price, updateFlag, err := utils.GetPriceRedisValueByKey(c.redisClient, key)
			if err != nil {
				c.log.Error("get redis cache error:", err, key)
			}
			if updateFlag && price == "" {
				createChains = append(createChains, chain)
			} else if updateFlag {
				updateChains = append(updateChains, chain)
			}
			if price == "" {
				price = "0"
			}
			if value, ok := chainMap[chain]; ok {
				for _, addr := range value {
					result[addr] = map[string]string{
						currency: price,
					}
				}
			} else {
				result[chain] = map[string]string{
					currency: price,
				}
			}
		}
	}
	if len(addresses) > 0 {
		//token price
		newAddressMap := utils.ParseCoinAddress(addresses)
		for chainAddress, _ := range newAddressMap {
			key := REDIS_PRICE_KEY + fmt.Sprintf("%s:%s", chainAddress, strings.ToLower(currency))
			price, updateFlag, err := utils.GetPriceRedisValueByKey(c.redisClient, key)
			if err != nil {
				c.log.Error("get redis cache error:", err, key)
			}
			if updateFlag && price == "" {
				createChains = append(createChains, chainAddress)
			} else if updateFlag {
				updateChains = append(updateChains, chainAddress)
			}
			if price == "" {
				price = "0"

			}
			for _, addr := range newAddressMap[chainAddress] {
				result[addr] = map[string]string{currency: price}
			}
		}
	}
	if len(updateChains) > 0 {
		go func() {
			updatePriceChannel <- updateChains
		}()
	}

	if len(createChains) > 0 {
		go func() {
			createPriceChannel <- createChains
		}()
	}
	return result
}

func handlerPriceMap(priceMap map[string]map[string]float32, addressIdMap map[string]string, newAddressMap map[string][]string,
	currency string, result map[string]map[string]string, isCG bool) {
	for id, prices := range priceMap {
		var address string
		if value, ok := addressIdMap[id]; ok {
			address = value
		} else {
			address = id
		}
		var price string
		priceCurrency := currency
		if isCG {
			priceCurrency = strings.ToLower(currency)
		}
		if value, ok := prices[priceCurrency]; ok {
			price = decimal.NewFromFloat32(value).String()
		}
		key := REDIS_PRICE_KEY + fmt.Sprintf("%s:%s", address, strings.ToLower(currency))
		//err := c.redisClient.Set(key, price, 1*time.Minute).Err()
		err := utils.SetPriceRedisKey(c.redisClient, key, price)
		if err != nil {
			c.log.Error("set redisClient error:", err, key)
		}
		if value, ok := newAddressMap[address]; ok {
			for _, addr := range value {
				result[addr] = map[string]string{currency: price}
			}
		} else {
			result[address] = map[string]string{currency: price}
		}
	}
}

func GetTokenList(chain string) ([]*v1.GetTokenListResp_Data, error) {
	chain = utils.GetChainNameByChain(chain)
	key := REDIS_LIST_KEY + chain
	//str, err := c.redisClient.Get(key).Result()
	str, updateFlag, err := utils.GetTokenListRedisValueByKey(c.redisClient, key)
	if err != nil {
		c.log.Error("get token list key error:", err, key)
	}
	var result []*v1.GetTokenListResp_Data
	if str != "" {
		err := json.Unmarshal([]byte(str), &result)
		if err != nil {
			c.log.Error("unmarshal error:", err)
		}
	}

	if !updateFlag {
		return result, nil
	}

	var tokenLists []models.TokenList
	err = c.db.Where("chain = ?", chain).Find(&tokenLists).Error
	if err != nil {
		c.log.Error("get token list error", err)
	}
	result = make([]*v1.GetTokenListResp_Data, len(tokenLists))
	for index, t := range tokenLists {
		result[index] = &v1.GetTokenListResp_Data{
			Name:     t.Name,
			Symbol:   t.Symbol,
			Address:  t.Address,
			Decimals: uint32(t.Decimals),
			LogoURI:  c.logoPrefix + t.LogoURI,
		}
	}
	b, _ := json.Marshal(result)
	//err = c.redisClient.Set(key, string(b), 24*time.Hour).Err()
	err = utils.SetTokenListRedisKey(c.redisClient, key, string(b))
	if err != nil {
		c.log.Error("set redisClient cache error:", err)
	}
	return result, nil
}

func GetDecimalsInfo() map[string]types.TokenInfo {
	result := make(map[string]types.TokenInfo)
	tokens := utils.ParseTokenListFile()
	for chain, token := range tokens {
		for _, t := range token {
			if strings.HasPrefix(t.Address, "0x") && chain != utils.STARCOIN_CHAIN && chain != utils.APTOS_CHAIN {
				t.Address = strings.ToLower(t.Address)
			}
			result[chain+":"+t.Address] = t
		}
	}
	return result
}

func GetDBTokenInfo(addressInfos []*v1.GetTokenInfoReq_Data) ([]*v1.GetTokenInfoResp_Data, error) {
	tokenInfos := make([]*v1.GetTokenInfoResp_Data, 0, len(addressInfos))
	params := make([][]interface{}, 0, len(addressInfos))
	addressMap := make(map[string]string, len(addressInfos))
	for _, addressInfo := range addressInfos {
		key := addressInfo.Chain + ":" + addressInfo.Address
		chain := utils.GetChainNameByChain(addressInfo.Chain)
		address := utils.GetUnificationAddress(chain, addressInfo.Address)
		//if strings.HasPrefix(addressInfo.Address, "0x") && chain != utils.STARCOIN_CHAIN &&
		//	chain != utils.APTOS_CHAIN && !strings.Contains(chain, "SUI") {
		//	address = strings.ToLower(addressInfo.Address)
		//} else if (strings.Contains(chain, utils.COSMOS_CHAIN) || strings.Contains(chain, utils.OSMOSIS_CHAIN)) &&
		//	strings.Contains(address, "/") {
		//	address = "ibc/" + strings.ToUpper(strings.Split(address, "/")[1])
		//}
		params = append(params, []interface{}{chain, address})
		addressMap[chain+":"+address] = key
	}

	// get token list
	var tokenLists []models.TokenList
	err := c.db.Where("(chain, address) IN ?", params).Find(&tokenLists).Error
	if err != nil {
		c.log.Error("get token list error:", err)
	}
	for _, tokenList := range tokenLists {
		chain := tokenList.Chain
		address := tokenList.Address
		key := chain + ":" + address
		if value, ok := addressMap[key]; ok {
			addressInfo := strings.SplitN(value, ":", 2)
			chain, address = addressInfo[0], addressInfo[1]
		}
		tokenInfos = append(tokenInfos, &v1.GetTokenInfoResp_Data{
			Chain:    chain,
			Address:  address,
			Decimals: uint32(tokenList.Decimals),
			Symbol:   tokenList.Symbol,
			Name:     tokenList.Name,
			LogoURI:  c.logoPrefix + tokenList.LogoURI,
		})
	}
	return tokenInfos, nil
}

func GetTokenListByChainAddress(chain, address string) ([]models.TokenList, error) {
	address = utils.GetUnificationAddress(chain, address)
	var tokenLists []models.TokenList
	err := c.db.Where("chain = ? AND address = ?", chain, address).Find(&tokenLists).Error
	if err != nil {
		c.log.Error("get token list error:", err)
	}
	return tokenLists, err
}

func GetTokenInfo(addressInfos []*v1.GetTokenInfoReq_Data) ([]*v1.GetTokenInfoResp_Data, error) {
	tokenInfos := make([]*v1.GetTokenInfoResp_Data, 0, len(addressInfos))
	params := make([][]interface{}, 0, len(addressInfos))
	addressMap := make(map[string]string, len(addressInfos))
	for _, addressInfo := range addressInfos {
		key := addressInfo.Chain + ":" + addressInfo.Address
		tokenInfo := utils.GetRedisTokenInfo(c.redisClient, REDIS_TOKEN_KEY+key)
		if tokenInfo != nil {
			tokenInfos = append(tokenInfos, tokenInfo)
			continue
		}
		chain := utils.GetChainNameByChain(addressInfo.Chain)
		address := utils.GetUnificationAddress(chain, addressInfo.Address)

		//if strings.HasPrefix(addressInfo.Address, "0x") && chain != utils.STARCOIN_CHAIN &&
		//	chain != utils.APTOS_CHAIN && !strings.Contains(chain, "SUI") {
		//	address = strings.ToLower(addressInfo.Address)
		//} else if (strings.Contains(chain, utils.COSMOS_CHAIN) || strings.Contains(chain, utils.OSMOSIS_CHAIN)) &&
		//	strings.Contains(address, "/") {
		//	address = "ibc/" + strings.ToUpper(strings.Split(address, "/")[1])
		//}
		params = append(params, []interface{}{chain, address})
		addressMap[chain+":"+address] = key
	}
	// get token list
	var tokenLists []models.TokenList
	err := c.db.Where("(chain, address) IN ?", params).Find(&tokenLists).Error
	if err != nil {
		c.log.Error("get token list error:", err)
	}
	for _, tokenList := range tokenLists {
		chain := tokenList.Chain
		address := tokenList.Address
		key := chain + ":" + address
		if value, ok := addressMap[key]; ok {
			addressInfo := strings.SplitN(value, ":", 2)
			chain, address = addressInfo[0], addressInfo[1]
			delete(addressMap, key)
		}
		tokenInfos = append(tokenInfos, &v1.GetTokenInfoResp_Data{
			Chain:    chain,
			Address:  address,
			Decimals: uint32(tokenList.Decimals),
			Symbol:   tokenList.Symbol,
			Name:     tokenList.Name,
			LogoURI:  c.logoPrefix + tokenList.LogoURI,
		})
	}
	if len(addressMap) > 0 {
		for _, value := range addressMap {
			addressInfo := strings.SplitN(value, ":", 2)
			chain, address := addressInfo[0], addressInfo[1]
			tokenInfo, err := platform.GetPlatformTokenInfo(chain, address)
			if err != nil {
				c.log.Error("get platform token info error:", err)
				tokenInfos = append(tokenInfos, &v1.GetTokenInfoResp_Data{
					Chain:    chain,
					Address:  address,
					Symbol:   "Unknown Token",
					Decimals: 0,
				})

				alarmMsg := fmt.Sprintf("请注意：%s链查询代币信息失败，tokenAddress:%s\n错误消息：%s", chain, address, err)
				alarmOpts := lark.WithMsgLevel("FATAL")
				lark.LarkClient.NotifyLark(alarmMsg, alarmOpts)
				continue
			}
			if tokenInfo != nil {
				if err := utils.SetRedisTokenInfo(c.redisClient, REDIS_TOKEN_KEY+value, tokenInfo); err != nil {
					c.log.Error("set redis token info error.", err)
					continue
				}
				tokenInfos = append(tokenInfos, tokenInfo)
			}
		}
	}

	return tokenInfos, nil
}

const imageHex = "https://static.openblock.com/download/token/"

func InsertLogURI() {
	count := 0
	filepath.Walk("images", func(path string, info fs.FileInfo, err error) error {
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".png") {
			logoURI := imageHex + path
			fileName := info.Name()[:len(info.Name())-4]
			chainAddress := strings.Split(fileName, "_")
			if len(chainAddress) == 2 {
				chain := chainAddress[0]
				address := chainAddress[1]
				err := c.db.Model(&models.TokenList{}).Where("chain = ? AND address = ?", chain, address).Update("logo_uri", logoURI).Error
				if err != nil {
					log.Error("update error:", zap.Error(err), zap.Any("logoURI", logoURI), zap.Any("infoname", info.Name()))
				} else {
					count++
				}
			}

		}

		return nil
	})

	fmt.Println("count==", count)
}

func AutoUpdateCGTokenList(ids []string) {
	// get update cg platform
	var coinGeckos []models.CoinGecko
	if len(ids) > 0 {
		coinGeckos = GetCoinGeckoByIds(ids)
	} else {
		coinGeckos = UpdateCoinGecko()
	}
	// get db all token list
	tempDBTokenList, _ := GetAllTokenList()

	tempDBTokenListMap := make(map[string]struct{}, len(tempDBTokenList))
	//tokenListMap := make(map[string]*models.TokenList)
	tokenLists := make([]models.TokenList, 0, len(coinGeckos))

	for _, t := range tempDBTokenList {
		tempDBTokenListMap[t.Chain+"_"+t.Address] = struct{}{}
	}
	for _, cgc := range coinGeckos {
		var platform map[string]types.CGDetailPlatformInfo
		json.Unmarshal([]byte(cgc.Platform), &platform)
		for key, value := range platform {
			if key != "" && value.ContractAddress != "" && value.DecimalPlace >= 0 {
				if cgcValue, cgcOk := utils.CGCNameChainMap[key]; cgcOk {
					key = cgcValue
				}
				chain := utils.GetPlatform(key)
				if support := utils.IsNotSupportChain(chain); support {
					continue
				}
				var address string
				if strings.HasPrefix(value.ContractAddress, "0x") && chain != utils.STARCOIN_CHAIN &&
					chain != utils.APTOS_CHAIN {
					address = strings.ToLower(value.ContractAddress)
				} else {
					address = value.ContractAddress
				}
				if _, dbOk := tempDBTokenListMap[chain+"_"+address]; !dbOk {
					tokenLists = append(tokenLists, models.TokenList{
						CgId:        cgc.Id,
						Name:        cgc.Name,
						Symbol:      cgc.Symbol,
						WebSite:     cgc.Homepage,
						Description: cgc.Description,
						Chain:       chain,
						Address:     address,
						Logo:        cgc.Image,
						Decimals:    value.DecimalPlace,
					})
				}
			}
		}
	}
	c.log.Info("update tokenLists==", len(tokenLists))
	if len(tokenLists) > 0 {
		for i := 0; i < len(tokenLists); i++ {
			if tokenLists[i].Address != "" {
				tokenLists[i].Address = strings.Trim(tokenLists[i].Address, " ")
			}
		}
		updateTokenChainMap := make(map[string]struct{})
		for i := 0; i < len(tokenLists); i++ {
			updateTokenChainMap[tokenLists[i].Chain] = struct{}{}
			result := c.db.Clauses(clause.OnConflict{
				UpdateAll: true,
			}).Create(&tokenLists[i])
			if result.Error != nil {
				c.log.Error("create token list error:", result.Error)
			}
			c.log.Info("insert token list db length:", result.RowsAffected)
		}

		//download images
		DownLoadImages(tokenLists)

		//upload images
		UpLoadImages()

		//update logo uri
		InsertLogoURI()

		//delete images path
		err := os.RemoveAll("images")
		if err != nil {
			c.log.Error("delete images path:", err)
		}
		upLoadchains := make([]string, 0, len(tokenLists))
		for chain, _ := range updateTokenChainMap {
			upLoadchains = append(upLoadchains, chain)
		}
		//upload token list json to cdn
		if len(upLoadchains) > 0 {
			UpLoadJsonToCDN(upLoadchains)
		}
	}
}

// AutoUpdatePrice 更新代币的价格
func AutoUpdatePrice() {
	c.log.Infof("AutoUpdatePrice start")
	//获取chain
	chains := make([]string, len(c.chains))
	for index, chain := range c.chains {
		chains[index] = utils.GetChainNameByChain(chain)
	}
	UpdatePriceByChains(chains, utils.GetChainPriceKey())
	c.log.Infof("AutoUpdatePrice end")
}

func UpdatePriceByChains(chains, chainPriceKey []string) {
	var tokenLists []models.TokenList
	err := c.db.Where("chain in ? and (cg_id != ? or cmc_id > ?)", chains, "", 0).Find(&tokenLists).Error
	if err != nil {
		c.log.Error("AutoUpdatePrice get  tokenList error:", err)
		return
	}
	//chainPriceKey := utils.GetChainPriceKey()
	cgIds := make([]string, 0, len(tokenLists)+len(chainPriceKey))
	cmcIds := make([]string, 0, len(tokenLists))
	addressIdMap := make(map[string]string, len(tokenLists))
	cgIds = append(cgIds, chainPriceKey...)
	for _, token := range tokenLists {
		var id string
		if token.CgId != "" {
			id = token.CgId
			cgIds = append(cgIds, id)
		} else if token.CmcId > 0 {
			id = fmt.Sprintf("%d", token.CmcId)
			cmcIds = append(cmcIds, id)
		}
		if id != "" {
			if handler := utils.GetHandlerByDBName(token.Chain); handler != "" {
				addressIdMap[id] = fmt.Sprintf("%s_%s", handler, token.Address)
			}
		}
	}

	//get coin gecko price
	if len(cgIds) > 0 {
		handlerCgPrice(cgIds, addressIdMap)
	}

	if len(cmcIds) > 0 {
		handlerCMCPrice(cmcIds, addressIdMap)
	}
}

func handlerCMCPrice(cmcIds []string, addressIdMap map[string]string) {
	currencys := []string{"USD", "CNY"}
	for _, currency := range currencys {
		cmcPriceMap, err := CMCSimplePrice(cmcIds, currency)
		if err != nil {
			continue
		}
		setAutoPrice(cmcPriceMap, addressIdMap, false)
	}
}

func handlerPriceByToken(tokenAddress []*v12.Tokens, addressIdMap map[string]string) {
	c.log.Infof("handlerPriceByToken tokenAddress:", tokenAddress)
	marketPricesMap, err := utils3.GetPriceByMarketToken(tokenAddress)
	if err != nil {
		c.log.Error("get price by market token Error:", err)
		//lark
		alarmMsg := fmt.Sprintf("请注意：行情中心通过地址获取价格失败：%s\n tokens:%v", err, tokenAddress)
		alarmOpts := lark.WithMsgLevel("FATAL")
		lark.LarkClient.NotifyLark(alarmMsg, alarmOpts)
	}
	setAutoPrice(marketPricesMap, addressIdMap, false)
}

func handlerCgPrice(cgIds []string, addressIdMap map[string]string) {
	c.log.Infof("CGSimplePrice cgIds:", cgIds)
	// get price by market
	marketPricesMap, err := utils3.GetPriceByMarket(cgIds)
	if err != nil {
		c.log.Error("get price by market Error:", err)
		//lark
		alarmMsg := fmt.Sprintf("请注意：行情中心获取价格失败：%s", err)
		alarmOpts := lark.WithMsgLevel("FATAL")
		lark.LarkClient.NotifyLark(alarmMsg, alarmOpts)
		// get price by coin gecko
		pageSize := 500
		pageEndIndex := 0
		for i := 0; i < len(cgIds); i += pageSize {
			if i+pageSize > len(cgIds) {
				pageEndIndex = len(cgIds)
			} else {
				pageEndIndex = i + pageSize
			}
			cgPricesMap, err := CGSimplePrice(cgIds[i:pageEndIndex], "USD,CNY")
			for j := 0; err != nil && j < 3; j++ {
				time.Sleep(1 * time.Minute)
				cgPricesMap, err = CGSimplePrice(cgIds[i:pageEndIndex], "USD,CNY")
			}
			if err != nil {
				c.log.Error("CGSimplePrice Error:", err)
				continue
			}
			setAutoPrice(cgPricesMap, addressIdMap, true)
		}
		return
	}
	c.log.Infof("marketPricesMap length:", len(marketPricesMap), marketPricesMap, addressIdMap)
	setAutoPrice(marketPricesMap, addressIdMap, true)
}

func setAutoPrice(priceMap map[string]map[string]float64, addressIdMap map[string]string, isCG bool) {
	for id, prices := range priceMap {
		var address string
		if value, ok := addressIdMap[id]; ok {
			address = value
		} else {
			address = id
		}
		for currency, price := range prices {
			priceStr := decimal.NewFromFloat(price).String()
			key := REDIS_PRICE_KEY + fmt.Sprintf("%s:%s", address, strings.ToLower(currency))
			err := utils.SetPriceRedisKey(c.redisClient, key, priceStr)
			if err != nil {
				c.log.Error("set redisClient error:", err, key)
			}
		}
	}
}

func UpdateTokenListByMarket() {
	tokenLists := make([]models.TokenList, 0, 50)
	for chain, dbChain := range utils.UpdateChainNameMap {
		//没有测试网
		if !strings.Contains(chain, "TEST") {
			//获取当前链token list
			tempDBTokenList, err := GetCGTokenListByChain(dbChain)
			if err != nil {
				c.log.Error("get db token list error:", err, ",chain=", dbChain)
				return
			}
			tempDBTokenListMap := make(map[string]struct{}, len(tempDBTokenList))
			for _, t := range tempDBTokenList {
				dbKey := t.Chain + "_*" + t.Address
				tempDBTokenListMap[dbKey] = struct{}{}
			}
			replay, err := utils3.GetCoinInfoByMarket(chain)
			if err != nil || replay == nil {
				c.log.Error("get coin info error:", err, ",chain=", chain)
				continue
			}
			count := 0
			for _, coinInfo := range replay {
				homePageByte, _ := json.Marshal(coinInfo.Homepage)
				dbKey := dbChain + "_*" + coinInfo.Address
				//fmt.Println("zql==1=dbKey=", dbKey)
				if _, dbOk := tempDBTokenListMap[dbKey]; !dbOk {
					count++
					tokenLists = append(tokenLists, models.TokenList{
						CgId:        coinInfo.CoinID,
						Name:        coinInfo.Name,
						Symbol:      strings.ToUpper(coinInfo.Symbol),
						WebSite:     string(homePageByte),
						Description: coinInfo.Description,
						Chain:       dbChain,
						Address:     coinInfo.Address,
						Logo:        coinInfo.Icon,
						Decimals:    int(coinInfo.DecimalPlace),
					})
				}
			}
		}
	}
	if len(tokenLists) > 0 {
		for _, tokenList := range tokenLists {
			if err := c.db.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "address"}, {Name: "chain"}},
				UpdateAll: true,
			}).Create(&tokenList).Error; err != nil {
				c.log.Error("UpdateTokenListByMarket insert db error ", err)
			}
		}
		//download images
		DownLoadImages(tokenLists)

		//upload images
		UpLoadImages()

		//update logo uri
		InsertLogoURI()

		//delete images path
		err := os.RemoveAll("images")
		if err != nil {
			c.log.Error("delete images path:", err)
		}
	}

	c.log.Info("UpdateTokenListByMarket End")
}

func AutoUpdateTokenList(cmcFlag, cgFlag, jsonFlag bool) {
	var wg sync.WaitGroup
	var coinMarketCaps []models.CoinMarketCap
	var coinGeckos []models.CoinGecko
	var tempDBTokenList []models.TokenList
	var decimalsInfo map[string]types.TokenInfo
	if cmcFlag {
		wg.Add(1)
		go func() {
			defer wg.Done()
			coinMarketCaps = UpdateCoinMarketCap()
		}()
	}
	if cgFlag {
		wg.Add(1)
		go func() {
			defer wg.Done()
			coinGeckos = UpdateCoinGecko()
		}()
	}
	if jsonFlag {
		wg.Add(1)
		go func() {
			defer wg.Done()
			decimalsInfo = GetDecimalsInfo()
		}()
	}
	tempDBTokenList, _ = GetAllTokenList()
	wg.Wait()
	tempDBTokenListMap := make(map[string]struct{}, len(tempDBTokenList))
	tokenListMap := make(map[string]*models.TokenList)
	noDecimals := make(map[string][]string)
	for _, t := range tempDBTokenList {
		tempDBTokenListMap[t.Chain+":"+t.Address] = struct{}{}
	}
	for _, cgc := range coinGeckos {
		var platform map[string]interface{}
		json.Unmarshal([]byte(cgc.Platform), &platform)
		for key, value := range platform {
			if key != "" && value != "" {
				if cgcValue, cgcOk := utils.CGCNameChainMap[key]; cgcOk {
					key = cgcValue
				}
				chain := utils.GetPlatform(key)
				var address string
				if strings.HasPrefix(value.(string), "0x") && chain != utils.STARCOIN_CHAIN {
					address = strings.ToLower(value.(string))
				} else {
					address = value.(string)
				}
				dbKey := chain + ":" + address
				if _, dbOk := tempDBTokenListMap[dbKey]; !dbOk {
					var decimals int
					if tokenInfo, ok := decimalsInfo[key]; ok {
						decimals = tokenInfo.Decimals
					}

					if decimals == 0 {
						noDecimals[chain] = append(noDecimals[chain], address)
					}
					tokenListMap[chain+":"+address] = &models.TokenList{
						CgId:        cgc.Id,
						Name:        cgc.Name,
						Symbol:      strings.ToUpper(cgc.Symbol),
						WebSite:     cgc.Homepage,
						Description: cgc.Description,
						Chain:       chain,
						Address:     address,
						Logo:        cgc.Image,
						Decimals:    decimals,
					}
				}

			}
		}
	}

	for _, cmc := range coinMarketCaps {
		var platform []types.Platform
		json.Unmarshal([]byte(cmc.Platform), &platform)
		for _, p := range platform {
			if p.ContractAddress != "" {
				var address, chain string
				if strings.HasPrefix(p.ContractAddress, "0x") && chain != utils.STARCOIN_CHAIN {
					address = strings.ToLower(p.ContractAddress)
				} else {
					address = p.ContractAddress
				}
				if value, ok := utils.CMCNameChainMap[p.Platform.Name]; ok {
					chain = value
				} else {
					chain = utils.GetPlatform(p.Platform.Coin.Slug)
				}
				key := chain + ":" + address
				if _, dbOk := tempDBTokenListMap[key]; !dbOk {
					if value, ok := tokenListMap[key]; ok {
						value.CmcId = cmc.Id
						value.Logo = cmc.Logo
						value.WebSite = cmc.WebSite
						//count++
					} else {
						var decimals int
						if tokenInfo, tok := decimalsInfo[chain+":"+address]; tok {
							decimals = tokenInfo.Decimals
						}
						if decimals == 0 {
							noDecimals[chain] = append(noDecimals[chain], address)
						}
						tokenListMap[key] = &models.TokenList{
							CmcId:       cmc.Id,
							Name:        cmc.Name,
							Symbol:      strings.ToUpper(cmc.Symbol),
							Logo:        cmc.Logo,
							Description: cmc.Description,
							WebSite:     cmc.WebSite,
							Chain:       chain,
							Address:     address,
							Decimals:    decimals,
						}
					}
				}
			}
		}
	}

	for key, token := range decimalsInfo {
		if _, dbOk := tempDBTokenListMap[key]; !dbOk {
			if _, ok := tokenListMap[key]; !ok {
				chain := strings.SplitN(key, ":", 2)[0]
				tokenListMap[key] = &models.TokenList{
					Name:     token.Name,
					Address:  token.Address,
					Chain:    chain,
					Symbol:   strings.ToUpper(token.Symbol),
					Logo:     token.LogoURI,
					Decimals: token.Decimals,
				}
			}
		}
	}

	decimalsMap := utils.GetDecimalsByMap(noDecimals)
	tokenLists := make([]models.TokenList, 0, len(tokenListMap))
	updateTokenListMap := make(map[string]struct{})
	for _, value := range tokenListMap {
		if decimal, ok := decimalsMap[value.Chain+":"+value.Address]; ok && decimal > 0 {
			value.Decimals = decimal

		}
		updateTokenListMap[value.Chain] = struct{}{}
		tokenLists = append(tokenLists, *value)
	}
	if len(tokenLists) > 0 {
		for i := 0; i < len(tokenLists); i++ {
			if tokenLists[i].Address != "" {
				tokenLists[i].Address = strings.Trim(tokenLists[i].Address, " ")
			}
		}

		for i := 0; i < len(tokenLists); i++ {
			result := c.db.Clauses(clause.OnConflict{
				UpdateAll: true,
			}).Create(&tokenLists[i])
			if result.Error != nil {
				c.log.Error("create token list error:", result.Error)
			}
			c.log.Info("insert token list db length:", result.RowsAffected)
		}

		//download images
		DownLoadImages(tokenLists)

		//upload images
		UpLoadImages()

		//update logo uri
		InsertLogoURI()

		//delete images path
		err := os.RemoveAll("images")
		if err != nil {
			c.log.Error("delete images path:", err)
		}
		upLoadchains := make([]string, 0, len(updateTokenListMap))
		for chain, _ := range updateTokenListMap {
			upLoadchains = append(upLoadchains, chain)
		}
		//upload token list json to cdn
		if len(upLoadchains) > 0 {
			UpLoadJsonToCDN(upLoadchains)
		}
	}

}

func RefreshLogoURIByAddress(chain string, addresses []string) {
	var tokenLists []models.TokenList
	err := c.db.Where("chain = ? AND address in ?", chain, addresses).Find(&tokenLists).Error
	if err != nil {
		c.log.Error("get token list error:", err)
		return
	}
	fmt.Println("tokenList length=", len(tokenLists))
	//download images
	DownLoadImages(tokenLists)

	//upload images
	UpLoadImages()

	//update logo uri
	InsertLogoURI()

	//delete images path
	err = os.RemoveAll("./images")
	if err != nil {
		c.log.Error("delete images path:", err)
	}
}

func RefreshLogoURI(chains []string) {
	//get all token list
	var tokenLists []models.TokenList
	var err error
	if len(chains) > 0 {
		err = c.db.Where("chain in ? AND logo_uri = ?", chains, "").Find(&tokenLists).Error
	} else {
		tokenLists, err = GetAllTokenList()
	}
	//err := c.db.Where("address = ?", "0x75231f58b43240c9718dd58b4967c5114342a86c").Find(&tokenLists).Error
	if err != nil {
		c.log.Error("get token list error:", err)
		return
	}
	fmt.Println("tokenList length=", len(tokenLists))
	//download images
	DownLoadImages(tokenLists)

	//upload images
	UpLoadImages()

	//update logo uri
	InsertLogoURI()

	//delete images path
	err = os.RemoveAll("./images")
	if err != nil {
		c.log.Error("delete images path:", err)
	}
}

func DeleteJsonCDN() {
	chainVersionMap := utils.GetCDNTokenList(c.logoPrefix + "tokenlist/tokenlist.json")
	newChainVersionMap := make(map[string]types.TokenInfoVersion, len(chainVersionMap))
	// 验证已上线的链
	for _, chain := range c.chains {
		if value, ok := chainVersionMap[chain]; ok {
			newChainVersionMap[chain] = value
		}
	}
	if len(newChainVersionMap) == len(chainVersionMap) {
		return
	}
	path := "tokenlist"
	exist, _ := utils.PathExists(path)
	if exist {
		err := os.RemoveAll(path)
		if err != nil {
			c.log.Error("RemoveAll error:", err)
			return
		}
	}
	os.MkdirAll(path, 0777)
	writeVersionInfo := make([]types.TokenInfoVersion, 0, len(chainVersionMap))
	for _, value := range newChainVersionMap {
		writeVersionInfo = append(writeVersionInfo, value)
	}
	c.log.Info("json info:", writeVersionInfo)
	err := utils.WriteJsonToFile(path+"/tokenlist.json", writeVersionInfo)
	if err != nil {
		c.log.Error("write json to file error:", err)
	}
	UpLoadToken()
	//删除目录
	err = os.RemoveAll(path)
	if err != nil {
		fmt.Println("error:", err)
	}
}

func UpLoadJsonToCDN(chains []string) {
	chainVersionMap := utils.GetCDNTokenList(c.logoPrefix + "tokenlist/tokenlist.json")
	// 验证已上线的链
	for _, chain := range c.chains {
		if _, ok := chainVersionMap[chain]; !ok {
			chains = append(chains, utils.GetChainNameByChain(chain))
		}
	}

	var tokenLists []models.TokenList
	var err error
	if len(chains) > 0 {
		err = c.db.Where("chain in ?", chains).Find(&tokenLists).Error
	} else {
		err = c.db.Find(&tokenLists).Error
	}
	if err != nil {
		c.log.Error("get token list error:", err)
		return
	}
	chainTokenList := make(map[string][]types.TokenInfo, len(tokenLists))
	for _, t := range tokenLists {
		if chain := utils.GetChainByDBChain(t.Chain); chain != "" {
			tokenInfo := types.TokenInfo{
				ChainId:  0,
				Address:  t.Address,
				Symbol:   t.Symbol,
				Decimals: t.Decimals,
				Name:     t.Name,
				LogoURI:  c.logoPrefix + t.LogoURI,
			}
			chainTokenList[chain] = append(chainTokenList[chain], tokenInfo)
		}
	}

	path := "tokenlist"
	exist, _ := utils.PathExists(path)
	if exist {
		err = os.RemoveAll(path)
	}
	os.MkdirAll(path, 0777)
	//tokenList.json

	tokenVersions := make([]types.TokenInfoVersion, 0, len(chainTokenList))
	for chain, tokenInfo := range chainTokenList {
		fileName := fmt.Sprintf("%s/%s.json", path, chain)
		file, err := os.OpenFile(fileName, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0777)
		if err != nil {
			c.log.Error("open file error:", err)
		}
		encoder := json.NewEncoder(file)
		err = encoder.Encode(tokenInfo)

		if err != nil {
			c.log.Error("编码错误", err.Error())
		}
		tokenVersions = append(tokenVersions, types.TokenInfoVersion{
			Chain:   chain,
			URL:     c.logoPrefix + fileName,
			Version: time.Now().Unix(),
		})
	}

	for _, info := range tokenVersions {
		chainVersionMap[info.Chain] = info
	}
	writeVersionInfo := make([]types.TokenInfoVersion, 0, len(chainVersionMap))
	for _, value := range chainVersionMap {
		writeVersionInfo = append(writeVersionInfo, value)
	}
	c.log.Info("json info:", writeVersionInfo)
	err = utils.WriteJsonToFile(path+"/tokenlist.json", writeVersionInfo)
	if err != nil {
		c.log.Error("write json to file error:", err)
	}
	UpLoadToken()
	//删除目录
	err = os.RemoveAll(path)
	if err != nil {
		fmt.Println("error:", err)
	}
}

func RefreshCDNDirs() {
	mac := qbox.NewMac(c.qiniu.AccessKey, c.qiniu.SecretKey)
	cdnManager := cdn.NewCdnManager(mac)
	_, err := cdnManager.RefreshDirs([]string{c.logoPrefix, c.awsLogoPrefie})
	if err != nil {
		c.log.Error("fetch dirs error:", err)
	}

	if c.aws != nil {
		for _, awsInfo := range c.aws {
			sess, err := session.NewSession(&aws.Config{
				Region: aws.String(awsInfo.Region), //桶所在的区域
				Credentials: credentials.NewStaticCredentials(
					awsInfo.AccessKey, // accessKey
					awsInfo.SecretKey, // secretKey
					""),               //sts的临时凭证
			})
			if err != nil {
				c.log.Error("new session error:", err)
				return
			}

			//refresh dir
			callerReference := time.Now().String()
			svc := cloudfront.New(sess)
			paths := &cloudfront.Paths{
				Items:    []*string{aws.String(awsInfo.KeyPrefix + "*")},
				Quantity: aws.Int64(1),
			}

			input := &cloudfront.CreateInvalidationInput{DistributionId: aws.String(awsInfo.DistributionId),
				InvalidationBatch: &cloudfront.InvalidationBatch{
					CallerReference: aws.String(callerReference),
					Paths:           paths,
				}}
			ret, err := svc.CreateInvalidation(input)
			if err != nil {
				c.log.Error("create invalidation error:", err)
			}
			c.log.Info("s3 refresh dir:", ret)
		}
	}
}

func UpLoadToken() {
	mac := qbox.NewMac(c.qiniu.AccessKey, c.qiniu.SecretKey)
	cdnManager := cdn.NewCdnManager(mac)
	var paths []string
	filepath.Walk("tokenlist", func(path string, info fs.FileInfo, err error) error {
		if !info.IsDir() {
			paths = append(paths, path)
		}
		return nil
	})
	for _, bucket := range c.qiniu.Bucket {
		//upToken := putPolicy.UploadToken(mac)
		cfg := storage.Config{
			UseHTTPS: true,
		}
		//bucketManager := storage.NewBucketManager(mac, &cfg)
		formUploader := storage.NewFormUploader(&cfg)
		ret := types.MyPutRet{}
		putExtra := storage.PutExtra{
			Params: map[string]string{
				"x:name": "github logo",
			},
		}
		for _, path := range paths {
			key := c.qiniu.KeyPrefix + path
			putPolicy := storage.PutPolicy{
				Scope: fmt.Sprintf("%s:%s", bucket, key),
			}
			upToken := putPolicy.UploadToken(mac)
			err := formUploader.PutFile(context.Background(), &ret, upToken, key, path, &putExtra)
			if err != nil {
				c.log.Error("PutFile Error:", err)
			}
			c.log.Info("upload info:", ret.Bucket, ret.Key, ret.Fsize, ret.Hash, ret.Name)
		}
	}

	_, err := cdnManager.RefreshDirs([]string{c.logoPrefix, c.awsLogoPrefie})
	if err != nil {
		c.log.Error("fetch dirs error:", err)
	}

	//upload file to s3
	UploadFileToS3(paths)
}

func DownLoadImages(tokenLists []models.TokenList) {
	var wg sync.WaitGroup
	var count int32 = 0
	var noCount int32 = 0
	c.log.Info("DownLoadImages start:length", len(tokenLists))
	for _, t := range tokenLists {
		var image string
		if t.CgId != "" && t.CmcId == 0 {
			var cgImage map[string]string
			if err := json.Unmarshal([]byte(t.Logo), &cgImage); err != nil {
				image = t.Logo
			} else {
				if value, ok := cgImage["large"]; ok {
					image = value
				} else {
					if strings.Contains(t.Logo, "/thumb/") {
						image = strings.Replace(t.Logo, "/thumb/", "/large/", 1)
					}
				}
			}
		} else {
			if strings.Contains(t.Logo, "/thumb/") {
				image = strings.Replace(t.Logo, "/thumb/", "/large/", 1)
			} else {
				image = t.Logo
			}
		}

		if image != "" && strings.HasPrefix(image, "https") {
			path := "./images/" + t.Chain
			exist, _ := utils.PathExists(path)
			if !exist {
				os.MkdirAll(path, 0777)
			}
			fileSuffix := ".png"
			if strings.Contains(image, ".svg") {
				fileSuffix = ".svg"
			}
			address := t.Address
			if (t.Chain == "osmosis" || t.Chain == "cosmos") && strings.Contains(t.Address, "/") {
				address = strings.Split(t.Address, "/")[1]
			} else if t.Chain == "Sei" && strings.Contains(t.Address, "/") {
				address = strings.Replace(t.Address, "/", "_", -1)
			}
			fileName := path + "/" + t.Chain + "_" + address + fileSuffix
			wg.Add(1)
			go func(f, i string) {
				defer wg.Done()
				err := utils.DownLoad(fileName, image)
				for j := 0; j < 3 && err != nil; j++ {
					time.Sleep(time.Duration(j) * time.Second)
					err = utils.DownLoad(fileName, image)
				}
				if err == nil {
					atomic.AddInt32(&count, 1)
				} else {
					c.log.Error("download error:", err, ",fileName:", f, ",image:", i)
				}
			}(fileName, image)

		} else {
			atomic.AddInt32(&noCount, 1)
		}

	}
	wg.Wait()
	c.log.Info("DownLoadImages End count:", count, noCount)
}

func UpLoadLocalImages(localFile string) {
	mac := qbox.NewMac(c.qiniu.AccessKey, c.qiniu.SecretKey)
	cdnManager := cdn.NewCdnManager(mac)
	for _, bucket := range c.qiniu.Bucket {
		cfg := storage.Config{
			UseHTTPS: true,
		}
		formUploader := storage.NewFormUploader(&cfg)
		ret := types.MyPutRet{}
		putExtra := storage.PutExtra{
			Params: map[string]string{
				"x:name": "github logo",
			},
		}

		key := c.qiniu.KeyPrefix + localFile
		putPolicy := storage.PutPolicy{
			Scope: fmt.Sprintf("%s:%s", bucket, key),
		}
		upToken := putPolicy.UploadToken(mac)
		err := formUploader.PutFile(context.Background(), &ret, upToken, key, localFile, &putExtra)
		if err != nil {
			c.log.Error("PutFile Error:", err)
		}
		c.log.Info("upload info:", ret.Bucket, ret.Key, ret.Fsize, ret.Hash, ret.Name)

	}

	_, err := cdnManager.RefreshDirs([]string{c.logoPrefix, c.awsLogoPrefie})
	if err != nil {
		c.log.Error("fetch dirs error:", err)
	}

	//upload file to s3
	UploadFileToS3([]string{localFile})

}

func UpLoadImages() {
	mac := qbox.NewMac(c.qiniu.AccessKey, c.qiniu.SecretKey)
	cdnManager := cdn.NewCdnManager(mac)
	var paths []string
	//var wg sync.WaitGroup
	filepath.Walk("images", func(path string, info fs.FileInfo, err error) error {
		if !info.IsDir() {
			paths = append(paths, path)
		}
		return nil
	})
	for _, bucket := range c.qiniu.Bucket {
		cfg := storage.Config{
			UseHTTPS: true,
		}
		formUploader := storage.NewFormUploader(&cfg)
		ret := types.MyPutRet{}
		putExtra := storage.PutExtra{
			Params: map[string]string{
				"x:name": "github logo",
			},
		}
		for _, path := range paths {
			//wg.Add(1)
			//go func(path string) {
			//	defer wg.Done()
			key := c.qiniu.KeyPrefix + path
			putPolicy := storage.PutPolicy{
				Scope: fmt.Sprintf("%s:%s", bucket, key),
			}

			upToken := putPolicy.UploadToken(mac)
			err := formUploader.PutFile(context.Background(), &ret, upToken, key, path, &putExtra)
			if err != nil {
				c.log.Error("PutFile Error:", err)
			}
			//c.log.Info("upload info:", ret.Bucket, ret.Key, ret.Fsize, ret.Hash, ret.Name)
			//}(p)
		}
	}
	//wg.Wait()
	_, err := cdnManager.RefreshDirs([]string{c.logoPrefix, c.awsLogoPrefie})
	if err != nil {
		c.log.Error("fetch dirs error:", err)
	}

	//upload file to s3
	UploadFileToS3(paths)

}

func InsertLogoURI() {
	count := 0
	filepath.Walk("images", func(path string, info fs.FileInfo, err error) error {
		if !info.IsDir() {
			logoURI := path
			fileName := info.Name()[:len(info.Name())-4]
			chainAddress := strings.SplitN(fileName, "_", 2)
			if len(chainAddress) == 2 {
				chain := chainAddress[0]
				address := chainAddress[1]
				if (chain == "osmosis" || chain == "cosmos") && len(address) == 64 {
					address = fmt.Sprintf("ibc/%s", address)
				} else if chain == "Sei" {
					address = strings.Replace(address, "_", "/", -1)
				}
				err := c.db.Model(&models.TokenList{}).Where("chain = ? AND address = ?", chain, address).Update("logo_uri", logoURI).Error
				if err != nil {
					c.log.Error("update token logo uri error:", err, ",logoURI:", logoURI, ",infoName:", info.Name())
				} else {
					count++
				}
			}

		}

		return nil
	})
	c.log.Info("InsertLogoURI count:", count)
}

func GetAllTokenList() ([]models.TokenList, error) {
	var tokenList []models.TokenList
	err := c.db.Find(&tokenList).Error

	return tokenList, err
}

func UpdateEVMDecimasl(chain string) {
	var tokenLists []models.TokenList
	err := c.db.Where("chain = ?", chain).Find(&tokenLists).Error
	if err != nil {
		c.log.Error("find token list error:", err)
	}
	fmt.Println("tokenList==", chain, len(tokenLists))
	//var wg sync.WaitGroup
	addresses := make([]string, 0, len(tokenLists))
	for _, t := range tokenLists {
		addresses = append(addresses, t.Address)
	}
	decimalMap := map[string][]string{
		chain: addresses,
	}
	result := utils.GetDecimalsByMap(decimalMap)
	for _, t := range tokenLists {
		if decimal, ok := result[t.Chain+":"+t.Address]; ok && decimal > 0 {
			//update decimal
			err = c.db.Model(&models.TokenList{}).Where("id = ?", t.ID).Update("decimals", decimal).Error
			if err != nil {
				c.log.Error("update token list decimal error:", t.ID, decimal, err)
			}
		}
	}
}

func UpdateDecimalsByChain(chain string) {
	var tokenLists []models.TokenList
	err := c.db.Where("chain = ? and decimals = ?", chain, "0").Find(&tokenLists).Error
	if err != nil {
		c.log.Error("find token list error:", err)
	}
	fmt.Println("token list lengt==", len(tokenLists))
	for _, t := range tokenLists {
		if t.Decimals == 0 {
			decimal, err := utils.GetDecimalsByChain(chain, t.Address)
			if err != nil {
				time.Sleep(1 * time.Minute)
				for i := 0; err != nil && strings.Contains(err.Error(), "Too many requests") && i < 3; i++ {
					decimal, err = utils.GetDecimalsByChain(chain, t.Address)
					time.Sleep(1 * time.Minute)
				}
			}
			err = c.db.Model(&models.TokenList{}).Where("id = ?", t.ID).Update("decimals", decimal).Error
			if err != nil {
				c.log.Error("update token list decimal error:", t.ID, decimal, err)
			}
		}
	}
}

func UploadFileToS3(localFiles []string) {
	if c.aws != nil {
		var wg sync.WaitGroup
		for _, awsInfo := range c.aws {
			sess, err := session.NewSession(&aws.Config{
				Region: aws.String(awsInfo.Region), //桶所在的区域
				Credentials: credentials.NewStaticCredentials(
					awsInfo.AccessKey, // accessKey
					awsInfo.SecretKey, // secretKey
					""),               //sts的临时凭证
			})
			if err != nil {
				c.log.Error("new session error:", err)
				return
			}
			//upload file
			for _, l := range localFiles {
				wg.Add(1)
				go func(localFile string) {
					defer wg.Done()
					exist, _ := utils.PathExists(localFile)
					if !exist {
						return
					}

					buffer, _ := os.ReadFile(localFile)
					key := awsInfo.KeyPrefix + localFile
					//fmt.Println("tokenList==", string(buffer))
					_, err := s3.New(sess).PutObject(&s3.PutObjectInput{
						Bucket: aws.String(awsInfo.Bucket), //桶名
						Key:    aws.String(key),
						Body:   bytes.NewReader(buffer),
					})
					if err != nil {
						c.log.Error("put file to s3 error:", err)
					}
					//c.log.Info("upload s3 info:", ret)
				}(l)
			}
			wg.Wait()
			//refresh dir
			callerReference := time.Now().String()
			svc := cloudfront.New(sess)
			paths := &cloudfront.Paths{
				Items:    []*string{aws.String(awsInfo.KeyPrefix + "*")},
				Quantity: aws.Int64(1),
			}

			input := &cloudfront.CreateInvalidationInput{DistributionId: aws.String(awsInfo.DistributionId),
				InvalidationBatch: &cloudfront.InvalidationBatch{
					CallerReference: aws.String(callerReference),
					Paths:           paths,
				}}
			ret, err := svc.CreateInvalidation(input)
			if err != nil {
				c.log.Error("create invalidation error:", err)
			}
			c.log.Info("s3 refresh dir:", ret)

		}
	}
}

func UpdateChainToken(chain string) {
	switch chain {
	case "nervos":
		UpdateNervosToken()
	case "aptos":
		UpdateAptosToken()
	case "cosmos":
		UpdateCosmosToken()
	case "osmosis":
		UpdateOsmosisToken()
	case "arbitrum-nova":
		UpdateArbitrumNovaToken()
	case "conflux":
		UpdateConfluxToken()
	case "binance-smart-chain":
		UpdateBSCToken()
	case "zkSync":
		UpdateZkSyncToken()
	case "SUITEST":
		UpdateSUITESTToken()
	case "SUI":
		UpdateSUIToken()
	case "evm210425":
		Updateevm210425Token()
	case "Linea":
		UpdateLineaToken()
	case "evm8453":
		UpdateEvm8453Token()
	case "Sei":
		UpdateSEIToken()
	case "arbitrum-one":
		UpdateArbitrumToken()
	case "ethereum":
		UpdateEthereumToken()

		//default:
		//	utils.GetCDNTokenList(c.logoPrefix + "tokenlist/tokenlist.json")
	}
}

func UpdateAptosToken() {
	var tokenLists = []models.TokenList{{
		Address:  "0x5e156f1207d0ebfa19a9eeff00d62a282278fb8719f4fab3a586a0a2c0fffbea::coin::T",
		Decimals: 6,
		Name:     "USD Coin (eth)",
		Symbol:   "USDC (eth)",
		Chain:    "aptos",
	}, {
		Address:  "0xdd89c0e695df0692205912fb69fc290418bed0dbe6e4573d744a6d5e6bab6c13::coin::T",
		Decimals: 8,
		Name:     "Wrapped SOL",
		Symbol:   "SOL",
		Chain:    "aptos",
	}, {
		Address:  "0xae478ff7d83ed072dbc5e264250e67ef58f57c99d89b447efd8a0a2e8b2be76e::coin::T",
		Decimals: 8,
		Name:     "Wrapped BTC",
		Symbol:   "WBTC",
		Chain:    "aptos",
	}, {
		Address:  "0x1000000fa32d122c18a6a31c009ce5e71674f22d06a581bb0a15575e6addadcc::usda::USDA",
		Decimals: 6,
		Name:     "Argo USD",
		Symbol:   "USDA",
		Chain:    "aptos",
	}, {
		Address:  "0xc91d826e29a3183eb3b6f6aa3a722089fdffb8e9642b94c5fcd4c48d035c0080::coin::T",
		Decimals: 6,
		Name:     "USD Coin (sol)",
		Symbol:   "USDC (sol)",
		Chain:    "aptos",
	},
	}
	result := c.db.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(&tokenLists)
	if result.Error != nil {
		c.log.Error("create db aptos error:", result.Error)
	}

}

func GetTopNTokenList(chain string, topN int) ([]*v1.TokenInfoData, error) {
	//oldChain := chain
	key := REDIS_TOP20_KEY + chain
	if topN == 50 {
		key = REDIS_TOP50_KEY + chain
	}
	result, err := getTopTokenByKey(chain, key, topN)
	if err != nil || len(result) == 0 {
		if topN == 50 {
			key = REDIS_TOP20_KEY + chain
			result, err = getTopTokenByKey(chain, key, topN)
		}
	}
	return result, err
}

func getTopTokenByKey(chain, key string, topN int) ([]*v1.TokenInfoData, error) {
	oldChain := chain
	//key := REDIS_TOP20_KEY + chain
	//get chain
	chain = utils.GetChainNameByChain(chain)
	//get redis cache
	var result []*v1.TokenInfoData
	str, updateFlag, err := utils.GetTokenTop20RedisValueByKey(c.redisClient, key)
	if err != nil {
		c.log.Error("get token top29 key error:", err, key)
	}
	if str != "" {
		if err := json.Unmarshal([]byte(str), &result); err != nil {
			c.log.Error("unmarshal error:", err)
		}
	}
	if updateFlag && !strings.Contains(chain, "TEST") {
		go func() {
			var tokenLists []models.TokenList
			err = c.db.Where("chain = ? AND name != ? AND cg_id != ?", chain, oldChain, "").Find(&tokenLists).Error
			if err != nil {
				c.log.Error("get token list error:", err)
			}
			c.log.Info("token list length=", len(tokenLists))
			if len(tokenLists) >= topN {
				tokenInfos := make(map[string]*v1.TokenInfoData, len(tokenLists))
				ids := make([]string, len(tokenLists))
				for i, t := range tokenLists {
					ids[i] = t.CgId
					tokenInfos[t.CgId] = &v1.TokenInfoData{
						Chain:    oldChain,
						Address:  t.Address,
						Symbol:   t.Symbol,
						Decimals: uint32(t.Decimals),
						Name:     t.Name,
						LogoURI:  c.logoPrefix + t.LogoURI,
					}
				}
				markets := make([]types.CGMarket, 0, len(tokenLists)+2)
				pageSize := 500
				endIndex := 0
				for i := 0; i < len(ids); i += pageSize {
					if i+pageSize > len(ids) {
						endIndex = len(ids)
					} else {
						endIndex = i + pageSize
					}
					cgMarkets, err := GetCGMarkets(ids[i:endIndex], "CNY", topN)
					if err != nil {
						c.log.Error("get cg markets error:", err)
					}
					for j := 0; err != nil && j < 3; j++ {
						time.Sleep(time.Duration(j) * time.Second)
						cgMarkets, err = GetCGMarkets(ids[i:endIndex], "CNY", topN)
					}
					if err != nil {
						continue
					}
					markets = append(markets, cgMarkets...)
				}
				//sort markets
				sort.Slice(markets, func(i, j int) bool {
					return markets[i].MarketCapRank <= markets[j].MarketCapRank
				})
				index := 0
				for ; index < len(markets) && markets[index].MarketCapRank == 0; index++ {
				}

				markets = append(markets[index:], markets[:index]...)

				if len(markets) > topN {
					markets = markets[:topN]
				}
				c.log.Info("cgMarkets=", markets)
				result = make([]*v1.TokenInfoData, 0, len(markets))
				for _, market := range markets {
					if value, ok := tokenInfos[market.ID]; ok {
						result = append(result, value)
					}
				}
				if len(result) > 0 {
					b, _ := json.Marshal(result)
					if err := utils.SetTokenTop20RedisKey(c.redisClient, key, string(b)); err != nil {
						c.log.Error("set redis client cache error:", err)
					}
				}
			}
		}()
	}

	return result, nil
}

func GetTop20TokenList(chain string) ([]*v1.TokenInfoData, error) {
	oldChain := chain
	key := REDIS_TOP20_KEY + chain
	//get chain
	chain = utils.GetChainNameByChain(chain)
	//get redis cache
	var result []*v1.TokenInfoData
	str, updateFlag, err := utils.GetTokenTop20RedisValueByKey(c.redisClient, key)
	if err != nil {
		c.log.Error("get token top29 key error:", err, key)
	}
	if str != "" {
		if err := json.Unmarshal([]byte(str), &result); err != nil {
			c.log.Error("unmarshal error:", err)
		}
	}
	if updateFlag && !strings.Contains(chain, "TEST") {
		go func() {
			var tokenLists []models.TokenList
			err = c.db.Where("chain = ? AND name != ? AND cg_id != ?", chain, oldChain, "").Find(&tokenLists).Error
			if err != nil {
				c.log.Error("get token list error:", err)
			}
			c.log.Info("token list length=", len(tokenLists))
			if len(tokenLists) >= 20 {
				tokenInfos := make(map[string]*v1.TokenInfoData, len(tokenLists))
				ids := make([]string, len(tokenLists))
				for i, t := range tokenLists {
					ids[i] = t.CgId
					tokenInfos[t.CgId] = &v1.TokenInfoData{
						Chain:    oldChain,
						Address:  t.Address,
						Symbol:   t.Symbol,
						Decimals: uint32(t.Decimals),
						Name:     t.Name,
						LogoURI:  c.logoPrefix + t.LogoURI,
					}
				}
				markets := make([]types.CGMarket, 0, len(tokenLists)+2)
				pageSize := 500
				endIndex := 0
				for i := 0; i < len(ids); i += pageSize {
					if i+pageSize > len(ids) {
						endIndex = len(ids)
					} else {
						endIndex = i + pageSize
					}
					cgMarkets, err := GetCGMarkets(ids[i:endIndex], "CNY", 20)
					if err != nil {
						c.log.Error("get cg markets error:", err)
					}
					for j := 0; err != nil && j < 3; j++ {
						time.Sleep(time.Duration(j) * time.Second)
						cgMarkets, err = GetCGMarkets(ids[i:endIndex], "CNY", 20)
					}
					if err != nil {
						continue
					}
					markets = append(markets, cgMarkets...)
				}
				//sort markets
				sort.Slice(markets, func(i, j int) bool {
					return markets[i].MarketCapRank <= markets[j].MarketCapRank
				})
				index := 0
				for ; index < len(markets) && markets[index].MarketCapRank == 0; index++ {
				}

				markets = append(markets[index:], markets[:index]...)
				if len(markets) > 20 {
					markets = markets[:20]
				}
				c.log.Info("cgMarkets=", markets)
				result = make([]*v1.TokenInfoData, 0, len(markets))
				for _, market := range markets {
					if value, ok := tokenInfos[market.ID]; ok {
						result = append(result, value)
					}
				}
				if len(result) > 0 {
					b, _ := json.Marshal(result)
					if err := utils.SetTokenTop20RedisKey(c.redisClient, key, string(b)); err != nil {
						c.log.Error("set redis client cache error:", err)
					}
				}
			}
		}()
	}

	return result, nil
}

func GetFakeCoinWhiteList(chain, address string) (*models.FakeCoinWhiteList, error) {
	var fakeCoinWhiteList *models.FakeCoinWhiteList
	err := c.db.Where("chain = ? AND address = ?", chain, address).
		First(&fakeCoinWhiteList).Error
	if err != nil {
		c.log.Error("get fake coin white list error:", err)
		return nil, err
	}
	return fakeCoinWhiteList, nil
}

func GetFakeCoinWhiteListBySymbol(chain, symbol string) (*models.FakeCoinWhiteList, error) {

	//get redis cache
	key := fmt.Sprintf("%s%s:%s", REDIS_FAKECOIN_KEY, chain, symbol)
	fakeCoinWhiteList, err := utils.GetFakeCoinWhiteList(c.redisClient, key)
	if err == nil {
		return fakeCoinWhiteList, nil
	}

	err = c.db.Where("chain = ? AND symbol = ?", chain, symbol).First(&fakeCoinWhiteList).Error
	if err != nil {
		c.log.Error("get fake coin white list error:", err)
		return nil, err
	}

	//set redis cache
	_ = utils.SetFakeCoinWhiteList(c.redisClient, key, fakeCoinWhiteList)
	return fakeCoinWhiteList, nil
}

func UpdateNervosToken() {
	count := 0
	page := 1
	pageSize := 20
	url := "https://mainnet-api.explorer.nervos.org/api/v1/udts"
	params := map[string]string{
		"page":      fmt.Sprintf("%d", page),
		"page_size": fmt.Sprintf("%d", pageSize),
	}
	heads := map[string]string{
		"Content-Type": "application/vnd.api+json",
		"Accept":       "application/vnd.api+json",
	}
	out := &types.NervosTokenList{}
	for {
		err := utils3.CommHttpsForm(url, "GET", params, heads, "", out)
		if err != nil {
			c.log.Error("get nervos token list error:", err)
		}
		count += len(out.Data)
		handlerNervosList(out)
		if len(out.Data) < pageSize {
			break
		}
		page += 1
		params["page"] = fmt.Sprintf("%d", page)
	}
}

func handlerNervosList(data *types.NervosTokenList) {
	for _, d := range data.Data {
		var decimal int
		if d.Attributes.Decimal != "" {
			decimal, _ = strconv.Atoi(d.Attributes.Decimal)
		}
		if d.Attributes.TypeScript.Args != "" {
			t := models.TokenList{
				Symbol:   d.Attributes.Symbol,
				Name:     d.Attributes.FullName,
				Address:  strings.ToLower(d.Attributes.TypeScript.Args),
				Logo:     d.Attributes.IconFile,
				Decimals: decimal,
				Chain:    "nervos",
			}
			result := c.db.Clauses(clause.OnConflict{
				UpdateAll: true,
			}).Create(&t)
			if result.Error != nil {
				c.log.Error("create db nervos error:", result.Error)
			}
		}
	}
}

func UpdateTokenTop20() {
	top20 := map[string]string{
		"ETH":       "[{\"chain\":\"ETH\",\"name\":\"Tether\",\"address\":\"0xdac17f958d2ee523a2206206994597c13d831ec7\",\"decimals\":6,\"symbol\":\"USDT\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/ethereum/ethereum_0xdac17f958d2ee523a2206206994597c13d831ec7.png\"},{\"chain\":\"ETH\",\"name\":\"USD Coin\",\"address\":\"0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48\",\"decimals\":6,\"symbol\":\"USDC\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/ethereum/ethereum_0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48.png\"},{\"chain\":\"ETH\",\"name\":\"Binance USD\",\"address\":\"0x4fabb145d64652a948d72533023f6e7a623c7c53\",\"decimals\":18,\"symbol\":\"BUSD\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/ethereum/ethereum_0x4fabb145d64652a948d72533023f6e7a623c7c53.png\"},{\"chain\":\"ETH\",\"name\":\"Polygon\",\"address\":\"0x7d1afa7b718fb893db30a3abc0cfc608aacfebb0\",\"decimals\":18,\"symbol\":\"MATIC\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/ethereum/ethereum_0x7d1afa7b718fb893db30a3abc0cfc608aacfebb0.png\"},{\"chain\":\"ETH\",\"name\":\"OKB\",\"address\":\"0x75231f58b43240c9718dd58b4967c5114342a86c\",\"decimals\":18,\"symbol\":\"OKB\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/ethereum/ethereum_0x75231f58b43240c9718dd58b4967c5114342a86c.png\"},{\"chain\":\"ETH\",\"name\":\"Lido Staked Ether\",\"address\":\"0xae7ab96520de3a18e5e111b5eaab095312d7fe84\",\"decimals\":18,\"symbol\":\"STETH\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/ethereum/ethereum_0xae7ab96520de3a18e5e111b5eaab095312d7fe84.png\"},{\"chain\":\"ETH\",\"name\":\"Shiba Inu\",\"address\":\"0x95ad61b0a150d79219dcf64e1e6cc01f0b64c4ce\",\"decimals\":18,\"symbol\":\"SHIB\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/ethereum/ethereum_0x95ad61b0a150d79219dcf64e1e6cc01f0b64c4ce.png\"},{\"chain\":\"ETH\",\"name\":\"Dai\",\"address\":\"0x6b175474e89094c44da98b954eedeac495271d0f\",\"decimals\":18,\"symbol\":\"DAI\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/ethereum/ethereum_0x6b175474e89094c44da98b954eedeac495271d0f.png\"},{\"chain\":\"ETH\",\"name\":\"Uniswap\",\"address\":\"0x1f9840a85d5af5bf1d1762f925bdaddc4201f984\",\"decimals\":18,\"symbol\":\"UNI\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/ethereum/ethereum_0x1f9840a85d5af5bf1d1762f925bdaddc4201f984.png\"},{\"chain\":\"ETH\",\"name\":\"Wrapped Bitcoin\",\"address\":\"0x2260fac5e5542a773aa44fbcfedf7c193bc2c599\",\"decimals\":8,\"symbol\":\"WBTC\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/ethereum/ethereum_0x2260fac5e5542a773aa44fbcfedf7c193bc2c599.png\"},{\"chain\":\"ETH\",\"name\":\"The Open Network\",\"address\":\"0x582d872a1b094fc48f5de31d3b73f2d9be47def1\",\"decimals\":9,\"symbol\":\"TON\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/ethereum/ethereum_0x582d872a1b094fc48f5de31d3b73f2d9be47def1.png\"},{\"chain\":\"ETH\",\"name\":\"Chainlink\",\"address\":\"0x514910771af9ca656af840dff83e8264ecf986ca\",\"decimals\":18,\"symbol\":\"LINK\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/ethereum/ethereum_0x514910771af9ca656af840dff83e8264ecf986ca.png\"},{\"chain\":\"ETH\",\"name\":\"LEO Token\",\"address\":\"0x2af5d2ad76741191d15dfe7bf6ac92d4bd912ca3\",\"decimals\":18,\"symbol\":\"LEO\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/ethereum/ethereum_0x2af5d2ad76741191d15dfe7bf6ac92d4bd912ca3.png\"},{\"chain\":\"ETH\",\"name\":\"Lido DAO\",\"address\":\"0x5a98fcbea516cf06857215779fd812ca3bef1b32\",\"decimals\":18,\"symbol\":\"LDO\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/ethereum/ethereum_0x5a98fcbea516cf06857215779fd812ca3bef1b32.png\"},{\"chain\":\"ETH\",\"name\":\"Quant\",\"address\":\"0x4a220e6096b25eadb88358cb44068a3248254675\",\"decimals\":18,\"symbol\":\"QNT\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/ethereum/ethereum_0x4a220e6096b25eadb88358cb44068a3248254675.png\"},{\"chain\":\"ETH\",\"name\":\"ApeCoin\",\"address\":\"0x4d224452801aced8b2f0aebe155379bb5d594381\",\"decimals\":18,\"symbol\":\"APE\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/ethereum/ethereum_0x4d224452801aced8b2f0aebe155379bb5d594381.png\"},{\"chain\":\"ETH\",\"name\":\"Cronos\",\"address\":\"0xa0b73e1ff0b80914ab6fe0444e65848c4c34450b\",\"decimals\":8,\"symbol\":\"CRO\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/ethereum/ethereum_0xa0b73e1ff0b80914ab6fe0444e65848c4c34450b.png\"},{\"chain\":\"ETH\",\"name\":\"The Graph\",\"address\":\"0xc944e90c64b2c07662a292be6244bdf05cda44a7\",\"decimals\":18,\"symbol\":\"GRT\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/ethereum/ethereum_0xc944e90c64b2c07662a292be6244bdf05cda44a7.png\"},{\"chain\":\"ETH\",\"name\":\"Fantom\",\"address\":\"0x4e15361fd6b4bb609fa63c81a2be19d873717870\",\"decimals\":18,\"symbol\":\"FTM\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/ethereum/ethereum_0x4e15361fd6b4bb609fa63c81a2be19d873717870.png\"},{\"chain\":\"ETH\",\"name\":\"The Sandbox\",\"address\":\"0x3845badade8e6dff049820680d1f14bd3903a5d0\",\"decimals\":18,\"symbol\":\"SAND\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/ethereum/ethereum_0x3845badade8e6dff049820680d1f14bd3903a5d0.png\"}]",
		"OEC":       "[{\"chain\":\"OEC\",\"name\":\"Tether\",\"address\":\"0x382bb369d343125bfb2117af9c149795c6c65c50\",\"decimals\":18,\"symbol\":\"USDT\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/okex-chain/okex-chain_0x382bb369d343125bfb2117af9c149795c6c65c50.png\"},{\"chain\":\"OEC\",\"name\":\"USD Coin\",\"address\":\"0xc946daf81b08146b1c7a8da2a851ddf2b3eaaf85\",\"decimals\":18,\"symbol\":\"USDC\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/okex-chain/okex-chain_0xc946daf81b08146b1c7a8da2a851ddf2b3eaaf85.png\"},{\"chain\":\"OEC\",\"name\":\"Binance USD\",\"address\":\"0x332730a4f6e03d9c55829435f10360e13cfa41ff\",\"decimals\":18,\"symbol\":\"BUSD\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/okex-chain/okex-chain_0x332730a4f6e03d9c55829435f10360e13cfa41ff.png\"},{\"chain\":\"OEC\",\"name\":\"OKB\",\"address\":\"0xdf54b6c6195ea4d948d03bfd818d365cf175cfc2\",\"decimals\":18,\"symbol\":\"OKB\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/okex-chain/okex-chain_0xdf54b6c6195ea4d948d03bfd818d365cf175cfc2.png\"},{\"chain\":\"OEC\",\"name\":\"WOO Network\",\"address\":\"0x5427a224e50a9af4d030aeecd2a686d41f348dfe\",\"decimals\":18,\"symbol\":\"WOO\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/okex-chain/okex-chain_0x5427a224e50a9af4d030aeecd2a686d41f348dfe.png\"},{\"chain\":\"OEC\",\"name\":\"Radio Caca\",\"address\":\"0x12bb890508c125661e03b09ec06e404bc9289040\",\"decimals\":18,\"symbol\":\"RACA\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/okex-chain/okex-chain_0x12bb890508c125661e03b09ec06e404bc9289040.png\"},{\"chain\":\"OEC\",\"name\":\"USDK\",\"address\":\"0xdcac52e001f5bd413aa6ea83956438f29098166b\",\"decimals\":18,\"symbol\":\"USDK\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/okex-chain/okex-chain_0xdcac52e001f5bd413aa6ea83956438f29098166b.png\"},{\"chain\":\"OEC\",\"name\":\"Crypto Gladiator Shards\",\"address\":\"0x81fde2721f556e402296b2a57e1871637c27d5e8\",\"decimals\":18,\"symbol\":\"CGS\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/okex-chain/okex-chain_0x81fde2721f556e402296b2a57e1871637c27d5e8.png\"},{\"chain\":\"OEC\",\"name\":\"Celestial\",\"address\":\"0x5ab622494ab7c5e81911558c9552dbd517f403fb\",\"decimals\":18,\"symbol\":\"CELT\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/okex-chain/okex-chain_0x5ab622494ab7c5e81911558c9552dbd517f403fb.png\"},{\"chain\":\"OEC\",\"name\":\"vEmpire DDAO\",\"address\":\"0x2c9a1d0e1226939edb7bbb68c43a080c28743c5c\",\"decimals\":18,\"symbol\":\"VEMP\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/okex-chain/okex-chain_0x2c9a1d0e1226939edb7bbb68c43a080c28743c5c.png\"},{\"chain\":\"OEC\",\"name\":\"O3 Swap\",\"address\":\"0xee9801669c6138e84bd50deb500827b776777d28\",\"decimals\":18,\"symbol\":\"O3\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/okex-chain/okex-chain_0xee9801669c6138e84bd50deb500827b776777d28.png\"},{\"chain\":\"OEC\",\"name\":\"CherrySwap\",\"address\":\"0x8179d97eb6488860d816e3ecafe694a4153f216c\",\"decimals\":18,\"symbol\":\"CHE\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/okex-chain/okex-chain_0x8179d97eb6488860d816e3ecafe694a4153f216c.png\"},{\"chain\":\"OEC\",\"name\":\"Jswap.Finance\",\"address\":\"0x5fac926bf1e638944bb16fb5b787b5ba4bc85b0a\",\"decimals\":18,\"symbol\":\"JF\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/okex-chain/okex-chain_0x5fac926bf1e638944bb16fb5b787b5ba4bc85b0a.png\"},{\"chain\":\"OEC\",\"name\":\"Lumiii\",\"address\":\"0xc9b6e036e94a316e4a9ea96f7d7fd8d632f66e7a\",\"decimals\":18,\"symbol\":\"LUMIII\",\"logoURI\":\"https://obstatic.243096.com/download/token/\"},{\"chain\":\"OEC\",\"name\":\"Wrapped OKT\",\"address\":\"0x8f8526dbfd6e38e3d8307702ca8469bae6c56c15\",\"decimals\":18,\"symbol\":\"WOKT\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/okex-chain/okex-chain_0x8f8526dbfd6e38e3d8307702ca8469bae6c56c15.png\"},{\"chain\":\"OEC\",\"name\":\"BABYOKX\",\"address\":\"0xe24b533d2170e808b64e7e22cc4006d19dfad70e\",\"decimals\":18,\"symbol\":\"BABYOKX\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/okex-chain/okex-chain_0xe24b533d2170e808b64e7e22cc4006d19dfad70e.png\"},{\"chain\":\"OEC\",\"name\":\"Flux Protocol\",\"address\":\"0xd0c6821aba4fcc65e8f1542589e64bae9de11228\",\"decimals\":18,\"symbol\":\"FLUX\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/okex-chain/okex-chain_0xd0c6821aba4fcc65e8f1542589e64bae9de11228.png\"},{\"chain\":\"OEC\",\"name\":\"Galaxy War\",\"address\":\"0x65511ce6980418db9db829b669e7dfd653dae420\",\"decimals\":18,\"symbol\":\"GWT\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/okex-chain/okex-chain_0x65511ce6980418db9db829b669e7dfd653dae420.png\"},{\"chain\":\"OEC\",\"name\":\"WePiggy Coin\",\"address\":\"0x6f620ec89b8479e97a6985792d0c64f237566746\",\"decimals\":18,\"symbol\":\"WPC\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/okex-chain/okex-chain_0x6f620ec89b8479e97a6985792d0c64f237566746.png\"},{\"chain\":\"OEC\",\"name\":\"AST.finance\",\"address\":\"0x493d8cbd9533e57d4befb17cc2ec1db76828261d\",\"decimals\":18,\"symbol\":\"AST\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/okex-chain/okex-chain_0x493d8cbd9533e57d4befb17cc2ec1db76828261d.png\"}]",
		"BSC":       "[{\"chain\":\"BSC\",\"name\":\"Tether\",\"address\":\"0x55d398326f99059ff775485246999027b3197955\",\"decimals\":18,\"symbol\":\"USDT\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/binance-smart-chain/binance-smart-chain_0x55d398326f99059ff775485246999027b3197955.png\"},{\"chain\":\"BSC\",\"name\":\"USD Coin\",\"address\":\"0x8ac76a51cc950d9822d68b83fe1ad97b32cd580d\",\"decimals\":18,\"symbol\":\"USDC\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/binance-smart-chain/binance-smart-chain_0x8ac76a51cc950d9822d68b83fe1ad97b32cd580d.png\"},{\"chain\":\"BSC\",\"name\":\"Binance USD\",\"address\":\"0xe9e7cea3dedca5984780bafc599bd69add087d56\",\"decimals\":18,\"symbol\":\"BUSD\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/binance-smart-chain/binance-smart-chain_0xe9e7cea3dedca5984780bafc599bd69add087d56.png\"},{\"chain\":\"BSC\",\"name\":\"Polygon\",\"address\":\"0xcc42724c6683b7e57334c4e856f4c9965ed682bd\",\"decimals\":18,\"symbol\":\"MATIC\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/binance-smart-chain/binance-smart-chain_0xcc42724c6683b7e57334c4e856f4c9965ed682bd.png\"},{\"chain\":\"BSC\",\"name\":\"Dai\",\"address\":\"0x1af3f329e8be154074d8769d1ffa4ee058b1dbc3\",\"decimals\":18,\"symbol\":\"DAI\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/binance-smart-chain/binance-smart-chain_0x1af3f329e8be154074d8769d1ffa4ee058b1dbc3.png\"},{\"chain\":\"BSC\",\"name\":\"Uniswap\",\"address\":\"0xbf5140a22578168fd562dccf235e5d43a02ce9b1\",\"decimals\":18,\"symbol\":\"UNI\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/binance-smart-chain/binance-smart-chain_0xbf5140a22578168fd562dccf235e5d43a02ce9b1.png\"},{\"chain\":\"BSC\",\"name\":\"Cosmos Hub\",\"address\":\"0x0eb3a705fc54725037cc9e008bdede697f62f335\",\"decimals\":18,\"symbol\":\"ATOM\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/binance-smart-chain/binance-smart-chain_0x0eb3a705fc54725037cc9e008bdede697f62f335.png\"},{\"chain\":\"BSC\",\"name\":\"The Open Network\",\"address\":\"0x76a797a59ba2c17726896976b7b3747bfd1d220f\",\"decimals\":9,\"symbol\":\"TON\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/binance-smart-chain/binance-smart-chain_0x76a797a59ba2c17726896976b7b3747bfd1d220f.png\"},{\"chain\":\"BSC\",\"name\":\"Chainlink\",\"address\":\"0xf8a0bf9cf54bb92f17374d9e9a321e6a111a51bd\",\"decimals\":18,\"symbol\":\"LINK\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/binance-smart-chain/binance-smart-chain_0xf8a0bf9cf54bb92f17374d9e9a321e6a111a51bd.png\"},{\"chain\":\"BSC\",\"name\":\"Fantom\",\"address\":\"0xad29abb318791d579433d831ed122afeaf29dcfe\",\"decimals\":18,\"symbol\":\"FTM\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/binance-smart-chain/binance-smart-chain_0xad29abb318791d579433d831ed122afeaf29dcfe.png\"},{\"chain\":\"BSC\",\"name\":\"Aave\",\"address\":\"0xfb6115445bff7b52feb98650c87f44907e58f802\",\"decimals\":18,\"symbol\":\"AAVE\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/binance-smart-chain/binance-smart-chain_0xfb6115445bff7b52feb98650c87f44907e58f802.png\"},{\"chain\":\"BSC\",\"name\":\"Axie Infinity\",\"address\":\"0x715d400f88c167884bbcc41c5fea407ed4d2f8a0\",\"decimals\":18,\"symbol\":\"AXS\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/binance-smart-chain/binance-smart-chain_0x715d400f88c167884bbcc41c5fea407ed4d2f8a0.png\"},{\"chain\":\"BSC\",\"name\":\"Frax\",\"address\":\"0x90c97f71e18723b0cf0dfa30ee176ab653e89f40\",\"decimals\":18,\"symbol\":\"FRAX\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/binance-smart-chain/binance-smart-chain_0x90c97f71e18723b0cf0dfa30ee176ab653e89f40.png\"},{\"chain\":\"BSC\",\"name\":\"TrueUSD\",\"address\":\"0x14016e85a25aeb13065688cafb43044c2ef86784\",\"decimals\":18,\"symbol\":\"TUSD\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/binance-smart-chain/binance-smart-chain_0x14016e85a25aeb13065688cafb43044c2ef86784.png\"},{\"chain\":\"BSC\",\"name\":\"Frax Share\",\"address\":\"0xe48a3d7d0bc88d552f730b62c006bc925eadb9ee\",\"decimals\":18,\"symbol\":\"FXS\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/binance-smart-chain/binance-smart-chain_0xe48a3d7d0bc88d552f730b62c006bc925eadb9ee.png\"},{\"chain\":\"BSC\",\"name\":\"USDD\",\"address\":\"0xd17479997f34dd9156deef8f95a52d81d265be9c\",\"decimals\":18,\"symbol\":\"USDD\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/binance-smart-chain/binance-smart-chain_0xd17479997f34dd9156deef8f95a52d81d265be9c.png\"},{\"chain\":\"BSC\",\"name\":\"PancakeSwap\",\"address\":\"0x0e09fabb73bd3ade0a17ecc321fd13a19e81ce82\",\"decimals\":18,\"symbol\":\"CAKE\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/binance-smart-chain/binance-smart-chain_0x0e09fabb73bd3ade0a17ecc321fd13a19e81ce82.png\"},{\"chain\":\"BSC\",\"name\":\"Baby Doge Coin\",\"address\":\"0xc748673057861a797275cd8a068abb95a902e8de\",\"decimals\":9,\"symbol\":\"BABYDOGE\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/binance-smart-chain/binance-smart-chain_0xc748673057861a797275cd8a068abb95a902e8de.png\"},{\"chain\":\"BSC\",\"name\":\"Trust Wallet\",\"address\":\"0x4b0f1812e5df2a09796481ff14017e6005508003\",\"decimals\":18,\"symbol\":\"TWT\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/binance-smart-chain/binance-smart-chain_0x4b0f1812e5df2a09796481ff14017e6005508003.png\"},{\"chain\":\"BSC\",\"name\":\"Zilliqa\",\"address\":\"0xb86abcb37c3a4b64f74f59301aff131a1becc787\",\"decimals\":12,\"symbol\":\"ZIL\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/binance-smart-chain/binance-smart-chain_0xb86abcb37c3a4b64f74f59301aff131a1becc787.png\"}]",
		"Polygon":   "[{\"chain\":\"Polygon\",\"name\":\"Tether\",\"address\":\"0xc2132d05d31c914a87c6611c10748aeb04b58e8f\",\"decimals\":6,\"symbol\":\"USDT\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/polygon-pos/polygon-pos_0xc2132d05d31c914a87c6611c10748aeb04b58e8f.png\"},{\"chain\":\"Polygon\",\"name\":\"USD Coin\",\"address\":\"0x2791bca1f2de4661ed88a30c99a7a9449aa84174\",\"decimals\":6,\"symbol\":\"USDC\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/polygon-pos/polygon-pos_0x2791bca1f2de4661ed88a30c99a7a9449aa84174.png\"},{\"chain\":\"Polygon\",\"name\":\"Binance USD\",\"address\":\"0xdab529f40e671a1d4bf91361c21bf9f0c9712ab7\",\"decimals\":18,\"symbol\":\"BUSD\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/polygon-pos/polygon-pos_0xdab529f40e671a1d4bf91361c21bf9f0c9712ab7.png\"},{\"chain\":\"Polygon\",\"name\":\"Avalanche\",\"address\":\"0x2c89bbc92bd86f8075d1decc58c7f4e0107f286b\",\"decimals\":18,\"symbol\":\"AVAX\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/polygon-pos/polygon-pos_0x2c89bbc92bd86f8075d1decc58c7f4e0107f286b.png\"},{\"chain\":\"Polygon\",\"name\":\"Dai\",\"address\":\"0x8f3cf7ad23cd3cadbd9735aff958023239c6a063\",\"decimals\":18,\"symbol\":\"DAI\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/polygon-pos/polygon-pos_0x8f3cf7ad23cd3cadbd9735aff958023239c6a063.png\"},{\"chain\":\"Polygon\",\"name\":\"Uniswap\",\"address\":\"0xb33eaad8d922b1083446dc23f610c2567fb5180f\",\"decimals\":18,\"symbol\":\"UNI\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/polygon-pos/polygon-pos_0xb33eaad8d922b1083446dc23f610c2567fb5180f.png\"},{\"chain\":\"Polygon\",\"name\":\"Wrapped Bitcoin\",\"address\":\"0x1bfd67037b42cf73acf2047067bd4f2c47d9bfd6\",\"decimals\":8,\"symbol\":\"WBTC\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/polygon-pos/polygon-pos_0x1bfd67037b42cf73acf2047067bd4f2c47d9bfd6.png\"},{\"chain\":\"Polygon\",\"name\":\"Chainlink\",\"address\":\"0x53e0bca35ec356bd5dddfebbd1fc0fd03fabad39\",\"decimals\":18,\"symbol\":\"LINK\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/polygon-pos/polygon-pos_0x53e0bca35ec356bd5dddfebbd1fc0fd03fabad39.png\"},{\"chain\":\"Polygon\",\"name\":\"Lido DAO\",\"address\":\"0xc3c7d422809852031b44ab29eec9f1eff2a58756\",\"decimals\":18,\"symbol\":\"LDO\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/polygon-pos/polygon-pos_0xc3c7d422809852031b44ab29eec9f1eff2a58756.png\"},{\"chain\":\"Polygon\",\"name\":\"The Graph\",\"address\":\"0x5fe2b58c013d7601147dcdd68c143a77499f5531\",\"decimals\":18,\"symbol\":\"GRT\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/polygon-pos/polygon-pos_0x5fe2b58c013d7601147dcdd68c143a77499f5531.png\"},{\"chain\":\"Polygon\",\"name\":\"The Sandbox\",\"address\":\"0xbbba073c31bf03b8acf7c28ef0738decf3695683\",\"decimals\":18,\"symbol\":\"SAND\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/polygon-pos/polygon-pos_0xbbba073c31bf03b8acf7c28ef0738decf3695683.png\"},{\"chain\":\"Polygon\",\"name\":\"Decentraland\",\"address\":\"0xa1c57f48f0deb89f569dfbe6e2b7f46d33606fd4\",\"decimals\":18,\"symbol\":\"MANA\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/polygon-pos/polygon-pos_0xa1c57f48f0deb89f569dfbe6e2b7f46d33606fd4.png\"},{\"chain\":\"Polygon\",\"name\":\"Aave\",\"address\":\"0xd6df932a45c0f255f85145f286ea0b292b21c90b\",\"decimals\":18,\"symbol\":\"AAVE\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/polygon-pos/polygon-pos_0xd6df932a45c0f255f85145f286ea0b292b21c90b.png\"},{\"chain\":\"Polygon\",\"name\":\"Frax\",\"address\":\"0x45c32fa6df82ead1e2ef74d17b76547eddfaff89\",\"decimals\":18,\"symbol\":\"FRAX\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/polygon-pos/polygon-pos_0x45c32fa6df82ead1e2ef74d17b76547eddfaff89.png\"},{\"chain\":\"Polygon\",\"name\":\"TrueUSD\",\"address\":\"0x2e1ad108ff1d8c782fcbbb89aad783ac49586756\",\"decimals\":18,\"symbol\":\"TUSD\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/polygon-pos/polygon-pos_0x2e1ad108ff1d8c782fcbbb89aad783ac49586756.png\"},{\"chain\":\"Polygon\",\"name\":\"Curve DAO\",\"address\":\"0x172370d5cd63279efa6d502dab29171933a610af\",\"decimals\":18,\"symbol\":\"CRV\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/polygon-pos/polygon-pos_0x172370d5cd63279efa6d502dab29171933a610af.png\"},{\"chain\":\"Polygon\",\"name\":\"Synthetix Network\",\"address\":\"0x50b728d8d964fd00c2d0aad81718b71311fef68a\",\"decimals\":18,\"symbol\":\"SNX\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/polygon-pos/polygon-pos_0x50b728d8d964fd00c2d0aad81718b71311fef68a.png\"},{\"chain\":\"Polygon\",\"name\":\"Frax Share\",\"address\":\"0x1a3acf6d19267e2d3e7f898f42803e90c9219062\",\"decimals\":18,\"symbol\":\"FXS\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/polygon-pos/polygon-pos_0x1a3acf6d19267e2d3e7f898f42803e90c9219062.png\"},{\"chain\":\"Polygon\",\"name\":\"Render\",\"address\":\"0x61299774020da444af134c82fa83e3810b309991\",\"decimals\":18,\"symbol\":\"RNDR\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/polygon-pos/polygon-pos_0x61299774020da444af134c82fa83e3810b309991.png\"},{\"chain\":\"Polygon\",\"name\":\"1inch\",\"address\":\"0x9c2c5fd7b07e95ee044ddeba0e97a665f142394f\",\"decimals\":18,\"symbol\":\"1INCH\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/polygon-pos/polygon-pos_0x9c2c5fd7b07e95ee044ddeba0e97a665f142394f.png\"}]",
		"Fantom":    "[{\"chain\":\"Fantom\",\"name\":\"Tether\",\"address\":\"0x049d68029688eabf473097a2fc38ef61633a3c7a\",\"decimals\":6,\"symbol\":\"USDT\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/fantom/fantom_0x049d68029688eabf473097a2fc38ef61633a3c7a.png\"},{\"chain\":\"Fantom\",\"name\":\"USD Coin\",\"address\":\"0x04068da6c83afcfa0e13ba15a6696662335d5b75\",\"decimals\":6,\"symbol\":\"USDC\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/fantom/fantom_0x04068da6c83afcfa0e13ba15a6696662335d5b75.png\"},{\"chain\":\"Fantom\",\"name\":\"Dai\",\"address\":\"0x8d11ec38a3eb5e956b052f67da8bdc9bef8abf3e\",\"decimals\":18,\"symbol\":\"DAI\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/fantom/fantom_0x8d11ec38a3eb5e956b052f67da8bdc9bef8abf3e.png\"},{\"chain\":\"Fantom\",\"name\":\"Wrapped Bitcoin\",\"address\":\"0x321162cd933e2be498cd2267a90534a804051b11\",\"decimals\":8,\"symbol\":\"WBTC\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/fantom/fantom_0x321162cd933e2be498cd2267a90534a804051b11.png\"},{\"chain\":\"Fantom\",\"name\":\"Chainlink\",\"address\":\"0xb3654dc3d10ea7645f8319668e8f54d2574fbdc8\",\"decimals\":18,\"symbol\":\"LINK\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/fantom/fantom_0xb3654dc3d10ea7645f8319668e8f54d2574fbdc8.png\"},{\"chain\":\"Fantom\",\"name\":\"Aave\",\"address\":\"0x6a07a792ab2965c72a5b8088d3a069a7ac3a993b\",\"decimals\":18,\"symbol\":\"AAVE\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/fantom/fantom_0x6a07a792ab2965c72a5b8088d3a069a7ac3a993b.png\"},{\"chain\":\"Fantom\",\"name\":\"Frax\",\"address\":\"0xdc301622e621166bd8e82f2ca0a26c13ad0be355\",\"decimals\":18,\"symbol\":\"FRAX\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/fantom/fantom_0xdc301622e621166bd8e82f2ca0a26c13ad0be355.png\"},{\"chain\":\"Fantom\",\"name\":\"Terra Luna Classic\",\"address\":\"0x95dd59343a893637be1c3228060ee6afbf6f0730\",\"decimals\":6,\"symbol\":\"LUNC\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/fantom/fantom_0x95dd59343a893637be1c3228060ee6afbf6f0730.png\"},{\"chain\":\"Fantom\",\"name\":\"TrueUSD\",\"address\":\"0x9879abdea01a879644185341f7af7d8343556b7a\",\"decimals\":18,\"symbol\":\"TUSD\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/fantom/fantom_0x9879abdea01a879644185341f7af7d8343556b7a.png\"},{\"chain\":\"Fantom\",\"name\":\"Curve DAO\",\"address\":\"0x1e4f97b9f9f913c46f1632781732927b9019c68b\",\"decimals\":18,\"symbol\":\"CRV\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/fantom/fantom_0x1e4f97b9f9f913c46f1632781732927b9019c68b.png\"},{\"chain\":\"Fantom\",\"name\":\"Synthetix Network\",\"address\":\"0x56ee926bd8c72b2d5fa1af4d9e4cbb515a1e3adc\",\"decimals\":18,\"symbol\":\"SNX\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/fantom/fantom_0x56ee926bd8c72b2d5fa1af4d9e4cbb515a1e3adc.png\"},{\"chain\":\"Fantom\",\"name\":\"Frax Share\",\"address\":\"0x7d016eec9c25232b01f23ef992d98ca97fc2af5a\",\"decimals\":18,\"symbol\":\"FXS\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/fantom/fantom_0x7d016eec9c25232b01f23ef992d98ca97fc2af5a.png\"},{\"chain\":\"Fantom\",\"name\":\"NEXO\",\"address\":\"0x7c598c96d02398d89fbcb9d41eab3df0c16f227d\",\"decimals\":18,\"symbol\":\"NEXO\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/fantom/fantom_0x7c598c96d02398d89fbcb9d41eab3df0c16f227d.png\"},{\"chain\":\"Fantom\",\"name\":\"WOO Network\",\"address\":\"0x6626c47c00f1d87902fc13eecfac3ed06d5e8d8a\",\"decimals\":18,\"symbol\":\"WOO\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/fantom/fantom_0x6626c47c00f1d87902fc13eecfac3ed06d5e8d8a.png\"},{\"chain\":\"Fantom\",\"name\":\"Synapse\",\"address\":\"0xe55e19fb4f2d85af758950957714292dac1e25b2\",\"decimals\":18,\"symbol\":\"SYN\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/fantom/fantom_0xe55e19fb4f2d85af758950957714292dac1e25b2.png\"},{\"chain\":\"Fantom\",\"name\":\"Sushi\",\"address\":\"0xae75a438b2e0cb8bb01ec1e1e376de11d44477cc\",\"decimals\":18,\"symbol\":\"SUSHI\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/fantom/fantom_0xae75a438b2e0cb8bb01ec1e1e376de11d44477cc.png\"},{\"chain\":\"Fantom\",\"name\":\"Band Protocol\",\"address\":\"0x46e7628e8b4350b2716ab470ee0ba1fa9e76c6c5\",\"decimals\":18,\"symbol\":\"BAND\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/fantom/fantom_0x46e7628e8b4350b2716ab470ee0ba1fa9e76c6c5.png\"},{\"chain\":\"Fantom\",\"name\":\"yearn.finance\",\"address\":\"0x29b0da86e484e1c0029b56e817912d778ac0ec69\",\"decimals\":18,\"symbol\":\"YFI\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/fantom/fantom_0x29b0da86e484e1c0029b56e817912d778ac0ec69.png\"},{\"chain\":\"Fantom\",\"name\":\"Celsius Network\",\"address\":\"0x2c78f1b70ccf63cdee49f9233e9faa99d43aa07e\",\"decimals\":4,\"symbol\":\"CEL\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/fantom/fantom_0x2c78f1b70ccf63cdee49f9233e9faa99d43aa07e.png\"},{\"chain\":\"Fantom\",\"name\":\"Multichain\",\"address\":\"0x9fb9a33956351cf4fa040f65a13b835a3c8764e3\",\"decimals\":18,\"symbol\":\"MULTI\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/fantom/fantom_0x9fb9a33956351cf4fa040f65a13b835a3c8764e3.png\"}]",
		"Avalanche": "[{\"chain\":\"Avalanche\",\"name\":\"Tether\",\"address\":\"0x9702230a8ea53601f5cd2dc00fdbc13d4df4a8c7\",\"decimals\":6,\"symbol\":\"USDT\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/avalanche/avalanche_0x9702230a8ea53601f5cd2dc00fdbc13d4df4a8c7.png\"},{\"chain\":\"Avalanche\",\"name\":\"USD Coin\",\"address\":\"0xb97ef9ef8734c71904d8002f8b6bc66dd9c48a6e\",\"decimals\":6,\"symbol\":\"USDC\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/avalanche/avalanche_0xb97ef9ef8734c71904d8002f8b6bc66dd9c48a6e.png\"},{\"chain\":\"Avalanche\",\"name\":\"Binance USD\",\"address\":\"0x19860ccb0a68fd4213ab9d8266f7bbf05a8dde98\",\"decimals\":18,\"symbol\":\"BUSD\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/avalanche/avalanche_0x19860ccb0a68fd4213ab9d8266f7bbf05a8dde98.png\"},{\"chain\":\"Avalanche\",\"name\":\"Dai\",\"address\":\"0xd586e7f844cea2f87f50152665bcbc2c279d8d70\",\"decimals\":18,\"symbol\":\"DAI\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/avalanche/avalanche_0xd586e7f844cea2f87f50152665bcbc2c279d8d70.png\"},{\"chain\":\"Avalanche\",\"name\":\"Uniswap\",\"address\":\"0x8ebaf22b6f053dffeaf46f4dd9efa95d89ba8580\",\"decimals\":18,\"symbol\":\"UNI\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/avalanche/avalanche_0x8ebaf22b6f053dffeaf46f4dd9efa95d89ba8580.png\"},{\"chain\":\"Avalanche\",\"name\":\"Wrapped Bitcoin\",\"address\":\"0x50b7545627a5162f82a992c33b87adc75187b218\",\"decimals\":8,\"symbol\":\"WBTC\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/avalanche/avalanche_0x50b7545627a5162f82a992c33b87adc75187b218.png\"},{\"chain\":\"Avalanche\",\"name\":\"Chainlink\",\"address\":\"0x5947bb275c521040051d82396192181b413227a3\",\"decimals\":18,\"symbol\":\"LINK\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/avalanche/avalanche_0x5947bb275c521040051d82396192181b413227a3.png\"},{\"chain\":\"Avalanche\",\"name\":\"The Graph\",\"address\":\"0x8a0cac13c7da965a312f08ea4229c37869e85cb9\",\"decimals\":18,\"symbol\":\"GRT\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/avalanche/avalanche_0x8a0cac13c7da965a312f08ea4229c37869e85cb9.png\"},{\"chain\":\"Avalanche\",\"name\":\"Aave\",\"address\":\"0x63a72806098bd3d9520cc43356dd78afe5d386d9\",\"decimals\":18,\"symbol\":\"AAVE\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/avalanche/avalanche_0x63a72806098bd3d9520cc43356dd78afe5d386d9.png\"},{\"chain\":\"Avalanche\",\"name\":\"Frax\",\"address\":\"0xd24c2ad096400b6fbcd2ad8b24e7acbc21a1da64\",\"decimals\":18,\"symbol\":\"FRAX\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/avalanche/avalanche_0xd24c2ad096400b6fbcd2ad8b24e7acbc21a1da64.png\"},{\"chain\":\"Avalanche\",\"name\":\"TrueUSD\",\"address\":\"0x1c20e891bab6b1727d14da358fae2984ed9b59eb\",\"decimals\":18,\"symbol\":\"TUSD\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/avalanche/avalanche_0x1c20e891bab6b1727d14da358fae2984ed9b59eb.png\"},{\"chain\":\"Avalanche\",\"name\":\"Synthetix Network\",\"address\":\"0xbec243c995409e6520d7c41e404da5deba4b209b\",\"decimals\":18,\"symbol\":\"SNX\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/avalanche/avalanche_0xbec243c995409e6520d7c41e404da5deba4b209b.png\"},{\"chain\":\"Avalanche\",\"name\":\"Frax Share\",\"address\":\"0x214db107654ff987ad859f34125307783fc8e387\",\"decimals\":18,\"symbol\":\"FXS\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/avalanche/avalanche_0x214db107654ff987ad859f34125307783fc8e387.png\"},{\"chain\":\"Avalanche\",\"name\":\"Maker\",\"address\":\"0x88128fd4b259552a9a1d457f435a6527aab72d42\",\"decimals\":18,\"symbol\":\"MKR\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/avalanche/avalanche_0x88128fd4b259552a9a1d457f435a6527aab72d42.png\"},{\"chain\":\"Avalanche\",\"name\":\"GMX\",\"address\":\"0x62edc0692bd897d2295872a9ffcac5425011c661\",\"decimals\":18,\"symbol\":\"GMX\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/avalanche/avalanche_0x62edc0692bd897d2295872a9ffcac5425011c661.png\"},{\"chain\":\"Avalanche\",\"name\":\"1inch\",\"address\":\"0xd501281565bf7789224523144fe5d98e8b28f267\",\"decimals\":18,\"symbol\":\"1INCH\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/avalanche/avalanche_0xd501281565bf7789224523144fe5d98e8b28f267.png\"},{\"chain\":\"Avalanche\",\"name\":\"Basic Attention\",\"address\":\"0x98443b96ea4b0858fdf3219cd13e98c7a4690588\",\"decimals\":18,\"symbol\":\"BAT\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/avalanche/avalanche_0x98443b96ea4b0858fdf3219cd13e98c7a4690588.png\"},{\"chain\":\"Avalanche\",\"name\":\"WOO Network\",\"address\":\"0xabc9547b534519ff73921b1fba6e672b5f58d083\",\"decimals\":18,\"symbol\":\"WOO\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/avalanche/avalanche_0xabc9547b534519ff73921b1fba6e672b5f58d083.png\"},{\"chain\":\"Avalanche\",\"name\":\"Compound\",\"address\":\"0xc3048e19e76cb9a3aa9d77d8c03c29fc906e2437\",\"decimals\":18,\"symbol\":\"COMP\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/avalanche/avalanche_0xc3048e19e76cb9a3aa9d77d8c03c29fc906e2437.png\"},{\"chain\":\"Avalanche\",\"name\":\"Synapse\",\"address\":\"0x1f1e7c893855525b303f99bdf5c3c05be09ca251\",\"decimals\":18,\"symbol\":\"SYN\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/avalanche/avalanche_0x1f1e7c893855525b303f99bdf5c3c05be09ca251.png\"}]",
		"Cronos":    "[{\"chain\":\"Cronos\",\"name\":\"Tether\",\"address\":\"0x66e428c3f67a68878562e79a0234c1f83c208770\",\"decimals\":6,\"symbol\":\"USDT\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/cronos/cronos_0x66e428c3f67a68878562e79a0234c1f83c208770.png\"},{\"chain\":\"Cronos\",\"name\":\"USD Coin\",\"address\":\"0xc21223249ca28397b4b6541dffaecc539bff0c59\",\"decimals\":6,\"symbol\":\"USDC\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/cronos/cronos_0xc21223249ca28397b4b6541dffaecc539bff0c59.png\"},{\"chain\":\"Cronos\",\"name\":\"Binance USD\",\"address\":\"0xc74d59a548ecf7fc1754bb7810d716e9ac3e3ae5\",\"decimals\":18,\"symbol\":\"BUSD\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/cronos/cronos_0xc74d59a548ecf7fc1754bb7810d716e9ac3e3ae5.png\"},{\"chain\":\"Cronos\",\"name\":\"Dai\",\"address\":\"0xf2001b145b43032aaf5ee2884e456ccd805f677d\",\"decimals\":18,\"symbol\":\"DAI\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/cronos/cronos_0xf2001b145b43032aaf5ee2884e456ccd805f677d.png\"},{\"chain\":\"Cronos\",\"name\":\"Wrapped Bitcoin\",\"address\":\"0x062e66477faf219f25d27dced647bf57c3107d52\",\"decimals\":8,\"symbol\":\"WBTC\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/cronos/cronos_0x062e66477faf219f25d27dced647bf57c3107d52.png\"},{\"chain\":\"Cronos\",\"name\":\"Alethea Artificial Liquid Intelligence\",\"address\":\"0x45c135c1cdce8d25a3b729a28659561385c52671\",\"decimals\":18,\"symbol\":\"ALI\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/cronos/cronos_0x45c135c1cdce8d25a3b729a28659561385c52671.png\"},{\"chain\":\"Cronos\",\"name\":\"VVS Finance\",\"address\":\"0x2d03bece6747adc00e1a131bba1469c15fd11e03\",\"decimals\":18,\"symbol\":\"VVS\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/cronos/cronos_0x2d03bece6747adc00e1a131bba1469c15fd11e03.png\"},{\"chain\":\"Cronos\",\"name\":\"MAI\",\"address\":\"0x2ae35c8e3d4bd57e8898ff7cd2bbff87166ef8cb\",\"decimals\":18,\"symbol\":\"MIMATIC\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/cronos/cronos_0x2ae35c8e3d4bd57e8898ff7cd2bbff87166ef8cb.png\"},{\"chain\":\"Cronos\",\"name\":\"Minted\",\"address\":\"0x0224010ba2d567ffa014222ed960d1fa43b8c8e1\",\"decimals\":18,\"symbol\":\"MTD\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/cronos/cronos_0x0224010ba2d567ffa014222ed960d1fa43b8c8e1.png\"},{\"chain\":\"Cronos\",\"name\":\"Beefy.Finance\",\"address\":\"0xe6801928061cdbe32ac5ad0634427e140efd05f9\",\"decimals\":18,\"symbol\":\"BIFI\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/cronos/cronos_0xe6801928061cdbe32ac5ad0634427e140efd05f9.png\"},{\"chain\":\"Cronos\",\"name\":\"Ferro\",\"address\":\"0x39bc1e38c842c60775ce37566d03b41a7a66c782\",\"decimals\":18,\"symbol\":\"FER\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/cronos/cronos_0x39bc1e38c842c60775ce37566d03b41a7a66c782.png\"},{\"chain\":\"Cronos\",\"name\":\"Tectonic\",\"address\":\"0xdd73dea10abc2bff99c60882ec5b2b81bb1dc5b2\",\"decimals\":18,\"symbol\":\"TONIC\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/cronos/cronos_0xdd73dea10abc2bff99c60882ec5b2b81bb1dc5b2.png\"},{\"chain\":\"Cronos\",\"name\":\"Cronos ID\",\"address\":\"0xcbf0adea24fd5f32c6e7f0474f0d1b94ace4e2e7\",\"decimals\":18,\"symbol\":\"CROID\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/cronos/cronos_0xcbf0adea24fd5f32c6e7f0474f0d1b94ace4e2e7.png\"},{\"chain\":\"Cronos\",\"name\":\"MMFinance\",\"address\":\"0x97749c9b61f878a880dfe312d2594ae07aed7656\",\"decimals\":18,\"symbol\":\"MMF\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/cronos/cronos_0x97749c9b61f878a880dfe312d2594ae07aed7656.png\"},{\"chain\":\"Cronos\",\"name\":\"Liquidus\",\"address\":\"0xabd380327fe66724ffda91a87c772fb8d00be488\",\"decimals\":18,\"symbol\":\"LIQ\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/cronos/cronos_0xabd380327fe66724ffda91a87c772fb8d00be488.png\"},{\"chain\":\"Cronos\",\"name\":\"VersaGames\",\"address\":\"0x00d7699b71290094ccb1a5884cd835bd65a78c17\",\"decimals\":18,\"symbol\":\"VERSA\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/cronos/cronos_0x00d7699b71290094ccb1a5884cd835bd65a78c17.png\"},{\"chain\":\"Cronos\",\"name\":\"RadioShack\",\"address\":\"0xf899e3909b4492859d44260e1de41a9e663e70f5\",\"decimals\":18,\"symbol\":\"RADIO\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/cronos/cronos_0xf899e3909b4492859d44260e1de41a9e663e70f5.png\"},{\"chain\":\"Cronos\",\"name\":\"Savanna\",\"address\":\"0x654bac3ec77d6db497892478f854cf6e8245dca9\",\"decimals\":18,\"symbol\":\"SVN\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/cronos/cronos_0x654bac3ec77d6db497892478f854cf6e8245dca9.png\"},{\"chain\":\"Cronos\",\"name\":\"Meerkat Shares\",\"address\":\"0xf8b9facb7b4410f5703eb29093302f2933d6e1aa\",\"decimals\":18,\"symbol\":\"MSHARE\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/cronos/cronos_0xf8b9facb7b4410f5703eb29093302f2933d6e1aa.png\"},{\"chain\":\"Cronos\",\"name\":\"Elk Finance\",\"address\":\"0xeeeeeb57642040be42185f49c52f7e9b38f8eeee\",\"decimals\":18,\"symbol\":\"ELK\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/cronos/cronos_0xeeeeeb57642040be42185f49c52f7e9b38f8eeee.png\"}]",
		"Arbitrum":  "[{\"chain\":\"Arbitrum\",\"name\":\"Tether\",\"address\":\"0xfd086bc7cd5c481dcc9c85ebe478a1c0b69fcbb9\",\"decimals\":6,\"symbol\":\"USDT\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/arbitrum-one/arbitrum-one_0xfd086bc7cd5c481dcc9c85ebe478a1c0b69fcbb9.png\"},{\"chain\":\"Arbitrum\",\"name\":\"USD Coin\",\"address\":\"0xff970a61a04b1ca14834a43f5de4533ebddb5cc8\",\"decimals\":6,\"symbol\":\"USDC\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/arbitrum-one/arbitrum-one_0xff970a61a04b1ca14834a43f5de4533ebddb5cc8.png\"},{\"chain\":\"Arbitrum\",\"name\":\"Dai\",\"address\":\"0xda10009cbd5d07dd0cecc66161fc93d7c9000da1\",\"decimals\":18,\"symbol\":\"DAI\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/arbitrum-one/arbitrum-one_0xda10009cbd5d07dd0cecc66161fc93d7c9000da1.png\"},{\"chain\":\"Arbitrum\",\"name\":\"Uniswap\",\"address\":\"0xfa7f8980b0f1e64a2062791cc3b0871572f1f7f0\",\"decimals\":18,\"symbol\":\"UNI\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/arbitrum-one/arbitrum-one_0xfa7f8980b0f1e64a2062791cc3b0871572f1f7f0.png\"},{\"chain\":\"Arbitrum\",\"name\":\"Wrapped Bitcoin\",\"address\":\"0x2f2a2543b76a4166549f7aab2e75bef0aefc5b0f\",\"decimals\":8,\"symbol\":\"WBTC\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/arbitrum-one/arbitrum-one_0x2f2a2543b76a4166549f7aab2e75bef0aefc5b0f.png\"},{\"chain\":\"Arbitrum\",\"name\":\"Chainlink\",\"address\":\"0xf97f4df75117a78c1a5a0dbb814af92458539fb4\",\"decimals\":18,\"symbol\":\"LINK\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/arbitrum-one/arbitrum-one_0xf97f4df75117a78c1a5a0dbb814af92458539fb4.png\"},{\"chain\":\"Arbitrum\",\"name\":\"The Graph\",\"address\":\"0x23a941036ae778ac51ab04cea08ed6e2fe103614\",\"decimals\":18,\"symbol\":\"GRT\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/arbitrum-one/arbitrum-one_0x23a941036ae778ac51ab04cea08ed6e2fe103614.png\"},{\"chain\":\"Arbitrum\",\"name\":\"Frax\",\"address\":\"0x17fc002b466eec40dae837fc4be5c67993ddbd6f\",\"decimals\":18,\"symbol\":\"FRAX\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/arbitrum-one/arbitrum-one_0x17fc002b466eec40dae837fc4be5c67993ddbd6f.png\"},{\"chain\":\"Arbitrum\",\"name\":\"TrueUSD\",\"address\":\"0x4d15a3a2286d883af0aa1b3f21367843fac63e07\",\"decimals\":18,\"symbol\":\"TUSD\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/arbitrum-one/arbitrum-one_0x4d15a3a2286d883af0aa1b3f21367843fac63e07.png\"},{\"chain\":\"Arbitrum\",\"name\":\"Curve DAO\",\"address\":\"0x11cdb42b0eb46d95f990bedd4695a6e3fa034978\",\"decimals\":18,\"symbol\":\"CRV\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/arbitrum-one/arbitrum-one_0x11cdb42b0eb46d95f990bedd4695a6e3fa034978.png\"},{\"chain\":\"Arbitrum\",\"name\":\"Frax Share\",\"address\":\"0x9d2f299715d94d8a7e6f5eaa8e654e8c74a988a7\",\"decimals\":18,\"symbol\":\"FXS\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/arbitrum-one/arbitrum-one_0x9d2f299715d94d8a7e6f5eaa8e654e8c74a988a7.png\"},{\"chain\":\"Arbitrum\",\"name\":\"Maker\",\"address\":\"0x2e9a6df78e42a30712c10a9dc4b1c8656f8f2879\",\"decimals\":18,\"symbol\":\"MKR\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/arbitrum-one/arbitrum-one_0x2e9a6df78e42a30712c10a9dc4b1c8656f8f2879.png\"},{\"chain\":\"Arbitrum\",\"name\":\"GMX\",\"address\":\"0xfc5a1a6eb076a2c7ad06ed22c90d7e710e35ad0a\",\"decimals\":18,\"symbol\":\"GMX\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/arbitrum-one/arbitrum-one_0xfc5a1a6eb076a2c7ad06ed22c90d7e710e35ad0a.png\"},{\"chain\":\"Arbitrum\",\"name\":\"Loopring\",\"address\":\"0x46d0ce7de6247b0a95f67b43b589b4041bae7fbe\",\"decimals\":18,\"symbol\":\"LRC\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/arbitrum-one/arbitrum-one_0x46d0ce7de6247b0a95f67b43b589b4041bae7fbe.png\"},{\"chain\":\"Arbitrum\",\"name\":\"WOO Network\",\"address\":\"0xcafcd85d8ca7ad1e1c6f82f651fa15e33aefd07b\",\"decimals\":18,\"symbol\":\"WOO\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/arbitrum-one/arbitrum-one_0xcafcd85d8ca7ad1e1c6f82f651fa15e33aefd07b.png\"},{\"chain\":\"Arbitrum\",\"name\":\"Compound\",\"address\":\"0x354a6da3fcde098f8389cad84b0182725c6c91de\",\"decimals\":18,\"symbol\":\"COMP\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/arbitrum-one/arbitrum-one_0x354a6da3fcde098f8389cad84b0182725c6c91de.png\"},{\"chain\":\"Arbitrum\",\"name\":\"Magic\",\"address\":\"0x539bde0d7dbd336b79148aa742883198bbf60342\",\"decimals\":18,\"symbol\":\"MAGIC\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/arbitrum-one/arbitrum-one_0x539bde0d7dbd336b79148aa742883198bbf60342.png\"},{\"chain\":\"Arbitrum\",\"name\":\"Gnosis\",\"address\":\"0xa0b862f60edef4452f25b4160f177db44deb6cf1\",\"decimals\":18,\"symbol\":\"GNO\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/arbitrum-one/arbitrum-one_0xa0b862f60edef4452f25b4160f177db44deb6cf1.png\"},{\"chain\":\"Arbitrum\",\"name\":\"Synapse\",\"address\":\"0x080f6aed32fc474dd5717105dba5ea57268f46eb\",\"decimals\":18,\"symbol\":\"SYN\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/arbitrum-one/arbitrum-one_0x080f6aed32fc474dd5717105dba5ea57268f46eb.png\"},{\"chain\":\"Arbitrum\",\"name\":\"Balancer\",\"address\":\"0x040d1edc9569d4bab2d15287dc5a4f10f56a56b8\",\"decimals\":18,\"symbol\":\"BAL\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/arbitrum-one/arbitrum-one_0x040d1edc9569d4bab2d15287dc5a4f10f56a56b8.png\"}]",
		"Klaytn":    "[{\"chain\":\"Klaytn\",\"name\":\"WEMIX\",\"address\":\"0x5096db80b21ef45230c9e423c373f1fc9c0198dd\",\"decimals\":18,\"symbol\":\"WEMIX\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/klay-token/klay-token_0x5096db80b21ef45230c9e423c373f1fc9c0198dd.png\"},{\"chain\":\"Klaytn\",\"name\":\"BORA\",\"address\":\"0x02cbe46fb8a1f579254a9b485788f2d86cad51aa\",\"decimals\":18,\"symbol\":\"BORA\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/klay-token/klay-token_0x02cbe46fb8a1f579254a9b485788f2d86cad51aa.png\"},{\"chain\":\"Klaytn\",\"name\":\"Marblex\",\"address\":\"0xd068c52d81f4409b9502da926ace3301cc41f623\",\"decimals\":18,\"symbol\":\"MBX\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/klay-token/klay-token_0xd068c52d81f4409b9502da926ace3301cc41f623.png\"},{\"chain\":\"Klaytn\",\"name\":\"KlayCity ORB\",\"address\":\"0x01ad62e0ff6dcaa72889fca155c7036c78ca1783\",\"decimals\":18,\"symbol\":\"ORB\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/klay-token/klay-token_0x01ad62e0ff6dcaa72889fca155c7036c78ca1783.png\"},{\"chain\":\"Klaytn\",\"name\":\"Pibble\",\"address\":\"0xafde910130c335fa5bd5fe991053e3e0a49dce7b\",\"decimals\":18,\"symbol\":\"PIB\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/klay-token/klay-token_0xafde910130c335fa5bd5fe991053e3e0a49dce7b.png\"},{\"chain\":\"Klaytn\",\"name\":\"Hiblocks\",\"address\":\"0xe06b40df899b9717b4e6b50711e1dc72d08184cf\",\"decimals\":18,\"symbol\":\"HIBS\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/klay-token/klay-token_0xe06b40df899b9717b4e6b50711e1dc72d08184cf.png\"},{\"chain\":\"Klaytn\",\"name\":\"SuperWalk\",\"address\":\"0x84f8c3c8d6ee30a559d73ec570d574f671e82647\",\"decimals\":18,\"symbol\":\"GRND\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/klay-token/klay-token_0x84f8c3c8d6ee30a559d73ec570d574f671e82647.png\"},{\"chain\":\"Klaytn\",\"name\":\"Project WITH\",\"address\":\"0x275f942985503d8ce9558f8377cc526a3aba3566\",\"decimals\":18,\"symbol\":\"WIKEN\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/klay-token/klay-token_0x275f942985503d8ce9558f8377cc526a3aba3566.png\"},{\"chain\":\"Klaytn\",\"name\":\"STAT\",\"address\":\"0x01987adc61782639ea3b8497e030b13a4510cfbe\",\"decimals\":18,\"symbol\":\"STAT\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/klay-token/klay-token_0x01987adc61782639ea3b8497e030b13a4510cfbe.png\"},{\"chain\":\"Klaytn\",\"name\":\"Ducato Finance\",\"address\":\"0xb7ab8c205dc282f83d0511b1dcd22a3ed4739597\",\"decimals\":18,\"symbol\":\"DUCATO\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/klay-token/klay-token_0xb7ab8c205dc282f83d0511b1dcd22a3ed4739597.png\"},{\"chain\":\"Klaytn\",\"name\":\"Klap Finance\",\"address\":\"0xd109065ee17e2dc20b3472a4d4fb5907bd687d09\",\"decimals\":18,\"symbol\":\"KLAP\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/klay-token/klay-token_0xd109065ee17e2dc20b3472a4d4fb5907bd687d09.png\"},{\"chain\":\"Klaytn\",\"name\":\"Galaxia\",\"address\":\"0xa80e96cceb1419f9bd9f1c67f7978f51b534a11b\",\"decimals\":18,\"symbol\":\"GXA\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/klay-token/klay-token_0xa80e96cceb1419f9bd9f1c67f7978f51b534a11b.png\"},{\"chain\":\"Klaytn\",\"name\":\"ClaimSwap\",\"address\":\"0xcf87f94fd8f6b6f0b479771f10df672f99eada63\",\"decimals\":18,\"symbol\":\"CLA\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/klay-token/klay-token_0xcf87f94fd8f6b6f0b479771f10df672f99eada63.png\"},{\"chain\":\"Klaytn\",\"name\":\"Artube\",\"address\":\"0x07aa7ae19b17579f7237ad72c616fecf4ccc787b\",\"decimals\":18,\"symbol\":\"ATT\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/klay-token/klay-token_0x07aa7ae19b17579f7237ad72c616fecf4ccc787b.png\"},{\"chain\":\"Klaytn\",\"name\":\"Influencer\",\"address\":\"0xdde2154f47e80c8721c2efbe02834ae056284368\",\"decimals\":18,\"symbol\":\"IMI\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/klay-token/klay-token_0xdde2154f47e80c8721c2efbe02834ae056284368.png\"},{\"chain\":\"Klaytn\",\"name\":\"Kokoa Finance\",\"address\":\"0xb15183d0d4d5e86ba702ce9bb7b633376e7db29f\",\"decimals\":18,\"symbol\":\"KOKOA\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/klay-token/klay-token_0xb15183d0d4d5e86ba702ce9bb7b633376e7db29f.png\"},{\"chain\":\"Klaytn\",\"name\":\"Grinbit\",\"address\":\"0xe2a10a2e55ece8a0fdf29dd8359d694f37b8ce17\",\"decimals\":18,\"symbol\":\"GRBT\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/klay-token/klay-token_0xe2a10a2e55ece8a0fdf29dd8359d694f37b8ce17.png\"},{\"chain\":\"Klaytn\",\"name\":\"IPVERSE\",\"address\":\"0x02e973155b1f5f60a1ff1c4e8e7f371c89526cbc\",\"decimals\":18,\"symbol\":\"IPV\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/klay-token/klay-token_0x02e973155b1f5f60a1ff1c4e8e7f371c89526cbc.png\"},{\"chain\":\"Klaytn\",\"name\":\"Cloudbric\",\"address\":\"0xdfb25178d7b59e33f7805c00c4a354ae1c46139a\",\"decimals\":18,\"symbol\":\"CLBK\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/klay-token/klay-token_0xdfb25178d7b59e33f7805c00c4a354ae1c46139a.png\"},{\"chain\":\"Klaytn\",\"name\":\"Cojam\",\"address\":\"0x7f223b1607171b81ebd68d22f1ca79157fd4a44b\",\"decimals\":18,\"symbol\":\"CT\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/klay-token/klay-token_0x7f223b1607171b81ebd68d22f1ca79157fd4a44b.png\"}]",
		"Aurora":    "[{\"chain\":\"Aurora\",\"name\":\"Tether\",\"address\":\"0x4988a896b1227218e4a686fde5eabdcabd91571f\",\"decimals\":6,\"symbol\":\"USDT\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/aurora/aurora_0x4988a896b1227218e4a686fde5eabdcabd91571f.png\"},{\"chain\":\"Aurora\",\"name\":\"USD Coin\",\"address\":\"0xb12bfca5a55806aaf64e99521918a4bf0fc40802\",\"decimals\":6,\"symbol\":\"USDC\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/aurora/aurora_0xb12bfca5a55806aaf64e99521918a4bf0fc40802.png\"},{\"chain\":\"Aurora\",\"name\":\"Dai\",\"address\":\"0xe3520349f477a5f6eb06107066048508498a291b\",\"decimals\":18,\"symbol\":\"DAI\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/aurora/aurora_0xe3520349f477a5f6eb06107066048508498a291b.png\"},{\"chain\":\"Aurora\",\"name\":\"Wrapped Bitcoin\",\"address\":\"0xf4eb217ba2454613b15dbdea6e5f22276410e89e\",\"decimals\":8,\"symbol\":\"WBTC\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/aurora/aurora_0xf4eb217ba2454613b15dbdea6e5f22276410e89e.png\"},{\"chain\":\"Aurora\",\"name\":\"DODO\",\"address\":\"0xe301ed8c7630c9678c39e4e45193d1e7dfb914f7\",\"decimals\":18,\"symbol\":\"DODO\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/aurora/aurora_0xe301ed8c7630c9678c39e4e45193d1e7dfb914f7.png\"},{\"chain\":\"Aurora\",\"name\":\"MAI\",\"address\":\"0xdfa46478f9e5ea86d57387849598dbfb2e964b02\",\"decimals\":18,\"symbol\":\"MIMATIC\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/aurora/aurora_0xdfa46478f9e5ea86d57387849598dbfb2e964b02.png\"},{\"chain\":\"Aurora\",\"name\":\"Beefy.Finance\",\"address\":\"0x218c3c3d49d0e7b37aff0d8bb079de36ae61a4c0\",\"decimals\":18,\"symbol\":\"BIFI\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/aurora/aurora_0x218c3c3d49d0e7b37aff0d8bb079de36ae61a4c0.png\"},{\"chain\":\"Aurora\",\"name\":\"FluxProtocol\",\"address\":\"0xea62791aa682d455614eaa2a12ba3d9a2fd197af\",\"decimals\":18,\"symbol\":\"FLX\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/aurora/aurora_0xea62791aa682d455614eaa2a12ba3d9a2fd197af.png\"},{\"chain\":\"Aurora\",\"name\":\"Pickle Finance\",\"address\":\"0x291c8fceaca3342b29cc36171deb98106f712c66\",\"decimals\":18,\"symbol\":\"PICKLE\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/aurora/aurora_0x291c8fceaca3342b29cc36171deb98106f712c66.png\"},{\"chain\":\"Aurora\",\"name\":\"Aurigami\",\"address\":\"0x09c9d464b58d96837f8d8b6f4d9fe4ad408d3a4f\",\"decimals\":18,\"symbol\":\"PLY\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/aurora/aurora_0x09c9d464b58d96837f8d8b6f4d9fe4ad408d3a4f.png\"},{\"chain\":\"Aurora\",\"name\":\"Chronicle\",\"address\":\"0x7ca1c28663b76cfde424a9494555b94846205585\",\"decimals\":18,\"symbol\":\"XNL\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/aurora/aurora_0x7ca1c28663b76cfde424a9494555b94846205585.png\"},{\"chain\":\"Aurora\",\"name\":\"Allbridge\",\"address\":\"0x2bae00c8bc1868a5f7a216e881bae9e662630111\",\"decimals\":18,\"symbol\":\"ABR\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/aurora/aurora_0x2bae00c8bc1868a5f7a216e881bae9e662630111.png\"},{\"chain\":\"Aurora\",\"name\":\"Trisolaris\",\"address\":\"0xfa94348467f64d5a457f75f8bc40495d33c65abb\",\"decimals\":18,\"symbol\":\"TRI\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/aurora/aurora_0xfa94348467f64d5a457f75f8bc40495d33c65abb.png\"},{\"chain\":\"Aurora\",\"name\":\"Bastion Protocol\",\"address\":\"0x9f1f933c660a1dc856f0e0fe058435879c5ccef0\",\"decimals\":18,\"symbol\":\"BSTN\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/aurora/aurora_0x9f1f933c660a1dc856f0e0fe058435879c5ccef0.png\"},{\"chain\":\"Aurora\",\"name\":\"SOLACE\",\"address\":\"0x501ace9c35e60f03a2af4d484f49f9b1efde9f40\",\"decimals\":18,\"symbol\":\"SOLACE\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/aurora/aurora_0x501ace9c35e60f03a2af4d484f49f9b1efde9f40.png\"},{\"chain\":\"Aurora\",\"name\":\"NearPad\",\"address\":\"0x885f8cf6e45bdd3fdcdc644efdcd0ac93880c781\",\"decimals\":18,\"symbol\":\"PAD\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/aurora/aurora_0x885f8cf6e45bdd3fdcdc644efdcd0ac93880c781.png\"},{\"chain\":\"Aurora\",\"name\":\"Baked\",\"address\":\"0x8973c9ec7b79fe880697cdbca744892682764c37\",\"decimals\":18,\"symbol\":\"BAKED\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/aurora/aurora_0x8973c9ec7b79fe880697cdbca744892682764c37.png\"},{\"chain\":\"Aurora\",\"name\":\"WannaSwap\",\"address\":\"0x7faa64faf54750a2e3ee621166635feaf406ab22\",\"decimals\":18,\"symbol\":\"WANNA\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/aurora/aurora_0x7faa64faf54750a2e3ee621166635feaf406ab22.png\"},{\"chain\":\"Aurora\",\"name\":\"Ethernal\",\"address\":\"0x17cbd9c274e90c537790c51b4015a65cd015497e\",\"decimals\":18,\"symbol\":\"ETHERNAL\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/aurora/aurora_0x17cbd9c274e90c537790c51b4015a65cd015497e.png\"},{\"chain\":\"Aurora\",\"name\":\"Empyrean\",\"address\":\"0xe9f226a228eb58d408fdb94c3ed5a18af6968fe1\",\"decimals\":9,\"symbol\":\"EMPYR\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/aurora/aurora_0xe9f226a228eb58d408fdb94c3ed5a18af6968fe1.png\"}]",
		"Optimism":  "[{\"chain\":\"Optimism\",\"name\":\"Tether\",\"address\":\"0x94b008aa00579c1307b0ef2c499ad98a8ce58e58\",\"decimals\":6,\"symbol\":\"USDT\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/optimistic-ethereum/optimistic-ethereum_0x94b008aa00579c1307b0ef2c499ad98a8ce58e58.png\"},{\"chain\":\"Optimism\",\"name\":\"USD Coin\",\"address\":\"0x7f5c764cbc14f9669b88837ca1490cca17c31607\",\"decimals\":6,\"symbol\":\"USDC\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/optimistic-ethereum/optimistic-ethereum_0x7f5c764cbc14f9669b88837ca1490cca17c31607.png\"},{\"chain\":\"Optimism\",\"name\":\"Dai\",\"address\":\"0xda10009cbd5d07dd0cecc66161fc93d7c9000da1\",\"decimals\":18,\"symbol\":\"DAI\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/optimistic-ethereum/optimistic-ethereum_0xda10009cbd5d07dd0cecc66161fc93d7c9000da1.png\"},{\"chain\":\"Optimism\",\"name\":\"Uniswap\",\"address\":\"0x6fd9d7ad17242c41f7131d257212c54a0e816691\",\"decimals\":18,\"symbol\":\"UNI\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/optimistic-ethereum/optimistic-ethereum_0x6fd9d7ad17242c41f7131d257212c54a0e816691.png\"},{\"chain\":\"Optimism\",\"name\":\"Wrapped Bitcoin\",\"address\":\"0x68f180fcce6836688e9084f035309e29bf0a2095\",\"decimals\":8,\"symbol\":\"WBTC\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/optimistic-ethereum/optimistic-ethereum_0x68f180fcce6836688e9084f035309e29bf0a2095.png\"},{\"chain\":\"Optimism\",\"name\":\"Chainlink\",\"address\":\"0x350a791bfc2c21f9ed5d10980dad2e2638ffa7f6\",\"decimals\":18,\"symbol\":\"LINK\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/optimistic-ethereum/optimistic-ethereum_0x350a791bfc2c21f9ed5d10980dad2e2638ffa7f6.png\"},{\"chain\":\"Optimism\",\"name\":\"Synthetix Network\",\"address\":\"0x8700daec35af8ff88c16bdf0418774cb3d7599b4\",\"decimals\":18,\"symbol\":\"SNX\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/optimistic-ethereum/optimistic-ethereum_0x8700daec35af8ff88c16bdf0418774cb3d7599b4.png\"},{\"chain\":\"Optimism\",\"name\":\"Stargate Finance\",\"address\":\"0x296f55f8fb28e498b858d0bcda06d955b2cb3f97\",\"decimals\":18,\"symbol\":\"STG\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/optimistic-ethereum/optimistic-ethereum_0x296f55f8fb28e498b858d0bcda06d955b2cb3f97.png\"},{\"chain\":\"Optimism\",\"name\":\"sUSD\",\"address\":\"0x8c6f28f2f1a3c87f0f938b96d27520d9751ec8d9\",\"decimals\":18,\"symbol\":\"SUSD\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/optimistic-ethereum/optimistic-ethereum_0x8c6f28f2f1a3c87f0f938b96d27520d9751ec8d9.png\"},{\"chain\":\"Optimism\",\"name\":\"Lyra Finance\",\"address\":\"0x50c5725949a6f0c72e6c4a641f24049a917db0cb\",\"decimals\":18,\"symbol\":\"LYRA\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/optimistic-ethereum/optimistic-ethereum_0x50c5725949a6f0c72e6c4a641f24049a917db0cb.png\"},{\"chain\":\"Optimism\",\"name\":\"Perpetual Protocol\",\"address\":\"0x9e1028f5f1d5ede59748ffcee5532509976840e0\",\"decimals\":18,\"symbol\":\"PERP\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/optimistic-ethereum/optimistic-ethereum_0x9e1028f5f1d5ede59748ffcee5532509976840e0.png\"},{\"chain\":\"Optimism\",\"name\":\"sETH\",\"address\":\"0xe405de8f52ba7559f9df3c368500b6e6ae6cee49\",\"decimals\":18,\"symbol\":\"SETH\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/optimistic-ethereum/optimistic-ethereum_0xe405de8f52ba7559f9df3c368500b6e6ae6cee49.png\"},{\"chain\":\"Optimism\",\"name\":\"Thales\",\"address\":\"0x217d47011b23bb961eb6d93ca9945b7501a5bb11\",\"decimals\":18,\"symbol\":\"THALES\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/optimistic-ethereum/optimistic-ethereum_0x217d47011b23bb961eb6d93ca9945b7501a5bb11.png\"},{\"chain\":\"Optimism\",\"name\":\"Velodrome Finance\",\"address\":\"0x3c8b650257cfb5f272f799f5e2b4e65093a11a05\",\"decimals\":18,\"symbol\":\"VELO\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/optimistic-ethereum/optimistic-ethereum_0x3c8b650257cfb5f272f799f5e2b4e65093a11a05.png\"},{\"chain\":\"Optimism\",\"name\":\"BLOCKv\",\"address\":\"0xe3c332a5dce0e1d9bc2cc72a68437790570c28a4\",\"decimals\":18,\"symbol\":\"VEE\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/optimistic-ethereum/optimistic-ethereum_0xe3c332a5dce0e1d9bc2cc72a68437790570c28a4.png\"},{\"chain\":\"Optimism\",\"name\":\"The Doge NFT\",\"address\":\"0x8f69ee043d52161fd29137aedf63f5e70cd504d5\",\"decimals\":18,\"symbol\":\"DOG\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/optimistic-ethereum/optimistic-ethereum_0x8f69ee043d52161fd29137aedf63f5e70cd504d5.png\"},{\"chain\":\"Optimism\",\"name\":\"Rai Reflex Index\",\"address\":\"0x7fb688ccf682d58f86d7e38e03f9d22e7705448b\",\"decimals\":18,\"symbol\":\"RAI\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/optimistic-ethereum/optimistic-ethereum_0x7fb688ccf682d58f86d7e38e03f9d22e7705448b.png\"},{\"chain\":\"Optimism\",\"name\":\"Kromatika\",\"address\":\"0xf98dcd95217e15e05d8638da4c91125e59590b07\",\"decimals\":18,\"symbol\":\"KROM\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/optimistic-ethereum/optimistic-ethereum_0xf98dcd95217e15e05d8638da4c91125e59590b07.png\"},{\"chain\":\"Optimism\",\"name\":\"0xBitcoin\",\"address\":\"0xe0bb0d3de8c10976511e5030ca403dbf4c25165b\",\"decimals\":8,\"symbol\":\"0XBTC\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/optimistic-ethereum/optimistic-ethereum_0xe0bb0d3de8c10976511e5030ca403dbf4c25165b.png\"},{\"chain\":\"Optimism\",\"name\":\"Dentacoin\",\"address\":\"0x1da650c3b2daa8aa9ff6f661d4156ce24d08a062\",\"symbol\":\"DCN\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/optimistic-ethereum/optimistic-ethereum_0x1da650c3b2daa8aa9ff6f661d4156ce24d08a062.png\"}]",
		"Oasis":     "[{\"chain\":\"Oasis\",\"name\":\"SOL (Wormhole)\",\"address\":\"0xd17ddac91670274f7ba1590a06eca0f2fd2b12bc\",\"decimals\":9,\"symbol\":\"SOL\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/oasis/oasis_0xd17ddac91670274f7ba1590a06eca0f2fd2b12bc.png\"},{\"chain\":\"Oasis\",\"name\":\"Terra Classic (Wormhole)\",\"address\":\"0x4f43717b20ae319aa50bc5b2349b93af5f7ac823\",\"decimals\":6,\"symbol\":\"LUNC\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/oasis/oasis_0x4f43717b20ae319aa50bc5b2349b93af5f7ac823.png\"},{\"chain\":\"Oasis\",\"name\":\"Fountain Protocol\",\"address\":\"0xd1df9ce4b6159441d18bd6887dbd7320a8d52a05\",\"decimals\":18,\"symbol\":\"FTP\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/oasis/oasis_0xd1df9ce4b6159441d18bd6887dbd7320a8d52a05.png\"},{\"chain\":\"Oasis\",\"name\":\"Ethereum (Wormhole)\",\"address\":\"0x3223f17957ba502cbe71401d55a0db26e5f7c68f\",\"decimals\":18,\"symbol\":\"ETH\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/oasis/oasis_0x3223f17957ba502cbe71401d55a0db26e5f7c68f.png\"},{\"chain\":\"Oasis\",\"name\":\"Avalanche (Wormhole)\",\"address\":\"0x32847e63e99d3a044908763056e25694490082f8\",\"decimals\":18,\"symbol\":\"AVAX\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/oasis/oasis_0x32847e63e99d3a044908763056e25694490082f8.png\"},{\"chain\":\"Oasis\",\"name\":\"Tether USD (PoS) (Wormhole)\",\"address\":\"0xfffd69e757d8220cea60dc80b9fe1a30b58c94f3\",\"decimals\":6,\"symbol\":\"USDTPO\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/oasis/oasis_0xfffd69e757d8220cea60dc80b9fe1a30b58c94f3.png\"},{\"chain\":\"Oasis\",\"name\":\"YuzuSwap\",\"address\":\"0xf02b3e437304892105992512539f769423a515cb\",\"decimals\":18,\"symbol\":\"YUZU\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/oasis/oasis_0xf02b3e437304892105992512539f769423a515cb.png\"},{\"chain\":\"Oasis\",\"name\":\"USD Coin (Wormhole from Ethereum)\",\"address\":\"0xe8a638b3b7565ee7c5eb9755e58552afc87b94dd\",\"decimals\":6,\"symbol\":\"USDCET\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/oasis/oasis_0xe8a638b3b7565ee7c5eb9755e58552afc87b94dd.png\"},{\"chain\":\"Oasis\",\"name\":\"USD Coin (PoS) (Wormhole)\",\"address\":\"0x3e62a9c3af8b810de79645c4579acc8f0d06a241\",\"decimals\":6,\"symbol\":\"USDCPO\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/oasis/oasis_0x3e62a9c3af8b810de79645c4579acc8f0d06a241.png\"},{\"chain\":\"Oasis\",\"name\":\"Tether USD (Wormhole from Ethereum)\",\"address\":\"0xdc19a122e268128b5ee20366299fc7b5b199c8e3\",\"decimals\":6,\"symbol\":\"USDTET\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/oasis/oasis_0xdc19a122e268128b5ee20366299fc7b5b199c8e3.png\"},{\"chain\":\"Oasis\",\"name\":\"Tether USD (Wormhole)\",\"address\":\"0x24285c5232ce3858f00bacb950cae1f59d1b2704\",\"decimals\":6,\"symbol\":\"USDTSO\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/oasis/oasis_0x24285c5232ce3858f00bacb950cae1f59d1b2704.png\"}]",
		"TRX":       "[{\"chain\":\"TRX\",\"name\":\"TrueUSD\",\"address\":\"TUpMhErZL2fhh4sVNULAbNKLokS4GjC1F4\",\"decimals\":18,\"symbol\":\"TUSD\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/tron/tron_TUpMhErZL2fhh4sVNULAbNKLokS4GjC1F4.png\"},{\"chain\":\"TRX\",\"name\":\"USDD\",\"address\":\"TPYmHEhy5n8TCEfYGqW2rPxsghSfzghPDn\",\"decimals\":18,\"symbol\":\"USDD\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/tron/tron_TPYmHEhy5n8TCEfYGqW2rPxsghSfzghPDn.png\"},{\"chain\":\"TRX\",\"name\":\"BitTorrent\",\"address\":\"TAFjULxiVgT4qWk6UZwjqwZXTSaGaqnVp4\",\"decimals\":18,\"symbol\":\"BTT\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/tron/tron_TAFjULxiVgT4qWk6UZwjqwZXTSaGaqnVp4.png\"},{\"chain\":\"TRX\",\"name\":\"JUST\",\"address\":\"TCFLL5dx5ZJdKnWuesXxi1VPwjLVmWZZy9\",\"decimals\":18,\"symbol\":\"JST\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/tron/tron_TCFLL5dx5ZJdKnWuesXxi1VPwjLVmWZZy9.png\"},{\"chain\":\"TRX\",\"name\":\"APENFT\",\"address\":\"TFczxzPhnThNSqr5by8tvxsdCFRRz6cPNq\",\"decimals\":6,\"symbol\":\"NFT\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/tron/tron_TFczxzPhnThNSqr5by8tvxsdCFRRz6cPNq.png\"},{\"chain\":\"TRX\",\"name\":\"WINkLink\",\"address\":\"TLa2f6VPqDgRE67v1736s7bJ8Ray5wYjU7\",\"decimals\":6,\"symbol\":\"WIN\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/tron/tron_TLa2f6VPqDgRE67v1736s7bJ8Ray5wYjU7.png\"},{\"chain\":\"TRX\",\"name\":\"Sun Token\",\"address\":\"TSSMHYeV2uE9qYH95DqyoCuNCzEL1NvU3S\",\"decimals\":18,\"symbol\":\"SUN\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/tron/tron_TSSMHYeV2uE9qYH95DqyoCuNCzEL1NvU3S.png\"},{\"chain\":\"TRX\",\"name\":\"Revain\",\"address\":\"TD4bVgcwj3FRbmAo283HxNvqZvY7T3uD8k\",\"decimals\":6,\"symbol\":\"REV\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/tron/tron_TD4bVgcwj3FRbmAo283HxNvqZvY7T3uD8k.png\"},{\"chain\":\"TRX\",\"name\":\"Klever\",\"address\":\"TVj7RNVHy6thbM7BWdSe9G6gXwKhjhdNZS\",\"decimals\":6,\"symbol\":\"KLV\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/tron/tron_TVj7RNVHy6thbM7BWdSe9G6gXwKhjhdNZS.png\"},{\"chain\":\"TRX\",\"name\":\"Pallapay\",\"address\":\"TCfLxS9xHxH8auBL3T8pf3NFTZhsxy4Ncg\",\"decimals\":8,\"symbol\":\"PALLA\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/tron/tron_TCfLxS9xHxH8auBL3T8pf3NFTZhsxy4Ncg.png\"},{\"chain\":\"TRX\",\"name\":\"CrossWallet\",\"address\":\"TY2Ge1YYphoAatwaBxa1zYfJVa8CqNyL6B\",\"decimals\":18,\"symbol\":\"CWT\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/tron/tron_TY2Ge1YYphoAatwaBxa1zYfJVa8CqNyL6B.png\"},{\"chain\":\"TRX\",\"name\":\"IBS\",\"address\":\"TKtj2eCBiRKZ4EBHnu4YgwrPfVftZz1Y4x\",\"decimals\":18,\"symbol\":\"IBS\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/tron/tron_TKtj2eCBiRKZ4EBHnu4YgwrPfVftZz1Y4x.png\"},{\"chain\":\"TRX\",\"name\":\"Bankroll Network\",\"address\":\"TNo59Khpq46FGf4sD7XSWYFNfYfbc8CqNK\",\"decimals\":6,\"symbol\":\"BNKR\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/tron/tron_TNo59Khpq46FGf4sD7XSWYFNfYfbc8CqNK.png\"},{\"chain\":\"TRX\",\"name\":\"SWAPZ.app\",\"address\":\"TKdpGm5wKPLz3stemW8gkGvAdhztJcDjcW\",\"decimals\":18,\"symbol\":\"SWAPZ\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/tron/tron_TKdpGm5wKPLz3stemW8gkGvAdhztJcDjcW.png\"},{\"chain\":\"TRX\",\"name\":\"StrongHands Finance\",\"address\":\"TV5MV8wjCY8xbJ6NfQSArP7xZRHHTw48No\",\"decimals\":18,\"symbol\":\"ISHND\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/tron/tron_TV5MV8wjCY8xbJ6NfQSArP7xZRHHTw48No.png\"},{\"chain\":\"TRX\",\"name\":\"UpStable\",\"address\":\"TYX2iy3i3793YgKU5vqKxDnLpiBMSa5EdV\",\"decimals\":6,\"symbol\":\"USTX\",\"logoURI\":\"https://obstatic.243096.com/download/token/\"},{\"chain\":\"TRX\",\"name\":\"PRivaCY Coin\",\"address\":\"TYV5eu6UgSPtxVLkPD9YfxmUEcXhum35yS\",\"decimals\":8,\"symbol\":\"PRCY\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/tron/tron_TYV5eu6UgSPtxVLkPD9YfxmUEcXhum35yS.png\"},{\"chain\":\"TRX\",\"name\":\"FINXFLO\",\"address\":\"TSNr126nQ8HKfXREqrqkBnxoQNHS5qJLg5\",\"decimals\":18,\"symbol\":\"FXF\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/tron/tron_TSNr126nQ8HKfXREqrqkBnxoQNHS5qJLg5.png\"},{\"chain\":\"TRX\",\"name\":\"TTcoin\",\"address\":\"TCMwzYUUCxLkTNpXjkYSBgXgqXwt7KJ82y\",\"decimals\":4,\"symbol\":\"TC\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/tron/tron_TCMwzYUUCxLkTNpXjkYSBgXgqXwt7KJ82y.png\"},{\"chain\":\"TRX\",\"name\":\"Black Phoenix\",\"address\":\"TXBcx59eDVndV5upFQnTR2xdvqFd5reXET\",\"decimals\":18,\"symbol\":\"BPX\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/tron/tron_TXBcx59eDVndV5upFQnTR2xdvqFd5reXET.png\"}]",
		"xDai":      "[{\"chain\":\"xDai\",\"name\":\"USD Coin\",\"address\":\"0xddafbb505ad214d7b80b1f830fccc89b60fb7a83\",\"decimals\":6,\"symbol\":\"USDC\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/xdai/xdai_0xddafbb505ad214d7b80b1f830fccc89b60fb7a83.png\"},{\"chain\":\"xDai\",\"name\":\"Uniswap\",\"address\":\"0x4537e328bf7e4efa29d05caea260d7fe26af9d74\",\"decimals\":18,\"symbol\":\"UNI\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/xdai/xdai_0x4537e328bf7e4efa29d05caea260d7fe26af9d74.png\"},{\"chain\":\"xDai\",\"name\":\"Wrapped Bitcoin\",\"address\":\"0x8e5bbbb09ed1ebde8674cda39a0c169401db4252\",\"decimals\":8,\"symbol\":\"WBTC\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/xdai/xdai_0x8e5bbbb09ed1ebde8674cda39a0c169401db4252.png\"},{\"chain\":\"xDai\",\"name\":\"Chainlink\",\"address\":\"0xe2e73a1c69ecf83f464efce6a5be353a37ca09b2\",\"decimals\":18,\"symbol\":\"LINK\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/xdai/xdai_0xe2e73a1c69ecf83f464efce6a5be353a37ca09b2.png\"},{\"chain\":\"xDai\",\"name\":\"Gnosis\",\"address\":\"0x9c58bacc331c9aa871afd802db6379a98e80cedb\",\"decimals\":18,\"symbol\":\"GNO\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/xdai/xdai_0x9c58bacc331c9aa871afd802db6379a98e80cedb.png\"},{\"chain\":\"xDai\",\"name\":\"yearn.finance\",\"address\":\"0xbf65bfcb5da067446cee6a706ba3fe2fb1a9fdfd\",\"decimals\":18,\"symbol\":\"YFI\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/xdai/xdai_0xbf65bfcb5da067446cee6a706ba3fe2fb1a9fdfd.png\"},{\"chain\":\"xDai\",\"name\":\"Energy Web\",\"address\":\"0x6a8cb6714b1ee5b471a7d2ec4302cb4f5ff25ec2\",\"decimals\":18,\"symbol\":\"EWT\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/xdai/xdai_0x6a8cb6714b1ee5b471a7d2ec4302cb4f5ff25ec2.png\"},{\"chain\":\"xDai\",\"name\":\"Badger DAO\",\"address\":\"0xdfc20ae04ed70bd9c7d720f449eedae19f659d65\",\"decimals\":18,\"symbol\":\"BADGER\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/xdai/xdai_0xdfc20ae04ed70bd9c7d720f449eedae19f659d65.png\"},{\"chain\":\"xDai\",\"name\":\"MAI\",\"address\":\"0x3f56e0c36d275367b8c502090edf38289b3dea0d\",\"decimals\":18,\"symbol\":\"MIMATIC\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/xdai/xdai_0x3f56e0c36d275367b8c502090edf38289b3dea0d.png\"},{\"chain\":\"xDai\",\"name\":\"Circuits of Value\",\"address\":\"0x8b8407c6184f1f0fd1082e83d6a3b8349caced12\",\"decimals\":8,\"symbol\":\"COVAL\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/xdai/xdai_0x8b8407c6184f1f0fd1082e83d6a3b8349caced12.png\"},{\"chain\":\"xDai\",\"name\":\"Streamr\",\"address\":\"0x256eb8a51f382650b2a1e946b8811953640ee47d\",\"decimals\":18,\"symbol\":\"DATA\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/xdai/xdai_0x256eb8a51f382650b2a1e946b8811953640ee47d.png\"},{\"chain\":\"xDai\",\"name\":\"r/CryptoCurrency Moons\",\"address\":\"0x1e16aa4df73d29c029d94ceda3e3114ec191e25a\",\"decimals\":18,\"symbol\":\"MOON\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/xdai/xdai_0x1e16aa4df73d29c029d94ceda3e3114ec191e25a.png\"},{\"chain\":\"xDai\",\"name\":\"DeFi Pulse Index\",\"address\":\"0xd3d47d5578e55c880505dc40648f7f9307c3e7a8\",\"decimals\":18,\"symbol\":\"DPI\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/xdai/xdai_0xd3d47d5578e55c880505dc40648f7f9307c3e7a8.png\"},{\"chain\":\"xDai\",\"name\":\"HOPR\",\"address\":\"0xd057604a14982fe8d88c5fc25aac3267ea142a08\",\"decimals\":18,\"symbol\":\"HOPR\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/xdai/xdai_0xd057604a14982fe8d88c5fc25aac3267ea142a08.png\"},{\"chain\":\"xDai\",\"name\":\"DXdao\",\"address\":\"0xb90d6bec20993be5d72a5ab353343f7a0281f158\",\"decimals\":18,\"symbol\":\"DXD\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/xdai/xdai_0xb90d6bec20993be5d72a5ab353343f7a0281f158.png\"},{\"chain\":\"xDai\",\"name\":\"UniCrypt\",\"address\":\"0x0116e28b43a358162b96f70b4de14c98a4465f25\",\"decimals\":18,\"symbol\":\"UNCX\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/xdai/xdai_0x0116e28b43a358162b96f70b4de14c98a4465f25.png\"},{\"chain\":\"xDai\",\"name\":\"ShapeShift FOX Token\",\"address\":\"0x21a42669643f45bc0e086b8fc2ed70c23d67509d\",\"decimals\":18,\"symbol\":\"FOX\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/xdai/xdai_0x21a42669643f45bc0e086b8fc2ed70c23d67509d.png\"},{\"chain\":\"xDai\",\"name\":\"Swash\",\"address\":\"0x84e2c67cbefae6b5148fca7d02b341b12ff4b5bb\",\"decimals\":18,\"symbol\":\"SWASH\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/xdai/xdai_0x84e2c67cbefae6b5148fca7d02b341b12ff4b5bb.png\"},{\"chain\":\"xDai\",\"name\":\"CoW Protocol\",\"address\":\"0x177127622c4a00f3d409b75571e12cb3c8973d3c\",\"decimals\":18,\"symbol\":\"COW\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/xdai/xdai_0x177127622c4a00f3d409b75571e12cb3c8973d3c.png\"},{\"chain\":\"xDai\",\"name\":\"Etherisc DIP\",\"address\":\"0x48b1b0d077b4919b65b4e4114806dd803901e1d9\",\"decimals\":18,\"symbol\":\"DIP\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/xdai/xdai_0x48b1b0d077b4919b65b4e4114806dd803901e1d9.png\"}]",
		"Solana":    "[{\"chain\":\"Solana\",\"name\":\"Tether\",\"address\":\"Es9vMFrzaCERmJfrF4H2FYD4KCoNkY11McCe8BenwNYB\",\"decimals\":6,\"symbol\":\"USDT\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/solana/solana_Es9vMFrzaCERmJfrF4H2FYD4KCoNkY11McCe8BenwNYB.png\"},{\"chain\":\"Solana\",\"name\":\"USD Coin\",\"address\":\"EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v\",\"decimals\":6,\"symbol\":\"USDC\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/solana/solana_EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v.png\"},{\"chain\":\"Solana\",\"name\":\"Avalanche\",\"address\":\"7JnHPPJBBKSTJ7iEmsiGSBcPJgbcKw28uCRXtQgimncp\",\"decimals\":8,\"symbol\":\"AVAX\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/solana/solana_7JnHPPJBBKSTJ7iEmsiGSBcPJgbcKw28uCRXtQgimncp.png\"},{\"chain\":\"Solana\",\"name\":\"Frax\",\"address\":\"FR87nWEUxVgerFGhZM8Y4AggKGLnaXswr1Pd8wZ4kZcp\",\"decimals\":8,\"symbol\":\"FRAX\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/solana/solana_FR87nWEUxVgerFGhZM8Y4AggKGLnaXswr1Pd8wZ4kZcp.png\"},{\"chain\":\"Solana\",\"name\":\"WOO Network\",\"address\":\"E5rk3nmgLUuKUiS94gg4bpWwWwyjCMtddsAXkTFLtHEy\",\"decimals\":6,\"symbol\":\"WOO\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/solana/solana_E5rk3nmgLUuKUiS94gg4bpWwWwyjCMtddsAXkTFLtHEy.png\"},{\"chain\":\"Solana\",\"name\":\"STEPN\",\"address\":\"7i5KKsX2weiTkry7jA4ZwSuXGhs5eJBEjY8vVxR4pfRx\",\"decimals\":9,\"symbol\":\"GMT\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/solana/solana_7i5KKsX2weiTkry7jA4ZwSuXGhs5eJBEjY8vVxR4pfRx.png\"},{\"chain\":\"Solana\",\"name\":\"Sushi\",\"address\":\"ChVzxWRmrTeSgwd3Ui3UumcN8KX7VK3WaD4KGeSKpypj\",\"decimals\":8,\"symbol\":\"SUSHI\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/solana/solana_ChVzxWRmrTeSgwd3Ui3UumcN8KX7VK3WaD4KGeSKpypj.png\"},{\"chain\":\"Solana\",\"name\":\"Serum\",\"address\":\"SRMuApVNdxXokk5GT7XD5cUUgXMBCoAz2LHeuAoKWRt\",\"decimals\":6,\"symbol\":\"SRM\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/solana/solana_SRMuApVNdxXokk5GT7XD5cUUgXMBCoAz2LHeuAoKWRt.png\"},{\"chain\":\"Solana\",\"name\":\"Marinade staked SOL\",\"address\":\"mSoLzYCxHdYgdzU16g5QSh3i5K3z3KZK7ytfqcJm7So\",\"decimals\":9,\"symbol\":\"MSOL\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/solana/solana_mSoLzYCxHdYgdzU16g5QSh3i5K3z3KZK7ytfqcJm7So.png\"},{\"chain\":\"Solana\",\"name\":\"Coin98\",\"address\":\"C98A4nkJXhpVZNAZdHUA95RpTF3T4whtQubL3YobiUX9\",\"decimals\":6,\"symbol\":\"C98\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/solana/solana_C98A4nkJXhpVZNAZdHUA95RpTF3T4whtQubL3YobiUX9.png\"},{\"chain\":\"Solana\",\"name\":\"Lido Staked SOL\",\"address\":\"7dHbWXmci3dT8UFYWYZweBLXgycu7Y3iL6trKn1Y7ARj\",\"decimals\":9,\"symbol\":\"STSOL\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/solana/solana_7dHbWXmci3dT8UFYWYZweBLXgycu7Y3iL6trKn1Y7ARj.png\"},{\"chain\":\"Solana\",\"name\":\"Star Atlas DAO\",\"address\":\"poLisWXnNRwC6oBu1vHiuKQzFjGL4XDSu4g9qjz9qVk\",\"decimals\":8,\"symbol\":\"POLIS\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/solana/solana_poLisWXnNRwC6oBu1vHiuKQzFjGL4XDSu4g9qjz9qVk.png\"},{\"chain\":\"Solana\",\"name\":\"Raydium\",\"address\":\"4k3Dyjzvzp8eMZWUXbBCjEvwSkkk59S5iCNLY3QrkX6R\",\"decimals\":6,\"symbol\":\"RAY\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/solana/solana_4k3Dyjzvzp8eMZWUXbBCjEvwSkkk59S5iCNLY3QrkX6R.png\"},{\"chain\":\"Solana\",\"name\":\"MAI\",\"address\":\"9mWRABuz2x6koTPCWiCPM49WUbcrNqGTHBV9T9k7y1o7\",\"decimals\":9,\"symbol\":\"MIMATIC\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/solana/solana_9mWRABuz2x6koTPCWiCPM49WUbcrNqGTHBV9T9k7y1o7.png\"},{\"chain\":\"Solana\",\"name\":\"Star Atlas\",\"address\":\"ATLASXmbPQxBUYbxPsV97usA3fPQYEqzQBUHgiFCUsXx\",\"decimals\":8,\"symbol\":\"ATLAS\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/solana/solana_ATLASXmbPQxBUYbxPsV97usA3fPQYEqzQBUHgiFCUsXx.png\"},{\"chain\":\"Solana\",\"name\":\"Bonk\",\"address\":\"DezXAZ8z7PnrnRJjz3wXBoRgixCa6xjnB7YaB1pPB263\",\"decimals\":5,\"symbol\":\"BONK\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/solana/solana_DezXAZ8z7PnrnRJjz3wXBoRgixCa6xjnB7YaB1pPB263.png\"},{\"chain\":\"Solana\",\"name\":\"UXD Protocol\",\"address\":\"UXPhBoR3qG4UCiGNJfV7MqhHyFqKN68g45GoYvAeL2M\",\"decimals\":9,\"symbol\":\"UXP\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/solana/solana_UXPhBoR3qG4UCiGNJfV7MqhHyFqKN68g45GoYvAeL2M.png\"},{\"chain\":\"Solana\",\"name\":\"Bonfida\",\"address\":\"EchesyfXePKdLtoiZSL8pBe8Myagyy8ZRqsACNCFGnvp\",\"decimals\":6,\"symbol\":\"FIDA\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/solana/solana_EchesyfXePKdLtoiZSL8pBe8Myagyy8ZRqsACNCFGnvp.png\"},{\"chain\":\"Solana\",\"name\":\"Mango\",\"address\":\"MangoCzJ36AjZyKwVj3VnYU4GTonjfVEnJmvvWaxLac\",\"decimals\":6,\"symbol\":\"MNGO\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/solana/solana_MangoCzJ36AjZyKwVj3VnYU4GTonjfVEnJmvvWaxLac.png\"},{\"chain\":\"Solana\",\"name\":\"DEAPCOIN\",\"address\":\"BgwQjVNMWvt2d8CN51CsbniwRWyZ9H9HfHkEsvikeVuZ\",\"decimals\":6,\"symbol\":\"DEP\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/solana/solana_BgwQjVNMWvt2d8CN51CsbniwRWyZ9H9HfHkEsvikeVuZ.png\"}]",
		"Osmosis":   "[{\"chain\":\"Osmosis\",\"name\":\"usdc\",\"address\":\"ibc/D189335C6E4A68B513C10AB227BF1C1D38C746766278BA3EEB4FB14124F1D858\",\"decimals\":6,\"symbol\":\"axlUSDC\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/osmosis/osmosis_D189335C6E4A68B513C10AB227BF1C1D38C746766278BA3EEB4FB14124F1D858.png\"},{\"chain\":\"Osmosis\",\"name\":\"dai\",\"address\":\"ibc/F292A17CF920E3462C816CBE6B042E779F676CAB59096904C4C1C966413E3DF5\",\"decimals\":18,\"symbol\":\"DAI\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/osmosis/osmosis_F292A17CF920E3462C816CBE6B042E779F676CAB59096904C4C1C966413E3DF5.png\"},{\"chain\":\"Osmosis\",\"name\":\"uatom\",\"address\":\"ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2\",\"decimals\":6,\"symbol\":\"ATOM\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/osmosis/osmosis_27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2.png\"},{\"chain\":\"Osmosis\",\"name\":\"wbtc\",\"address\":\"ibc/D1542AA8762DB13087D8364F3EA6509FD6F009A34F00426AF9E4F9FA85CBBF1F\",\"decimals\":8,\"symbol\":\"axlWBTC\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/osmosis/osmosis_D1542AA8762DB13087D8364F3EA6509FD6F009A34F00426AF9E4F9FA85CBBF1F.png\"},{\"chain\":\"Osmosis\",\"name\":\"link\",\"address\":\"ibc/D3327A763C23F01EC43D1F0DB3CEFEC390C362569B6FD191F40A5192F8960049\",\"decimals\":18,\"symbol\":\"axlLINK\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/osmosis/osmosis_D3327A763C23F01EC43D1F0DB3CEFEC390C362569B6FD191F40A5192F8960049.png\"},{\"chain\":\"Osmosis\",\"name\":\"basecro\",\"address\":\"ibc/E6931F78057F7CC5DA0FD6CEF82FF39373A6E0452BF1FD76910B93292CF356C1\",\"decimals\":8,\"symbol\":\"CRO\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/osmosis/osmosis_E6931F78057F7CC5DA0FD6CEF82FF39373A6E0452BF1FD76910B93292CF356C1.png\"},{\"chain\":\"Osmosis\",\"name\":\"mkr\",\"address\":\"ibc/D27DDDF34BB47E5D5A570742CC667DE53277867116CCCA341F27785E899A70F3\",\"decimals\":18,\"symbol\":\"axlMKR\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/osmosis/osmosis_D27DDDF34BB47E5D5A570742CC667DE53277867116CCCA341F27785E899A70F3.png\"},{\"chain\":\"Osmosis\",\"name\":\"afet\",\"address\":\"ibc/5D1F516200EE8C6B2354102143B78A2DEDA25EDE771AC0F8DC3C1837C8FD4447\",\"decimals\":18,\"symbol\":\"FET\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/osmosis/osmosis_5D1F516200EE8C6B2354102143B78A2DEDA25EDE771AC0F8DC3C1837C8FD4447.png\"},{\"chain\":\"Osmosis\",\"name\":\"ukava\",\"address\":\"ibc/57AA1A70A4BC9769C525EBF6386F7A21536E04A79D62E1981EFCEF9428EBB205\",\"decimals\":6,\"symbol\":\"KAVA\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/osmosis/osmosis_57AA1A70A4BC9769C525EBF6386F7A21536E04A79D62E1981EFCEF9428EBB205.png\"},{\"chain\":\"Osmosis\",\"name\":\"inj\",\"address\":\"ibc/64BA6E31FE887D66C6F8F31C7B1A80C7CA179239677B4088BB55F5EA07DBE273\",\"decimals\":18,\"symbol\":\"INJ\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/osmosis/osmosis_64BA6E31FE887D66C6F8F31C7B1A80C7CA179239677B4088BB55F5EA07DBE273.png\"},{\"chain\":\"Osmosis\",\"name\":\"uband\",\"address\":\"ibc/F867AE2112EFE646EC71A25CD2DFABB8927126AC1E19F1BBF0FF693A4ECA05DE\",\"decimals\":6,\"symbol\":\"BAND\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/osmosis/osmosis_F867AE2112EFE646EC71A25CD2DFABB8927126AC1E19F1BBF0FF693A4ECA05DE.png\"},{\"chain\":\"Osmosis\",\"name\":\"aevmos\",\"address\":\"ibc/6AE98883D4D5D5FF9E50D7130F1305DA2FFA0C652D1DD9C123657C6B4EB2DF8A\",\"decimals\":18,\"symbol\":\"EVMOS\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/osmosis/osmosis_6AE98883D4D5D5FF9E50D7130F1305DA2FFA0C652D1DD9C123657C6B4EB2DF8A.png\"},{\"chain\":\"Osmosis\",\"name\":\"uscrt\",\"address\":\"ibc/0954E1C28EB7AF5B72D24F3BC2B47BBB2FDF91BDDFD57B74B99E133AED40972A\",\"decimals\":6,\"symbol\":\"SCRT\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/osmosis/osmosis_0954E1C28EB7AF5B72D24F3BC2B47BBB2FDF91BDDFD57B74B99E133AED40972A.png\"},{\"chain\":\"Osmosis\",\"name\":\"uaxl\",\"address\":\"ibc/903A61A498756EA560B85A85132D3AEE21B5DEDD41213725D22ABF276EA6945E\",\"decimals\":6,\"symbol\":\"AXL\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/osmosis/osmosis_903A61A498756EA560B85A85132D3AEE21B5DEDD41213725D22ABF276EA6945E.png\"},{\"chain\":\"Osmosis\",\"name\":\"uakt\",\"address\":\"ibc/1480B8FD20AD5FCAE81EA87584D269547DD4D436843C1D20F15E00EB64743EF4\",\"decimals\":6,\"symbol\":\"AKT\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/osmosis/osmosis_1480B8FD20AD5FCAE81EA87584D269547DD4D436843C1D20F15E00EB64743EF4.png\"},{\"chain\":\"Osmosis\",\"name\":\"umed\",\"address\":\"ibc/3BCCC93AD5DF58D11A6F8A05FA8BC801CBA0BA61A981F57E91B8B598BF8061CB\",\"decimals\":6,\"symbol\":\"MED\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/osmosis/osmosis_3BCCC93AD5DF58D11A6F8A05FA8BC801CBA0BA61A981F57E91B8B598BF8061CB.png\"},{\"chain\":\"Osmosis\",\"name\":\"ujuno\",\"address\":\"ibc/46B44899322F3CD854D2D46DEEF881958467CDD4B3B10086DA49296BBED94BED\",\"decimals\":6,\"symbol\":\"JUNO\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/osmosis/osmosis_46B44899322F3CD854D2D46DEEF881958467CDD4B3B10086DA49296BBED94BED.png\"},{\"chain\":\"Osmosis\",\"name\":\"uxprt\",\"address\":\"ibc/A0CC0CF735BFB30E730C70019D4218A1244FF383503FF7579C9201AB93CA9293\",\"decimals\":6,\"symbol\":\"XPRT\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/osmosis/osmosis_A0CC0CF735BFB30E730C70019D4218A1244FF383503FF7579C9201AB93CA9293.png\"},{\"chain\":\"Osmosis\",\"name\":\"ukuji\",\"address\":\"ibc/BB6BCDB515050BAE97516111873CCD7BCF1FD0CCB723CC12F3C4F704D6C646CE\",\"decimals\":6,\"symbol\":\"KUJI\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/osmosis/osmosis_BB6BCDB515050BAE97516111873CCD7BCF1FD0CCB723CC12F3C4F704D6C646CE.png\"},{\"chain\":\"Osmosis\",\"name\":\"uctk\",\"address\":\"ibc/7ED954CFFFC06EE8419387F3FC688837FF64EF264DE14219935F724EEEDBF8D3\",\"decimals\":6,\"symbol\":\"CTK\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/osmosis/osmosis_7ED954CFFFC06EE8419387F3FC688837FF64EF264DE14219935F724EEEDBF8D3.png\"}]",
		"Harmony":   "[{\"chain\":\"Harmony\",\"name\":\"Tether\",\"address\":\"0x3c2b8be99c50593081eaa2a724f0b8285f5aba8f\",\"decimals\":6,\"symbol\":\"USDT\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/harmony-shard-0/harmony-shard-0_0x3c2b8be99c50593081eaa2a724f0b8285f5aba8f.png\"},{\"chain\":\"Harmony\",\"name\":\"USD Coin\",\"address\":\"0x985458e523db3d53125813ed68c274899e9dfab4\",\"decimals\":6,\"symbol\":\"USDC\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/harmony-shard-0/harmony-shard-0_0x985458e523db3d53125813ed68c274899e9dfab4.png\"},{\"chain\":\"Harmony\",\"name\":\"Binance USD\",\"address\":\"0xe176ebe47d621b984a73036b9da5d834411ef734\",\"decimals\":18,\"symbol\":\"BUSD\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/harmony-shard-0/harmony-shard-0_0xe176ebe47d621b984a73036b9da5d834411ef734.png\"},{\"chain\":\"Harmony\",\"name\":\"Polygon\",\"address\":\"0x301259f392b551ca8c592c9f676fcd2f9a0a84c5\",\"decimals\":18,\"symbol\":\"MATIC\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/harmony-shard-0/harmony-shard-0_0x301259f392b551ca8c592c9f676fcd2f9a0a84c5.png\"},{\"chain\":\"Harmony\",\"name\":\"Dai\",\"address\":\"0xef977d2f931c1978db5f6747666fa1eacb0d0339\",\"decimals\":18,\"symbol\":\"DAI\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/harmony-shard-0/harmony-shard-0_0xef977d2f931c1978db5f6747666fa1eacb0d0339.png\"},{\"chain\":\"Harmony\",\"name\":\"Uniswap\",\"address\":\"0x90d81749da8867962c760414c1c25ec926e889b6\",\"decimals\":18,\"symbol\":\"UNI\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/harmony-shard-0/harmony-shard-0_0x90d81749da8867962c760414c1c25ec926e889b6.png\"},{\"chain\":\"Harmony\",\"name\":\"Wrapped Bitcoin\",\"address\":\"0x3095c7557bcb296ccc6e363de01b760ba031f2d9\",\"decimals\":8,\"symbol\":\"WBTC\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/harmony-shard-0/harmony-shard-0_0x3095c7557bcb296ccc6e363de01b760ba031f2d9.png\"},{\"chain\":\"Harmony\",\"name\":\"Chainlink\",\"address\":\"0x218532a12a389a4a92fc0c5fb22901d1c19198aa\",\"decimals\":18,\"symbol\":\"LINK\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/harmony-shard-0/harmony-shard-0_0x218532a12a389a4a92fc0c5fb22901d1c19198aa.png\"},{\"chain\":\"Harmony\",\"name\":\"Cronos\",\"address\":\"0x2672b791d23879995aabdf51bc7d3df54bb4e266\",\"decimals\":8,\"symbol\":\"CRO\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/harmony-shard-0/harmony-shard-0_0x2672b791d23879995aabdf51bc7d3df54bb4e266.png\"},{\"chain\":\"Harmony\",\"name\":\"The Graph\",\"address\":\"0x002fa662f2e09de7c306d2bab0085ee9509488ff\",\"decimals\":18,\"symbol\":\"GRT\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/harmony-shard-0/harmony-shard-0_0x002fa662f2e09de7c306d2bab0085ee9509488ff.png\"},{\"chain\":\"Harmony\",\"name\":\"Fantom\",\"address\":\"0x39ab439897380ed10558666c4377facb0322ad48\",\"decimals\":18,\"symbol\":\"FTM\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/harmony-shard-0/harmony-shard-0_0x39ab439897380ed10558666c4377facb0322ad48.png\"},{\"chain\":\"Harmony\",\"name\":\"The Sandbox\",\"address\":\"0x35de8649e1e4fd1a7bd3b14f7e24e5e7887174ed\",\"decimals\":18,\"symbol\":\"SAND\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/harmony-shard-0/harmony-shard-0_0x35de8649e1e4fd1a7bd3b14f7e24e5e7887174ed.png\"},{\"chain\":\"Harmony\",\"name\":\"Aave\",\"address\":\"0xcf323aad9e522b93f11c352caa519ad0e14eb40f\",\"decimals\":18,\"symbol\":\"AAVE\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/harmony-shard-0/harmony-shard-0_0xcf323aad9e522b93f11c352caa519ad0e14eb40f.png\"},{\"chain\":\"Harmony\",\"name\":\"Axie Infinity\",\"address\":\"0x14a7b318fed66ffdcc80c1517c172c13852865de\",\"decimals\":18,\"symbol\":\"AXS\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/harmony-shard-0/harmony-shard-0_0x14a7b318fed66ffdcc80c1517c172c13852865de.png\"},{\"chain\":\"Harmony\",\"name\":\"Frax\",\"address\":\"0xfa7191d292d5633f702b0bd7e3e3bccc0e633200\",\"decimals\":18,\"symbol\":\"FRAX\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/harmony-shard-0/harmony-shard-0_0xfa7191d292d5633f702b0bd7e3e3bccc0e633200.png\"},{\"chain\":\"Harmony\",\"name\":\"TrueUSD\",\"address\":\"0x553a1151f3df3620fc2b5a75a6edda629e3da350\",\"decimals\":18,\"symbol\":\"TUSD\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/harmony-shard-0/harmony-shard-0_0x553a1151f3df3620fc2b5a75a6edda629e3da350.png\"},{\"chain\":\"Harmony\",\"name\":\"Huobi\",\"address\":\"0xbaa0974354680b0e8146d64bb27fb92c03c4a2f2\",\"decimals\":18,\"symbol\":\"HT\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/harmony-shard-0/harmony-shard-0_0xbaa0974354680b0e8146d64bb27fb92c03c4a2f2.png\"},{\"chain\":\"Harmony\",\"name\":\"Synthetix Network\",\"address\":\"0x7b9c523d59aefd362247bd5601a89722e3774dd2\",\"decimals\":18,\"symbol\":\"SNX\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/harmony-shard-0/harmony-shard-0_0x7b9c523d59aefd362247bd5601a89722e3774dd2.png\"},{\"chain\":\"Harmony\",\"name\":\"Frax Share\",\"address\":\"0x0767d8e1b05efa8d6a301a65b324b6b66a1cc14c\",\"decimals\":18,\"symbol\":\"FXS\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/harmony-shard-0/harmony-shard-0_0x0767d8e1b05efa8d6a301a65b324b6b66a1cc14c.png\"},{\"chain\":\"Harmony\",\"name\":\"PAX Gold\",\"address\":\"0x7afb0e2eba6dc938945fe0f42484d3b8f442d0ac\",\"decimals\":18,\"symbol\":\"PAXG\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/harmony-shard-0/harmony-shard-0_0x7afb0e2eba6dc938945fe0f42484d3b8f442d0ac.png\"}]",
		"HECO":      "[{\"chain\":\"HECO\",\"name\":\"Tether\",\"address\":\"0xa71edc38d189767582c38a3145b5873052c3e47a\",\"decimals\":18,\"symbol\":\"USDT\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/huobi-token/huobi-token_0xa71edc38d189767582c38a3145b5873052c3e47a.png\"},{\"chain\":\"HECO\",\"name\":\"Uniswap\",\"address\":\"0x22c54ce8321a4015740ee1109d9cbc25815c46e6\",\"decimals\":18,\"symbol\":\"UNI\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/huobi-token/huobi-token_0x22c54ce8321a4015740ee1109d9cbc25815c46e6.png\"},{\"chain\":\"HECO\",\"name\":\"Chainlink\",\"address\":\"0x9e004545c59d359f6b7bfb06a26390b087717b42\",\"decimals\":18,\"symbol\":\"LINK\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/huobi-token/huobi-token_0x9e004545c59d359f6b7bfb06a26390b087717b42.png\"},{\"chain\":\"HECO\",\"name\":\"Aave\",\"address\":\"0x202b4936fe1a82a4965220860ae46d7d3939bb25\",\"decimals\":18,\"symbol\":\"AAVE\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/huobi-token/huobi-token_0x202b4936fe1a82a4965220860ae46d7d3939bb25.png\"},{\"chain\":\"HECO\",\"name\":\"Synthetix Network\",\"address\":\"0x777850281719d5a96c29812ab72f822e0e09f3da\",\"decimals\":18,\"symbol\":\"SNX\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/huobi-token/huobi-token_0x777850281719d5a96c29812ab72f822e0e09f3da.png\"},{\"chain\":\"HECO\",\"name\":\"WOO Network\",\"address\":\"0x3befb2308bce92da97264077faf37dcd6c8a75e6\",\"decimals\":18,\"symbol\":\"WOO\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/huobi-token/huobi-token_0x3befb2308bce92da97264077faf37dcd6c8a75e6.png\"},{\"chain\":\"HECO\",\"name\":\"Balancer\",\"address\":\"0x045de15ca76e76426e8fc7cba8392a3138078d0f\",\"decimals\":18,\"symbol\":\"BAL\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/huobi-token/huobi-token_0x045de15ca76e76426e8fc7cba8392a3138078d0f.png\"},{\"chain\":\"HECO\",\"name\":\"yearn.finance\",\"address\":\"0xb4f019beac758abbee2f906033aaa2f0f6dacb35\",\"decimals\":18,\"symbol\":\"YFI\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/huobi-token/huobi-token_0xb4f019beac758abbee2f906033aaa2f0f6dacb35.png\"},{\"chain\":\"HECO\",\"name\":\"Huobi BTC\",\"address\":\"0x66a79d23e58475d2738179ca52cd0b41d73f0bea\",\"decimals\":18,\"symbol\":\"HBTC\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/huobi-token/huobi-token_0x66a79d23e58475d2738179ca52cd0b41d73f0bea.png\"},{\"chain\":\"HECO\",\"name\":\"Mdex\",\"address\":\"0x25d2e80cb6b86881fd7e07dd263fb79f4abe033c\",\"decimals\":18,\"symbol\":\"MDX\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/huobi-token/huobi-token_0x25d2e80cb6b86881fd7e07dd263fb79f4abe033c.png\"},{\"chain\":\"HECO\",\"name\":\"Beefy.Finance\",\"address\":\"0x765277eebeca2e31912c9946eae1021199b39c61\",\"decimals\":18,\"symbol\":\"BIFI\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/huobi-token/huobi-token_0x765277eebeca2e31912c9946eae1021199b39c61.png\"},{\"chain\":\"HECO\",\"name\":\"HUSD\",\"address\":\"0x0298c2b32eae4da002a15f36fdf7615bea3da047\",\"decimals\":8,\"symbol\":\"HUSD\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/huobi-token/huobi-token_0x0298c2b32eae4da002a15f36fdf7615bea3da047.png\"},{\"chain\":\"HECO\",\"name\":\"Pocket TPT\",\"address\":\"0x9ef1918a9bee17054b35108bd3e2665e2af1bb1b\",\"decimals\":4,\"symbol\":\"TPT\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/huobi-token/huobi-token_0x9ef1918a9bee17054b35108bd3e2665e2af1bb1b.png\"},{\"chain\":\"HECO\",\"name\":\"Elastos\",\"address\":\"0xa1ecfc2bec06e4b43ddd423b94fef84d0dbc8f5c\",\"decimals\":18,\"symbol\":\"ELA\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/huobi-token/huobi-token_0xa1ecfc2bec06e4b43ddd423b94fef84d0dbc8f5c.png\"},{\"chain\":\"HECO\",\"name\":\"FUSION\",\"address\":\"0xa790b07796abed3cdaf81701b4535014bf5e1a65\",\"decimals\":18,\"symbol\":\"FSN\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/huobi-token/huobi-token_0xa790b07796abed3cdaf81701b4535014bf5e1a65.png\"},{\"chain\":\"HECO\",\"name\":\"Deri Protocol\",\"address\":\"0x2bda3e331cf735d9420e41567ab843441980c4b8\",\"decimals\":18,\"symbol\":\"DERI\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/huobi-token/huobi-token_0x2bda3e331cf735d9420e41567ab843441980c4b8.png\"},{\"chain\":\"HECO\",\"name\":\"TigerCash\",\"address\":\"0x5ecc4b299e23f526980c33fe35eff531a54aedb1\",\"decimals\":18,\"symbol\":\"TCH\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/huobi-token/huobi-token_0x5ecc4b299e23f526980c33fe35eff531a54aedb1.png\"},{\"chain\":\"HECO\",\"name\":\"Minto\",\"address\":\"0x410a56541bd912f9b60943fcb344f1e3d6f09567\",\"decimals\":18,\"symbol\":\"BTCMT\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/huobi-token/huobi-token_0x410a56541bd912f9b60943fcb344f1e3d6f09567.png\"},{\"chain\":\"HECO\",\"name\":\"O3 Swap\",\"address\":\"0xee9801669c6138e84bd50deb500827b776777d28\",\"decimals\":18,\"symbol\":\"O3\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/huobi-token/huobi-token_0xee9801669c6138e84bd50deb500827b776777d28.png\"},{\"chain\":\"HECO\",\"name\":\"Xend Finance\",\"address\":\"0xa649325aa7c5093d12d6f98eb4378deae68ce23f\",\"decimals\":18,\"symbol\":\"XEND\",\"logoURI\":\"https://obstatic.243096.com/download/token/images/huobi-token/huobi-token_0xa649325aa7c5093d12d6f98eb4378deae68ce23f.png\"}]",
	}
	for chain, value := range top20 {
		key := REDIS_TOP20_KEY + chain
		if err := utils.SetTokenTop20RedisKey(c.redisClient, key, value); err != nil {
			c.log.Error("set redis client cache error:", err)
		}
	}
}

func GetCGTokenListByChain(chain string) ([]models.TokenList, error) {
	var tokenLists []models.TokenList
	err := c.db.Where("chain = ? and cg_id != ?", chain, "").Find(&tokenLists).Error
	if err != nil {
		c.log.Error("get token list error:", err)
	}
	return tokenLists, err
}
