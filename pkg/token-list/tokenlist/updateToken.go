package tokenlist

import (
	"encoding/json"
	"fmt"
	"gitlab.bixin.com/mili/node-proxy/internal/data/models"
	"gitlab.bixin.com/mili/node-proxy/pkg/nft"
	"gitlab.bixin.com/mili/node-proxy/pkg/token-list/types"
	"gorm.io/gorm/clause"
	"strings"
)

func UpdateOsmosisToken() {
	out := &types.OsmosisTokenInfo{}
	url := "https://api.mintscan.io/v2/assets/osmosis"
	body, err := nft.DoWebRequest(url)
	if err != nil {
		fmt.Println("error==", err)
		return
	}
	if err := json.Unmarshal([]byte(body), out); err != nil {
		fmt.Println("json error=", err)
		return
	}

	//err := utils.HttpsGet(url, nil, nil, out)
	//if err != nil {
	//	fmt.Println("error==", err)
	//	return
	//}
	if out != nil && len(out.Assets) > 0 {
		tokenLists := make([]models.TokenList, 0, len(out.Assets))
		for _, asset := range out.Assets {
			if asset.DpDenom == "OSMO" {
				continue
			}
			address := asset.Denom
			if strings.Contains(address, "/") {
				addressInfo := strings.Split(address, "/")
				address = fmt.Sprintf("%s/%s", strings.ToLower(addressInfo[0]), strings.ToUpper(addressInfo[1]))
			}
			tokenLists = append(tokenLists, models.TokenList{
				Chain:    asset.Chain,
				Address:  address,
				Name:     asset.BaseDenom,
				Symbol:   asset.DpDenom,
				Logo:     "https://raw.githubusercontent.com/cosmostation/cosmostation_token_resource/master/assets/images/" + asset.Image,
				Decimals: asset.Decimal,
				CgId:     asset.CoinGeckoID,
			})
		}
		result := c.db.Clauses(clause.OnConflict{
			UpdateAll: true,
		}).Create(&tokenLists)
		if result.Error != nil {
			c.log.Error("create db aptos error:", result.Error)
		}
	}

}

func UpdateCosmosToken() {
	//https://api.mintscan.io/v2/assets/cosmos

	var tokenLists = []models.TokenList{
		{
			Chain:    "cosmos",
			Address:  "uatom",
			Name:     "Cosmos Hub",
			Symbol:   "ATOM",
			Logo:     "https://raw.githubusercontent.com/cosmostation/cosmostation_token_resource/master/assets/images/common/atom.png",
			Decimals: 6,
			CgId:     "cosmos",
			WebSite:  "https://cosmos.network/",
		},
		{
			Chain:    "cosmos",
			Address:  "uosmo",
			Name:     "Osmosis",
			Symbol:   "OSMO",
			Logo:     "https://raw.githubusercontent.com/cosmostation/cosmostation_token_resource/master/assets/images/common/osmo.png",
			Decimals: 6,
			CgId:     "osmosis",
			WebSite:  "https://osmosis.zone/",
		},
		{
			Chain:    "cosmos",
			Address:  "uakt",
			Name:     "Akash Network",
			Symbol:   "AKT",
			Logo:     "https://raw.githubusercontent.com/cosmostation/cosmostation_token_resource/master/assets/images/common/akt.png",
			Decimals: 6,
			CgId:     "akash-network",
			WebSite:  "https://akash.network/",
		},
		{
			Chain:    "cosmos",
			Address:  "uixo",
			Name:     "IXO",
			Symbol:   "IXO",
			Logo:     "https://raw.githubusercontent.com/cosmostation/cosmostation_token_resource/master/assets/images/common/ixo.png",
			Decimals: 6,
			CgId:     "ixo",
			WebSite:  "https://www.ixo.world/",
		},
		{
			Chain:    "cosmos",
			Address:  "ubtsg",
			Name:     "BitSong",
			Symbol:   "BTSG",
			Logo:     "https://raw.githubusercontent.com/cosmostation/cosmostation_token_resource/master/assets/images/common/btsg.png",
			Decimals: 6,
			CgId:     "bitsong",
			WebSite:  "https://bitsong.io/",
		},
		{
			Chain:    "cosmos",
			Address:  "ubcna",
			Name:     "BitCanna",
			Symbol:   "BCNA",
			Logo:     "https://raw.githubusercontent.com/cosmostation/cosmostation_token_resource/master/assets/images/common/bcna.png",
			Decimals: 6,
			CgId:     "bitcanna",
			WebSite:  "https://www.bitcanna.io/",
		},
		{
			Chain:    "cosmos",
			Address:  "uregen",
			Name:     "Regen Network",
			Symbol:   "REGEN",
			Logo:     "https://raw.githubusercontent.com/cosmostation/cosmostation_token_resource/master/assets/images/common/regen.png",
			Decimals: 6,
			CgId:     "regen",
			WebSite:  "https://www.regen.network/",
		},
		{
			Chain:    "cosmos",
			Address:  "uxprt",
			Name:     "Persistence",
			Symbol:   "XPRT",
			Logo:     "https://raw.githubusercontent.com/cosmostation/cosmostation_token_resource/master/assets/images/common/xprt.png",
			Decimals: 6,
			CgId:     "persistence",
			WebSite:  "https://persistence.one/",
		},
		{
			Chain:    "cosmos",
			Address:  "uxki",
			Name:     "Ki",
			Symbol:   "XKI",
			Logo:     "https://raw.githubusercontent.com/cosmostation/cosmostation_token_resource/master/assets/images/common/xki.png",
			Decimals: 6,
			CgId:     "ki",
			WebSite:  "https://foundation.ki/",
		},
		{
			Chain:    "cosmos",
			Address:  "uiris",
			Name:     "IRISnet",
			Symbol:   "IRIS",
			Logo:     "https://raw.githubusercontent.com/cosmostation/cosmostation_token_resource/master/assets/images/common/iris.png",
			Decimals: 6,
			CgId:     "iris-network",
			WebSite:  "https://www.irisnet.org/",
		},
		{
			Chain:    "cosmos",
			Address:  "uion",
			Name:     "Ion",
			Symbol:   "ION",
			Logo:     "https://raw.githubusercontent.com/cosmostation/cosmostation_token_resource/master/assets/images/osmosis/ion.png",
			Decimals: 6,
			CgId:     "ion",
			WebSite:  "",
		},
		{
			Chain:    "cosmos",
			Address:  "basecro",
			Name:     "Cronos",
			Symbol:   "CRO",
			Logo:     "https://raw.githubusercontent.com/cosmostation/cosmostation_token_resource/master/assets/images/common/cro.png",
			Decimals: 8,
			CgId:     "crypto-com-chain",
			WebSite:  "https://crypto.org/",
		},
		{
			Chain:    "cosmos",
			Address:  "uiov",
			Name:     "Starname",
			Symbol:   "IOV",
			Logo:     "https://raw.githubusercontent.com/cosmostation/cosmostation_token_resource/master/assets/images/common/iov.png",
			Decimals: 6,
			CgId:     "starname",
			WebSite:  "https://www.starname.me/",
		},
		{
			Chain:    "cosmos",
			Address:  "aevmos",
			Name:     "Evmos",
			Symbol:   "EVMOS",
			Logo:     "https://raw.githubusercontent.com/cosmostation/cosmostation_token_resource/master/assets/images/common/evmos.png",
			Decimals: 18,
			CgId:     "evmos",
			WebSite:  "https://evmos.org/",
		},
		{
			Chain:    "cosmos",
			Address:  "udvpn",
			Name:     "Sentinel",
			Symbol:   "DVPN",
			Logo:     "https://raw.githubusercontent.com/cosmostation/cosmostation_token_resource/master/assets/images/common/dvpn.png",
			Decimals: 6,
			CgId:     "sentinel",
			WebSite:  "https://sentinel.co/",
		},
		{
			Chain:    "cosmos",
			Address:  "uscrt",
			Name:     "Secret",
			Symbol:   "SCRT",
			Logo:     "https://raw.githubusercontent.com/cosmostation/cosmostation_token_resource/master/assets/images/common/scrt.png",
			Decimals: 6,
			CgId:     "secret",
			WebSite:  "https://scrt.network/",
		},
		{
			Chain:    "cosmos",
			Address:  "nanolike",
			Name:     "LikeCoin",
			Symbol:   "LIKE",
			Logo:     "https://raw.githubusercontent.com/cosmostation/cosmostation_token_resource/master/assets/images/common/like.png",
			Decimals: 9,
			CgId:     "likecoin",
			WebSite:  "https://about.like.co/",
		},
		{
			Chain:    "cosmos",
			Address:  "ujuno",
			Name:     "JUNO",
			Symbol:   "JUNO",
			Logo:     "https://raw.githubusercontent.com/cosmostation/cosmostation_token_resource/master/assets/images/common/juno.png",
			Decimals: 6,
			CgId:     "juno-network",
			WebSite:  "https://www.junonetwork.io/",
		},
		{
			Chain:    "cosmos",
			Address:  "ungm",
			Name:     "e-Money",
			Symbol:   "NGM",
			Logo:     "https://raw.githubusercontent.com/cosmostation/cosmostation_token_resource/master/assets/images/common/ngm.png",
			Decimals: 6,
			CgId:     "e-money",
			WebSite:  "https://e-money.com/",
		},
		{
			Chain:    "cosmos",
			Address:  "eeur",
			Name:     "e-Money EUR",
			Symbol:   "EEUR",
			Logo:     "https://raw.githubusercontent.com/cosmostation/cosmostation_token_resource/master/assets/images/emoney/eeur.png",
			Decimals: 6,
			CgId:     "e-money-eur",
			WebSite:  "",
		},
		{
			Chain:    "cosmos",
			Address:  "ucre",
			Name:     "Crescent Network",
			Symbol:   "CRE",
			Logo:     "https://raw.githubusercontent.com/cosmostation/cosmostation_token_resource/master/assets/images/common/cre.png",
			Decimals: 6,
			CgId:     "crescent-network",
			WebSite:  "https://crescent.network/",
		},
		{
			Chain:    "cosmos",
			Address:  "rowan",
			Name:     "Sifchain",
			Symbol:   "ROWAN",
			Logo:     "https://raw.githubusercontent.com/cosmostation/cosmostation_token_resource/master/assets/images/common/rowan.png",
			Decimals: 18,
			CgId:     "sifchain",
			WebSite:  "https://www.sifchain.finance/",
		},
		{
			Chain:    "cosmos",
			Address:  "ukava",
			Name:     "Kava",
			Symbol:   "KAVA",
			Logo:     "https://raw.githubusercontent.com/cosmostation/cosmostation_token_resource/master/assets/images/common/kava.png",
			Decimals: 6,
			CgId:     "kava",
			WebSite:  "https://www.kava.io/",
		},
		{
			Chain:    "cosmos",
			Address:  "uumee",
			Name:     "Umee",
			Symbol:   "UMEE",
			Logo:     "https://raw.githubusercontent.com/cosmostation/cosmostation_token_resource/master/assets/images/common/umee.png",
			Decimals: 6,
			CgId:     "umee",
			WebSite:  "https://www.umee.cc/",
		},
		{
			Chain:    "cosmos",
			Address:  "inj",
			Name:     "Injective",
			Symbol:   "INJ",
			Logo:     "https://raw.githubusercontent.com/cosmostation/cosmostation_token_resource/master/assets/images/common/inj.png",
			Decimals: 18,
			CgId:     "injective-protocol",
			WebSite:  "https://injective.com/",
		},
		{
			Chain:    "cosmos",
			Address:  "ubcre",
			Name:     "Liquid Staking Crescent",
			Symbol:   "bCRE",
			Logo:     "https://raw.githubusercontent.com/cosmostation/cosmostation_token_resource/master/assets/images/crescent/bcre.png",
			Decimals: 6,
			CgId:     "liquid-staking-crescent",
			WebSite:  "",
		},
		{
			Chain:    "cosmos",
			Address:  "uatolo",
			Name:     "RIZON",
			Symbol:   "ATOLO",
			Logo:     "https://raw.githubusercontent.com/cosmostation/cosmostation_token_resource/master/assets/images/common/atolo.png",
			Decimals: 6,
			CgId:     "rizon",
			WebSite:  "https://rizon.world/",
		},
	}
	result := c.db.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(&tokenLists)
	if result.Error != nil {
		c.log.Error("create db aptos error:", result.Error)
	}
}