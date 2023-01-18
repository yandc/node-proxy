package types

type ETHGasEstimate struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Result  string `json:"result"` //seconds
}

type EVMGasEstimate struct {
	Status  string             `json:"status"`
	Message string             `json:"message"`
	Result  EVMGasEstimateData `json:"result"`
}

type EVMGasEstimateData struct {
	Pendingcount              string  `json:"pendingcount"`
	Avgminingblocktxcountsize int     `json:"avgminingblocktxcountsize"`
	Avgtxnsperblock           int     `json:"avgtxnsperblock"`
	Mingaspricegwei           float64 `json:"mingaspricegwei"`
	Avgnetworkutilization     float64 `json:"avgnetworkutilization"`
	Rapidgaspricegwei         int     `json:"rapidgaspricegwei"`
	Fastgaspricegwei          int     `json:"fastgaspricegwei"`
	Standardgaspricegwei      int     `json:"standardgaspricegwei"`
	Data                      string  `json:"data"`
}
