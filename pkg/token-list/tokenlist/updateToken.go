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
			Description: `{"en":""}`,
		},
		//add
		{
			Chain:       "arbitrum-nova",
			Address:     "0x6ab6d61428fde76768d7b45d8bfeec19c6ef91a8",
			Name:        "DPS Rum",
			Symbol:      "RUM",
			Logo:        `{"thumb":"https://assets.coingecko.com/coins/images/27656/thumb/rum.png?1665099688","small":"https://assets.coingecko.com/coins/images/27656/small/rum.png?1665099688","large":"https://assets.coingecko.com/coins/images/27656/large/rum.png?1665099688"}`,
			Decimals:    18,
			CgId:        "dps-rum",
			WebSite:     "https://damnedpiratessociety.io/",
			Description: `{"en":"Damned Pirates Society RUM is an LP reward which allows you to mint exclusive NFT's and cosmetics for the Pirateverse"}`,
		},
		{
			Chain:       "arbitrum-nova",
			Address:     "0xefaeee334f0fd1712f9a8cc375f427d9cdd40d73",
			Name:        "DPSDoubloon",
			Symbol:      "DOUBLOON",
			Logo:        `{"thumb":"https://assets.coingecko.com/coins/images/27642/thumb/6lCO7C9y_400x400.jpeg?1665019747","small":"https://assets.coingecko.com/coins/images/27642/small/6lCO7C9y_400x400.jpeg?1665019747","large":"https://assets.coingecko.com/coins/images/27642/large/6lCO7C9y_400x400.jpeg?1665019747"}`,
			Decimals:    18,
			CgId:        "dps-doubloon",
			WebSite:     "https://damnedpiratessociety.io",
			Description: `{"en":""}`,
		},
		{
			Chain:       "arbitrum-nova",
			Address:     "0x80a16016cc4a2e6a2caca8a4a498b1699ff0f844",
			Name:        "TreasureMaps",
			Symbol:      "TMAP",
			Logo:        `{"thumb":"https://assets.coingecko.com/coins/images/27655/thumb/tmap.png?1665099449","small":"https://assets.coingecko.com/coins/images/27655/small/tmap.png?1665099449","large":"https://assets.coingecko.com/coins/images/27655/large/tmap.png?1665099449"}`,
			Decimals:    18,
			CgId:        "dps-treasuremaps",
			WebSite:     "https://damnedpiratessociety.io/",
			Description: `{"en":"Treasure Map Tokens are the tickets to voyage in the Damned Pirates Society universe. Exchange them at the cartographer for a voyage and set sail in search of chests of Doubloons. Ye be warned dangers lurk on the high seas."}`,
		},

		//db old
		{
			Chain:       "arbitrum-nova",
			Address:     "0x6dcb98f460457fe4952e12779ba852f82ecc62c1",
			Name:        "r/FortNiteBR Bricks",
			Symbol:      "BRICK",
			Logo:        `{"thumb":"https://assets.coingecko.com/coins/images/11223/thumb/Brick.png?1589620469","small":"https://assets.coingecko.com/coins/images/11223/small/Brick.png?1589620469","large":"https://assets.coingecko.com/coins/images/11223/large/Brick.png?1589620469"}`,
			Decimals:    18,
			CgId:        "brick",
			WebSite:     "https://www.reddit.com/r/FortNiteBR/",
			Description: `{"en":"Bricks are ERC-20 Tokens given as rewards for an individuals contributions to r/Fortnite either via posts or comments etc. They can be freely transferred, tipped and spent in r/CryptoCurrency. Moons are distributed monthly using Reddit Karma as a basis for contributions.\r\n\r\nBricks can be traded freely and used for any number of purposes within the community. At this time, they can be used to display reputation within the subreddit, unlock exclusive features like badges and GIFs in comments with a Special Membership, and add weight to votes in polls."}`,
		},
		{
			Chain:       "arbitrum-nova",
			Address:     "0x750ba8b76187092b0d1e87e28daaf484d1b5273b",
			Name:        "USD Coin",
			Symbol:      "USDC",
			Logo:        `{"thumb":"https://assets.coingecko.com/coins/images/6319/thumb/USD_Coin_icon.png?1547042389","small":"https://assets.coingecko.com/coins/images/6319/small/USD_Coin_icon.png?1547042389","large":"https://assets.coingecko.com/coins/images/6319/large/USD_Coin_icon.png?1547042389"}`,
			Decimals:    6,
			CgId:        "usd-coin",
			WebSite:     "https://www.circle.com/en/usdc",
			Description: `{"en":"USDC is a fully collateralized US dollar stablecoin. USDC is the bridge between dollars and trading on cryptocurrency exchanges. The technology behind CENTRE makes it possible to exchange value between people, businesses and financial institutions just like email between mail services and texts between SMS providers. We believe by removing artificial economic borders, we can create a more inclusive global economy."}`,
		},
		{
			Chain:       "arbitrum-nova",
			Address:     "0x722e8bdd2ce80a4422e880164f2079488e115365",
			Name:        "WETH",
			Symbol:      "WETH",
			Logo:        `{"thumb":"https://assets.coingecko.com/coins/images/2518/thumb/weth.png?1628852295","small":"https://assets.coingecko.com/coins/images/2518/small/weth.png?1628852295","large":"https://assets.coingecko.com/coins/images/2518/large/weth.png?1628852295"}`,
			Decimals:    18,
			CgId:        "weth",
			WebSite:     "https://weth.io/",
			Description: `{"en":"What is WETH (Wrapped ETH)?\r\nWETH is the tokenized/packaged form of ETH that you use to pay for items when you interact with Ethereum dApps. WETH follows the ERC-20 token standards, enabling it to achieve interoperability with other ERC-20 tokens. \r\n\r\nThis offers more utility to holders as they can use it across networks and dApps. You can stake, yield farm, lend, and provide liquidity to various liquidity pools with WETH. \r\n\r\nAlso, unlike ETH, which doesn’t conform to its own ERC-20 standard and thus has lower interoperability as it can’t be used on other chains besides Ethereum, WETH can be used on cheaper and high throughput alternatives like Binance, Polygon, Solana, and Cardano.\r\n\r\nThe price of WETH will always be the same as ETH because it maintains a 1:1 wrapping ratio.\r\n\r\nHow to Wrap ETH?\r\nCustodians wrap and unwrap ETH. To wrap ETH, you send ETH to a custodian. This can be a multi-sig wallet, a Decentralized Autonomous Organization (DAO), or a smart contract. After connecting your web3 wallet to a DeFi exchange, you enter the amount of ETH you wish to wrap and click the swap function. Once the transaction is confirmed, you will receive WETH tokens equivalent to the ETH that you’ve swapped.\r\n\r\nOn a centralized exchange, the exchange burns the deposited ETH and mints a wrapped form for you. And when you want to unwrap it, the exchange will burn the wrapped version and mint the ETH on your behalf.\r\n\r\nWhat’s Next for WETH?\r\nAccording to the developers, hopefully there will be no future for WETH. According to the website, steps are being taken to update ETH to make it compliant with its own ERC-20 standards."}`,
		},
		{
			Chain:       "arbitrum-nova",
			Address:     "0x0057ac2d777797d31cd3f8f13bf5e927571d6ad0",
			Name:        "r/CryptoCurrency Moons",
			Symbol:      "MOON",
			Logo:        `{"thumb":"https://assets.coingecko.com/coins/images/11222/thumb/Moons.png?1589620193","small":"https://assets.coingecko.com/coins/images/11222/small/Moons.png?1589620193","large":"https://assets.coingecko.com/coins/images/11222/large/Moons.png?1589620193"}`,
			Decimals:    18,
			CgId:        "moon",
			WebSite:     "https://www.reddit.com/r/CryptoCurrency/",
			Description: `{"en":"Moons are ERC-20 Tokens given as rewards for an individuals contributions to r/CryptoCurrency either via posts or comments etc. They can be freely transferred, tipped and spent in r/CryptoCurrency. Moons are distributed monthly using Reddit Karma as a basis for contributions.\r\n\r\nMoons can be traded freely and used for any number of purposes within the community. At this time, they can be used to display reputation within the subreddit, unlock exclusive features like badges and GIFs in comments with a Special Membership, and add weight to votes in polls."}`,
		},
	}
	for _, t := range tokenLists {
		c.db.Clauses(clause.OnConflict{
			UpdateAll: true,
		}).Create(&t)
	}

}

func UpdateConfluxToken() {
	var tokenLists = []models.TokenList{
		{
			Chain:       "conflux",
			Address:     "0xfe197e7968807b311d476915db585831b43a7e3b",
			Name:        "Nucleon Governance Token",
			Symbol:      "NUT",
			Logo:        `https://scan-icons.oss-cn-hongkong.aliyuncs.com/mainnet/net1030%3Aad9bw9x3rcah0pj7j7yvn042na25jsx8hpf19ma3zp.png`,
			Decimals:    18,
			CgId:        "",
			WebSite:     "",
			Description: "",
		},
		{
			Chain:       "conflux",
			Address:     "0x889138644274a7dc602f25a7e7d53ff40e6d0091",
			Name:        "X nucleon CFX",
			Symbol:      "xCFX",
			Logo:        `https://scan-icons.oss-cn-hongkong.aliyuncs.com/mainnet/net1030%3Aacekcsdejk4mt1daf6w4t38zh94a65jaweg04t3z9h.png`,
			Decimals:    18,
			CgId:        "",
			WebSite:     "https://www.nucleon.space/",
			Description: "",
		},
		{
			Chain:       "conflux",
			Address:     "0xff33b107a0e2c0794ac43c3ffaf637fcea3697cf",
			Name:        "AUSD Stablecoin",
			Symbol:      "AUSD",
			Logo:        `https://scan-icons.oss-cn-hongkong.aliyuncs.com/mainnet/net1030%3Aad9xhpjhydvpa8mm2u8d9810g98syry13689ujkg7r.png`,
			Decimals:    18,
			CgId:        "",
			WebSite:     "https://www.triangledao.finance/",
			Description: "",
		},
	}
	for _, t := range tokenLists {
		c.db.Clauses(clause.OnConflict{
			UpdateAll: true,
		}).Create(&t)
	}
}
