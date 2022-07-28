package tokenlist

import (
	"encoding/json"
	"fmt"
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
	key     string
	db      *gorm.DB
	log     *log.Helper
}

var cmc = cmcConf{
	baseUrl: "https://pro-api.coinmarketcap.com/v1",
	key:     "4a77c975-64ff-4c2e-8f70-b195bfb07be5",
}

func InitCMC(baseURL, key string, db *gorm.DB, logger log.Logger) {
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
	params := map[string]string{
		"CMC_PRO_API_KEY": cmc.key,
	}
	out := &types.CMCList{}
	err := utils.HttpsGetForm(url, params, out)
	return out, err
}

func CMCCoinsId(id string) (*types.CMCCoinsID, error) {
	if len(id) == 0 {
		return nil, fmt.Errorf("id is required")
	}
	url := fmt.Sprintf("%s/cryptocurrency/info", cmc.baseUrl)
	params := map[string]string{
		"CMC_PRO_API_KEY": cmc.key,
		"id":              id,
	}
	out := &types.CMCCoinsID{}
	err := utils.HttpsGetForm(url, params, out)
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

//func GetCoinMarketCapPlatform() {
//	var cmcs []models.CoinMarketCap
//	result := c.db.Find(&cmcs)
//	fmt.Println("result-", result.RowsAffected, len(cmcs))
//	fileName := "cmc.txt"
//	file, err := os.Create(fileName)
//	if err != nil {
//		fmt.Println("error:", err)
//	}
//	defer file.Close()
//	platforms := make(map[string]struct{})
//	for _, c := range cmcs {
//		var platform []types.Platform
//		json.Unmarshal([]byte(c.Platform), &platform)
//		for _, p := range platform {
//			if p.ContractAddress != "" {
//				platforms[strings.ToLower(p.Platform.Coin.Name)] = struct{}{}
//			}
//		}
//	}
//	fmt.Println("platforms length:", len(platforms))
//	for key, _ := range platforms {
//		file.WriteString(key + "\n")
//	}
//
//}

func GetAllCoinMarketCap() ([]models.CoinMarketCap, error) {
	var coinMarketCaps []models.CoinMarketCap
	err := c.db.Find(&coinMarketCaps).Error
	return coinMarketCaps, err
}

func CMCSimplePrice(ids []string, currency string) (map[string]map[string]float32, error) {
	result := make(map[string]map[string]float32, len(ids))
	for _, id := range ids {
		price, err := getCmcPrice(id, currency)
		if err != nil {
			return result, err
		}
		result[id] = map[string]float32{
			currency: price,
		}
	}
	return result, nil
}

func getCmcPrice(id string, currency string) (float32, error) {
	url := fmt.Sprintf("%s/tools/price-conversion", cmc.baseUrl)
	params := map[string]string{
		"CMC_PRO_API_KEY": cmc.key,
		"id":              id,
		"amount":          "1",
		"convert":         currency,
	}
	out := &types.CMCPriceResp{}
	err := utils.HttpsGetForm(url, params, out)

	return out.Data.Quote[currency].Price, err
}
