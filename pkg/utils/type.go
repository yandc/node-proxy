package utils

type GetPriceV2Req struct {
	CoinName    []string `json:"coin_name"`
	CoinAddress []string `json:"coin_address"`
	Currency    string   `json:"currency"`
}

type GetABIReq struct {
	Chain    string `json:"chain"`
	Contract string `json:"contract"`
	MethodId string `json:"method_id"`
}

type ParseDataByABIReq struct {
	Chain    string `json:"chain"`
	Contract string `json:"contract"`
	Data     string `json:"data"`
}

type PretreatmentReq struct {
	Chain string `json:"chain"`
	From  string `json:"from"`
	To    string `json:"to"`
	Data  string `json:"data"`
	Value string `json:"value"`
}

type IsContractReq struct {
	Chain   string `json:"chain"`
	Address string `json:"address"`
}

type GasDefaultsReq struct {
}
