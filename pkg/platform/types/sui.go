package types

type SuiObjectInfo struct {
	ObjectID string `json:"objectId"`
	Version  int    `json:"version"`
	Digest   string `json:"digest"`
	Type     string `json:"type"`
	Owner    struct {
		AddressOwner string `json:"AddressOwner"`
	} `json:"owner"`
	PreviousTransaction string `json:"previousTransaction"`
}

type SuiObjectRead struct {
	Status  string `json:"status"`
	Details struct {
		Data struct {
			DataType          string `json:"dataType"`
			Type              string `json:"type"`
			HasPublicTransfer bool   `json:"has_public_transfer"`
			Fields            struct {
				Balance string `json:"balance"`
				ID      struct {
					ID string `json:"id"`
				} `json:"id"`
			} `json:"fields"`
		} `json:"data"`
		Owner struct {
			AddressOwner string `json:"AddressOwner"`
		} `json:"owner"`
		PreviousTransaction string `json:"previousTransaction"`
		StorageRebate       int    `json:"storageRebate"`
		Reference           struct {
			ObjectID string `json:"objectId"`
			Version  int    `json:"version"`
			Digest   string `json:"digest"`
		} `json:"reference"`
	} `json:"details"`
}

type SuiTransactionEffects struct {
	Effects struct {
		Status struct {
			Status string `json:"status"`
		} `json:"status"`
	} `json:"effects"`
}

type SuiTransactionResponse struct {
	EffectsCert struct {
		Certificate struct {
			TransactionDigest string `json:"transactionDigest"`
		} `json:"certificate"`
		Effects struct {
			TransactionEffectsDigest string `json:"transactionEffectsDigest"`
			Effects                  struct {
				Status struct {
					Status string `json:"status"`
				} `json:"status"`
				TransactionDigest string   `json:"transactionDigest"`
				Dependencies      []string `json:"dependencies"`
			} `json:"effects"`
		} `json:"effects"`
	} `json:"EffectsCert"`
}
