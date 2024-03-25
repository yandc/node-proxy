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
	url := "https://api.mintscan.io/v3/assets/osmosis"
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
			if asset.Symbol == "OSMO" {
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
				Name:     asset.OriginDenom,
				Symbol:   asset.Symbol,
				Logo:     "https://raw.githubusercontent.com/cosmostation/chainlist/master/chain/" + asset.Image,
				Decimals: asset.Decimal,
				CgId:     asset.CoinGeckoID,
			})
		}
		for _, token := range tokenLists {
			c.db.Clauses(clause.OnConflict{
				UpdateAll: true,
			}).Create(&token)

		}

	}

}

func UpdateCosmosToken() {
	//https://api.mintscan.io/v2/assets/cosmos

	var tokenLists = []models.TokenList{
		{
			Chain:   "cosmos",
			Address: "ibc/D79E7D83AB399BFFF93433E54FAA480C191248FC556924A2A8351AE2638B3877",
			Name:    "Celestia",
			Symbol:  "TIA",
			Logo:    "https://assets.coingecko.com/coins/images/31967/large/tia.jpg?1696530772",
			//Decimals: 6,
			CgId:    "celestia",
			WebSite: "https://celestia.org/",
		},
		//{
		//	Chain:    "cosmos",
		//	Address:  "ibc/2181AAB0218EAC24BC9F86BD1364FBBFA3E6E3FCC25E88E3E68C15DC6E752D86",
		//	Name:     "Akash Network",
		//	Symbol:   "AKT",
		//	Logo:     "https://raw.githubusercontent.com/cosmostation/cosmostation_token_resource/master/assets/images/common/akt.png",
		//	Decimals: 6,
		//	CgId:     "akash-network",
		//	WebSite:  "https://akash.network/",
		//},
		//{
		//	Chain:    "cosmos",
		//	Address:  "ibc/E7D5E9D0E9BF8B7354929A817DD28D4D017E745F638954764AA88522A7A409EC",
		//	Name:     "BitSong",
		//	Symbol:   "BTSG",
		//	Logo:     "https://raw.githubusercontent.com/cosmostation/cosmostation_token_resource/master/assets/images/common/btsg.png",
		//	Decimals: 6,
		//	CgId:     "bitsong",
		//	WebSite:  "https://bitsong.io/",
		//},
		//{
		//	Chain:    "cosmos",
		//	Address:  "ibc/ADBEC1A7AC2FEF73E06B066A1C94DAB6C27924EF7EA3F5A43378150009620284",
		//	Name:     "BitCanna",
		//	Symbol:   "BCNA",
		//	Logo:     "https://raw.githubusercontent.com/cosmostation/cosmostation_token_resource/master/assets/images/common/bcna.png",
		//	Decimals: 6,
		//	CgId:     "bitcanna",
		//	WebSite:  "https://www.bitcanna.io/",
		//},
		//{
		//	Chain:    "cosmos",
		//	Address:  "ibc/1FBDD58D438B4D04D26CBFB2E722C18984A0F1A52468C4F42F37D102F3D3F399",
		//	Name:     "Regen Network",
		//	Symbol:   "REGEN",
		//	Logo:     "https://raw.githubusercontent.com/cosmostation/cosmostation_token_resource/master/assets/images/common/regen.png",
		//	Decimals: 6,
		//	CgId:     "regen",
		//	WebSite:  "https://www.regen.network/",
		//},
		//{
		//	Chain:    "cosmos",
		//	Address:  "ibc/81D08BC39FB520EBD948CF017910DD69702D34BF5AC160F76D3B5CFC444EBCE0",
		//	Name:     "Persistence",
		//	Symbol:   "XPRT",
		//	Logo:     "https://raw.githubusercontent.com/cosmostation/cosmostation_token_resource/master/assets/images/common/xprt.png",
		//	Decimals: 6,
		//	CgId:     "persistence",
		//	WebSite:  "https://persistence.one/",
		//},
		//{
		//	Chain:    "cosmos",
		//	Address:  "ibc/533E5FFC606FD11B8DCA309C66AFD6A1F046EF784A73F323A332CF6823F0EA87",
		//	Name:     "Ki",
		//	Symbol:   "XKI",
		//	Logo:     "https://raw.githubusercontent.com/cosmostation/cosmostation_token_resource/master/assets/images/common/xki.png",
		//	Decimals: 6,
		//	CgId:     "ki",
		//	WebSite:  "https://foundation.ki/",
		//},
		//{
		//	Chain:    "cosmos",
		//	Address:  "ibc/12DA42304EE1CE96071F712AA4D58186AD11C3165C0DCDA71E017A54F3935E66",
		//	Name:     "IRISnet",
		//	Symbol:   "IRIS",
		//	Logo:     "https://raw.githubusercontent.com/cosmostation/cosmostation_token_resource/master/assets/images/common/iris.png",
		//	Decimals: 6,
		//	CgId:     "iris-network",
		//	WebSite:  "https://www.irisnet.org/",
		//},
		//{
		//	Chain:    "cosmos",
		//	Address:  "ibc/5BB694D466CCF099EF73F165F88472AF51D9C4991EAA42BD1168C5304712CC0D",
		//	Name:     "Ion",
		//	Symbol:   "ION",
		//	Logo:     "https://raw.githubusercontent.com/cosmostation/cosmostation_token_resource/master/assets/images/osmosis/ion.png",
		//	Decimals: 6,
		//	CgId:     "ion",
		//	WebSite:  "",
		//},
		//{
		//	Chain:    "cosmos",
		//	Address:  "ibc/C932ADFE2B4216397A4F17458B6E4468499B86C3BC8116180F85D799D6F5CC1B",
		//	Name:     "Cronos",
		//	Symbol:   "CRO",
		//	Logo:     "https://raw.githubusercontent.com/cosmostation/cosmostation_token_resource/master/assets/images/common/cro.png",
		//	Decimals: 8,
		//	CgId:     "crypto-com-chain",
		//	WebSite:  "https://crypto.org/",
		//},
		//{
		//	Chain:    "cosmos",
		//	Address:  "ibc/68A333688E5B07451F95555F8FE510E43EF9D3D44DF0909964F92081EF9BE5A7",
		//	Name:     "Starname",
		//	Symbol:   "IOV",
		//	Logo:     "https://raw.githubusercontent.com/cosmostation/cosmostation_token_resource/master/assets/images/common/iov.png",
		//	Decimals: 6,
		//	CgId:     "starname",
		//	WebSite:  "https://www.starname.me/",
		//},
		//{
		//	Chain:    "cosmos",
		//	Address:  "ibc/19DD710119533524061885A6F190B18AF28D9537E2BAE37F32A62C1A25979287",
		//	Name:     "Evmos",
		//	Symbol:   "EVMOS",
		//	Logo:     "https://raw.githubusercontent.com/cosmostation/cosmostation_token_resource/master/assets/images/common/evmos.png",
		//	Decimals: 18,
		//	CgId:     "evmos",
		//	WebSite:  "https://evmos.org/",
		//},
		//{
		//	Chain:    "cosmos",
		//	Address:  "ibc/42E47A5BA708EBE6E0C227006254F2784E209F4DBD3C6BB77EDC4B29EF875E8E",
		//	Name:     "Sentinel",
		//	Symbol:   "DVPN",
		//	Logo:     "https://raw.githubusercontent.com/cosmostation/cosmostation_token_resource/master/assets/images/common/dvpn.png",
		//	Decimals: 6,
		//	CgId:     "sentinel",
		//	WebSite:  "https://sentinel.co/",
		//},
		//{
		//	Chain:    "cosmos",
		//	Address:  "ibc/1542F8DC70E7999691E991E1EDEB1B47E65E3A217B1649D347098EE48ACB580F",
		//	Name:     "Secret",
		//	Symbol:   "SCRT",
		//	Logo:     "https://raw.githubusercontent.com/cosmostation/cosmostation_token_resource/master/assets/images/common/scrt.png",
		//	Decimals: 6,
		//	CgId:     "secret",
		//	WebSite:  "https://scrt.network/",
		//},
		//{
		//	Chain:    "cosmos",
		//	Address:  "ibc/1D5826F7EDE6E3B13009FEF994DC9CAAF15CC24CA7A9FF436FFB2E56FD72F54F",
		//	Name:     "LikeCoin",
		//	Symbol:   "LIKE",
		//	Logo:     "https://raw.githubusercontent.com/cosmostation/cosmostation_token_resource/master/assets/images/common/like.png",
		//	Decimals: 9,
		//	CgId:     "likecoin",
		//	WebSite:  "https://about.like.co/",
		//},
		//{
		//	Chain:    "cosmos",
		//	Address:  "ibc/CDAB23DA5495290063363BD1C3499E26189036302DC689985A7E23F8DF8D8DB0",
		//	Name:     "JUNO",
		//	Symbol:   "JUNO",
		//	Logo:     "https://raw.githubusercontent.com/cosmostation/cosmostation_token_resource/master/assets/images/common/juno.png",
		//	Decimals: 6,
		//	CgId:     "juno-network",
		//	WebSite:  "https://www.junonetwork.io/",
		//},
		//{
		//	Chain:    "cosmos",
		//	Address:  "ibc/E070CE91CC4BD15AEC9B5788C0826755AAD35052A3037E9AC62BE70B4C9A7DBB",
		//	Name:     "e-Money",
		//	Symbol:   "NGM",
		//	Logo:     "https://raw.githubusercontent.com/cosmostation/cosmostation_token_resource/master/assets/images/common/ngm.png",
		//	Decimals: 6,
		//	CgId:     "e-money",
		//	WebSite:  "https://e-money.com/",
		//},
		//{
		//	Chain:    "cosmos",
		//	Address:  "ibc/B93F321238F7BB15AB5B882660AAE72286C8E9035DE34E2B30F60E54C623C63C",
		//	Name:     "e-Money EUR",
		//	Symbol:   "EEUR",
		//	Logo:     "https://raw.githubusercontent.com/cosmostation/cosmostation_token_resource/master/assets/images/emoney/eeur.png",
		//	Decimals: 6,
		//	CgId:     "e-money-eur",
		//	WebSite:  "",
		//},
		//{
		//	Chain:    "cosmos",
		//	Address:  "ibc/3F18D520CE791A40357D061FAD657CED6B21D023F229EAF131D7FE7CE6F488BD",
		//	Name:     "Crescent Network",
		//	Symbol:   "CRE",
		//	Logo:     "https://raw.githubusercontent.com/cosmostation/cosmostation_token_resource/master/assets/images/common/cre.png",
		//	Decimals: 6,
		//	CgId:     "crescent-network",
		//	WebSite:  "https://crescent.network/",
		//},
		//{
		//	Chain:    "cosmos",
		//	Address:  "ibc/F5ED5F3DC6F0EF73FA455337C027FE91ABCB375116BF51A228E44C493E020A09",
		//	Name:     "Sifchain",
		//	Symbol:   "ROWAN",
		//	Logo:     "https://raw.githubusercontent.com/cosmostation/cosmostation_token_resource/master/assets/images/common/rowan.png",
		//	Decimals: 18,
		//	CgId:     "sifchain",
		//	WebSite:  "https://www.sifchain.finance/",
		//},
		//{
		//	Chain:    "cosmos",
		//	Address:  "ibc/8870C4203CEBF2279BA065E3DE95FC3F8E05A4A93424E7DC707A21514BE353A0",
		//	Name:     "Kava",
		//	Symbol:   "KAVA",
		//	Logo:     "https://raw.githubusercontent.com/cosmostation/cosmostation_token_resource/master/assets/images/common/kava.png",
		//	Decimals: 6,
		//	CgId:     "kava",
		//	WebSite:  "https://www.kava.io/",
		//},
		//{
		//	Chain:    "cosmos",
		//	Address:  "ibc/DEC41A02E47658D40FC71E5A35A9C807111F5A6662A3FB5DA84C4E6F53E616B3",
		//	Name:     "Umee",
		//	Symbol:   "UMEE",
		//	Logo:     "https://raw.githubusercontent.com/cosmostation/cosmostation_token_resource/master/assets/images/common/umee.png",
		//	Decimals: 6,
		//	CgId:     "umee",
		//	WebSite:  "https://www.umee.cc/",
		//},
		//{
		//	Chain:    "cosmos",
		//	Address:  "ibc/6469BDA6F62C4F4B8F76629FA1E72A02A3D1DD9E2B22DDB3C3B2296DEAD29AB8",
		//	Name:     "Injective",
		//	Symbol:   "INJ",
		//	Logo:     "https://raw.githubusercontent.com/cosmostation/cosmostation_token_resource/master/assets/images/common/inj.png",
		//	Decimals: 18,
		//	CgId:     "injective-protocol",
		//	WebSite:  "https://injective.com/",
		//},
		//{
		//	Chain:    "cosmos",
		//	Address:  "ibc/835EE9D00C35D72128F195B50F8A89EB83E5011C43EA0AA00D16348E2208FEBB",
		//	Name:     "Liquid Staking Crescent",
		//	Symbol:   "bCRE",
		//	Logo:     "https://raw.githubusercontent.com/cosmostation/cosmostation_token_resource/master/assets/images/crescent/bcre.png",
		//	Decimals: 6,
		//	CgId:     "liquid-staking-crescent",
		//	WebSite:  "",
		//},
		//{
		//	Chain:    "cosmos",
		//	Address:  "ibc/20A7DC8E24709E6F1EE0F4E832C2ED345ADD77425890482A349AE3C43CAC6B2C",
		//	Name:     "RIZON",
		//	Symbol:   "ATOLO",
		//	Logo:     "https://raw.githubusercontent.com/cosmostation/cosmostation_token_resource/master/assets/images/common/atolo.png",
		//	Decimals: 6,
		//	CgId:     "rizon",
		//	WebSite:  "https://rizon.world/",
		//},
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

func UpdateBSCToken() {
	var tokenLists = []models.TokenList{
		{
			Chain:       "binance-smart-chain",
			Address:     "0x2dff88a56767223a5529ea5960da7a3f5f766406",
			Name:        "Space ID",
			Symbol:      "id",
			Logo:        `{"thumb":"https://assets.coingecko.com/coins/images/29468/thumb/id.png?1679112555","small":"https://assets.coingecko.com/coins/images/29468/small/id.png?1679112555","large":"https://assets.coingecko.com/coins/images/29468/large/id.png?1679112555"}`,
			Decimals:    18,
			CgId:        "space-id",
			WebSite:     "https://space.id/",
			Description: `{"en":"SPACE ID is building a universal name service network with a one-stop identity platform to discover, register, trade, manage web3 domains. It also includes a Web3 Name SDK \u0026 API for developers across blockchains and provides a multi-chain name service for everyone to easily build and create a web3 identity."}`,
			LogoURI:     "images/binance-smart-chain/binance-smart-chain_0x2dff88a56767223a5529ea5960da7a3f5f766406.png",
		},
	}
	for _, t := range tokenLists {
		c.db.Clauses(clause.OnConflict{
			UpdateAll: true,
		}).Create(&t)
	}
}

func UpdateZkSyncToken() {
	var tokenLists = []models.TokenList{
		{
			Chain:       "zkSync",
			Address:     "0x3355df6d4c9c3035724fd0e3914de96a5a83aaf4",
			Name:        "USD Coin",
			Symbol:      "USDC",
			Logo:        "https://s2.coinmarketcap.com/static/img/coins/64x64/3408.png",
			Decimals:    6,
			CmcId:       3408,
			CgId:        "usd-coin",
			WebSite:     "https://www.centre.io/usdc",
			Description: `{"en":"USDC is a fully collateralized US dollar stablecoin. USDC is the bridge between dollars and trading on cryptocurrency exchanges. The technology behind CENTRE makes it possible to exchange value between people, businesses and financial institutions just like email between mail services and texts between SMS providers. We believe by removing artificial economic borders, we can create a more inclusive global economy."}`,
		},
		{
			Chain:    "zkSync",
			Address:  "0x0e97c7a0f8b2c9885c8ac9fc6136e829cbc21d42",
			Name:     "Mute.io",
			Symbol:   "MUTE",
			Logo:     "https://s2.coinmarketcap.com/static/img/coins/64x64/8795.png",
			Decimals: 18,
			CmcId:    8795,
			CgId:     "mute",
			WebSite:  "https://mute.io",
			Description: `{"en":"MUTE is one half of the dual-token mechanics powering the Mute.io ZK-Rollup ecosystem.\r\n\r\nMUTE is the gas that facilitates growth of the ecosystem via the DAO, funding proposals and benefitting directly through a 'buyback and make' initiative. Liquidity providers are also rewarded thanks to a 1% transaction fee, guarding against impermeant loss and paid out incrementally via the MuteVault contract. \r\n\r\nMute Switch, an easy to use swap-style ZK-Rollup DEX, is the first Dapp in the ecosystem. This will be running on ZK-Rollup infrastructure meaning trading will be cheaper and more scalable, but not just that - transactions will be zero knowledge, ensuring nobody can see into your wallet history. Power users of the DEX will qualify for reduced fees by locking
a specified amount of MUTE in their wallet.\r\n\r\nThe non-inflationary supply is complimented by an innovative economic model that includes the combination of buy backs, coin burns, a smart treasury and vaults."}`,
		},
		{
			Chain:       "zkSync",
			Address:     "0xc2b13bb90e33f1e191b8aa8f44ce11534d5698e3",
			Name:        "Furucombo",
			Symbol:      "COMBO",
			Logo:        `{"thumb":"https://assets.coingecko.com/coins/images/13629/thumb/COMBO_token_ol.png?1610701537","small":"https://assets.coingecko.com/coins/images/13629/small/COMBO_token_ol.png?1610701537","large":"https://assets.coingecko.com/coins/images/13629/large/COMBO_token_ol.png?1610701537"}`,
			Decimals:    18,
			CgId:        "furucombo",
			WebSite:     "https://furucombo.app/",
			Description: ` {"en":"Furucombo is a drag-and-drop tool that allows users to build and customize different DeFi combinations (‘combos’ or ‘cubes’). These combos/cubes represent multiple protocol actions bundled into one transaction executed by Furucombo. \r\n\r\nAs one of the most comprehensive DeFi aggregators, Furucombo connects different DeFi protocols such as Uniswap, Compound, and Aave in one place. Its lego-like interface simplifies the complexity in DeFi, allowing many lay-man users to reap the benefits of DeFi using Furucombo without coding knowledge.\r\n \r\nInstead of clicking and confirming five Ethereum transactions, users of Furucombo will need to confirm only one - saving time and steps and simultaneously optimizing users’ actions to save on gas fees. "}`,
		},
		{
			Chain:       "zkSync",
			Address:     "0x42c1c56be243c250ab24d2ecdcc77f9ccaa59601",
			Name:        "Perpetual",
			Symbol:      "PERP",
			Logo:        "https://s2.coinmarketcap.com/static/img/coins/64x64/6950.png",
			Decimals:    18,
			CmcId:       6950,
			CgId:        "perpetual-protocol",
			WebSite:     "https://perp.com",
			Description: "",
		},
		{
			Chain:       "zkSync",
			Address:     "0x5aea5775959fbc2557cc8789bc1bf90a239d9a91",
			Name:        "Wrapped Ether",
			Symbol:      "WETH",
			Logo:        "https://s2.coinmarketcap.com/static/img/coins/64x64/2396.png",
			Decimals:    18,
			CmcId:       2396,
			CgId:        "weth",
			WebSite:     "https://weth.io/",
			Description: `{"en":"W-ETH is \"wrapped ETH\" but let's start by introducing the players. First, there's Ether token. Ether or ETH is the native currency built on the Ethereum blockchain.\r\nSecond, there are alt tokens. When a dApp (decentralized app) is built off of the Ethereum Blockchain it usually implements it’s own form of Token. Think Augur’s REP Token, or Bancor's BNT Token. Finally the ERC-20 standard. ERC20 is a standard developed after the release of ETH that defines how tokens are transferred and how to keep a consistent record of those transfers among tokens in the Ethereum Network."}`,
		},
	}
	for _, t := range tokenLists {
		c.db.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "address"}, {Name: "chain"}},
			UpdateAll: true,
		}).Create(&t)
	}
}

func UpdateSUITESTToken() {
	var tokenLists = []models.TokenList{
		{
			Chain:    "SUITEST",
			Address:  "0xe158e6df182971bb6c85eb9de9fbfb460b68163d19afc45873c8672b5cc521b2::TOKEN::TestUSDT",
			Name:     "TestUSDT",
			Symbol:   "TestUSDT",
			Decimals: 6,
		},
		{
			Chain:    "SUITEST",
			Address:  "0xe158e6df182971bb6c85eb9de9fbfb460b68163d19afc45873c8672b5cc521b2::TOKEN::TestDAI",
			Name:     "TestDAI",
			Symbol:   "TestDAI",
			Decimals: 6,
		},
	}
	for _, t := range tokenLists {
		c.db.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "address"}, {Name: "chain"}},
			UpdateAll: true,
		}).Create(&t)
	}
}

func UpdateSUIToken() {
	var tokenLists = []models.TokenList{
		{
			Chain:       "SUI",
			Address:     "0xdbe380b13a6d0f5cdedd58de8f04625263f113b3f9db32b3e1983f49e2841676::coin::COIN",
			Name:        "Wrapped Matic",
			Symbol:      "WMATIC",
			Logo:        "https://s2.coinmarketcap.com/static/img/coins/64x64/8925.png",
			Decimals:    8,
			CmcId:       8925,
			CgId:        "wmatic",
			WebSite:     "https://matic.network/",
			Description: `Wrapped Matic on Polygon (PoS) chain.`,
		}, {
			Chain:       "SUI",
			Address:     "0xb848cce11ef3a8f62eccea6eb5b35a12c4c2b1ee1af7755d02d7bd6218e8226f::coin::COIN",
			Name:        "Wrapped BNB",
			Symbol:      "WBNB",
			Logo:        "https://s2.coinmarketcap.com/static/img/coins/64x64/7192.png",
			Decimals:    8,
			CmcId:       7192,
			CgId:        "wbnb",
			WebSite:     "https://www.binance.org/",
			Description: `{"en":"Wrapped BNB a wrapped version of the BNB native tokens on the BEP-20 standard on the Binance Smart Chain and other EVM-compatible chains. Not to be confused with BNB Native Token on the BSC Chain."}`,
		}, {
			Chain:       "SUI",
			Address:     "0xaf8cd5edc19c4512f4259f0bee101a40d41ebed738ade5874359610ef8eeced5::coin::COIN",
			Name:        "Wrapped Ether",
			Symbol:      "WETH",
			Logo:        "https://s2.coinmarketcap.com/static/img/coins/64x64/2396.png",
			Decimals:    8,
			CmcId:       2396,
			CgId:        "weth",
			WebSite:     "https://weth.io/",
			Description: `{"en":"W-ETH is \"wrapped ETH\" but let's start by introducing the players. First, there's Ether token. Ether or ETH is the native currency built on the Ethereum blockchain.\r\nSecond, there are alt tokens. When a dApp (decentralized app) is built off of the Ethereum Blockchain it usually implements it’s own form of Token. Think Augur’s REP Token, or Bancor's BNT Token. Finally the ERC-20 standard. ERC20 is a standard developed after the release of ETH that defines how tokens are transferred and how to keep a consistent record of those transfers among tokens in the Ethereum Network."}`,
		}, {
			Chain:       "SUI",
			Address:     "0x5d4b302506645c37ff133b98c4b50a5ae14841659738d6d733d59d0d217a93bf::coin::COIN",
			Name:        "USD Coin",
			Symbol:      "USDC",
			Logo:        "https://raw.githubusercontent.com/cosmostation/cosmostation_token_resource/master/assets/images/ethereum/usdc.png",
			Decimals:    6,
			CmcId:       3408,
			CgId:        "usd-coin",
			WebSite:     "https://www.centre.io/usdc",
			Description: `{"en":"USDC is a fully collateralized US dollar stablecoin. USDC is the bridge between dollars and trading on cryptocurrency exchanges. The technology behind CENTRE makes it possible to exchange value between people, businesses and financial institutions just like email between mail services and texts between SMS providers. We believe by removing artificial economic borders, we can create a more inclusive global economy."}`,
		}, {
			Chain:    "SUI",
			Address:  "0xc060006111016b8a020ad5b33834984a437aaa7d3c74c18e09a95d48aceab08c::coin::COIN",
			Name:     "Tether USD",
			Symbol:   "USDT",
			Logo:     `{"thumb":"https://assets.coingecko.com/coins/images/325/thumb/Tether.png?1668148663","small":"https://assets.coingecko.com/coins/images/325/small/Tether.png?1668148663","large":"https://assets.coingecko.com/coins/images/325/large/Tether.png?1668148663"}`,
			Decimals: 6,
			CgId:     "tether",
			WebSite:  "https://tether.to/",
			Description: `{"en":"Tether (USDT) is a cryptocurrency with a value meant to mirror the value of the U.S. dollar. The idea was to create a stable cryptocurrency that can be used like digital dollars. Coins that serve this purpose of being a stable dollar substitute are called “stable coins.” Tether is the most popular stable coin and even acts as a dollar replacement on many popular exchanges! According to their site, Tether converts cash into digital currency, to anchor or “tether” the value of the coin to the price of national currencies like the US dollar, the Euro, and the Yen. Like other cryptos it uses blockchain. Unlike other cryptos, it is [according to the official Tether site] “100% backed by USD” (USD is held in reserve). The primary use of Tether is that it offers some stability to the otherwise volatile crypto space and offers liquidity to exchanges who can’t deal in dollars and with banks (for example to the sometimes controversial but leading exchange \u003ca href=\"https://www.coingecko.com/en/exchanges/bitfinex\"\u003eBitfinex\u003c/a\u003e).\r\n\r\nThe digital coins are issued by a company called Tether Limited that is governed by the laws of the British Virgin Islands, according to the legal part of its website. It is incorporated in Hong Kong. It has emerged that Jan Ludovicus van der Velde is the CEO of cryptocurrency exchange Bitfinex, which has been accused of being involved in the price manipulation of bitcoin, as well as tether. Many people trading on exchanges, including Bitfinex, will use tether to buy other cryptocurrencies like bitcoin. Tether Limited argues that using this method to buy virtual currencies allows users to move fiat in and out of an exchange more quickly and cheaply. Also, exchanges typically have rocky relationships with banks, and using Tether is a way to circumvent that.\r\n\r\nUSDT is fairly simple to use. Once on exchanges like \u003ca href=\"https://www.coingecko.com/en/exchanges/poloniex\"\u003ePoloniex\u003c/a\u003e or Bittrex, it can be used to purchase Bitcoin and other cryptocurrencies. It can be easily transferred from an exchange to any Omni La
yer enabled wallet. Tether has no transaction fees, although external wallets and exchanges may charge one. In order to convert USDT to USD and vise versa through the Tether.to Platform, users must pay a small fee. Buying and selling Tether for Bitcoin can be done through a variety of exchanges like the ones mentioned previously or through the Tether.to platform, which also allows the conversion between USD to and from your bank account."}`,
		}, {
			Chain:       "SUI",
			Address:     "0x27792d9fed7f9844eb4839566001bb6f6cb4804f66aa2da6fe1ee242d896881::coin::COIN",
			Name:        "Wrapped BTC",
			Symbol:      "WBTC",
			Logo:        `{"thumb":"https://assets.coingecko.com/coins/images/7598/thumb/wrapped_bitcoin_wbtc.png?1548822744","small":"https://assets.coingecko.com/coins/images/7598/small/wrapped_bitcoin_wbtc.png?1548822744","large":"https://assets.coingecko.com/coins/images/7598/large/wrapped_bitcoin_wbtc.png?1548822744"}`,
			Decimals:    8,
			CgId:        "wrapped-bitcoin",
			WebSite:     "https://www.wbtc.network/c",
			Description: `{"en":""}`,
		}, {
			Chain:       "SUI",
			Address:     "0x1e8b532cca6569cab9f9b9ebc73f8c13885012ade714729aa3b450e0339ac766::coin::COIN",
			Name:        "Wrapped AVAX",
			Symbol:      "WAVAX",
			Logo:        "https://s2.coinmarketcap.com/static/img/coins/64x64/9462.png",
			Decimals:    8,
			CmcId:       9462,
			CgId:        "wrapped-avax",
			WebSite:     "https://www.avalabs.org/",
			Description: `{"en":""}`,
		}, {
			Chain:       "SUI",
			Address:     "0x6081300950a4f1e2081580e919c210436a1bed49080502834950d31ee55a2396::coin::COIN",
			Name:        "Wrapped Fantom",
			Symbol:      "WFTM",
			Logo:        "https://s2.coinmarketcap.com/static/img/coins/64x64/10240.png",
			Decimals:    8,
			CmcId:       10240,
			CgId:        "wrapped-fantom",
			WebSite:     "http://fantom.foundation",
			Description: `{"en":""}`,
		}, {
			Chain:       "SUI",
			Address:     "0xb7844e289a8410e50fb3ca48d69eb9cf29e27d223ef90353fe1bd8e27ff8f3f8::coin::COIN",
			Name:        "Wrapped SOL",
			Symbol:      "SOL",
			Logo:        "https://s2.coinmarketcap.com/static/img/coins/64x64/16116.png",
			Decimals:    8,
			CmcId:       16116,
			CgId:        "wrapped-solana",
			WebSite:     "https://solana.com/",
			Description: `{"en":"Wrapped Solana "}`,
		}, {
			Chain:       "SUI",
			Address:     "0xb231fcda8bbddb31f2ef02e6161444aec64a514e2c89279584ac9806ce9cf037::coin::COIN",
			Name:        "Celo native asset",
			Symbol:      "CELO",
			Logo:        "https://s2.coinmarketcap.com/static/img/coins/64x64/5567.png",
			Decimals:    8,
			CmcId:       5567,
			CgId:        "celo",
			WebSite:     "https://celo.org/",
			Description: `{"en":"Celo enables participation on the Platform, with the opportunity to earn rewards through network participation. Celo’s stability mechanism and token economics are designed in such a way that demand for cGLD directly increases as demand for Celo Dollars (cUSD) and other stable value assets increases.\r\n\r\ncGLD is a native cryptographic digital asset created at the mainnet release of the Celo Platform. cGLD has no relationship to physical gold.\r\n\r\ncGLD is a utility and governance asset required to participate on the Celo Platform.Some uses include:\r\n\r\nRunning a validator to secure and operate aspects of the Celo Platform\r\nVoting for validators working to secure and operate the Celo Platform\r\nParticipating in governance decisions to influence the future of the Celo Platform\r\nSupporting applications onthe platform\r\ncGLD forms part of the overcollateralized reserve that supports the Celo stable value assets (initially Celo Dollar or cUSD).\r\n\r\nThe Celo Protocol automatically adds cGLD to the reserve whenever the Celo stable value asset supply increases."}`,
		},
	}
	for _, t := range tokenLists {
		c.db.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "address"}, {Name: "chain"}},
			UpdateAll: true,
		}).Create(&t)
	}
}

func Updateevm210425Token() {
	var tokenLists = []models.TokenList{
		{
			Chain:       "evm210425",
			Address:     "0xda396a3c7fc762643f658b47228cd51de6ce936d",
			Name:        "USD Coin",
			Symbol:      "USDC",
			Logo:        "https://raw.githubusercontent.com/cosmostation/cosmostation_token_resource/master/assets/images/ethereum/usdc.png",
			Decimals:    6,
			CmcId:       3408,
			CgId:        "usd-coin",
			WebSite:     "https://www.centre.io/usdc",
			Description: `{"en":"USDC is a fully collateralized US dollar stablecoin. USDC is the bridge between dollars and trading on cryptocurrency exchanges. The technology behind CENTRE makes it possible to exchange value between people, businesses and financial institutions just like email between mail services and texts between SMS providers. We believe by removing artificial economic borders, we can create a more inclusive global economy."}`,
		}, {
			Chain:    "evm210425",
			Address:  "0xeac734fb7581d8eb2ce4949b0896fc4e76769509",
			Name:     "Tether USD",
			Symbol:   "USDT",
			Logo:     `{"thumb":"https://assets.coingecko.com/coins/images/325/thumb/Tether.png?1668148663","small":"https://assets.coingecko.com/coins/images/325/small/Tether.png?1668148663","large":"https://assets.coingecko.com/coins/images/325/large/Tether.png?1668148663"}`,
			Decimals: 6,
			CgId:     "tether",
			WebSite:  "https://tether.to/",
			Description: `{"en":"Tether (USDT) is a cryptocurrency with a value meant to mirror the value of the U.S. dollar. The idea was to create a stable cryptocurrency that can be used like digital dollars. Coins that serve this purpose of being a stable dollar substitute are called “stable coins.” Tether is the most popular stable coin and even acts as a dollar replacement on many popular exchanges! According to their site, Tether converts cash into digital currency, to anchor or “tether” the value of the coin to the price of national currencies like the US dollar, the Euro, and the Yen. Like other cryptos it uses blockchain. Unlike other cryptos, it is [according to the official Tether site] “100% backed by USD” (USD is held in reserve). The primary use of Tether is that it offers some stability to the otherwise volatile crypto space and offers liquidity to exchanges who can’t deal in dollars and with banks (for example to the sometimes controversial but leading exchange \u003ca href=\"https://www.coingecko.com/en/exchanges/bitfinex\"\u003eBitfinex\u003c/a\u003e).\r\n\r\nThe digital coins are issued by a company called Tether Limited that is governed by the laws of the British Virgin Islands, according to the legal part of its website. It is incorporated in Hong Kong. It has emerged that Jan Ludovicus van der Velde is the CEO of cryptocurrency exchange Bitfinex, which has been accused of being involved in the price manipulation of bitcoin, as well as tether. Many people trading on exchanges, including Bitfinex, will use tether to buy other cryptocurrencies like bitcoin. Tether Limited argues that using this method to buy virtual currencies allows users to move fiat in and out of an exchange more quickly and cheaply. Also, exchanges typically have rocky relationships with banks, and using Tether is a way to circumvent that.\r\n\r\nUSDT is fairly simple to use. Once on exchanges like \u003ca href=\"https://www.coingecko.com/en/exchanges/poloniex\"\u003ePoloniex\u003c/a\u003e or Bittrex, it can be used to purchase Bitcoin and other cryptocurrencies. It can be easily transferred from an exchange to any Omni La
yer enabled wallet. Tether has no transaction fees, although external wallets and exchanges may charge one. In order to convert USDT to USD and vise versa through the Tether.to Platform, users must pay a small fee. Buying and selling Tether for Bitcoin can be done through a variety of exchanges like the ones mentioned previously or through the Tether.to platform, which also allows the conversion between USD to and from your bank account."}`,
		},
	}
	for _, t := range tokenLists {
		c.db.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "address"}, {Name: "chain"}},
			UpdateAll: true,
		}).Create(&t)
	}
}

func UpdateLineaToken() {
	var tokenLists = []models.TokenList{
		{
			Chain:       "Linea",
			Address:     strings.ToLower("0x66627F389ae46D881773B7131139b2411980E09E"),
			Name:        "USD Coin",
			Symbol:      "USDC",
			Logo:        "https://raw.githubusercontent.com/cosmostation/cosmostation_token_resource/master/assets/images/ethereum/usdc.png",
			Decimals:    6,
			CmcId:       3408,
			CgId:        "usd-coin",
			WebSite:     "https://www.centre.io/usdc",
			Description: `{"en":"USDC is a fully collateralized US dollar stablecoin. USDC is the bridge between dollars and trading on cryptocurrency exchanges. The technology behind CENTRE makes it possible to exchange value between people, businesses and financial institutions just like email between mail services and texts between SMS providers. We believe by removing artificial economic borders, we can create a more inclusive global economy."}`,
		}, {
			Chain:    "Linea",
			Address:  strings.ToLower("0x6C6470898882b65E0275723d883a0d877aade43f"),
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
			Chain:       "Linea",
			Address:     strings.ToLower("0xf5C6825015280CdfD0b56903F9F8B5A2233476F5"),
			Name:        "Binance Coin",
			Symbol:      "BNB",
			Logo:        `https://s2.coinmarketcap.com/static/img/coins/64x64/7192.png`,
			Decimals:    18,
			CmcId:       7192,
			CgId:        "binance-coin-wormhole",
			WebSite:     "https://www.binance.org/",
			Description: `{"en":""}`,
		},
		{
			Chain:       "Linea",
			Address:     strings.ToLower("0x7d43aabc515c356145049227cee54b608342c0ad"),
			Name:        "Binance USD",
			Symbol:      "BUSD",
			Logo:        `https://s2.coinmarketcap.com/static/img/coins/64x64/4687.png`,
			Decimals:    18,
			CmcId:       4687,
			CgId:        "binance-usd",
			WebSite:     "https://www.binance.com/en/busd",
			Description: `{"en":"Binance USD (BUSD) is a stable coin pegged to USD that has received approval from the New York State Department of Financial Services (NYDFS). BUSD will be available for direct purchase and redemption at a rate of 1 BUSD = 1 USD."}`,
		},
		{
			Chain:       "Linea",
			Address:     strings.ToLower("0xe5D7C2a44FfDDf6b295A15c148167daaAf5Cf34f"),
			Name:        "Wrapped Ether",
			Symbol:      "WETH",
			Logo:        `{"thumb":"https://assets.coingecko.com/coins/images/2518/thumb/weth.png?1628852295","small":"https://assets.coingecko.com/coins/images/2518/small/weth.png?1628852295","large":"https://assets.coingecko.com/coins/images/2518/large/weth.png?1628852295"}`,
			Decimals:    18,
			CgId:        "weth",
			WebSite:     "https://weth.io/",
			Description: `{"en":"What is WETH (Wrapped ETH)?\r\nWETH is the tokenized/packaged form of ETH that you use to pay for items when you interact with Ethereum dApps. WETH follows the ERC-20 token standards, enabling it to achieve interoperability with other ERC-20 tokens. \r\n\r\nThis offers more utility to holders as they can use it across networks and dApps. You can stake, yield farm, lend, and provide liquidity to various liquidity pools with WETH. \r\n\r\nAlso, unlike ETH, which doesn’t conform to its own ERC-20 standard and thus has lower interoperability as it can’t be used on other chains besides Ethereum, WETH can be used on cheaper and high throughput alternatives like Binance, Polygon, Solana, and Cardano.\r\n\r\nThe price of WETH will always be the same as ETH because it maintains a 1:1 wrapping ratio.\r\n\r\nHow to Wrap ETH?\r\nCustodians wrap and unwrap ETH. To wrap ETH, you send ETH to a custodian. This can be a multi-sig wallet, a Decentralized Autonomous Organization (DAO), or a smart contract. After connecting your web3 wallet to a DeFi exchange, you enter the amount of ETH you wish to wrap and click the swap function. Once the transaction is confirmed, you will receive WETH tokens equivalent to the ETH that you’ve swapped.\r\n\r\nOn a centralized exchange, the exchange burns the deposited ETH and mints a wrapped form for you. And when you want to unwrap it, the exchange will burn the wrapped version and mint the ETH on your behalf.\r\n\r\nWhat’s Next for WETH?\r\nAccording to the developers, hopefully there will be no future for WETH. According to the website, steps are being taken to update ETH to make it compliant with its own ERC-20 standards."}`,
		},
		{
			Chain:       "Linea",
			Address:     strings.ToLower("0x265b25e22bcd7f10a5bd6e6410f10537cc7567e8"),
			Name:        "Matic Token",
			Symbol:      "MATIC",
			Logo:        `https://s2.coinmarketcap.com/static/img/coins/64x64/3890.png`,
			Decimals:    18,
			CmcId:       3890,
			CgId:        "matic-network",
			WebSite:     "https://polygon.technology/",
			Description: `{"en":""}`,
		},
		{
			Chain:       "Linea",
			Address:     strings.ToLower("0x5471ea8f739dd37E9B81Be9c5c77754D8AA953E4"),
			Name:        "Avalanche",
			Symbol:      "AVAX",
			Logo:        `https://s2.coinmarketcap.com/static/img/coins/64x64/5805.png`,
			Decimals:    18,
			CmcId:       5805,
			CgId:        "avalanche-2",
			WebSite:     "https://avax.network/",
			Description: `{"en":""}`,
		},
		{
			Chain:       "Linea",
			Address:     strings.ToLower("0x60D01EC2D5E98Ac51C8B4cF84DfCCE98D527c747"),
			Name:        "izumi Token",
			Symbol:      "IZI",
			Logo:        "https://s2.coinmarketcap.com/static/img/coins/64x64/16305.png",
			Decimals:    18,
			CmcId:       16305,
			CgId:        "izumi-finance",
			WebSite:     "https://izumi.finance/home",
			Description: `{"en":"izumi Finance is a platform providing liquidity as a Service (LaaS) with Uniswap V3 and a built-in multi-chain dex. LiquidBox proposes innovative liquidity mining protocols to help protocols attract liquidity efficiently by distributing incentives in certain price ranges."}`,
		},
		{
			Chain:       "Linea",
			Address:     strings.ToLower("0x0A3BB08b3a15A19b4De82F8AcFc862606FB69A2D"),
			Name:        "iZUMi Bond USD",
			Symbol:      "IUSD",
			Logo:        "https://s2.coinmarketcap.com/static/img/coins/64x64/19799.png",
			Decimals:    18,
			CmcId:       19799,
			CgId:        "izumi-bond-usd",
			WebSite:     "https://izumi.finance",
			Description: `{"en":"iUSD, in its full name, iZUMi Bond USD, is 100% backed by iZUMi’s collaterals and future revenues. iUSD is 1:1 pegged to USD, issued by iZUMi Finance and sold to private investors as a bond to raise funds for future development of iZUMi’s ecosystem."}`,
		},
	}
	for _, t := range tokenLists {
		c.db.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "address"}, {Name: "chain"}},
			UpdateAll: true,
		}).Create(&t)
	}
}

func UpdateEvm8453Token() {
	var tokenLists = []models.TokenList{
		{
			Chain:       "evm8453",
			Address:     strings.ToLower("0x50c5725949A6F0c72E6C4a641F24049A917DB0Cb"),
			Name:        "Dai Stablecoin",
			Symbol:      "DAI",
			Logo:        `{"thumb":"https://assets.coingecko.com/coins/images/9956/thumb/4943.png?1636636734","small":"https://assets.coingecko.com/coins/images/9956/small/4943.png?1636636734","large":"https://assets.coingecko.com/coins/images/9956/large/4943.png?1636636734"}`,
			Decimals:    18,
			CgId:        "dai",
			WebSite:     "https://makerdao.com/",
			Description: `{"en":"MakerDAO has launched Multi-collateral DAI (MCD). This token refers to the new DAI that is collaterized by multiple assets.\r\n"}`,
		},
		{
			Chain:       "evm8453",
			Address:     strings.ToLower("0x4A3A6Dd60A34bB2Aba60D73B4C88315E9CeB6A3D"),
			Name:        "Magic Internet Money",
			Symbol:      "MIM",
			Logo:        `https://s2.coinmarketcap.com/static/img/coins/64x64/162.png`,
			Decimals:    18,
			CgId:        "magic-internet-money",
			CmcId:       162,
			WebSite:     "https://abracadabra.money/",
			Description: `{"en":"Abracadabra.money is a lending platform that allows users to mint the Stablecoin MIM using interest bearing tokens as collaterals!"}`,
		},
		{
			Chain:       "evm8453",
			Address:     strings.ToLower("0x8Fbd0648971d56f1f2c35Fa075Ff5Bc75fb0e39D"),
			Name:        "MBS",
			Symbol:      "MBS",
			Logo:        `https://s2.coinmarketcap.com/static/img/coins/64x64/13011.png`,
			Decimals:    18,
			CgId:        "monkeyball",
			CmcId:       13011,
			WebSite:     "https://www.monkeyleague.io/",
			Description: `{"en":"MonkeyLeague is the next-gen esports metaverse that enables players to Create, Play, Compete, and Earn.\r\n\r\nMonkeyLeague combines high-production-value, multiplayer gaming with Solana blockchain, NFTs, and decentralized finance to deliver an exciting, turn-based, play-to-earn soccer game that’s easy to learn yet hard to master.\r\n\r\nPlay in Three Modes:\r\n1. Player vs Environment: Training mode played against the computer\r\n2. Player vs Player: Classic game where each team is controlled by users\r\n3. Team vs Team: Each team is controlled and played by multiple users"}`,
		},
		{
			Chain:       "evm8453",
			Address:     strings.ToLower("0xbd2DBb8eceA9743CA5B16423b4eAa26bDcfE5eD2"),
			Name:        "Synth Token",
			Symbol:      "SYNTH",
			Logo:        `https://assets.coingecko.com/coins/images/31190/large/synth.png?1691145281`,
			Decimals:    18,
			CgId:        "synthswap",
			WebSite:     "https://synthswap.io",
			Description: `{"en": "What is the project about?\r\nFirst fully audited and native DEX on Base. We will also offer V3 and V4 tech alongside other DeFi products\r\n\r\n\r\nWhat makes your project unique?\r\nSynthswap is one of the first decentralized exchanges (DEX) with an automated market-maker (AMM) in the Base ecosystem. \r\nCompared to its competitors, Synthswap will enable trading with the lowest fees! Rewards from Staking and Yield Farming will be among the most lucrative.\r\n\r\nHistory of your project.\r\nLaunched today\r\n\r\nWhat’s next for your project?\r\nBring more features like V3, V4, more token utility and grow the dexes volume and userbase\r\n\r\nWhat can your token be used for?\r\nStake for real yield (platform fee sharing), Governance, Launchpad allocation.\r\n"}`,
		},
		{
			Chain:       "evm8453",
			Address:     strings.ToLower("0x01CC6b33c63CeE896521D63451896C14D42D05Ea"),
			Name:        "Synth escrowed token",
			Symbol:      "XSYNTH",
			Logo:        `https://assets.coingecko.com/coins/images/31190/large/synth.png?1691145281`,
			Decimals:    18,
			CgId:        "synthswap",
			WebSite:     "https://synthswap.io",
			Description: `{"en": "What is the project about?\r\nFirst fully audited and native DEX on Base. We will also offer V3 and V4 tech alongside other DeFi products\r\n\r\n\r\nWhat makes your project unique?\r\nSynthswap is one of the first decentralized exchanges (DEX) with an automated market-maker (AMM) in the Base ecosystem. \r\nCompared to its competitors, Synthswap will enable trading with the lowest fees! Rewards from Staking and Yield Farming will be among the most lucrative.\r\n\r\nHistory of your project.\r\nLaunched today\r\n\r\nWhat’s next for your project?\r\nBring more features like V3, V4, more token utility and grow the dexes volume and userbase\r\n\r\nWhat can your token be used for?\r\nStake for real yield (platform fee sharing), Governance, Launchpad allocation.\r\n"}`,
		},
		{
			Chain:       "evm8453",
			Address:     strings.ToLower("0x4E0dA40b9063dC48364C1C0fFB4AE9d091fc2270"),
			Name:        "Edgeware",
			Symbol:      "EDG",
			Logo:        `https://s2.coinmarketcap.com/static/img/coins/64x64/1596.png`,
			Decimals:    18,
			CgId:        "edgeless",
			CmcId:       1596,
			WebSite:     "https://edgeware.community/",
			Description: `{"en":"EDG is a utility token used in Edgeless casino. It is the branded gaming currency of the casino and one of the best altcoins on the Smartchain and Ethereum blockchain.\r\nErc20 token will switch from Ethereum network to Binance Smartchain (Bsc) network on 15.09.2121.\r\nTotal and maximum supply is 100000 Edgeless Tokens, It can be staked.\r\nEdgeless has Jackpots unmatched by any other casino in the industry: all Jackpots are publicly viewed on the blockchain.\r\nThe EDG token is used as a utility token on the Edgeless platform.\r\nWith EDG token you can play 6+ games in the casino."}`,
		},
		{
			Chain:       "evm8453",
			Address:     strings.ToLower("0x2Ae3F1Ec7F1F5012CFEab0185bfc7aa3cf0DEc22"),
			Name:        "Coinbase Wrapped Staked ETH",
			Symbol:      "CBETH",
			Logo:        `https://s2.coinmarketcap.com/static/img/coins/64x64/21535.png`,
			Decimals:    18,
			CgId:        "coinbase-wrapped-staked-eth",
			CmcId:       21535,
			WebSite:     "https://www.coinbase.com/ ",
			Description: `{"en":""}`,
		},
		{
			Chain:       "evm8453",
			Address:     strings.ToLower("0xd9aAEc86B65D86f6A7B5B1b0c42FFA531710b6CA"),
			Name:        "USD Base Coin",
			Symbol:      "USDBC",
			Logo:        `https://assets.coingecko.com/coins/images/31164/large/baseusdc.jpg?1691075428`,
			Decimals:    6,
			CgId:        "bridged-usd-coin-base",
			WebSite:     "https://help.coinbase.com/en/coinbase/getting-started/crypto-education/usd-base-coin",
			Description: `{"en":"Bridged USD Coin (Base)"}`,
		},
		{
			Chain:       "evm8453",
			Address:     strings.ToLower("0x0963a1aBAF36Ca88C21032b82e479353126A1C4b"),
			Name:        "leetswap.finance",
			Symbol:      "LEET",
			Logo:        `https://assets.coingecko.com/coins/images/31090/large/leet.png?1690964578`,
			Decimals:    18,
			CgId:        "leet",
			WebSite:     "https://leetswap.finance/",
			Description: `{"en":"LeetSwap is an emerging DEX and DeFi ecosystem built on the most-prominent blockchains with a focus on retaining a secure, fast, and user-friendly experience while introducing novel approaches and ideas to the space. LEET is our signature DEX token used to power liquidity mining rewards. Currently live on Linea, Polygon zkEVM, and Canto."}`,
		},
		{
			Chain:       "evm8453",
			Address:     strings.ToLower("0xB3CE4Ce692e035720B25880f678871CfC5a863CA"),
			Name:        "BaseTools",
			Symbol:      "BASE",
			Logo:        `https://assets.coingecko.com/coins/images/31127/large/logo_%2811%29_%281%29.png?1690778566`,
			Decimals:    18,
			CgId:        "basetools",
			WebSite:     "https://basetools.tech/",
			Description: `{"en":"Base Tools is a token that provides powerful Telegram bots such as the Decompiler, Basechain New Tokens, honeypot scanner and Sniper Bot."}`,
		},
		{
			Chain:       "evm8453",
			Address:     strings.ToLower("0x8544FE9D190fD7EC52860abBf45088E81Ee24a8c"),
			Name:        "Toshi",
			Symbol:      "TOSHI",
			Logo:        `https://assets.coingecko.com/coins/images/31126/large/5dajOmhM_400x400.jpg?1690777236`,
			Decimals:    18,
			CgId:        "toshi",
			WebSite:     "https://www.toshithecat.com/",
			Description: `{"en":""}`,
		},
		{
			Chain:       "evm8453",
			Address:     strings.ToLower("0x6653dD4B92a0e5Bf8ae570A98906d9D6fD2eEc09"),
			Name:        "RocketSwap",
			Symbol:      "RCKT",
			Logo:        `https://basescan.org/token/images/rocketswap_32.png`,
			Decimals:    18,
			WebSite:     "https://app.rocketswap.cc/",
			Description: `{"en":""}`,
		},
	}
	for _, t := range tokenLists {
		c.db.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "address"}, {Name: "chain"}},
			UpdateAll: true,
		}).Create(&t)
	}
}

func UpdateSEIToken() {
	var tokenLists = []models.TokenList{
		{
			Chain:       "Sei",
			Address:     "factory/sei189adguawugk3e55zn63z8r9ll29xrjwca636ra7v7gxuzn98sxyqwzt47l/4tLQqCLaoKKfNFuPjA9o39YbKUwhR1F8N29Tz3hEbfP2",
			Name:        "Wrapped Ether",
			Symbol:      "WETH",
			Logo:        "https://s2.coinmarketcap.com/static/img/coins/64x64/2396.png",
			Decimals:    18,
			CmcId:       2396,
			CgId:        "weth",
			WebSite:     "https://weth.io/",
			Description: `{"en":"W-ETH is \"wrapped ETH\" but let's start by introducing the players. First, there's Ether token. Ether or ETH is the native currency built on the Ethereum blockchain.\r\nSecond, there are alt tokens. When a dApp (decentralized app) is built off of the Ethereum Blockchain it usually implements it’s own form of Token. Think Augur’s REP Token, or Bancor's BNT Token. Finally the ERC-20 standard. ERC20 is a standard developed after the release of ETH that defines how tokens are transferred and how to keep a consistent record of those transfers among tokens in the Ethereum Network."}`,
		},
		{
			Chain:       "Sei",
			Address:     "factory/sei189adguawugk3e55zn63z8r9ll29xrjwca636ra7v7gxuzn98sxyqwzt47l/Hq4tuDzhRBnxw3tFA5n6M52NVMVcC19XggbyDiJKCD6H",
			Name:        "USD Coin",
			Symbol:      "USDCet",
			Logo:        `{"thumb":"https://assets.coingecko.com/coins/images/6319/thumb/USD_Coin_icon.png?1547042389","small":"https://assets.coingecko.com/coins/images/6319/small/USD_Coin_icon.png?1547042389","large":"https://assets.coingecko.com/coins/images/6319/large/USD_Coin_icon.png?1547042389"}`,
			Decimals:    6,
			CgId:        "usd-coin",
			WebSite:     "https://www.circle.com/en/usdc",
			Description: `{"en":"USDC is a fully collateralized US dollar stablecoin. USDC is the bridge between dollars and trading on cryptocurrency exchanges. The technology behind CENTRE makes it possible to exchange value between people, businesses and financial institutions just like email between mail services and texts between SMS providers. We believe by removing artificial economic borders, we can create a more inclusive global economy."}`,
		},
		{
			Chain:       "Sei",
			Address:     "factory/sei189adguawugk3e55zn63z8r9ll29xrjwca636ra7v7gxuzn98sxyqwzt47l/7edDfnf4mku8So3t4Do215GNHwASEwCWrdhM5GqD51xZ",
			Name:        "USD Coin",
			Symbol:      "USDCar",
			Logo:        `{"thumb":"https://assets.coingecko.com/coins/images/6319/thumb/USD_Coin_icon.png?1547042389","small":"https://assets.coingecko.com/coins/images/6319/small/USD_Coin_icon.png?1547042389","large":"https://assets.coingecko.com/coins/images/6319/large/USD_Coin_icon.png?1547042389"}`,
			Decimals:    6,
			CgId:        "usd-coin",
			WebSite:     "https://www.circle.com/en/usdc",
			Description: `{"en":"USDC is a fully collateralized US dollar stablecoin. USDC is the bridge between dollars and trading on cryptocurrency exchanges. The technology behind CENTRE makes it possible to exchange value between people, businesses and financial institutions just like email between mail services and texts between SMS providers. We believe by removing artificial economic borders, we can create a more inclusive global economy."}`,
		},
		{
			Chain:       "Sei",
			Address:     "factory/sei189adguawugk3e55zn63z8r9ll29xrjwca636ra7v7gxuzn98sxyqwzt47l/DUVFMY2neJdL8aE4d3stcpttDDm5aoyfGyVvm29iA9Yp",
			Name:        "USD Coin",
			Symbol:      "USDCpo",
			Logo:        `{"thumb":"https://assets.coingecko.com/coins/images/6319/thumb/USD_Coin_icon.png?1547042389","small":"https://assets.coingecko.com/coins/images/6319/small/USD_Coin_icon.png?1547042389","large":"https://assets.coingecko.com/coins/images/6319/large/USD_Coin_icon.png?1547042389"}`,
			Decimals:    6,
			CgId:        "usd-coin",
			WebSite:     "https://www.circle.com/en/usdc",
			Description: `{"en":"USDC is a fully collateralized US dollar stablecoin. USDC is the bridge between dollars and trading on cryptocurrency exchanges. The technology behind CENTRE makes it possible to exchange value between people, businesses and financial institutions just like email between mail services and texts between SMS providers. We believe by removing artificial economic borders, we can create a more inclusive global economy."}`,
		},
		{
			Chain:       "Sei",
			Address:     "factory/sei189adguawugk3e55zn63z8r9ll29xrjwca636ra7v7gxuzn98sxyqwzt47l/3VKKYtbQ9iq8f9CaZfgR6Cr3TUj6ypXPAn6kco6wjcAu",
			Name:        "USD Coin",
			Symbol:      "USDCop",
			Logo:        `{"thumb":"https://assets.coingecko.com/coins/images/6319/thumb/USD_Coin_icon.png?1547042389","small":"https://assets.coingecko.com/coins/images/6319/small/USD_Coin_icon.png?1547042389","large":"https://assets.coingecko.com/coins/images/6319/large/USD_Coin_icon.png?1547042389"}`,
			Decimals:    6,
			CgId:        "usd-coin",
			WebSite:     "https://www.circle.com/en/usdc",
			Description: `{"en":"USDC is a fully collateralized US dollar stablecoin. USDC is the bridge between dollars and trading on cryptocurrency exchanges. The technology behind CENTRE makes it possible to exchange value between people, businesses and financial institutions just like email between mail services and texts between SMS providers. We believe by removing artificial economic borders, we can create a more inclusive global economy."}`,
		},
		{
			Chain:       "Sei",
			Address:     "factory/sei189adguawugk3e55zn63z8r9ll29xrjwca636ra7v7gxuzn98sxyqwzt47l/9fELvUhFo6yWL34ZaLgPbCPzdk9MD1tAzMycgH45qShH",
			Name:        "USD Coin",
			Symbol:      "USDCso",
			Logo:        `{"thumb":"https://assets.coingecko.com/coins/images/6319/thumb/USD_Coin_icon.png?1547042389","small":"https://assets.coingecko.com/coins/images/6319/small/USD_Coin_icon.png?1547042389","large":"https://assets.coingecko.com/coins/images/6319/large/USD_Coin_icon.png?1547042389"}`,
			Decimals:    6,
			CgId:        "usd-coin",
			WebSite:     "https://www.circle.com/en/usdc",
			Description: `{"en":"USDC is a fully collateralized US dollar stablecoin. USDC is the bridge between dollars and trading on cryptocurrency exchanges. The technology behind CENTRE makes it possible to exchange value between people, businesses and financial institutions just like email between mail services and texts between SMS providers. We believe by removing artificial economic borders, we can create a more inclusive global economy."}`,
		},
		{
			Chain:    "Sei",
			Address:  "factory/sei189adguawugk3e55zn63z8r9ll29xrjwca636ra7v7gxuzn98sxyqwzt47l/871jbn9unTavWsAe83f2Ma9GJWSv6BKsyWYLiQ6z3Pva",
			Name:     "Binance-Peg BSC-USD",
			Symbol:   "USDTbs",
			Logo:     `{"thumb":"https://assets.coingecko.com/coins/images/325/thumb/Tether.png?1668148663","small":"https://assets.coingecko.com/coins/images/325/small/Tether.png?1668148663","large":"https://assets.coingecko.com/coins/images/325/large/Tether.png?1668148663"}`,
			Decimals: 6,
			CgId:     "tether",
			WebSite:  "https://tether.to/",
			Description: `{"en":"Tether (USDT) is a cryptocurrency with a value meant to mirror the value of the U.S. dollar. The idea was to create a stable cryptocurrency that can be used like digital dollars. Coins that serve this purpose of being a stable dollar substitute are called “stable coins.” Tether is the most popular stable coin and even acts as a dollar replacement on many popular exchanges! According to their site, Tether converts cash into digital currency, to anchor or “tether” the value of the coin to the price of national currencies like the US dollar, the Euro, and the Yen. Like other cryptos it uses blockchain. Unlike other cryptos, it is [according to the official Tether site] “100% backed by USD” (USD is held in reserve). The primary use of Tether is that it offers some stability to the otherwise volatile crypto space and offers liquidity to exchanges who can’t deal in dollars and with banks (for example to the sometimes controversial but leading exchange \u003ca href=\"https://www.coingecko.com/en/exchanges/bitfinex\"\u003eBitfinex\u003c/a\u003e).\r\n\r\nThe digital coins are issued by a company called Tether Limited that is governed by the laws of the British Virgin Islands, according to the legal part of its website. It is incorporated in Hong Kong. It has emerged that Jan Ludovicus van der Velde is the CEO of cryptocurrency exchange Bitfinex, which has been accused of being involved in the price manipulation of bitcoin, as well as tether. Many people trading on exchanges, including Bitfinex, will use tether to buy other cryptocurrencies like bitcoin. Tether Limited argues that using this method to buy virtual currencies allows users to move fiat in and out of an exchange more quickly and cheaply. Also, exchanges typically have rocky relationships with banks, and using Tether is a way to circumvent that.\r\n\r\nUSDT is fairly simple to use. Once on exchanges like \u003ca href=\"https://www.coingecko.com/en/exchanges/poloniex\"\u003ePoloniex\u003c/a\u003e or Bittrex, it can be used to purchase Bitcoin and other cryptocurrencies. It can be easily transferred from an exchange to any Omni La
yer enabled wallet. Tether has no transaction fees, although external wallets and exchanges may charge one. In order to convert USDT to USD and vise versa through the Tether.to Platform, users must pay a small fee. Buying and selling Tether for Bitcoin can be done through a variety of exchanges like the ones mentioned previously or through the Tether.to platform, which also allows the conversion between USD to and from your bank account."}`,
		},
		{
			Chain:       "Sei",
			Address:     "factory/sei189adguawugk3e55zn63z8r9ll29xrjwca636ra7v7gxuzn98sxyqwzt47l/7omXa4gryZ5NiBmLep7JsTtTtANCVKXwT9vbN91aS1br",
			Name:        "Wrapped BTC",
			Symbol:      "WBTC",
			Logo:        `{"thumb":"https://assets.coingecko.com/coins/images/7598/thumb/wrapped_bitcoin_wbtc.png?1548822744","small":"https://assets.coingecko.com/coins/images/7598/small/wrapped_bitcoin_wbtc.png?1548822744","large":"https://assets.coingecko.com/coins/images/7598/large/wrapped_bitcoin_wbtc.png?1548822744"}`,
			Decimals:    8,
			CgId:        "wrapped-bitcoin",
			WebSite:     "https://www.wbtc.network/c",
			Description: `{"en":""}`,
		},
	}
	for _, t := range tokenLists {
		c.db.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "address"}, {Name: "chain"}},
			UpdateAll: true,
		}).Create(&t)
	}
}

func UpdateArbitrumToken() {
	var tokenLists = []models.TokenList{
		{
			Chain:       "arbitrum-one",
			Address:     "0xaf88d065e77c8cc2239327c5edb3a432268e5831",
			Name:        "USD Coin",
			Symbol:      "USDC",
			Logo:        "https://s2.coinmarketcap.com/static/img/coins/64x64/3408.png",
			Decimals:    6,
			CmcId:       3408,
			CgId:        "usd-coin",
			WebSite:     "https://www.centre.io/usdc",
			Description: `{"en":"USDC is a fully collateralized US dollar stablecoin. USDC is the bridge between dollars and trading on cryptocurrency exchanges. The technology behind CENTRE makes it possible to exchange value between people, businesses and financial institutions just like email between mail services and texts between SMS providers. We believe by removing artificial economic borders, we can create a more inclusive global economy."}`,
		},
	}
	for _, t := range tokenLists {
		c.db.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "address"}, {Name: "chain"}},
			UpdateAll: true,
		}).Create(&t)
	}
}

func UpdateEthereumToken() {
	var tokenLists = []models.TokenList{
		{
			Chain:       "ethereum",
			Address:     "0x64bc2ca1be492be7185faa2c8835d9b824c8a194",
			Name:        "Big Time",
			Symbol:      "BIGTIME",
			Logo:        "https://assets.coingecko.com/coins/images/32251/large/-6136155493475923781_121.jpg?1696998691",
			Decimals:    18,
			CgId:        "big-time",
			WebSite:     "https://bigtime.gg/",
			Description: `{"en": "Big Time is a free-to-play, multiplayer action RPG game that combines fast-action combat and an adventure through time and space.\r\n \r\nExplore ancient mysteries and futuristic civilizations as you battle your way through history. Pick up rare collectibles Loot, Cosmetics and Tokens as you fight and defeat enemies. Collect and trade your Collectibles to decorate your avatar and personal metaverse, where you can hang out with your friends.\r\n\r\nExpand your personal metaverse and production capabilities with SPACE to join our in-game creator economy. \r\n\r\nPlay for free, collect in-game items and tokens, produce Collectibles, or hang out with friends...  Limitless environments and Adventure Instances give you INFINITE possibilities. The gameplay options are endless. "}`,
		},
	}
	for _, t := range tokenLists {
		c.db.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "address"}, {Name: "chain"}},
			UpdateAll: true,
		}).Create(&t)
	}
}

func UpdateScrollToken() {
	var tokenLists = []models.TokenList{
		{
			Chain:       "scroll",
			Address:     "0x06efdbff2a14a7c8e15944d1f4a48f9f95f663a4",
			Name:        "USD Coin",
			Symbol:      "USDC",
			Logo:        "https://s2.coinmarketcap.com/static/img/coins/64x64/3408.png",
			Decimals:    6,
			CmcId:       3408,
			CgId:        "usd-coin",
			WebSite:     "https://www.centre.io/usdc",
			Description: `{"en":"USDC is a fully collateralized US dollar stablecoin. USDC is the bridge between dollars and trading on cryptocurrency exchanges. The technology behind CENTRE makes it possible to exchange value between people, businesses and financial institutions just like email between mail services and texts between SMS providers. We believe by removing artificial economic borders, we can create a more inclusive global economy."}`,
		},
		{
			Chain:    "scroll",
			Address:  "0xf55bec9cafdbe8730f096aa55dad6d22d44099df",
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
			Chain:       "scroll",
			Address:     "0x5300000000000000000000000000000000000004",
			Name:        "Wrapped Ether",
			Symbol:      "WETH",
			Logo:        `{"thumb":"https://assets.coingecko.com/coins/images/2518/thumb/weth.png?1628852295","small":"https://assets.coingecko.com/coins/images/2518/small/weth.png?1628852295","large":"https://assets.coingecko.com/coins/images/2518/large/weth.png?1628852295"}`,
			Decimals:    18,
			CgId:        "weth",
			WebSite:     "https://weth.io/",
			Description: `{"en":"What is WETH (Wrapped ETH)?\r\nWETH is the tokenized/packaged form of ETH that you use to pay for items when you interact with Ethereum dApps. WETH follows the ERC-20 token standards, enabling it to achieve interoperability with other ERC-20 tokens. \r\n\r\nThis offers more utility to holders as they can use it across networks and dApps. You can stake, yield farm, lend, and provide liquidity to various liquidity pools with WETH. \r\n\r\nAlso, unlike ETH, which doesn’t conform to its own ERC-20 standard and thus has lower interoperability as it can’t be used on other chains besides Ethereum, WETH can be used on cheaper and high throughput alternatives like Binance, Polygon, Solana, and Cardano.\r\n\r\nThe price of WETH will always be the same as ETH because it maintains a 1:1 wrapping ratio.\r\n\r\nHow to Wrap ETH?\r\nCustodians wrap and unwrap ETH. To wrap ETH, you send ETH to a custodian. This can be a multi-sig wallet, a Decentralized Autonomous Organization (DAO), or a smart contract. After connecting your web3 wallet to a DeFi exchange, you enter the amount of ETH you wish to wrap and click the swap function. Once the transaction is confirmed, you will receive WETH tokens equivalent to the ETH that you’ve swapped.\r\n\r\nOn a centralized exchange, the exchange burns the deposited ETH and mints a wrapped form for you. And when you want to unwrap it, the exchange will burn the wrapped version and mint the ETH on your behalf.\r\n\r\nWhat’s Next for WETH?\r\nAccording to the developers, hopefully there will be no future for WETH. According to the website, steps are being taken to update ETH to make it compliant with its own ERC-20 standards."}`,
		},
		{
			Chain:       "scroll",
			Address:     "0x36f983124b027781216adc94c4d81ef4026ffcdd",
			Name:        "Scroll Doge",
			Symbol:      "ZKDOGE",
			Logo:        `{"thumb":"https://assets.coingecko.com/coins/images/32271/thumb/8d7453f7-340b-4327-9529-180572813a7a.jpeg?1697179075","small":"https://assets.coingecko.com/coins/images/32271/small/8d7453f7-340b-4327-9529-180572813a7a.jpeg?1697179075","large":"https://assets.coingecko.com/coins/images/32271/large/8d7453f7-340b-4327-9529-180572813a7a.jpeg?1697179075"}`,
			Decimals:    9,
			CgId:        "scroll-doge",
			WebSite:     "https://scrolldog3.xyz/",
			Description: `{"en": "Scroll Doge is first doge on Scroll chain with cryptocurrency based on the popular Dogecoin. It operates on the Scroll Chain blockchain, offering fast and low-cost transactions. With its fun and playful branding is perfect for fun-loving enthusiasts."}`,
		},
		{
			Chain:       "scroll",
			Address:     "0x9daea97fd467d704c583beda2454e3da27097b60",
			Name:        "Scrollswap Finance",
			Symbol:      "SF",
			Logo:        `{"thumb":"https://assets.coingecko.com/coins/images/32269/thumb/h4lCoyvo_400x400.jpg?1697177957","small":"https://assets.coingecko.com/coins/images/32269/small/h4lCoyvo_400x400.jpg?1697177957","large":https://assets.coingecko.com/coins/images/32269/large/h4lCoyvo_400x400.jpg?1697177957"}`,
			Decimals:    18,
			CgId:        "scrollswap-finance",
			WebSite:     "https://dex.scrollswap.app/",
			Description: `{"en": ""}`,
		},
		{
			Chain:       "scroll",
			Address:     "0x3c1bca5a656e69edcd0d4e36bebb3fcdaca60cf1",
			Name:        "Wrapped BTC",
			Symbol:      "WBTC",
			Logo:        `{"thumb":"https://assets.coingecko.com/coins/images/7598/thumb/wrapped_bitcoin_wbtc.png?1548822744","small":"https://assets.coingecko.com/coins/images/7598/small/wrapped_bitcoin_wbtc.png?1548822744","large":"https://assets.coingecko.com/coins/images/7598/large/wrapped_bitcoin_wbtc.png?1548822744"}`,
			Decimals:    8,
			CgId:        "wrapped-bitcoin",
			WebSite:     "https://www.wbtc.network/c",
			Description: `{"en":""}`,
		},
		{
			Chain:       "scroll",
			Address:     "0xca77eb3fefe3725dc33bccb54edefc3d9f764f97",
			Name:        "Dai Stablecoin",
			Symbol:      "DAI",
			Logo:        `{"thumb":"https://assets.coingecko.com/coins/images/9956/thumb/4943.png?1636636734","small":"https://assets.coingecko.com/coins/images/9956/small/4943.png?1636636734","large":"https://assets.coingecko.com/coins/images/9956/large/4943.png?1636636734"}`,
			Decimals:    18,
			CgId:        "dai",
			WebSite:     "https://makerdao.com/",
			Description: `{"en":"MakerDAO has launched Multi-collateral DAI (MCD). This token refers to the new DAI that is collaterized by multiple assets.\r\n"}`,
		},
		{
			Chain:       "scroll",
			Address:     "0x466da81a766ff502fe8ea318d4d1e044b1bf84c1",
			Name:        "cUSDC",
			Symbol:      "CUSDC",
			Logo:        `{"thumb":"https://assets.coingecko.com/coins/images/9442/thumb/Compound_USDC.png?1696509534","small":"https://assets.coingecko.com/coins/images/9442/small/Compound_USDC.png?1696509534","large":"https://assets.coingecko.com/coins/images/9442/large/Compound_USDC.png?1696509534"}`,
			Decimals:    8,
			CgId:        "compound-usd-coin",
			WebSite:     "https://compound.finance/",
			Description: `{"en":"Compound is an open-source, autonomous protocol built for developers, to unlock a universe of new financial applications. Interest and borrowing, for the open financial system."}`,
		},
	}
	for _, t := range tokenLists {
		c.db.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "address"}, {Name: "chain"}},
			UpdateAll: true,
		}).Create(&t)
	}
}

func UpdateBenfenTESTToken() {
	var tokenLists = []models.TokenList{
		//{
		//	Chain:       "BenfenTEST",
		//	Address:     "BFC00000000000000000000000000000000000000000000000000000000000000c8e30a::busd::BUSD",
		//	Name:        "Benfen USD",
		//	Symbol:      "BUSD",
		//	LogoURI:     "images/BenfenTEST/BFC00000000000000000000000000000000000000000000000000000000000000c8e30a::busd::BUSD.png",
		//	Decimals:    9,
		//	WebSite:     "",
		//	Description: "",
		//},
		//{
		//	Chain:       "BenfenTEST",
		//	Address:     "BFC00000000000000000000000000000000000000000000000000000000000000c8e30a::bjpy::BJPY",
		//	Name:        "Benfen JPY",
		//	Symbol:      "BJPY",
		//	LogoURI:     "images/BenfenTEST/BFC00000000000000000000000000000000000000000000000000000000000000c8e30a::bjpy::BJPY.png",
		//	Decimals:    9,
		//	WebSite:     "",
		//	Description: "",
		//},
		{
			Chain:       "BenfenTEST",
			Address:     "BFCe88253dcc3eaced8168f5a87f8d5cb78f2663655fce2246951c7df6ea1b8cca677d6::long::LONG",
			Name:        "Benfen Long Coin",
			Symbol:      "LONG",
			LogoURI:     "images/BenfenTEST/BFCe88253dcc3eaced8168f5a87f8d5cb78f2663655fce2246951c7df6ea1b8cca677d6::long::LONG.png",
			Decimals:    9,
			WebSite:     "",
			Description: "",
		},
	}
	for _, t := range tokenLists {
		c.db.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "address"}, {Name: "chain"}},
			UpdateAll: true,
		}).Create(&t)
	}
}

func UpdateStrideToken() {
	var tokenLists = []models.TokenList{
		{
			Chain:    "Stride",
			Address:  "ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2",
			Name:     "Cosmos Hub",
			Symbol:   "ATOM",
			Logo:     "https://assets.coingecko.com/coins/images/1481/large/cosmos_hub.png?1696502525",
			Decimals: 6,
			CgId:     "cosmos",
			WebSite:  "http://cosmos.network",
		},

		{
			Chain:    "Stride",
			Address:  "ibc/7EAE5BEF3A26B64AFBD89828AFDDB1DC7024A0276D22745201632C40E6E634D0",
			Name:     "Stargaze",
			Symbol:   "STARS",
			Logo:     "https://assets.coingecko.com/coins/images/22363/large/stargaze_star_pink.png?1708079583",
			Decimals: 6,
			CgId:     "stargaze",
			WebSite:  "https://stargaze.zone/",
		},
		{
			Chain:    "Stride",
			Address:  "ibc/DA356E369C3E5CF6A9F1DCD99CE8ED55FBD595E676A5CF033CE784C060492D5A",
			Name:     "JUNO",
			Symbol:   "JUNO",
			Logo:     "https://assets.coingecko.com/coins/images/19249/large/Juno_Logo_%28Salmon%29_%282%29.png?1696518694",
			Decimals: 6,
			CgId:     "juno-network",
			WebSite:  "https://www.junonetwork.io/",
		},
		{
			Chain:    "Stride",
			Address:  "ibc/D24B4564BCD51D3D02D9987D92571EAC5915676A9BD6D9B0C1D0254CB8A5EA34",
			Name:     "Osmosis",
			Symbol:   "OSMO",
			Logo:     "https://assets.coingecko.com/coins/images/16724/large/osmo.png?1696516298",
			Decimals: 6,
			CgId:     "osmosis",
			WebSite:  "https://osmosis.zone/",
		},
		{
			Chain:    "Stride",
			Address:  "ibc/2475196A6E87EEA47C8878EDE62E99C0FAA14A81B83165E4121E776622EF59BE",
			Name:     "Cronos",
			Symbol:   "CRO",
			Logo:     "https://assets.coingecko.com/coins/images/7310/large/cro_token_logo.png?1696507599",
			Decimals: 8,
			CgId:     "crypto-com-chain",
			WebSite:  "https://cronos-pos.org/",
		},
		{
			Chain:    "Stride",
			Address:  "ibc/D54D8A7843F08148FEEFA1DDCC7C9EAF94A4FC4F7E42EE90BBB18E157A31FDF1",
			Name:     "Arable Protocol",
			Symbol:   "ACRE",
			Logo:     "https://assets.coingecko.com/coins/images/23659/large/acre_token-02.png?1696522862",
			Decimals: 18,
			CgId:     "arable-protocol",
			WebSite:  "https://arable.finance/",
		},
		{
			Chain:    "Stride",
			Address:  "ibc/DFAE6C8B4F9BC2FE094442BF3BD7F94551BE9C337EFA6A24B0DD6CF54AEED093",
			Name:     "CNTO",
			Symbol:   "CNTO",
			Logo:     "https://raw.githubusercontent.com/cosmostation/chainlist/master/chain/acrechain/asset/cnto.png",
			Decimals: 18,
			CgId:     "",
			WebSite:  "",
		},
		{
			Chain:    "Stride",
			Address:  "ibc/4B322204B4F59D770680FE4D7A565DDC3F37BFF035474B717476C66A4F83DD72",
			Name:     "Evmos",
			Symbol:   "EVMOS",
			Logo:     "https://assets.coingecko.com/coins/images/24023/large/evmos.png?1696523216",
			Decimals: 18,
			CgId:     "evmos",
			WebSite:  "http://evmos.org/",
		},
		{
			Chain:    "Stride",
			Address:  "ibc/E61BCB1126F42A2ED73B4CEA2221C9635BC2102F0417543C38071779F991942E",
			Name:     "LUNA",
			Symbol:   "LUNA",
			Logo:     "https://www.mintscan.io/assets/tokens/_rendered/uluna.png",
			Decimals: 6,
			CgId:     "",
			WebSite:  "",
		},
		{
			Chain:    "Stride",
			Address:  "ibc/A7454562FF29FE068F42F9DE4805ABEF54F599D1720B345D6518D9B5C64EA6D2",
			Name:     "Injective",
			Symbol:   "INJ",
			Logo:     "https://assets.coingecko.com/coins/images/12882/large/Secondary_Symbol.png?1696512670",
			Decimals: 18,
			CgId:     "injective-protocol",
			WebSite:  "https://injective.com",
		},
		{
			Chain:    "Stride",
			Address:  "ibc/1A2271226209D309902AFF4F21BD21237CB514DD24EA2EE0423BF74C6135D8B8",
			Name:     "UX Chain",
			Symbol:   "UMEE",
			Logo:     "https://assets.coingecko.com/coins/images/20635/large/yArLkbL.png?1696520039",
			Decimals: 6,
			CgId:     "umee",
			WebSite:  "https://umee.cc/",
		},
		{
			Chain:    "Stride",
			Address:  "ibc/EB66980014602E6BD50A1CB9FFB8FA694DC3EC10A48D2C1C649D732954F88D4A",
			Name:     "COMDEX",
			Symbol:   "CMDX",
			Logo:     "https://assets.coingecko.com/coins/images/21540/large/CMDX_Token.png?1696520899",
			Decimals: 6,
			CgId:     "comdex",
			WebSite:  "https://comdex.one/",
		},
		{
			Chain:    "Stride",
			Address:  "ibc/B86EFF6D227E3E33D7E3B5E65D0C1BB5BD79CCB56D35A9D824F0DD5D52CA43BA",
			Name:     "Sommelier",
			Symbol:   "SOMM",
			Logo:     "https://assets.coingecko.com/coins/images/23308/large/somm_new.png?1696522525",
			Decimals: 6,
			CgId:     "sommelier",
			WebSite:  "https://sommelier.finance/",
		},
		{
			Chain:    "Stride",
			Address:  "ibc/BF3B4F53F3694B66E13C23107C84B6485BD2B96296BB7EC680EA77BBA75B4801",
			Name:     "Celestia",
			Symbol:   "TIA",
			Logo:     "https://assets.coingecko.com/coins/images/31967/large/tia.jpg?1696530772",
			Decimals: 6,
			CgId:     "celestia",
			WebSite:  "https://celestia.org/",
		},
		{
			Chain:    "Stride",
			Address:  "stuatom",
			Name:     "Stride Staked Atom",
			Symbol:   "STATOM",
			Logo:     "https://assets.coingecko.com/coins/images/26316/large/stATOM.png?1696525398",
			Decimals: 6,
			CgId:     "stride-staked-atom",
			WebSite:  "https://www.stride.zone/",
		},
		{
			Chain:    "Stride",
			Address:  "stustars",
			Name:     "Stride Staked Stars",
			Symbol:   "STSTARS",
			Logo:     "https://assets.coingecko.com/coins/images/29952/large/ststars.png?1696528879",
			Decimals: 6,
			CgId:     "stride-staked-stars",
			WebSite:  "https://app.stride.zone/",
		},
		{
			Chain:    "Stride",
			Address:  "stujuno",
			Name:     "Stride Staked Juno",
			Symbol:   "STJUNO",
			Logo:     "https://assets.coingecko.com/coins/images/28700/large/stJUNO_800.png?1696527682",
			Decimals: 6,
			CgId:     "stride-staked-juno",
			WebSite:  "https://www.stride.zone/",
		},
		{
			Chain:    "Stride",
			Address:  "stuosmo",
			Name:     "Stride Staked Osmo",
			Symbol:   "STOSMO",
			Logo:     "https://assets.coingecko.com/coins/images/28698/large/stOSMO_800.png?1696527680",
			Decimals: 6,
			CgId:     "stride-staked-osmo",
			WebSite:  "https://www.stride.zone/",
		},
		{
			Chain:    "Stride",
			Address:  "stuluna",
			Name:     "Stride Staked Luna",
			Symbol:   "STLUNA",
			Logo:     "https://assets.coingecko.com/coins/images/29380/large/stLuna.jpg?1696528327",
			Decimals: 6,
			CgId:     "stride-staked-luna",
			WebSite:  "https://www.stride.zone/",
		},
		{
			Chain:    "Stride",
			Address:  "staevmos",
			Name:     "Stride Staked Evmos",
			Symbol:   "STEVMOS",
			Logo:     "https://assets.coingecko.com/coins/images/29951/large/stevmos.png?1696528878",
			Decimals: 18,
			CgId:     "stride-staked-evmos",
			WebSite:  "https://www.stride.zone/",
		},
		{
			Chain:    "Stride",
			Address:  "stuumee",
			Name:     "Stride Staked Umee",
			Symbol:   "STUMEE",
			Logo:     "https://assets.coingecko.com/coins/images/33121/large/stUMEE.png?1700733183",
			Decimals: 6,
			CgId:     "stride-staked-umee",
			WebSite:  "https://www.stride.zone/",
		},
		{
			Chain:    "Stride",
			Address:  "stinj",
			Name:     "Stride Staked Injective",
			Symbol:   "STINJ",
			Logo:     "https://assets.coingecko.com/coins/images/29953/large/stinj.png?1696528880",
			Decimals: 18,
			CgId:     "stride-staked-injective",
			WebSite:  "https://www.stride.zone/",
		},
		{
			Chain:    "Stride",
			Address:  "stucmdx",
			Name:     "Stride Staked CMDX",
			Symbol:   "STCMDX",
			Logo:     "https://raw.githubusercontent.com/cosmostation/chainlist/master/chain/stride/asset/stcmdx.png",
			Decimals: 6,
			CgId:     "",
			WebSite:  "https://www.stride.zone/",
		},
		{
			Chain:    "Stride",
			Address:  "stusomm",
			Name:     "Stride Staked SOMM",
			Symbol:   "STSOMM",
			Logo:     "https://raw.githubusercontent.com/cosmostation/chainlist/master/chain/stride/asset/stsomm.png",
			Decimals: 6,
			CgId:     "",
			WebSite:  "https://www.stride.zone/",
		},
		{
			Chain:    "Stride",
			Address:  "stadydx",
			Name:     "Stride Staked DYDX",
			Symbol:   "STDYDX",
			Logo:     "https://assets.coingecko.com/coins/images/33539/large/stDYDX.png?1707226741",
			Decimals: 18,
			CgId:     "stride-staked-dydx",
			WebSite:  "https://www.stride.zone/",
		},
		{
			Chain:    "Stride",
			Address:  "stutia",
			Name:     "Stride Staked TIA",
			Symbol:   "STTIA",
			Logo:     "https://assets.coingecko.com/coins/images/34919/large/stTIA.png?1706956373",
			Decimals: 6,
			CgId:     "stride-staked-tia",
			WebSite:  "https://www.stride.zone/",
		},
	}
	result := c.db.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(&tokenLists)
	if result.Error != nil {
		c.log.Error("create db aptos error:", result.Error)
	}
}

func UpdateNeutronToken() {
	var tokenLists = []models.TokenList{
		{
			Chain:    "Neutron",
			Address:  "ibc/C4CFF46FD6DE35CA4CF4CE031E643C8FDC9BA4B99AE598E9B0ED98FE3A2319F9",
			Name:     "Cosmos Hub",
			Symbol:   "ATOM",
			Logo:     "https://assets.coingecko.com/coins/images/1481/large/cosmos_hub.png?1696502525",
			Decimals: 6,
			CgId:     "cosmos",
			WebSite:  "http://cosmos.network",
		},
		{
			Chain:    "Neutron",
			Address:  "ibc/C0E66D1C81D8AAF0E6896E05190FDFBC222367148F86AC3EA679C28327A763CD",
			Name:     "Axelar",
			Symbol:   "AXL",
			Logo:     "https://assets.coingecko.com/coins/images/27277/large/V-65_xQ1_400x400.jpeg?1696526329",
			Decimals: 6,
			CgId:     "axelar",
			WebSite:  "https://axelar.network/",
		},
		{
			Chain:    "Neutron",
			Address:  "ibc/F082B65C88E4B6D5EF1DB243CDA1D331D002759E938A0F5CD3FFDC5D53B3E349",
			Name:     "Bridged USD Coin (Axelar)",
			Symbol:   "AXLUSDC",
			Logo:     "https://assets.coingecko.com/coins/images/26476/large/uausdc_D_3x.png?1696525548",
			Decimals: 6,
			CgId:     "axlusdc",
			WebSite:  "https://axelar.network/",
		},
		{
			Chain:    "Neutron",
			Address:  "ibc/57503D7852EF4E1899FE6D71C5E81D7C839F76580F86F21E39348FC2BC9D7CE2",
			Name:     "Bridged Tether (Axelar)",
			Symbol:   "AXLUSDT",
			Logo:     "https://assets.coingecko.com/coins/images/31002/large/uusdt_D_3x.png?1696529840",
			Decimals: 6,
			CgId:     "axelar-usdt",
			WebSite:  "https://axelar.network/",
		},
		{
			Chain:    "Neutron",
			Address:  "ibc/A585C2D15DCD3B010849B453A2CFCB5E213208A5AB665691792684C26274304D",
			Name:     "Axelar Wrapped Ether",
			Symbol:   "AXLWETH",
			Logo:     "https://assets.coingecko.com/coins/images/28171/large/weth-wei_D_3x.png?1696527174",
			Decimals: 18,
			CgId:     "axlweth",
			WebSite:  "https://axelarscan.io/assets",
		},
		{
			Chain:    "Neutron",
			Address:  "ibc/B7864B03E1B9FD4F049243E92ABD691586F682137037A9F3FCA5222815620B3C",
			Name:     "Stride Staked Atom",
			Symbol:   "STATOM",
			Logo:     "https://assets.coingecko.com/coins/images/26316/large/stATOM.png?1696525398",
			Decimals: 6,
			CgId:     "stride-staked-atom",
			WebSite:  "https://www.stride.zone/",
		},
		{
			Chain:    "Neutron",
			Address:  "ibc/2CB87BCE0937B1D1DFCEE79BE4501AAF3C265E923509AEAC410AD85D27F35130",
			Name:     "dYdX",
			Symbol:   "DYDX",
			Logo:     "https://assets.coingecko.com/coins/images/32594/large/dydx.png?1698673495",
			Decimals: 18,
			CgId:     "dydx-chain",
			WebSite:  "https://dydx.trade/",
		},
		{
			Chain:    "Neutron",
			Address:  "ibc/3649CE0C8A2C79048D8C6F31FF18FA69C9BC7EB193512E0BD03B733011290445",
			Name:     "stkATOM",
			Symbol:   "STKATOM",
			Logo:     "https://assets.coingecko.com/coins/images/28289/large/B5712581-7546-47FA-8ECA-B256CD8F8F34.png?1696527288",
			Decimals: 6,
			CgId:     "stkatom",
			WebSite:  "https://pstake.finance/atom",
		},
		{
			Chain:    "Neutron",
			Address:  "ibc/58923AAE6E879D7CB5FB0F2F05550FD4F696099AB0F5CDF0A05CC0309DD8BC78",
			Name:     "Cerberus",
			Symbol:   "CRBRUS",
			Logo:     "https://assets.coingecko.com/coins/images/24509/large/AUdZyrmO_400x400.png?1696523688",
			Decimals: 6,
			CgId:     "cerberus-2",
			WebSite:  "https://cerberus.zone/",
		},
		{
			Chain:    "Neutron",
			Address:  "ibc/376222D6D9DAE23092E29740E56B758580935A6D77C24C2ABD57A6A78A1F3955",
			Name:     "Osmosis",
			Symbol:   "OSMO",
			Logo:     "https://assets.coingecko.com/coins/images/16724/large/osmo.png?1696516298",
			Decimals: 6,
			CgId:     "osmosis",
			WebSite:  "https://osmosis.zone/",
		},
		{
			Chain:    "Neutron",
			Address:  "ibc/773B4D0A3CD667B2275D5A4A7A2F0909C0BA0F4059C0B9181E680DDF4965DCC7",
			Name:     "Celestia",
			Symbol:   "TIA",
			Logo:     "https://assets.coingecko.com/coins/images/31967/large/tia.jpg?1696530772",
			Decimals: 6,
			CgId:     "celestia",
			WebSite:  "https://celestia.org/",
		},
		{
			Chain:    "Neutron",
			Address:  "neutron1uqvse8fdrd9tam47f2jhy9m6al6xxtqpc83f9pdnz5gdle4swc0spfnctv",
			Name:     "Astro BOY",
			Symbol:   "BOY",
			Logo:     "https://raw.githubusercontent.com/cosmostation/chainlist/master/chain/neutron/asset/boy.png",
			Decimals: 6,
			CgId:     "",
			WebSite:  "",
		},
	}
	result := c.db.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(&tokenLists)
	if result.Error != nil {
		c.log.Error("create db aptos error:", result.Error)
	}
}

func UpdateKujiraToken() {
	var tokenLists = []models.TokenList{
		{
			Chain:    "Kujira",
			Address:  "ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2",
			Name:     "Cosmos Hub",
			Symbol:   "ATOM",
			Logo:     "https://assets.coingecko.com/coins/images/1481/large/cosmos_hub.png?1696502525",
			Decimals: 6,
			CgId:     "cosmos",
			WebSite:  "http://cosmos.network",
		},
		{
			Chain:    "Kujira",
			Address:  "ibc/BBC45F1B65B6D3C11C3C56A9428D38C3A8D03944473791C52DFB7CD3F8342CBC",
			Name:     "Cronos",
			Symbol:   "CRO",
			Logo:     "https://assets.coingecko.com/coins/images/7310/large/cro_token_logo.png?1696507599",
			Decimals: 8,
			CgId:     "crypto-com-chain",
			WebSite:  "https://cronos-pos.org/",
		},
		{
			Chain:    "Kujira",
			Address:  "ibc/47BD209179859CDE4A2806763D7189B6E6FE13A17880FE2B42DE1E6C1E329E23",
			Name:     "Osmosis",
			Symbol:   "OSMO",
			Logo:     "https://assets.coingecko.com/coins/images/16724/large/osmo.png?1696516298",
			Decimals: 6,
			CgId:     "osmosis",
			WebSite:  "https://osmosis.zone/",
		},
		{
			Chain:    "Kujira",
			Address:  "ibc/EFF323CC632EC4F747C61BCE238A758EFDB7699C3226565F7C20DA06509D59A5",
			Name:     "JUNO",
			Symbol:   "JUNO",
			Logo:     "https://assets.coingecko.com/coins/images/19249/large/Juno_Logo_%28Salmon%29_%282%29.png?1696518694",
			Decimals: 6,
			CgId:     "juno-network",
			WebSite:  "https://www.junonetwork.io/",
		},
		{
			Chain:    "Kujira",
			Address:  "ibc/4F393C3FCA4190C0A6756CE7F6D897D5D1BE57D6CCB80D0BC87393566A7B6602",
			Name:     "Stargaze",
			Symbol:   "STARS",
			Logo:     "https://assets.coingecko.com/coins/images/22363/large/stargaze_star_pink.png?1708079583",
			Decimals: 6,
			CgId:     "stargaze",
			WebSite:  "https://stargaze.zone/",
		},
		{
			Chain:    "Kujira",
			Address:  "ibc/2F1447818CF99498AE62D9FB4D5E0C9FD48C68FC118C34D2ECFFFED0AD082196",
			Name:     "Graviton",
			Symbol:   "GRAV",
			Logo:     "https://assets.coingecko.com/coins/images/24818/large/Graviton_-_Blue_200x200i.png?1696523976",
			Decimals: 6,
			CgId:     "graviton",
			WebSite:  "https://gravitybridge.net",
		},
		{
			Chain:    "Kujira",
			Address:  "ibc/3607EB5B5E64DD1C0E12E07F077FF470D5BC4706AFCBC98FE1BA960E5AE4CE07",
			Name:     "COMDEX",
			Symbol:   "CMDX",
			Logo:     "https://assets.coingecko.com/coins/images/21540/large/CMDX_Token.png?1696520899",
			Decimals: 6,
			CgId:     "comdex",
			WebSite:  "https://comdex.one/",
		},
		{
			Chain:    "Kujira",
			Address:  "ibc/A358D7F19237777AF6D8AD0E0F53268F8B18AE8A53ED318095C14D6D7F3B2DB5",
			Name:     "Secret",
			Symbol:   "SCRT",
			Logo:     "https://raw.githubusercontent.com/cosmostation/cosmostation_token_resource/master/assets/images/common/scrt.png",
			Decimals: 6,
			CgId:     "secret",
			WebSite:  "https://scrt.network/",
		},
		{
			Chain:    "Kujira",
			Address:  "ibc/F3AA7EF362EC5E791FE78A0F4CCC69FEE1F9A7485EB1A8CAB3F6601C00522F10",
			Name:     "Evmos",
			Symbol:   "EVMOS",
			Logo:     "https://assets.coingecko.com/coins/images/24023/large/evmos.png?1696523216",
			Decimals: 18,
			CgId:     "evmos",
			WebSite:  "http://evmos.org/",
		},
		{
			Chain:    "Kujira",
			Address:  "ibc/C01154C2547F4CB10A985EA78E7CD4BA891C1504360703A37E1D7043F06B5E1F",
			Name:     "Axelar",
			Symbol:   "AXL",
			Logo:     "https://assets.coingecko.com/coins/images/27277/large/V-65_xQ1_400x400.jpeg?1696526329",
			Decimals: 6,
			CgId:     "axelar",
			WebSite:  "https://axelar.network/",
		},
		{
			Chain:    "Kujira",
			Address:  "ibc/F2331645B9683116188EF36FC04A809C28BD36B54555E8705A37146D0182F045",
			Name:     "Tether USD",
			Symbol:   "AXLUSDT",
			Logo:     `{"thumb":"https://assets.coingecko.com/coins/images/325/thumb/Tether.png?1668148663","small":"https://assets.coingecko.com/coins/images/325/small/Tether.png?1668148663","large":"https://assets.coingecko.com/coins/images/325/large/Tether.png?1668148663"}`,
			Decimals: 6,
			CgId:     "tether",
			WebSite:  "https://tether.to/",
		},
		{
			Chain:    "Kujira",
			Address:  "ibc/295548A78785A1007F232DE286149A6FF512F180AF5657780FC89C009E2C348F",
			Name:     "USD Coin",
			Symbol:   "axlUSDC",
			Logo:     `{"thumb":"https://assets.coingecko.com/coins/images/6319/thumb/USD_Coin_icon.png?1547042389","small":"https://assets.coingecko.com/coins/images/6319/small/USD_Coin_icon.png?1547042389","large":"https://assets.coingecko.com/coins/images/6319/large/USD_Coin_icon.png?1547042389"}`,
			Decimals: 6,
			CgId:     "usd-coin",
			WebSite:  "https://www.circle.com/en/usdc",
		},
		{
			Chain:    "Kujira",
			Address:  "ibc/004EBF085BBED1029326D56BE8A2E67C08CECE670A94AC1947DF413EF5130EB2",
			Name:     "Wrapped AVAX",
			Symbol:   "AXLWAVAX",
			Logo:     "https://assets.coingecko.com/coins/images/15075/large/wrapped-avax.png?1696514734",
			Decimals: 18,
			CgId:     "wrapped-avax",
			WebSite:  "https://www.avalabs.org/",
		},
		{
			Chain:    "Kujira",
			Address:  "ibc/1B38805B1C75352B28169284F96DF56BDEBD9E8FAC005BDCC8CF0378C82AA8E7",
			Name:     "WETH",
			Symbol:   "AXLWETH",
			Logo:     "https://assets.coingecko.com/coins/images/2518/large/weth.png?1696503332",
			Decimals: 18,
			CgId:     "weth",
			WebSite:  "https://weth.io/",
		},
		{
			Chain:    "Kujira",
			Address:  "ibc/3914BDEF46F429A26917E4D8D434620EC4817DC6B6E68FB327E190902F1E9242",
			Name:     "Dai",
			Symbol:   "AXLDAI",
			Logo:     "https://assets.coingecko.com/coins/images/9956/large/Badge_Dai.png?1696509996",
			Decimals: 18,
			CgId:     "dai",
			WebSite:  "https://makerdao.com/",
		},
		{
			Chain:    "Kujira",
			Address:  "ibc/4925CCFAD4FFBC63D7C52BB9AE2DE8CF7D0809124FBA1F44F4F2B7B4267D5E5A",
			Name:     "CUB",
			Symbol:   "CUB",
			Logo:     "",
			Decimals: 6,
			CgId:     "",
			WebSite:  "",
		},
		{
			Chain:    "Kujira",
			Address:  "ibc/3766EE7E4A42BA83C4AB0DD26B144B26B7758BDAB4017E0924C95FBFC69A2589",
			Name:     "BLUE",
			Symbol:   "BLUE",
			Logo:     "",
			Decimals: 6,
			CgId:     "",
			WebSite:  "",
		},
		{
			Chain:    "Kujira",
			Address:  "ibc/DBF6ABDB5F3D4267C039967515594453F4A31007FD838A566F563A01D2C2FB80",
			Name:     "Mars Protocol",
			Symbol:   "MARS",
			Logo:     "https://assets.coingecko.com/coins/images/28896/large/mars.png?1696527872",
			Decimals: 6,
			CgId:     "mars-protocol-a7fcbcfb-fd61-4017-92f0-7ee9f9cc6da3",
			WebSite:  "https://www.marsprotocol.io",
		},
		{
			Chain:    "Kujira",
			Address:  "ibc/FE98AAD68F02F03565E9FA39A5E627946699B2B07115889ED812D8BA639576A9",
			Name:     "USD Coin",
			Symbol:   "USDC",
			Logo:     `{"thumb":"https://assets.coingecko.com/coins/images/6319/thumb/USD_Coin_icon.png?1547042389","small":"https://assets.coingecko.com/coins/images/6319/small/USD_Coin_icon.png?1547042389","large":"https://assets.coingecko.com/coins/images/6319/large/USD_Coin_icon.png?1547042389"}`,
			Decimals: 6,
			CgId:     "usd-coin",
			WebSite:  "https://www.circle.com/en/usdc",
		},
		{
			Chain:    "Kujira",
			Address:  "ibc/D5F668AA43F9DE1F5E8A443EC5F9DD2D71AB3F3DC42CCD374F8751615A13B70B",
			Name:     "Cerberus",
			Symbol:   "CRBRUS",
			Logo:     "https://assets.coingecko.com/coins/images/24509/large/AUdZyrmO_400x400.png?1696523688",
			Decimals: 6,
			CgId:     "cerberus-2",
			WebSite:  "https://cerberus.zone/",
		},
		{
			Chain:    "Kujira",
			Address:  "ibc/086C1B2E9E611F23D3C9159C858D3F4C6A911D348DFB916A11AE306A865A77A8",
			Name:     "stkATOM",
			Symbol:   "STKATOM",
			Logo:     "https://assets.coingecko.com/coins/images/28289/large/B5712581-7546-47FA-8ECA-B256CD8F8F34.png?1696527288",
			Decimals: 6,
			CgId:     "stkatom",
			WebSite:  "https://pstake.finance/atom",
		},
	}
	result := c.db.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(&tokenLists)
	if result.Error != nil {
		c.log.Error("create db aptos error:", result.Error)
	}
}
