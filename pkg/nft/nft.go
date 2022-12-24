package nft

import (
	"github.com/Danny-Dasilva/CycleTLS/cycletls"
	"github.com/RomainMichau/cloudscraper_go/cloudscraper"
	"github.com/go-kratos/kratos/v2/log"
	"gitlab.bixin.com/mili/node-proxy/internal/conf"
	"gorm.io/gorm"
)

type config struct {
	db   *gorm.DB
	log  *log.Helper
	ipfs string
}

const (
	BASEURL      = "https://api.opensea.io/api/v1/"
	TESTBASEURL  = "https://testnets-api.opensea.io/api/v1/"
	MAXLISTLIMIT = 200
	OPENSEA_KEY  = "207e09c24d49409ca949578d7e3bde27"
)

var nftConfig config

func InitNFT(db *gorm.DB, logger log.Logger, nftList *conf.NFTList) {
	log := log.NewHelper(log.With(logger, "module", "nft/nftList"))
	nftConfig = config{
		db:   db,
		log:  log,
		ipfs: nftList.Ipfs,
	}
}

var chainFullName = map[string]string{
	"ETH":           "Ethereum",
	"ETHGoerliTEST": "Ethereum",
	"Aptos":         "Aptos",
	"AptosTEST":     "Aptos",
	"Arbitrum":      "Arbitrum",
	"ArbitrumTEST":  "Arbitrum",
	"BSC":           "BSC",
	"BSCTEST":       "BSC",
	"Polygon":       "Polygon",
	"PolygonTEST":   "Polygon",
}

func GetFullName(chain string) string {
	if value, ok := chainFullName[chain]; ok {
		return value
	}
	return chain
}

func GetNFTDb() *gorm.DB {
	return nftConfig.db
}

func GetNFTLog() *log.Helper {
	return nftConfig.log
}

func GetIPFS() string {
	return nftConfig.ipfs
}

func DoWebRequest(url string) (string, error) {
	client, err := cloudscraper.Init(false, false)
	options := cycletls.Options{
		Headers: map[string]string{"User-Agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 12_2_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/98.0.4758.102 Safari/537.36",
			"Accept":       "application/json",
			"Content-Type": "application/json"},

		//Proxy:           "http://127.0.0.1:1087",
		Timeout:         10,
		DisableRedirect: true,
	}
	resp, err := client.Do(url, options, "GET")
	return resp.Body, err
}
