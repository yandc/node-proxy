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

type GasOracleReq struct {
	Key       string `json:"key"`
	CacheTime int64  `json:"cache_time"`
}

type GasOracleRes struct {
	SafeGasPrice    string `json:"safe_gas_price"`
	ProposeGasPrice string `json:"propose_gas_price"`
	FastGasPrice    string `json:"fast_gas_price"`
	SuggestBaseFee  string `json:"suggest_base_fee"`
}

type GasOracleOkex struct {
	Data map[string]interface{} `json:"data"`
}

type GasOracleResult struct {
	Result map[string]interface{} `json:"result"`
}

const (
	ERC20_TYPE_ERR = "token address is not erc20"
)

type CheckHMTokenAddress struct {
	Chain        string   `json:"chain"`
	TokenAddress []string `json:"token_address"`
}
