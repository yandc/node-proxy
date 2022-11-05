package types

type Token struct {
	Tokens []TokenInfo `json:"tokens"`
}

type TokenInfo struct {
	ChainId  int    `json:"chainId"`
	Name     string `json:"name"`
	Symbol   string `json:"symbol"`
	Address  string `json:"address"`
	Decimals int    `json:"decimals"`
	LogoURI  string `json:"logoURI"`
}

type TokenInfoVersion struct {
	Chain   string
	URL     string
	Version int64
}

type KlaytnTokenInfo struct {
	Id       int    `json:"id"`
	Address  string `json:"address"`
	Symbol   string `json:"symbol"`
	Decimals int    `json:"decimals"`
	NameEn   string `json:"name_en"`
	Icon     string `json:"icon"`
}

type KlaytnToken struct {
	Tokens map[string]KlaytnTokenInfo
}

type CMCListItem struct {
	ID int `json:"id"`
}

type CMCList struct {
	Data []CMCListItem `json:"data"`
}

type CMCCoinsID struct {
	Data map[string]CoinsIDInfo `json:"data"`
}

type CoinsIDInfo struct {
	ID          int    `json:"id"`
	Symbol      string `json:"symbol"`
	Name        string `json:"name"`
	Category    string `json:"category"`
	Description string `json:"description"`
	Logo        string `json:"logo"`
	Urls        struct {
		Website []string `json:"website"`
		Twitter []string `json:"twitter"`
	} `json:"urls"`
	ContractAddress []struct {
		ContractAddress string `json:"contract_address"`
		Platform        struct {
			Name string `json:"name"`
			Coin struct {
				ID     string `json:"id"`
				Name   string `json:"name"`
				Symbol string `json:"symbol"`
				Slug   string `json:"slug"`
			} `json:"coin"`
		} `json:"platform"`
	} `json:"contract_address"`
}

type Platform struct {
	ContractAddress string `json:"contract_address"`
	Platform        struct {
		Name string `json:"name"`
		Coin struct {
			ID     string `json:"id"`
			Name   string `json:"name"`
			Symbol string `json:"symbol"`
			Slug   string `json:"slug"`
		} `json:"coin"`
	} `json:"platform"`
}

type CMCPriceResp struct {
	Data struct {
		Quote map[string]CMCPriceQuote `json:"quote"`
	} `json:"data"`
}

type CMCPriceQuote struct {
	Price       float32 `json:"price"`
	LastUpdated string  `json:"last_updated"`
}

type CoinsListItem struct {
	ID string `json:"id"`
}

type CGCoinsID struct {
	ID            string                 `json:"id"`
	Symbol        string                 `json:"symbol"`
	Name          string                 `json:"name"`
	Platforms     map[string]string      `json:"platforms"`
	Description   DescriptionItem        `json:"description"`
	Links         map[string]interface{} `json:"links"`
	Image         ImageItem              `json:"image"`
	CoinGeckoRank uint16                 `json:"coingecko_rank"`
}
type ImageItem struct {
	Thumb string `json:"thumb"`
	Small string `json:"small"`
	Large string `json:"large"`
}

type MyPutRet struct {
	Key    string
	Hash   string
	Fsize  int
	Bucket string
	Name   string
}

type QiNiuConf struct {
	AccessKey string
	SecretKey string
	Bucket    []string
	KeyPrefix string
}

type TronTokenInfo struct {
	Data []struct {
		TokenInfo struct {
			TokenID      string `json:"tokenId"`
			TokenAbbr    string `json:"tokenAbbr"`
			TokenName    string `json:"tokenName"`
			TokenDecimal int    `json:"tokenDecimal"`
			TokenCanShow int    `json:"tokenCanShow"`
			TokenType    string `json:"tokenType"`
		} `json:"tokenInfo"`
	} `json:"data"`
}

type DescriptionItem map[string]string
type LinksItem map[string]interface{}

type SolanaTokenInfo struct {
	Value struct {
		Amount         string  `json:"amount"`
		Decimals       int     `json:"decimals"`
		UIAmount       float64 `json:"uiAmount"`
		UIAmountString string  `json:"uiAmountString"`
	} `json:"value"`
}

type CGMarket struct {
	ID            string `json:"id"`
	MarketCapRank int    `json:"market_cap_rank"`
}
