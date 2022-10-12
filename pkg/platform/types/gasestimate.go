package types

type ETHGasEstimate struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Result  string `json:"result"` //seconds
}
