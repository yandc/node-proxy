package data

import (
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-redis/redis"
	"github.com/google/wire"
	"gitlab.bixin.com/mili/node-proxy/internal/conf"
	"gitlab.bixin.com/mili/node-proxy/internal/data/models"
	"gitlab.bixin.com/mili/node-proxy/pkg/utils"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewDB, NewRedis, NewTokenListRepo, NewChainListRepo, NewPlatformRepo, NewNFTRepo, NewCommRPCRepo)

// Data .
type Data struct {
	db  *gorm.DB
	rdb *redis.Client
	log *log.Helper
}

// NewDB connect to postgres and auto migrate
func NewDB(conf *conf.Data, logger log.Logger) *gorm.DB {
	log := log.NewHelper(log.With(logger, "module", "data/gorm"))

	db, err := gorm.Open(postgres.Open(conf.Database.Source), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed opening connection to postgres: %v", err)
	}
	if err := db.AutoMigrate(&models.CoinGecko{}, &models.CoinMarketCap{}, &models.TokenList{}, &models.NftList{}, &models.BlockChain{}, &models.ChainNodeUrl{},
		&models.NftCollection{}); err != nil {
		log.Fatal(err)
	}
	return db
}

func NewRedis(conf *conf.Data) *redis.Client {
	clint := redis.NewClient(&redis.Options{
		Addr: conf.Redis.Addr,
		DB:   int(conf.Redis.Db),
	})
	utils.SetRedisClient(clint)
	return clint
}

// NewData .
func NewData(db *gorm.DB, rdb *redis.Client, logger log.Logger) (*Data, func(), error) {
	log := log.NewHelper(log.With(logger, "module", "data/new"))
	d := &Data{
		db:  db,
		rdb: rdb,
		log: log,
	}

	cleanup := func() {
		log.Info("closing the data resources")
		// TODO 关闭 gorm 数据库连接？
	}

	return d, cleanup, nil

}
