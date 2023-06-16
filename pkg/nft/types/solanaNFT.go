package types

type SolanaNFTData struct {
	Results struct {
		MintAddress         string `json:"mintAddress"`
		Supply              int    `json:"supply"`
		Title               string `json:"title"`
		Content             string `json:"content"`
		PrimarySaleHappened bool   `json:"primarySaleHappened"`
		UpdateAuthority     string `json:"updateAuthority"`
		OnChainCollection   struct {
		} `json:"onChainCollection"`
		SellerFeeBasisPoints int `json:"sellerFeeBasisPoints"`
		Creators             []struct {
			Address  string `json:"address"`
			Verified int    `json:"verified"`
			Share    int    `json:"share"`
		} `json:"creators"`
		Owner              string `json:"owner"`
		ID                 string `json:"id"`
		TokenDelegateValid bool   `json:"tokenDelegateValid"`
		IsFrozen           bool   `json:"isFrozen"`
		TokenStandard      int    `json:"tokenStandard"`
		Img                string `json:"img"`
		Attributes         []struct {
			TraitType string      `json:"trait_type"`
			Value     interface{} `json:"value"`
		} `json:"attributes"`
		Properties struct {
			Files []struct {
				URI  string `json:"uri"`
				Type string `json:"type"`
			} `json:"files"`
			Category string `json:"category"`
			Creators []struct {
				Share   int    `json:"share"`
				Address string `json:"address"`
			} `json:"creators"`
		} `json:"properties"`
		PropertyCategory string `json:"propertyCategory"`
		AnimationURL     string `json:"animationURL"`
		ExternalURL      string `json:"externalURL"`
		CollectionName   string `json:"collectionName"`
		CollectionTitle  string `json:"collectionTitle"`
		IsTradeable      bool   `json:"isTradeable"`
		Rarity           struct {
		} `json:"rarity"`
	} `json:"results"`
}
