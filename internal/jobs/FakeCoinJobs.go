package jobs

import (
	"context"
	"fmt"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-redis/redis"
	v1 "gitlab.bixin.com/mili/node-proxy/api/market/v1"
	"gitlab.bixin.com/mili/node-proxy/internal/data/models"
	"gitlab.bixin.com/mili/node-proxy/pkg/token-list/tokenlist"
	"gitlab.bixin.com/mili/node-proxy/pkg/token-list/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"sort"
	"time"
)

type TopCoinJob struct {
	execTime string
	db       *gorm.DB
	rdb      *redis.Client
	log      *log.Helper
	client   v1.MarketClient
}

func NewTopCoinJob(db *gorm.DB, rdb *redis.Client, client v1.MarketClient, logger log.Logger) *TopCoinJob {
	t := &TopCoinJob{
		db:       db,
		rdb:      rdb,
		execTime: "0 0 1 1/3 *", //从1月份开始 每3月 1号 0点执行
		log:      log.NewHelper(logger),
		client:   client,
	}
	return t
}

func (j *TopCoinJob) Run() {
	jobName := "TopCoinJob"

	//分布式锁
	muxKey := fmt.Sprintf("mux:%s", jobName)
	if ok, _ := j.rdb.SetNX(muxKey, true, time.Second*3).Result(); !ok {
		return
	}

	j.log.Infof("任务执行开始：%s", jobName)

	//遍历支持的所有链
	for chain, chainName := range utils.ChainNameMap {

		//查询所有的coin
		var tokenLists []models.TokenList
		err := j.db.Where("chain = ? AND name != ? AND cg_id != ?", chainName, chain, "").Find(&tokenLists).Error
		if err != nil {
			j.log.Error("get token list error:", err)
			return
		}

		if tokenLists != nil && len(tokenLists) == 0 {
			continue
		}

		ids := make([]string, len(tokenLists))
		tokenMap := make(map[string]models.TokenList, len(tokenLists))
		for i, token := range tokenLists {
			ids[i] = token.CgId
			tokenMap[token.CgId] = token
		}

		markets := make([]*v1.DescribeCexCoinsReply_Coin, 0, len(tokenLists)+2)
		pageSize := 500
		endIndex := 0
		for i := 0; i < len(ids); i += pageSize {
			if i+pageSize > len(ids) {
				endIndex = len(ids)
			} else {
				endIndex = i + pageSize
			}

			reply, err := j.client.DescribeCexCoins(context.Background(), &v1.DescribeCexCoinsRequest{
				EventId:   fmt.Sprintf("%d", time.Now().Unix()),
				CoinIDs:   ids[i:endIndex],
				Currency:  "cny",
				PageSize:  50,
				Page:      1,
				SortField: v1.DescribeCexCoinsRequest_MarketCap,
			})

			if err != nil {
				j.log.Error("get cg markets error:", err)
				continue
			}

			markets = append(markets, reply.Coins...)
		}
		//sort markets
		sort.Slice(markets, func(i, j int) bool {
			return markets[i].Rank <= markets[j].Rank
		})

		var fakeCoinWhiteList []*models.FakeCoinWhiteList
		var end int
		if len(markets) > 50 {
			end = 50
		} else {
			end = len(markets)
		}
		for _, market := range markets[0:end] {
			if token, ok := tokenMap[market.CoinID]; ok {
				fakeCoinWhiteList = append(fakeCoinWhiteList, &models.FakeCoinWhiteList{
					Symbol:  token.Symbol,
					Name:    token.Name,
					Chain:   token.Chain,
					Address: token.Address,
				})
			}
		}

		if err = j.db.Clauses(clause.OnConflict{DoNothing: true}).Create(fakeCoinWhiteList).Error; err != nil {
			j.log.Warnf("save fake coin white list error:", err)
		}

	}

	allFakeCoinWhiteList := make([]*models.FakeCoinWhiteList, 0)
	err := j.db.Find(&allFakeCoinWhiteList).Error
	if err != nil {
		j.log.Error("get all fake coin white list error:", err)
		return
	}

	//更新缓存
	for _, coinWhiteList := range allFakeCoinWhiteList {
		key := fmt.Sprintf("%s%s:%s", tokenlist.REDIS_FAKECOIN_KEY, coinWhiteList.Chain, coinWhiteList.Symbol)
		_ = utils.SetFakeCoinWhiteList(j.rdb, key, coinWhiteList)
	}

	j.log.Infof("任务执行完成：%s", jobName)
}
