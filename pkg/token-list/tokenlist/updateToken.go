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

func UpdateArbitrumNovaToken() {
	var tokenLists = []models.TokenList{
		{
			Chain:       "arbitrum-nova",
			Address:     "0xda10009cbd5d07dd0cecc66161fc93d7c9000da1",
			Name:        "Dai",
			Symbol:      "DAI",
			Logo:        `{"thumb":"https://assets.coingecko.com/coins/images/9956/thumb/4943.png?1636636734","small":"https://assets.coingecko.com/coins/images/9956/small/4943.png?1636636734","large":"https://assets.coingecko.com/coins/images/9956/large/4943.png?1636636734"}`,
			Decimals:    18,
			CgId:        "dai",
			WebSite:     "https://makerdao.com/",
			Description: `{"en":"MakerDAO has launched Multi-collateral DAI (MCD). This token refers to the new DAI that is collaterized by multiple assets.\r\n"}`,
		},
		{
			Chain:       "arbitrum-nova",
			Address:     "0xfe60a48a0bcf4636afecc9642a145d2f241a7011",
			Name:        "SushiToken",
			Symbol:      "SUSHI",
			Logo:        `{"thumb":"https://assets.coingecko.com/coins/images/22977/thumb/SUSHI_wh_small.png?1644224030","small":"https://assets.coingecko.com/coins/images/22977/small/SUSHI_wh_small.png?1644224030","large":"https://assets.coingecko.com/coins/images/22977/large/SUSHI_wh_small.png?1644224030"}`,
			Decimals:    18,
			CgId:        "sushi",
			WebSite:     "https://sushi.com/",
			Description: `{"en":"Sushi is a DeFi protocol that is completely community-driven, serving up delicious interest for your held crypto assets.\r\n\r\nOn Sushi, you can take advantage of passive-income providing DeFi tools such as liquidity providing, yield farming and staking. Furthermore, due to the decentralized nature of being an AMM (Automated Market Maker), Sushi has fewer hurdles to execute your cryptocurrency trades and all fees are paid to the users who provided liquidity, just as it should be!"}`,
		},
		{
			Chain:    "arbitrum-nova",
			Address:  "0x52484e1ab2e2b22420a25c20fa49e173a26202cd",
			Name:     "Tether USD",
			Symbol:   "USDT",
			Logo:     `{"thumb":"https://assets.coingecko.com/coins/images/325/thumb/Tether.png?1668148663","small":"https://assets.coingecko.com/coins/images/325/small/Tether.png?1668148663","large":"https://assets.coingecko.com/coins/images/325/large/Tether.png?1668148663"}`,
			Decimals: 6,
			CgId:     "tether",
			WebSite:  "https://tether.to/",
			Description: `{"en":"Tether (USDT) is a cryptocurrency with a value meant to mirror the value of the U.S. dollar. The idea was to create a stable cryptocurrency that can be used like digital dollars. Coins that serve this purpose of being a stable dollar substitute are called “stable coins.” Tether is the most popular stable coin and even acts as a dollar replacement on many popular exchanges! According to their site, Tether converts cash into digital currency, to anchor or “tether” the value of the coin to the price of national currencies like the US dollar, the Euro, and the Yen. Like other cryptos it uses blockchain. Unlike other cryptos, it is [according to the official Tether site] “100% backed by USD” (USD is held in reserve). The primary use of Tether is that it offers some stability to the otherwise volatile crypto space and offers liquidity to exchanges who can’t deal in dollars and with banks (for example to the sometimes controversial but leading exchange \u003ca href=\"https://www.coingecko.com/en/exchanges/bitfinex\"\u003eBitfinex\u003c/a\u003e).\r\n\r\nThe digital coins are issued by a company called Tether Limited that is governed by the laws of the British Virgin Islands, according to the legal part of its website. It is incorporated in Hong Kong. It has emerged that Jan Ludovicus van der Velde is the CEO of cryptocurrency exchange Bitfinex, which has been accused of being involved in the price manipulation of bitcoin, as well as tether. Many people trading on exchanges, including Bitfinex, will use tether to buy other cryptocurrencies like bitcoin. Tether Limited argues that using this method to buy virtual currencies allows users to move fiat in and out of an exchange more quickly and cheaply. Also, exchanges typically have rocky relationships with banks, and using Tether is a way to circumvent that.\r\n\r\nUSDT is fairly simple to use. Once on exchanges like \u003ca href=\"https://www.coingecko.com/en/exchanges/poloniex\"\u003ePoloniex\u003c/a\u003e or Bittrex, it can be used to purchase Bitcoin and other cryptocurrencies. It can be easily transferred from an exchange to any Omni La
yer enabled wallet. Tether has no transaction fees, although external wallets and exchanges may charge one. In order to convert USDT to USD and vise versa through the Tether.to Platform, users must pay a small fee. Buying and selling Tether for Bitcoin can be done through a variety of exchanges like the ones mentioned previously or through the Tether.to platform, which also allows the conversion between USD to and from your bank account."}`,
		},
		{
			Chain:       "arbitrum-nova",
			Address:     "0x1d05e4e72cd994cdf976181cfb0707345763564d",
			Name:        "Wrapped BTC",
			Symbol:      "WBTC",
			Logo:        `{"thumb":"https://assets.coingecko.com/coins/images/7598/thumb/wrapped_bitcoin_wbtc.png?1548822744","small":"https://assets.coingecko.com/coins/images/7598/small/wrapped_bitcoin_wbtc.png?1548822744","large":"https://assets.coingecko.com/coins/images/7598/large/wrapped_bitcoin_wbtc.png?1548822744"}`,
			Decimals:    8,
			CgId:        "wrapped-bitcoin",
			WebSite:     "https://www.wbtc.network/c",
			Description: ` {"en":""}`,
		},
	}
	result := c.db.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(&tokenLists)
	if result.Error != nil {
		c.log.Error("create db aptos error:", result.Error)
	}
}
