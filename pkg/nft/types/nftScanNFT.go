package types

type NFTScanAsset struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		ContractAddress      string              `json:"contract_address"`
		ContractName         string              `json:"contract_name"`
		ContractTokenID      string              `json:"contract_token_id"`
		TokenID              string              `json:"token_id"`
		ErcType              string              `json:"erc_type"`
		Amount               string              `json:"amount"`
		Minter               string              `json:"minter"`
		Owner                string              `json:"owner"`
		OwnTimestamp         int64               `json:"own_timestamp"`
		MintTimestamp        int64               `json:"mint_timestamp"`
		MintTransactionHash  string              `json:"mint_transaction_hash"`
		MintPrice            float64             `json:"mint_price"`
		TokenURI             string              `json:"token_uri"`
		MetadataJSON         string              `json:"metadata_json"`
		Name                 string              `json:"name"`
		ContentType          string              `json:"content_type"`
		ContentURI           string              `json:"content_uri"`
		ImageURI             string              `json:"image_uri"`
		ExternalLink         string              `json:"external_link"`
		LatestTradePrice     interface{}         `json:"latest_trade_price"`
		LatestTradeSymbol    interface{}         `json:"latest_trade_symbol"`
		LatestTradeTimestamp interface{}         `json:"latest_trade_timestamp"`
		NftscanID            string              `json:"nftscan_id"`
		NftscanURI           interface{}         `json:"nftscan_uri"`
		Attributes           []NFTScanAttributes `json:"attributes"`
	} `json:"data"`
}

type NFTScanNFTAssetMateData struct {
	Name         string `json:"name"`
	Image        string `json:"image"`
	Description  string `json:"description"`
	ExternalUrl  string `json:"external_url"`
	attributes   string `json:"attributes"`
	AnimationUrl string `json:"animation_url"`
}

type NFTScanAttributes struct {
	AttributeName  string      `json:"attribute_name"`
	AttributeValue interface{} `json:"attribute_value"`
	Percentage     interface{} `json:"percentage"`
}

type NFTScanCollection struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		ContractAddress string      `json:"contract_address"`
		Name            string      `json:"name"`
		Symbol          string      `json:"symbol"`
		Description     string      `json:"description"`
		Website         interface{} `json:"website"`
		Email           interface{} `json:"email"`
		Twitter         string      `json:"twitter"`
		Discord         string      `json:"discord"`
		Telegram        interface{} `json:"telegram"`
		Github          interface{} `json:"github"`
		Instagram       interface{} `json:"instagram"`
		Medium          interface{} `json:"medium"`
		LogoURL         string      `json:"logo_url"`
		BannerURL       string      `json:"banner_url"`
		FeaturedURL     interface{} `json:"featured_url"`
		LargeImageURL   interface{} `json:"large_image_url"`
		Attributes      []struct {
			AttributesName   string `json:"attributes_name"`
			AttributesValues []struct {
				AttributesValue string `json:"attributes_value"`
				Total           int    `json:"total"`
			} `json:"attributes_values"`
			Total int `json:"total"`
		} `json:"attributes"`
		ErcType                 string        `json:"erc_type"`
		DeployBlockNumber       int           `json:"deploy_block_number"`
		Owner                   string        `json:"owner"`
		Verified                bool          `json:"verified"`
		OpenseaVerified         bool          `json:"opensea_verified"`
		Royalty                 interface{}   `json:"royalty"`
		ItemsTotal              int           `json:"items_total"`
		AmountsTotal            int           `json:"amounts_total"`
		OwnersTotal             int           `json:"owners_total"`
		OpenseaFloorPrice       interface{}   `json:"opensea_floor_price"`
		FloorPrice              interface{}   `json:"floor_price"`
		CollectionsWithSameName []interface{} `json:"collections_with_same_name"`
		PriceSymbol             string        `json:"price_symbol"`
	} `json:"data"`
}
