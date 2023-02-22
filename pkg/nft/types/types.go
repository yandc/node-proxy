package types

import "time"

type AssetList struct {
	Next     string      `json:"next"`
	Previous interface{} `json:"previous"`
	Assets   []Asset     `json:"assets"`
}

type Asset struct {
	ID                   int         `json:"id"`
	NumSales             int         `json:"num_sales"`
	BackgroundColor      interface{} `json:"background_color"`
	ImageURL             string      `json:"image_url"`
	ImagePreviewURL      string      `json:"image_preview_url"`
	ImageThumbnailURL    string      `json:"image_thumbnail_url"`
	ImageOriginalURL     string      `json:"image_original_url"`
	AnimationURL         string      `json:"animation_url"`
	AnimationOriginalURL interface{} `json:"animation_original_url"`
	Name                 string      `json:"name"`
	Description          string      `json:"description"`
	ExternalLink         interface{} `json:"external_link"`
	AssetContract        struct {
		Address                     string      `json:"address"`
		AssetContractType           string      `json:"asset_contract_type"`
		CreatedDate                 string      `json:"created_date"`
		Name                        string      `json:"name"`
		NftVersion                  interface{} `json:"nft_version"`
		OpenseaVersion              string      `json:"opensea_version"`
		Owner                       int         `json:"owner"`
		SchemaName                  string      `json:"schema_name"`
		Symbol                      string      `json:"symbol"`
		TotalSupply                 interface{} `json:"total_supply"`
		Description                 string      `json:"description"`
		ExternalLink                interface{} `json:"external_link"`
		ImageURL                    string      `json:"image_url"`
		DefaultToFiat               bool        `json:"default_to_fiat"`
		DevBuyerFeeBasisPoints      int         `json:"dev_buyer_fee_basis_points"`
		DevSellerFeeBasisPoints     int         `json:"dev_seller_fee_basis_points"`
		OnlyProxiedTransfers        bool        `json:"only_proxied_transfers"`
		OpenseaBuyerFeeBasisPoints  int         `json:"opensea_buyer_fee_basis_points"`
		OpenseaSellerFeeBasisPoints int         `json:"opensea_seller_fee_basis_points"`
		BuyerFeeBasisPoints         int         `json:"buyer_fee_basis_points"`
		SellerFeeBasisPoints        int         `json:"seller_fee_basis_points"`
		PayoutAddress               interface{} `json:"payout_address"`
	} `json:"asset_contract"`
	Permalink  string `json:"permalink"`
	Collection struct {
		BannerImageURL          string      `json:"banner_image_url"`
		ChatURL                 interface{} `json:"chat_url"`
		CreatedDate             time.Time   `json:"created_date"`
		DefaultToFiat           bool        `json:"default_to_fiat"`
		Description             string      `json:"description"`
		DevBuyerFeeBasisPoints  string      `json:"dev_buyer_fee_basis_points"`
		DevSellerFeeBasisPoints string      `json:"dev_seller_fee_basis_points"`
		DiscordURL              interface{} `json:"discord_url"`
		DisplayData             struct {
			CardDisplayStyle string `json:"card_display_style"`
		} `json:"display_data"`
		ExternalURL                interface{} `json:"external_url"`
		Featured                   bool        `json:"featured"`
		FeaturedImageURL           string      `json:"featured_image_url"`
		Hidden                     bool        `json:"hidden"`
		SafelistRequestStatus      string      `json:"safelist_request_status"`
		ImageURL                   string      `json:"image_url"`
		IsSubjectToWhitelist       bool        `json:"is_subject_to_whitelist"`
		LargeImageURL              string      `json:"large_image_url"`
		MediumUsername             interface{} `json:"medium_username"`
		Name                       string      `json:"name"`
		OnlyProxiedTransfers       bool        `json:"only_proxied_transfers"`
		OpenseaBuyerFeeBasisPoints string      `json:"opensea_buyer_fee_basis_points"`
		PayoutAddress              interface{} `json:"payout_address"`
		RequireEmail               bool        `json:"require_email"`
		ShortDescription           interface{} `json:"short_description"`
		Slug                       string      `json:"slug"`
		TelegramURL                interface{} `json:"telegram_url"`
		TwitterUsername            interface{} `json:"twitter_username"`
		InstagramUsername          interface{} `json:"instagram_username"`
		WikiURL                    interface{} `json:"wiki_url"`
		IsNsfw                     bool        `json:"is_nsfw"`
		Fees                       struct {
			SellerFees struct {
			} `json:"seller_fees"`
			OpenseaFees struct {
				ZeroX0000A26B00C1F0Df003000390027140000Faa719 int `json:"0x0000a26b00c1f0df003000390027140000faa719"`
			} `json:"opensea_fees"`
		} `json:"fees"`
		IsRarityEnabled bool `json:"is_rarity_enabled"`
	} `json:"collection"`
	Decimals      interface{} `json:"decimals"`
	TokenMetadata interface{} `json:"token_metadata"`
	IsNsfw        bool        `json:"is_nsfw"`
	Owner         struct {
		User struct {
			Username string `json:"username"`
		} `json:"user"`
		ProfileImgURL string `json:"profile_img_url"`
		Address       string `json:"address"`
		Config        string `json:"config"`
	} `json:"owner"`
	SeaportSellOrders interface{} `json:"seaport_sell_orders"`
	Creator           struct {
		User struct {
			Username string `json:"username"`
		} `json:"user"`
		ProfileImgURL string `json:"profile_img_url"`
		Address       string `json:"address"`
		Config        string `json:"config"`
	} `json:"creator"`
	Traits                  []TraitsInfo `json:"traits"`
	LastSale                interface{}  `json:"last_sale"`
	TopBid                  interface{}  `json:"top_bid"`
	ListingDate             interface{}  `json:"listing_date"`
	IsPresale               bool         `json:"is_presale"`
	SupportsWyvern          bool         `json:"supports_wyvern"`
	RarityData              interface{}  `json:"rarity_data"`
	TransferFee             interface{}  `json:"transfer_fee"`
	TransferFeePaymentToken interface{}  `json:"transfer_fee_payment_token"`
	TokenID                 string       `json:"token_id"`
}

type OpenSeaCollection struct {
	Collection struct {
		BannerImageURL          string      `json:"banner_image_url"`
		ChatURL                 string      `json:"chat_url"`
		CreatedDate             time.Time   `json:"created_date"`
		DefaultToFiat           bool        `json:"default_to_fiat"`
		Description             string      `json:"description"`
		DevBuyerFeeBasisPoints  string      `json:"dev_buyer_fee_basis_points"`
		DevSellerFeeBasisPoints string      `json:"dev_seller_fee_basis_points"`
		DiscordURL              interface{} `json:"discord_url"`
		DisplayData             struct {
			CardDisplayStyle string `json:"card_display_style"`
		} `json:"display_data"`
		ExternalURL                 interface{} `json:"external_url"`
		Featured                    bool        `json:"featured"`
		FeaturedImageURL            interface{} `json:"featured_image_url"`
		Hidden                      bool        `json:"hidden"`
		SafelistRequestStatus       string      `json:"safelist_request_status"`
		ImageURL                    string      `json:"image_url"`
		IsSubjectToWhitelist        bool        `json:"is_subject_to_whitelist"`
		LargeImageURL               interface{} `json:"large_image_url"`
		MediumUsername              interface{} `json:"medium_username"`
		Name                        string      `json:"name"`
		OnlyProxiedTransfers        bool        `json:"only_proxied_transfers"`
		OpenseaBuyerFeeBasisPoints  string      `json:"opensea_buyer_fee_basis_points"`
		OpenseaSellerFeeBasisPoints string      `json:"opensea_seller_fee_basis_points"`
		PayoutAddress               interface{} `json:"payout_address"`
		RequireEmail                bool        `json:"require_email"`
		ShortDescription            interface{} `json:"short_description"`
		Slug                        string      `json:"slug"`
		TelegramURL                 interface{} `json:"telegram_url"`
		TwitterUsername             interface{} `json:"twitter_username"`
		InstagramUsername           interface{} `json:"instagram_username"`
		WikiURL                     interface{} `json:"wiki_url"`
		IsNsfw                      bool        `json:"is_nsfw"`
		Fees                        struct {
			SellerFees struct {
			} `json:"seller_fees"`
			OpenseaFees struct {
				ZeroX0000A26B00C1F0Df003000390027140000Faa719 int `json:"0x0000a26b00c1f0df003000390027140000faa719"`
			} `json:"opensea_fees"`
		} `json:"fees"`
		IsRarityEnabled bool `json:"is_rarity_enabled"`
	} `json:"collection"`
	Address                     string      `json:"address"`
	AssetContractType           string      `json:"asset_contract_type"`
	CreatedDate                 string      `json:"created_date"`
	Name                        string      `json:"name"`
	NftVersion                  interface{} `json:"nft_version"`
	OpenseaVersion              string      `json:"opensea_version"`
	Owner                       int         `json:"owner"`
	SchemaName                  string      `json:"schema_name"`
	Symbol                      string      `json:"symbol"`
	TotalSupply                 interface{} `json:"total_supply"`
	Description                 string      `json:"description"`
	ExternalLink                interface{} `json:"external_link"`
	ImageURL                    string      `json:"image_url"`
	DefaultToFiat               bool        `json:"default_to_fiat"`
	DevBuyerFeeBasisPoints      int         `json:"dev_buyer_fee_basis_points"`
	DevSellerFeeBasisPoints     int         `json:"dev_seller_fee_basis_points"`
	OnlyProxiedTransfers        bool        `json:"only_proxied_transfers"`
	OpenseaBuyerFeeBasisPoints  int         `json:"opensea_buyer_fee_basis_points"`
	OpenseaSellerFeeBasisPoints int         `json:"opensea_seller_fee_basis_points"`
	BuyerFeeBasisPoints         int         `json:"buyer_fee_basis_points"`
	SellerFeeBasisPoints        int         `json:"seller_fee_basis_points"`
	PayoutAddress               interface{} `json:"payout_address"`
}

type CollectionList struct {
	Collections []struct {
		PrimaryAssetContracts []interface{} `json:"primary_asset_contracts"`
		Traits                struct {
		} `json:"traits"`
		Stats struct {
			OneHourVolume         float64 `json:"one_hour_volume"`
			OneHourChange         float64 `json:"one_hour_change"`
			OneHourSales          float64 `json:"one_hour_sales"`
			OneHourSalesChange    float64 `json:"one_hour_sales_change"`
			OneHourAveragePrice   float64 `json:"one_hour_average_price"`
			OneHourDifference     float64 `json:"one_hour_difference"`
			SixHourVolume         float64 `json:"six_hour_volume"`
			SixHourChange         float64 `json:"six_hour_change"`
			SixHourSales          float64 `json:"six_hour_sales"`
			SixHourSalesChange    float64 `json:"six_hour_sales_change"`
			SixHourAveragePrice   float64 `json:"six_hour_average_price"`
			SixHourDifference     float64 `json:"six_hour_difference"`
			OneDayVolume          float64 `json:"one_day_volume"`
			OneDayChange          float64 `json:"one_day_change"`
			OneDaySales           float64 `json:"one_day_sales"`
			OneDaySalesChange     float64 `json:"one_day_sales_change"`
			OneDayAveragePrice    float64 `json:"one_day_average_price"`
			OneDayDifference      float64 `json:"one_day_difference"`
			SevenDayVolume        float64 `json:"seven_day_volume"`
			SevenDayChange        float64 `json:"seven_day_change"`
			SevenDaySales         float64 `json:"seven_day_sales"`
			SevenDayAveragePrice  float64 `json:"seven_day_average_price"`
			SevenDayDifference    float64 `json:"seven_day_difference"`
			ThirtyDayVolume       float64 `json:"thirty_day_volume"`
			ThirtyDayChange       float64 `json:"thirty_day_change"`
			ThirtyDaySales        float64 `json:"thirty_day_sales"`
			ThirtyDayAveragePrice float64 `json:"thirty_day_average_price"`
			ThirtyDayDifference   float64 `json:"thirty_day_difference"`
			TotalVolume           float64 `json:"total_volume"`
			TotalSales            float64 `json:"total_sales"`
			TotalSupply           float64 `json:"total_supply"`
			Count                 float64 `json:"count"`
			NumOwners             int     `json:"num_owners"`
			AveragePrice          float64 `json:"average_price"`
			NumReports            int     `json:"num_reports"`
			MarketCap             float64 `json:"market_cap"`
			FloorPrice            int     `json:"floor_price"`
		} `json:"stats"`
		BannerImageURL          interface{} `json:"banner_image_url"`
		ChatURL                 interface{} `json:"chat_url"`
		CreatedDate             time.Time   `json:"created_date"`
		DefaultToFiat           bool        `json:"default_to_fiat"`
		Description             string      `json:"description"`
		DevBuyerFeeBasisPoints  string      `json:"dev_buyer_fee_basis_points"`
		DevSellerFeeBasisPoints string      `json:"dev_seller_fee_basis_points"`
		DiscordURL              interface{} `json:"discord_url"`
		DisplayData             struct {
			CardDisplayStyle string `json:"card_display_style"`
		} `json:"display_data"`
		ExternalURL                 interface{} `json:"external_url"`
		Featured                    bool        `json:"featured"`
		FeaturedImageURL            interface{} `json:"featured_image_url"`
		Hidden                      bool        `json:"hidden"`
		SafelistRequestStatus       string      `json:"safelist_request_status"`
		ImageURL                    string      `json:"image_url"`
		IsSubjectToWhitelist        bool        `json:"is_subject_to_whitelist"`
		LargeImageURL               interface{} `json:"large_image_url"`
		MediumUsername              interface{} `json:"medium_username"`
		Name                        string      `json:"name"`
		OnlyProxiedTransfers        bool        `json:"only_proxied_transfers"`
		OpenseaBuyerFeeBasisPoints  string      `json:"opensea_buyer_fee_basis_points"`
		OpenseaSellerFeeBasisPoints string      `json:"opensea_seller_fee_basis_points"`
		PayoutAddress               interface{} `json:"payout_address"`
		RequireEmail                bool        `json:"require_email"`
		ShortDescription            interface{} `json:"short_description"`
		Slug                        string      `json:"slug"`
		TelegramURL                 interface{} `json:"telegram_url"`
		TwitterUsername             interface{} `json:"twitter_username"`
		InstagramUsername           interface{} `json:"instagram_username"`
		WikiURL                     interface{} `json:"wiki_url"`
		IsNsfw                      bool        `json:"is_nsfw"`
		Fees                        struct {
			SellerFees struct {
			} `json:"seller_fees"`
			OpenseaFees struct {
				ZeroX0000A26B00C1F0Df003000390027140000Faa719 int `json:"0x0000a26b00c1f0df003000390027140000faa719"`
			} `json:"opensea_fees"`
		} `json:"fees"`
		IsRarityEnabled bool `json:"is_rarity_enabled"`
	} `json:"collections"`
}

type NFTGONFTInfo struct {
	Blockchain            string      `json:"blockchain"`
	CollectionName        string      `json:"collection_name"`
	CollectionSlug        string      `json:"collection_slug"`
	CollectionOpenseaSlug string      `json:"collection_opensea_slug"`
	ContractAddress       string      `json:"contract_address"`
	TokenID               string      `json:"token_id"`
	Name                  string      `json:"name"`
	Description           string      `json:"description"`
	Image                 string      `json:"image"`
	AnimationURL          interface{} `json:"animation_url"`
	OwnerAddresses        []string    `json:"owner_addresses"`
	Traits                []struct {
		Type       string  `json:"type"`
		Value      string  `json:"value"`
		Percentage float64 `json:"percentage"`
	} `json:"traits"`
	Rarity struct {
		Score float64 `json:"score"`
		Rank  int     `json:"rank"`
		Total int     `json:"total"`
	} `json:"rarity"`
}

type TraitsInfo struct {
	TraitType  string      `json:"trait_type"`
	Value      interface{} `json:"value"`
	TraitCount interface{} `json:"trait_count"`
}

type NFTScanInfo struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		ContractAddress      string      `json:"contract_address"`
		ContractName         string      `json:"contract_name"`
		ContractTokenID      string      `json:"contract_token_id"`
		TokenID              string      `json:"token_id"`
		ErcType              string      `json:"erc_type"`
		Amount               string      `json:"amount"`
		Minter               string      `json:"minter"`
		Owner                string      `json:"owner"`
		MintTimestamp        int64       `json:"mint_timestamp"`
		MintTransactionHash  string      `json:"mint_transaction_hash"`
		TokenURI             string      `json:"token_uri"`
		MetadataJSON         string      `json:"metadata_json"`
		Name                 string      `json:"name"`
		ContentType          string      `json:"content_type"`
		ContentURI           string      `json:"content_uri"`
		ImageURI             string      `json:"image_uri"`
		ExternalLink         string      `json:"external_link"`
		LatestTradePrice     float64     `json:"latest_trade_price"`
		LatestTradeSymbol    string      `json:"latest_trade_symbol"`
		LatestTradeTimestamp int64       `json:"latest_trade_timestamp"`
		NftscanID            string      `json:"nftscan_id"`
		NftscanURI           interface{} `json:"nftscan_uri"`
		Attributes           interface{} `json:"attributes"`
	} `json:"data"`
}

type NFTScanMetaData struct {
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Attributes  []TraitsInfo `json:"attributes"`
}

type NFTPortAssetInfo struct {
	Response string `json:"response"`
	Nft      struct {
		Chain           string `json:"chain"`
		ContractAddress string `json:"contract_address"`
		TokenID         string `json:"token_id"`
		MetadataURL     string `json:"metadata_url"`
		Metadata        struct {
			Name        string `json:"name"`
			Description string `json:"description"`
			Image       string `json:"image"`
			Dna         string `json:"dna"`
			Edition     int    `json:"edition"`
			Date        int64  `json:"date"`
			Attributes  []struct {
				TraitType  string  `json:"trait_type"`
				Value      string  `json:"value"`
				TraitCount float64 `json:"trait_count"`
			} `json:"attributes"`
		} `json:"metadata"`
		FileInformation struct {
			Height   int `json:"height"`
			Width    int `json:"width"`
			FileSize int `json:"file_size"`
		} `json:"file_information"`
		FileURL            string      `json:"file_url"`
		AnimationURL       string      `json:"animation_url"`
		CachedFileURL      string      `json:"cached_file_url"`
		CachedAnimationURL interface{} `json:"cached_animation_url"`
		MintDate           string      `json:"mint_date"`
		UpdatedDate        string      `json:"updated_date"`
	} `json:"nft"`
	Owner    string `json:"owner"`
	Contract struct {
		Name     string `json:"name"`
		Symbol   string `json:"symbol"`
		Type     string `json:"type"`
		Metadata struct {
			Description        string `json:"description"`
			ThumbnailURL       string `json:"thumbnail_url"`
			CachedThumbnailURL string `json:"cached_thumbnail_url"`
			BannerURL          string `json:"banner_url"`
			CachedBannerURL    string `json:"cached_banner_url"`
		} `json:"metadata"`
	} `json:"contract"`
	Error struct {
		StatusCode int    `json:"status_code"`
		Code       string `json:"code"`
		Message    string `json:"message"`
	} `json:"error"`
}
