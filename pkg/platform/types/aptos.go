package types

type AptosBalanceResp struct {
	Data struct {
		Coin struct {
			Value string `json:"value"`
		} `json:"coin"`
	} `json:"data"`
	AptosBadResp
}

type AptosBadResp struct {
	ErrorCode          string `json:"error_code"`
	Message            string `json:"message"`
	AptosLedgerVersion string `json:"aptos_ledger_version"`
}

type AptosResourceResp struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

type AptosTokenInfo struct {
	Decimals int    `json:"decimals"`
	Name     string `json:"name"`
	Symbol   string `json:"symbol"`
}
