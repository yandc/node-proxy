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
	v1 "gitlab.bixin.com/mili/node-proxy/api/tokenlist/v1"
	"gitlab.bixin.com/mili/node-proxy/internal/conf"
	"gitlab.bixin.com/mili/node-proxy/internal/data/models"
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
	REDIS_PRICE_KEY = "tokenlist:price:"
	REDIS_LIST_KEY  = "tokenlist:list:"
	REDIS_TOKEN_KEY = "tokenlist:tokeninfo:"
	REDIS_TOP20_KEY = "tokenlist:top20:"
)

var c config

func InitTokenList(conf *conf.TokenList, db *gorm.DB, client *redis.Client, logger log.Logger) {
	log := log.NewHelper(log.With(logger, "module", "tokenlist/InittokenList"))
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
				tempChainMap[cm.Id] = struct{}{}
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
		needUpdateChain := make([]string, 0, len(chains))
		for _, chain := range tempChain {
			//get id by chain
			key := REDIS_PRICE_KEY + fmt.Sprintf("%s:%s", chain, strings.ToLower(currency))
			//price, err := c.redisClient.Get(key).Result()
			price, updateFlag, err := utils.GetPriceRedisValueByKey(c.redisClient, key)
			if err != nil {
				c.log.Error("get redis cache error:", err, key)
			}
			if price != "" {
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
				if updateFlag {
					needUpdateChain = append(needUpdateChain, chain)
				}
			} else {
				needUpdateChain = append(needUpdateChain, chain)
			}
		}
		if len(needUpdateChain) > 0 {
			pricesMap, err := CGSimplePrice(needUpdateChain, currency)
			if err != nil {
				c.log.Error("CGSimplePrice Error:", err)
			}
			handlerPriceMap(pricesMap, map[string]string{}, chainMap, currency, result, true)
		}
	}
	if len(addresses) > 0 {
		//token price
		newAddressMap := utils.ParseCoinAddress(addresses)
		//addressMap := make(map[string]string, len(addresses))
		needUpdateAddress := make([]string, 0, len(addresses))
		for chainAddress, _ := range newAddressMap {
			key := REDIS_PRICE_KEY + fmt.Sprintf("%s:%s", chainAddress, strings.ToLower(currency))
			//price, err := c.redisClient.Get(key).Result()
			price, updateFlag, err := utils.GetPriceRedisValueByKey(c.redisClient, key)
			if err != nil {
				c.log.Error("get redis cache error:", err, key)
			}
			if price != "" {
				for _, addr := range newAddressMap[chainAddress] {
					result[addr] = map[string]string{currency: price}
				}

				if updateFlag {
					needUpdateAddress = append(needUpdateAddress, chainAddress)
				}
			} else {
				//fmt.Println("add address", chainAddress)
				needUpdateAddress = append(needUpdateAddress, chainAddress)
				//addressMap[address] = addresses[i]
			}
		}
		if len(needUpdateAddress) > 0 {
			//get redisClient price
			tokenList := make([]models.TokenList, 0, len(needUpdateAddress))
			for _, chainAddress := range needUpdateAddress {
				if strings.Contains(chainAddress, "_") {
					addressInfo := strings.Split(chainAddress, "_")
					chain, address := addressInfo[0], addressInfo[1]
					var tempTokenList models.TokenList
					err := c.db.Where("chain = ? AND address = ?", chain, address).First(&tempTokenList).Error
					if err != nil {
						c.log.Error("get token list error:", err)
						for _, addr := range newAddressMap[chainAddress] {
							result[addr] = map[string]string{currency: "0"}
						}
						continue
					}
					//addressMap[tempTokenList.CgId] = chainAddress
					tokenList = append(tokenList, tempTokenList)
				}
			}
			var cgIds, cmcIds []string
			addressIdMap := make(map[string]string, len(tokenList))
			for _, t := range tokenList {
				var id string
				if t.CgId != "" {
					id = t.CgId
					cgIds = append(cgIds, id)
				} else if t.CmcId > 0 {
					id = fmt.Sprintf("%d", t.CmcId)
					cmcIds = append(cmcIds, id)
				}
				addressIdMap[id] = fmt.Sprintf("%s_%s", t.Chain, t.Address)
			}
			//coin gecko price
			if len(cgIds) > 0 {
				cgPricesMap, err := CGSimplePrice(cgIds, currency)
				if err != nil {
					c.log.Error("CGSimplePrice Error:", err)
				} else {
					handlerPriceMap(cgPricesMap, addressIdMap, newAddressMap, currency, result, true)

				}
			}

			if len(cmcIds) > 0 {
				cmcPriceMap, err := CMCSimplePrice(cmcIds, currency)
				if err != nil {
					c.log.Error("get cmc price error:", err)
				} else {
					handlerPriceMap(cmcPriceMap, addressIdMap, newAddressMap, currency, result, false)
				}
			}

		}
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
		address := addressInfo.Address
		if strings.HasPrefix(addressInfo.Address, "0x") && chain != utils.STARCOIN_CHAIN &&
			chain != utils.APTOS_CHAIN {
			address = strings.ToLower(addressInfo.Address)
		}
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
		address := addressInfo.Address

		if (strings.HasPrefix(addressInfo.Address, "0x") && chain != utils.STARCOIN_CHAIN &&
			chain != utils.APTOS_CHAIN) || (strings.Contains(chain, utils.COSMOS_CHAIN)) {
			address = strings.ToLower(addressInfo.Address)
		}
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

func RefreshLogoURI(chains []string) {
	//get all token list
	var tokenLists []models.TokenList
	var err error
	if len(chains) > 0 {
		err = c.db.Where("chain in ?", chains).Find(&tokenLists).Error
	} else {
		tokenLists, err = GetAllTokenList()
	}
	//err := c.db.Where("address = ?", "0x75231f58b43240c9718dd58b4967c5114342a86c").Find(&tokenLists).Error
	if err != nil {
		c.log.Error("get token list error:", err)
		return
	}
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
			json.Unmarshal([]byte(t.Logo), &cgImage)
			if value, ok := cgImage["large"]; ok {
				image = value
			} else {
				if strings.Contains(t.Logo, "/thumb/") {
					image = strings.Replace(t.Logo, "/thumb/", "/large/", 1)
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
			if t.Chain == "osmosis" && strings.Contains(t.Address, "/") {
				address = strings.Split(t.Address, "/")[1]
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
	var wg sync.WaitGroup
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
		for _, p := range paths {
			wg.Add(1)
			go func(path string) {
				defer wg.Done()
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
			}(p)
		}
	}
	wg.Wait()
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
			chainAddress := strings.Split(fileName, "_")
			if len(chainAddress) == 2 {
				chain := chainAddress[0]
				address := chainAddress[1]
				if chain == "osmosis" && len(address) == 64 {
					address = fmt.Sprintf("ibc/%s", address)
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
					ret, err := s3.New(sess).PutObject(&s3.PutObjectInput{
						Bucket: aws.String(awsInfo.Bucket), //桶名
						Key:    aws.String(key),
						Body:   bytes.NewReader(buffer),
					})
					if err != nil {
						c.log.Error("put file to s3 error:", err)
					}
					c.log.Info("upload s3 info:", ret)
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
	if !updateFlag {
		return result, nil
	}
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
		b, _ := json.Marshal(result)
		if err := utils.SetTokenTop20RedisKey(c.redisClient, key, string(b)); err != nil {
			c.log.Error("set redis client cache error:", err)
		}
		return result, nil
	}
	return result, nil
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
