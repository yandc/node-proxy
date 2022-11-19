package types

type AptosNFTData struct {
	Data struct {
		CurrentTokenDatas []struct {
			TokenDataIDHash   string `json:"token_data_id_hash"`
			Name              string `json:"name"`
			CollectionName    string `json:"collection_name"`
			Description       string `json:"description"`
			CreatorAddress    string `json:"creator_address"`
			DefaultProperties struct {
				TOKENBURNABLEBYOWNER   string `json:"TOKEN_BURNABLE_BY_OWNER"`
				TOKENBURNABLEBYCREATOR string `json:"TOKEN_BURNABLE_BY_CREATOR"`
			} `json:"default_properties"`
			LargestPropertyVersion   int    `json:"largest_property_version"`
			Maximum                  int    `json:"maximum"`
			MetadataURI              string `json:"metadata_uri"`
			PayeeAddress             string `json:"payee_address"`
			RoyaltyPointsDenominator int    `json:"royalty_points_denominator"`
			RoyaltyPointsNumerator   int    `json:"royalty_points_numerator"`
			Supply                   int    `json:"supply"`
			Typename                 string `json:"__typename"`
		} `json:"current_token_datas"`
	} `json:"data"`
}

type AptosNFTReq struct {
	OperationName string      `json:"operationName"`
	Variables     interface{} `json:"variables"`
	Query         string      `json:"query"`
}

type AptosCollectionData struct {
	Data struct {
		CurrentCollectionDatas []struct {
			CollectionName string `json:"collection_name"`
			Description    string `json:"description"`
			CreatorAddress string `json:"creator_address"`
			MetadataURI    string `json:"metadata_uri"`
			Typename       string `json:"__typename"`
		} `json:"current_collection_datas"`
	} `json:"data"`
}

type AptosNFTSourceData struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Image       string `json:"image"`
	Dna         string `json:"dna"`
	Edition     int    `json:"edition"`
	Date        int64  `json:"date"`
	Attributes  []struct {
		TraitType string `json:"trait_type"`
		Value     string `json:"value"`
	} `json:"attributes"`
	Compiler string `json:"compiler"`
}

type AptosTopazNFTData struct {
	Error interface{} `json:"error"`
	Data  struct {
		TokenID        string `json:"token_id"`
		Creator        string `json:"creator"`
		Collection     string `json:"collection"`
		Name           string `json:"name"`
		Description    string `json:"description"`
		URI            string `json:"uri"`
		CollectionSlug string `json:"collection_slug"`
	} `json:"data"`
	Count      interface{} `json:"count"`
	Status     int         `json:"status"`
	StatusText string      `json:"statusText"`
	Body       struct {
		TokenID        string `json:"token_id"`
		Creator        string `json:"creator"`
		Collection     string `json:"collection"`
		Name           string `json:"name"`
		Description    string `json:"description"`
		URI            string `json:"uri"`
		CollectionSlug string `json:"collection_slug"`
	} `json:"body"`
}

type AptosTopazCollection struct {
	Data struct {
		Collection struct {
			CollectionID string `json:"collection_id"`
			Creator      string `json:"creator"`
			Name         string `json:"name"`
			Description  string `json:"description"`
			URI          string `json:"uri"`
			Slug         string `json:"slug"`
			LogoURI      string `json:"logo_uri"`
		} `json:"collection"`
	} `json:"data"`
	Error interface{} `json:"error"`
}
