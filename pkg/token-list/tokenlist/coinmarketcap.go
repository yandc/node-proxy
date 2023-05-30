package tokenlist

import (
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"math/rand"
	"strings"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"gitlab.bixin.com/mili/node-proxy/internal/data/models"
	"gitlab.bixin.com/mili/node-proxy/pkg/token-list/types"
	"gitlab.bixin.com/mili/node-proxy/pkg/token-list/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type cmcConf struct {
	baseUrl string
	key     []string
	db      *gorm.DB
	log     *log.Helper
}

var cmc = cmcConf{
	baseUrl: "https://pro-api.coinmarketcap.com/v1",
	key:     []string{"4a77c975-64ff-4c2e-8f70-b195bfb07be5"},
}

func InitCMC(baseURL string, key []string, db *gorm.DB, logger log.Logger) {
	log := log.NewHelper(log.With(logger, "module", "tokenlist/InitCMC"))
	cmc = cmcConf{
		baseUrl: baseURL,
		key:     key,
		db:      db,
		log:     log,
	}
}

func CMCCoinsList() (*types.CMCList, error) {
	url := fmt.Sprintf("%s/cryptocurrency/map", cmc.baseUrl)
	index := rand.Intn(len(cmc.key))
	params := map[string]string{
		"CMC_PRO_API_KEY": cmc.key[index],
	}
	out := &types.CMCList{}
	err := utils.HttpsGetForm(url, params, out)
	if err != nil {
		for i := 0; i < len(cmc.key) && err != nil; i++ {
			params["CMC_PRO_API_KEY"] = cmc.key[i]
			err = utils.HttpsGetForm(url, params, out)
		}
	}
	return out, err
}

func CMCCoinsId(id string) (*types.CMCCoinsID, error) {
	if len(id) == 0 {
		return nil, fmt.Errorf("id is required")
	}
	url := fmt.Sprintf("%s/cryptocurrency/info", cmc.baseUrl)
	index := rand.Intn(len(cmc.key))
	params := map[string]string{
		"CMC_PRO_API_KEY": cmc.key[index],
		"id":              id,
	}
	out := &types.CMCCoinsID{}
	err := utils.HttpsGetForm(url, params, out)
	if err != nil {
		for i := 0; i < len(cmc.key) && err != nil; i++ {
			params["CMC_PRO_API_KEY"] = cmc.key[i]
			err = utils.HttpsGetForm(url, params, out)
		}
	}
	return out, err
}

func GetCoinsId(ids []types.CMCListItem) string {
	var id strings.Builder
	for i := 0; i < len(ids); i++ {
		id.WriteString(fmt.Sprintf("%d", ids[i].ID))
		id.WriteString(",")
	}
	return id.String()[:id.Len()-1]
}

func CreateCoinMarketCap() {
	//fmt.Println("CreateCoinMarketCap start")
	coinsList, err := CMCCoinsList()
	if err != nil {
		//log.Error("CreateCoinMarketCap coinsList error:", zap.Error(err))
		for err != nil {
			coinsList, err = CMCCoinsList()
			time.Sleep(1 * time.Minute)
		}
	}
	//log.Info("coinsList length", zap.Int("length", len(coinsList.Data)))
	var allIds []string
	pageEndIndex := 0
	pageSize := 800
	for i := 0; i < len(coinsList.Data); i += pageSize {
		if i+pageSize >= len(coinsList.Data) {
			pageEndIndex = len(coinsList.Data)
		} else {
			pageEndIndex = i + pageSize
		}
		allIds = append(allIds, GetCoinsId(coinsList.Data[i:pageEndIndex]))
	}
	//log.Info("allIds", zap.Int("length", len(allIds)))
	for _, ids := range allIds {
		coinsId, err := CMCCoinsId(ids)
		if err != nil {
			//log.Error("coinsId error:", zap.Error(err))
			for err != nil {
				time.Sleep(1 * time.Minute)
				coinsId, err = CMCCoinsId(ids)
			}
		}
		if coinsId != nil {
			coinMarketCaps := make([]models.CoinMarketCap, 0, len(coinsId.Data))
			for _, value := range coinsId.Data {
				contractAddress, _ := json.Marshal(value.ContractAddress)
				twitter, _ := json.Marshal(value.Urls.Twitter)
				website, _ := json.Marshal(value.Urls.Website)
				coinMarketCaps = append(coinMarketCaps, models.CoinMarketCap{
					Id:          value.ID,
					Symbol:      value.Symbol,
					Name:        value.Symbol,
					Platform:    string(contractAddress),
					Category:    value.Category,
					Twitter:     string(twitter),
					Logo:        value.Logo,
					WebSite:     string(website),
					Description: value.Description,
				})
			}
			//log.Info("cmc insert db:", zap.Int("", len(coinMarketCaps)))
			c.db.Clauses(clause.OnConflict{
				UpdateAll: true,
			}).Create(&coinMarketCaps)
		}
		time.Sleep(60 * time.Second)
	}
}

func UpdateCoinMarketCap() (result []models.CoinMarketCap) {
	c.log.Info("UpdateCoinMarketCap start")
	coinsList, err := CMCCoinsList()
	if err != nil {
		log.Error("UpdateCoinMarketCap coinsList error:", zap.Error(err))
		for err != nil {
			coinsList, err = CMCCoinsList()
			time.Sleep(1 * time.Minute)
		}
	}
	cmcDBList, err := GetAllCoinMarketCap()
	if err != nil {
		c.log.Error("get all coin market cap error:", err)
	}
	c.log.Info("source length:", len(coinsList.Data), ",db length:", len(cmcDBList))
	cmcDBMap := make(map[int]struct{}, len(cmcDBList))
	for _, cmc := range cmcDBList {
		cmcDBMap[cmc.Id] = struct{}{}
	}
	tempCoinsList := make([]types.CMCListItem, 0, len(coinsList.Data))
	for i := 0; i < len(coinsList.Data); i++ {
		if _, ok := cmcDBMap[coinsList.Data[i].ID]; !ok {
			tempCoinsList = append(tempCoinsList, coinsList.Data[i])
		}
	}
	if len(tempCoinsList) > 0 {
		result = make([]models.CoinMarketCap, 0, len(tempCoinsList))
		var allIds []string
		pageEndIndex := 0
		pageSize := 800
		for i := 0; i < len(tempCoinsList); i += pageSize {
			if i+pageSize >= len(tempCoinsList) {
				pageEndIndex = len(tempCoinsList)
			} else {
				pageEndIndex = i + pageSize
			}
			allIds = append(allIds, GetCoinsId(tempCoinsList[i:pageEndIndex]))
		}
		for _, ids := range allIds {
			coinsId, err := CMCCoinsId(ids)
			if err != nil {
				c.log.Error("CMC coins id error", err)
				for err != nil {
					time.Sleep(1 * time.Minute)
					coinsId, err = CMCCoinsId(ids)
				}
			}
			if coinsId != nil {
				coinMarketCaps := make([]models.CoinMarketCap, 0, len(coinsId.Data))
				for _, value := range coinsId.Data {
					contractAddress, _ := json.Marshal(value.ContractAddress)
					twitter, _ := json.Marshal(value.Urls.Twitter)
					website, _ := json.Marshal(value.Urls.Website)
					coinMarketCaps = append(coinMarketCaps, models.CoinMarketCap{
						Id:          value.ID,
						Symbol:      value.Symbol,
						Name:        value.Symbol,
						Platform:    string(contractAddress),
						Category:    value.Category,
						Twitter:     string(twitter),
						Logo:        value.Logo,
						WebSite:     string(website),
						Description: value.Description,
					})
				}
				c.log.Info("cmc insert db:", len(coinMarketCaps))
				c.db.Clauses(clause.OnConflict{
					UpdateAll: true,
				}).Create(&coinMarketCaps)
				result = append(result, coinMarketCaps...)
			}
			time.Sleep(60 * time.Second)
		}
	}
	c.log.Info("UpdateCoinMarketCap length:", len(result))
	return
}

func GetAllCoinMarketCap() ([]models.CoinMarketCap, error) {
	var coinMarketCaps []models.CoinMarketCap
	err := c.db.Find(&coinMarketCaps).Error
	return coinMarketCaps, err
}

func CMCSimplePrice(ids []string, currency string) (map[string]map[string]float64, error) {
	result := make(map[string]map[string]float64, len(ids))
	for _, id := range ids {
		price, err := getCmcPrice(id, currency)
		if err != nil {
			return result, err
		}
		result[id] = map[string]float64{
			currency: price,
		}
	}
	return result, nil
}

func getCmcPrice(id string, currency string) (float64, error) {
	url := fmt.Sprintf("%s/tools/price-conversion", cmc.baseUrl)
	index := rand.Intn(len(cmc.key))
	params := map[string]string{
		"CMC_PRO_API_KEY": cmc.key[index],
		"id":              id,
		"amount":          "1",
		"convert":         currency,
	}
	out := &types.CMCPriceResp{}
	err := utils.HttpsGetForm(url, params, out)
	if err != nil {
		for i := 0; i < len(cmc.key) && err != nil; i++ {
			params["CMC_PRO_API_KEY"] = cmc.key[i]
			err = utils.HttpsGetForm(url, params, out)
		}
	}
	return out.Data.Quote[strings.ToUpper(currency)].Price, err
}
