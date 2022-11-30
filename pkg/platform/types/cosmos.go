package types

type CosmosBlockHeight struct {
	Block struct {
		Header struct {
			Version struct {
				Block string `json:"block"`
			} `json:"version"`
			ChainID string `json:"chain_id"`
			Height  string `json:"height"`
		} `json:"header"`
	} `json:"block"`
}

type CosmosTxHash struct {
	Code       int    `json:"code"`
	Message    string `json:"message"`
	TxResponse struct {
		Height string `json:"height"`
		Txhash string `json:"txhash"`
		RawLog string `json:"raw_log"`
	} `json:"tx_response"`
}

type CosmosAccount struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Account struct {
		Type    string `json:"@type"`
		Address string `json:"address"`
		PubKey  struct {
			Type string `json:"@type"`
			Key  string `json:"key"`
		} `json:"pub_key"`
		AccountNumber string `json:"account_number"`
		Sequence      string `json:"sequence"`
	} `json:"account"`
}

type CosmosBalances struct {
	Code     int    `json:"code"`
	Message  string `json:"message"`
	Balances []struct {
		Denom  string `json:"denom"`
		Amount string `json:"amount"`
	} `json:"balances"`
}
