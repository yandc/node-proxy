package utils

type GetPriceV2Req struct {
	CoinName    []string `json:"coin_name"`
	CoinAddress []string `json:"coin_address"`
	Currency    string   `json:"currency"`
}
