package tokenlist

import (
	"encoding/json"
	"fmt"
	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"node-proxy/internal/data/models"
	"node-proxy/pkg/token-list/types"
	"node-proxy/pkg/token-list/utils"
	"strings"
	"sync"
	"time"
)

type cgconf struct {
	baseURL string
	db      *gorm.DB
	log     *log.Helper
}

var cg = cgconf{
	baseURL: "https://api.coingecko.com/api/v3",
}

func InitCG(baseURL string, db *gorm.DB, logger log.Logger) {
	log := log.NewHelper(log.With(logger, "module", "tokenlist/InitCG"))
	cg = cgconf{baseURL: baseURL, db: db, log: log}
}

func CGCoinsList() ([]types.CoinsListItem, error) {
	url := fmt.Sprintf("%s/coins/list", cg.baseURL)
	params := map[string]string{
		"include_platform": "true",
	}
	var out []types.CoinsListItem
	err := utils.HttpsGetForm(url, params, &out)
	return out, err
}

func CGSimplePrice(ids []string, currency string) (map[string]map[string]float32, error) {
	url := fmt.Sprintf("%s/simple/price", cg.baseURL)
	id := strings.Join(ids, ",")
	params := map[string]string{
		"ids":           id,
		"vs_currencies": currency,
	}
	result := make(map[string]map[string]float32)
	err := utils.HttpsGetForm(url, params, &result)
	return result, err
}

func CGCoinsId(id string) (*types.CGCoinsID, error) {
	if len(id) == 0 {
		return nil, fmt.Errorf("id is required")
	}
	url := fmt.Sprintf("%s/coins/%s", cg.baseURL, id)
	params := map[string]string{
		"localization":   "false",
		"tickers":        "false",
		"market_data":    "false",
		"community_data": "false",
		"developer_data": "false",
	}
	out := &types.CGCoinsID{}
	err := utils.HttpsGetForm(url, params, out)
	return out, err
}

func PatchDo(items []types.CoinsListItem) (errorItems []types.CoinsListItem) {
	coinGeckos := make([]models.CoinGecko, 0, len(items))
	var wg sync.WaitGroup
	var lock sync.RWMutex
	for _, value := range items {
		wg.Add(1)
		go func(item types.CoinsListItem) {
			defer wg.Done()
			coinsId, err := CGCoinsId(item.ID)
			if err != nil {
				//log.Error("coinsId error:", zap.Error(err))
				return
			}
			if len(coinsId.ID) == 0 {
				lock.Lock()
				defer lock.Unlock()
				errorItems = append(errorItems, item)
				return
			}
			p, _ := json.Marshal(coinsId.Platforms)
			image, _ := json.Marshal(coinsId.Image)
			b, _ := json.Marshal(coinsId.Description)
			var homepage string
			if value, ok := coinsId.Links["homepage"]; ok {
				homepage = value.([]interface{})[0].(string)
			}
			lock.Lock()
			defer lock.Unlock()
			coinGeckos = append(coinGeckos, models.CoinGecko{
				Id:            coinsId.ID,
				Symbol:        coinsId.Symbol,
				Name:          coinsId.Name,
				Platform:      string(p),
				Image:         string(image),
				Description:   string(b),
				Homepage:      homepage,
				CoinGeckoRank: coinsId.CoinGeckoRank,
			})
		}(value)
	}
	wg.Wait()
	c.db.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(&coinGeckos)
	return
}

func CreateCoinGecko() {
	//get coin list
	coinsList, err := CGCoinsList()
	if err != nil {
		//log.Error("CreateCoinGecko error", zap.Error(err))
		for err != nil {
			time.Sleep(1 * time.Minute)
			coinsList, err = CGCoinsList()
		}
	}
	//log.Info("coinsList length", zap.Int("length", len(coinsList)))
	var ids [][]types.CoinsListItem
	pageSize := 50
	pageEndIndex := 0
	for i := 0; i < len(coinsList); i += pageSize {
		if i+pageSize > len(coinsList) {
			pageEndIndex = len(coinsList)
		} else {
			pageEndIndex = i + pageSize
		}
		ids = append(ids, coinsList[i:pageEndIndex])
	}
	for i := 0; i < len(ids); i++ {
		errorItems := PatchDo(ids[i])
		for len(errorItems) != 0 {
			time.Sleep(70 * time.Second)
			errorItems = PatchDo(errorItems)
		}
		//log.Info("errorItems==", zap.Int("length", len(errorItems)))
		time.Sleep(70 * time.Second)
	}
}

//func GetCoinGeckoPlatform() {
//	var coinGeckos []models.CoinGecko
//	result := c.db.Find(&coinGeckos)
//	fmt.Println("result-", result.RowsAffected, len(coinGeckos))
//	fileName := "coingecko.txt"
//	file, err := os.Create(fileName)
//	if err != nil {
//		fmt.Println("error:", err)
//	}
//	defer file.Close()
//	platforms := make(map[string]struct{})
//	for _, c := range coinGeckos {
//		var platform map[string]interface{}
//		json.Unmarshal([]byte(c.Platform), &platform)
//		for key, _ := range platform {
//			if key != "" {
//				platforms[key] = struct{}{}
//			}
//		}
//	}
//	fmt.Println("platforms length:", len(platforms))
//	for key, _ := range platforms {
//		file.WriteString(key + "\n")
//	}
//
//}

func GetAllCoinGecko() ([]models.CoinGecko, error) {
	var coinGeckos []models.CoinGecko
	err := c.db.Find(&coinGeckos).Error
	return coinGeckos, err
}
