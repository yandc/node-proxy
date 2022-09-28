package types

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
