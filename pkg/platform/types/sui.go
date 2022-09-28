package types

const (
	RESPONSE_BALANCE    = "balance"
	RESPONSE_TXHASH     = "txHash"
	RESPONSE_OBJECTID   = "objectId"
	RESPONSE_TXPARAMS   = "txParams"
	RESPONSE_OBJECTREAD = "objectRead"
	RESPONSE_TXSTATUS   = "txStatus"
	RESPONSE_HEIGHT     = "height"
)

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
				Balance int `json:"balance"`
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

type SuiTransactionResponse struct {
	Certificate struct {
		AuthSignInfo struct {
			Epoch      int           `json:"epoch"`
			Signature  []interface{} `json:"signature"`
			SignersMap []int         `json:"signers_map"`
		} `json:"authSignInfo"`
		Data struct {
			GasBudget  int `json:"gasBudget"`
			GasPayment struct {
				Digest   string `json:"digest"`
				ObjectID string `json:"objectId"`
				Version  int    `json:"version"`
			} `json:"gasPayment"`
			Sender       string `json:"sender"`
			Transactions []struct {
				TransferObject struct {
					ObjectRef struct {
						Digest   string `json:"digest"`
						ObjectID string `json:"objectId"`
						Version  int    `json:"version"`
					} `json:"objectRef"`
					Recipient string `json:"recipient"`
				} `json:"TransferObject"`
			} `json:"transactions"`
		} `json:"data"`
		TransactionDigest string `json:"transactionDigest"`
		TxSignature       string `json:"txSignature"`
	} `json:"certificate"`
	Effects struct {
		GasObject struct {
			Owner struct {
				ObjectOwner string `json:"ObjectOwner"`
			} `json:"owner"`
			Reference struct {
				Digest   string `json:"digest"`
				ObjectID string `json:"objectId"`
				Version  int    `json:"version"`
			} `json:"reference"`
		} `json:"gasObject"`
		GasUsed struct {
			ComputationCost int `json:"computationCost"`
			StorageCost     int `json:"storageCost"`
			StorageRebate   int `json:"storageRebate"`
		} `json:"gasUsed"`
		Mutated []struct {
			Owner struct {
				AddressOwner string `json:"AddressOwner"`
			} `json:"owner"`
			Reference struct {
				Digest   string `json:"digest"`
				ObjectID string `json:"objectId"`
				Version  int    `json:"version"`
			} `json:"reference"`
		} `json:"mutated"`
		Status struct {
			Status string `json:"status"`
		} `json:"status"`
		TransactionDigest string `json:"transactionDigest"`
	} `json:"effects"`
	ParsedData  interface{} `json:"parsed_data"`
	TimestampMs interface{} `json:"timestamp_ms"`
}
