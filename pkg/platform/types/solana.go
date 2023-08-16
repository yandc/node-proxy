package types

type SolanaTokenInfo struct {
	Value struct {
		Amount         string  `json:"amount"`
		Decimals       int     `json:"decimals"`
		UIAmount       float64 `json:"uiAmount"`
		UIAmountString string  `json:"uiAmountString"`
	} `json:"value"`
}

type SolanaBalance struct {
	Context struct {
		ApiVersion string `json:"apiVersion"`
		Slot       int64  `json:"slot"`
	}
	Value int64 `json:"value"`
}

type SolanaTokenAccount struct {
	Context struct {
		APIVersion string `json:"apiVersion"`
		Slot       int    `json:"slot"`
	} `json:"context"`
	Value []struct {
		Account struct {
			Data struct {
				Parsed struct {
					Info struct {
						IsNative    bool   `json:"isNative"`
						Mint        string `json:"mint"`
						Owner       string `json:"owner"`
						State       string `json:"state"`
						TokenAmount struct {
							Amount         string  `json:"amount"`
							Decimals       int     `json:"decimals"`
							UIAmount       float64 `json:"uiAmount"`
							UIAmountString string  `json:"uiAmountString"`
						} `json:"tokenAmount"`
					} `json:"info"`
					Type string `json:"type"`
				} `json:"parsed"`
				Program string `json:"program"`
				Space   int    `json:"space"`
			} `json:"data"`
			Executable bool   `json:"executable"`
			Lamports   int    `json:"lamports"`
			Owner      string `json:"owner"`
			RentEpoch  int    `json:"rentEpoch"`
		} `json:"account"`
		Pubkey string `json:"pubkey"`
	} `json:"value"`
}

type SolanaRecentBlockHash struct {
	Context struct {
		APIVersion string `json:"apiVersion"`
		Slot       int    `json:"slot"`
	} `json:"context"`
	Value struct {
		Blockhash     string `json:"blockhash"`
		FeeCalculator struct {
			LamportsPerSignature int `json:"lamportsPerSignature"`
		} `json:"feeCalculator"`
	} `json:"value"`
}

type SolanaAccountInfo struct {
	Context struct {
		APIVersion string `json:"apiVersion"`
		Slot       int    `json:"slot"`
	} `json:"context"`
	Value struct {
		Data       []string `json:"data"`
		Executable bool     `json:"executable"`
		Lamports   int64    `json:"lamports"`
		Owner      string   `json:"owner"`
		RentEpoch  int      `json:"rentEpoch"`
	} `json:"value"`
}

type SolanaTokenType struct {
	Succcess bool `json:"succcess"`
	Data     struct {
		Symbol   string `json:"symbol"`
		Address  string `json:"address"`
		Name     string `json:"name"`
		Icon     string `json:"icon"`
		Decimals uint32 `json:"decimals"`
		Holder   int    `json:"holder"`
	} `json:"data"`
}
