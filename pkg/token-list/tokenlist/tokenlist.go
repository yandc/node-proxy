package tokenlist

import (
	"encoding/json"
	"fmt"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-redis/redis"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"io/fs"
	v1 "node-proxy/api/tokenlist/v1"
	"node-proxy/internal/conf"
	"node-proxy/internal/data/models"
	"node-proxy/pkg/token-list/types"
	"node-proxy/pkg/token-list/utils"
	"os"
	"path/filepath"
	"sync/atomic"
	"time"

	"strings"
	"sync"
)

type config struct {
	db          *gorm.DB
	log         *log.Helper
	redisClient *redis.Client
	logoPrefix  string
}

const REDIS_PRICE_KEY = "tokenlist:price:"
const REDIS_LIST_KEY = "tokenlsit:list:"

var c config

func InitTokenList(conf *conf.TokenList, db *gorm.DB, client *redis.Client, logger log.Logger) {
	log := log.NewHelper(log.With(logger, "module", "tokenlist/InittokenList"))
	c = config{
		db:          db,
		log:         log,
		redisClient: client,
		logoPrefix:  conf.LogoPrefix,
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
				var address string
				if strings.HasPrefix(p.ContractAddress, "0x") {
					address = strings.ToLower(p.ContractAddress)
				} else {
					address = p.ContractAddress
				}

				chain := utils.GetPlatform(p.Platform.Coin.Slug)
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
			chain := strings.Split(key, ":")[0]
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
		chainMap := make(map[string]string)
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
			for _, cm := range cmLists {
				tempChain = append(tempChain, cm.Id)
				if cm.Name == "Huobi" {
					chainMap[cm.Id] = "Huobi-Token"
				} else {
					chainMap[cm.Id] = cm.Name
				}
			}
		} else {
			tempChain = append(tempChain, chains...)
		}
		needUpdateChain := make([]string, 0, len(chains))
		for _, chain := range tempChain {
			//get id by chain
			key := REDIS_PRICE_KEY + fmt.Sprintf("%s:%s", chain, strings.ToLower(currency))
			price, err := c.redisClient.Get(key).Result()
			if err != nil {
				c.log.Error("get redis cache error:", err, key)
			}
			if price != "" {
				var resultKey string

				if value, ok := chainMap[chain]; ok {
					resultKey = value
				} else {
					resultKey = chain
				}

				result[resultKey] = map[string]string{
					currency: price,
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
			for address, prices := range pricesMap {
				var p string
				if value, ok := prices[strings.ToLower(currency)]; ok {
					p = decimal.NewFromFloat32(value).String()
				}
				tempKey := REDIS_PRICE_KEY + fmt.Sprintf("%s:%s", address, strings.ToLower(currency))
				err := c.redisClient.Set(tempKey, p, 1*time.Minute).Err()
				if err != nil {
					c.log.Error("set redisClient cache error:", err, tempKey)
				}
				var resultKey string

				if value, ok := chainMap[address]; ok {
					resultKey = value

				} else {
					resultKey = address
				}

				result[resultKey] = map[string]string{
					currency: p,
				}
			}
		}
	}
	if len(addresses) > 0 {
		//token price
		newAddressMap := utils.ParseCoinAddress(addresses)
		//addressMap := make(map[string]string, len(addresses))
		needUpdateAddress := make([]string, 0, len(addresses))
		for chainAddress, _ := range newAddressMap {
			key := REDIS_PRICE_KEY + fmt.Sprintf("%s:%s", chainAddress, strings.ToLower(currency))
			price, err := c.redisClient.Get(key).Result()
			if err != nil {
				c.log.Error("get redis cache error:", err, key)
			}
			if price != "" {
				result[newAddressMap[chainAddress]] = map[string]string{currency: price}
			} else {
				fmt.Println("add address", chainAddress)
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
						result[newAddressMap[chainAddress]] = map[string]string{currency: "0"}
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
				}
				for id, prices := range cgPricesMap {
					var price string
					if value, ok := prices[strings.ToLower(currency)]; ok {
						price = decimal.NewFromFloat32(value).String()
					}
					address := addressIdMap[id]
					key := REDIS_PRICE_KEY + fmt.Sprintf("%s:%s", address, strings.ToLower(currency))
					err := c.redisClient.Set(key, price, 1*time.Minute).Err()
					if err != nil {
						c.log.Error("set redisClient error:", err, key)
					}
					if value, ok := newAddressMap[address]; ok {
						result[value] = map[string]string{currency: price}
					} else {
						result[address] = map[string]string{currency: price}
					}
				}
			}

			if len(cmcIds) > 0 {
				cmcPriceMap, err := CMCSimplePrice(cmcIds, currency)
				if err != nil {
					c.log.Error("get cmc price error:", err)
				}
				for id, prices := range cmcPriceMap {
					address := addressIdMap[id]
					var price string
					if value, ok := prices[currency]; ok {
						price = decimal.NewFromFloat32(value).String()
					}
					key := REDIS_PRICE_KEY + fmt.Sprintf("%s:%s", address, strings.ToLower(currency))
					err := c.redisClient.Set(key, price, 1*time.Minute).Err()
					if err != nil {
						c.log.Error("set redisClient error:", err, key)
					}
					if value, ok := newAddressMap[address]; ok {
						result[value] = map[string]string{currency: price}
					} else {
						result[address] = map[string]string{currency: price}
					}
				}
			}

		}
	}

	return result
}

func GetTokenList(chain string) ([]*v1.GetTokenListResp_Data, error) {
	chain = utils.GetChainNameByChain(chain)
	key := REDIS_LIST_KEY + chain
	str, err := c.redisClient.Get(key).Result()
	if err != nil {
		c.log.Error("get token list key error:", err, key)
	}
	var result []*v1.GetTokenListResp_Data
	if str != "" {
		err := json.Unmarshal([]byte(str), &result)
		if err != nil {
			c.log.Error("unmarshal error:", err)
		}
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
	err = c.redisClient.Set(key, string(b), 24*time.Hour).Err()
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
			if strings.HasPrefix(t.Address, "0x") {
				t.Address = strings.ToLower(t.Address)
			}
			result[chain+":"+t.Address] = t
		}
	}
	return result
}

func DownLoadImages() {
	var tokenLists []models.TokenList
	err := c.db.Find(&tokenLists).Error
	if err != nil {
		c.log.Error("find token list error:", err)
	}
	var wg sync.WaitGroup
	var count int32 = 0
	for _, t := range tokenLists {
		var image string
		if t.CgId != "" && t.CmcId == 0 {
			var cgImage map[string]string
			json.Unmarshal([]byte(t.Logo), &cgImage)
			if value, ok := cgImage["small"]; ok {
				image = value
			}
		} else {
			image = t.Logo
		}
		if image != "" {
			path := "./images/" + t.Chain
			exist, _ := utils.PathExists(path)
			if !exist {
				os.MkdirAll(path, 0777)
			}
			fileName := path + "/" + t.Chain + "_" + t.Address + ".png"
			wg.Add(1)
			go func() {
				defer wg.Done()
				err := utils.DownLoad(fileName, image)
				if err == nil {
					atomic.AddInt32(&count, 1)
				}
			}()

		}
	}
	wg.Wait()
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
