package types

//type SuiObjectInfo struct {
//	ObjectID string `json:"objectId"`
//	Version  int    `json:"version"`
//	Digest   string `json:"digest"`
//	Type     string `json:"type"`
//	Owner    struct {
//		AddressOwner string `json:"AddressOwner"`
//	} `json:"owner"`
//	PreviousTransaction string `json:"previousTransaction"`
//}

type SuiObjectInfo struct {
	Data []struct {
		Data struct {
			ObjectID string      `json:"objectId"`
			Version  interface{} `json:"version"`
			Digest   string      `json:"digest"`
		} `json:"data"`
	} `json:"data"`
	NextCursor  string `json:"nextCursor"`
	HasNextPage bool   `json:"hasNextPage"`
}

type SuiObjectRead struct {
	Status string `json:"status"`
	Error  struct {
		Code     string `json:"code"`
		ObjectId string `json:"object_id"`
	} `json:"error"`
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

type SuiObjectResponse struct {
	Data struct {
		ObjectID string      `json:"objectId"`
		Version  interface{} `json:"version"`
		Digest   string      `json:"digest"`
		Type     string      `json:"type"`
		Content  struct {
			DataType          string `json:"dataType"`
			Type              string `json:"type"`
			HasPublicTransfer bool   `json:"hasPublicTransfer"`
			Fields            struct {
				Balance string `json:"balance"`
				ID      struct {
					ID string `json:"id"`
				} `json:"id"`
			} `json:"fields"`
		} `json:"content"`
	} `json:"data"`
}

type SuiTransactionEffects struct {
	Effects struct {
		Status struct {
			Status string `json:"status"`
		} `json:"status"`
	} `json:"effects"`
}

type SuiTransactionResponse struct {
	Digest      string `json:"digest"`
	Transaction struct {
		Data struct {
			MessageVersion string `json:"messageVersion"`
			Transaction    struct {
				Kind   string `json:"kind"`
				Inputs []struct {
					Type       string      `json:"type"`
					ValueType  string      `json:"valueType,omitempty"`
					Value      interface{} `json:"value,omitempty"`
					ObjectType string      `json:"objectType,omitempty"`
					ObjectID   string      `json:"objectId,omitempty"`
					Version    string      `json:"version,omitempty"`
					Digest     string      `json:"digest,omitempty"`
				} `json:"inputs"`
				Transactions []struct {
					TransferObjects []interface{} `json:"TransferObjects"`
				} `json:"transactions"`
			} `json:"transaction"`
			Sender  string `json:"sender"`
			GasData struct {
				Payment []struct {
					ObjectID string `json:"objectId"`
					Version  int    `json:"version"`
					Digest   string `json:"digest"`
				} `json:"payment"`
				Owner  string `json:"owner"`
				Price  string `json:"price"`
				Budget string `json:"budget"`
			} `json:"gasData"`
		} `json:"data"`
		TxSignatures []string `json:"txSignatures"`
	} `json:"transaction"`
	RawTransaction string `json:"rawTransaction"`
	Effects        struct {
		MessageVersion string `json:"messageVersion"`
		Status         struct {
			Status string `json:"status"`
		} `json:"status"`
		ExecutedEpoch string `json:"executedEpoch"`
		GasUsed       struct {
			ComputationCost         string `json:"computationCost"`
			StorageCost             string `json:"storageCost"`
			StorageRebate           string `json:"storageRebate"`
			NonRefundableStorageFee string `json:"nonRefundableStorageFee"`
		} `json:"gasUsed"`
		TransactionDigest string `json:"transactionDigest"`
		Mutated           []struct {
			Owner struct {
				AddressOwner string `json:"AddressOwner"`
			} `json:"owner"`
			Reference struct {
				ObjectID string `json:"objectId"`
				Version  int    `json:"version"`
				Digest   string `json:"digest"`
			} `json:"reference"`
		} `json:"mutated"`
		GasObject struct {
			Owner struct {
				ObjectOwner string `json:"ObjectOwner"`
			} `json:"owner"`
			Reference struct {
				ObjectID string `json:"objectId"`
				Version  int    `json:"version"`
				Digest   string `json:"digest"`
			} `json:"reference"`
		} `json:"gasObject"`
		EventsDigest string `json:"eventsDigest"`
	} `json:"effects"`
	ObjectChanges []struct {
		Type      string `json:"type"`
		Sender    string `json:"sender"`
		Recipient struct {
			AddressOwner string `json:"AddressOwner"`
		} `json:"recipient"`
		ObjectType string `json:"objectType"`
		ObjectID   string `json:"objectId"`
		Version    string `json:"version"`
		Digest     string `json:"digest"`
	} `json:"objectChanges"`
}

type SUIBalance struct {
	CoinType        string `json:"coinType"`
	CoinObjectCount int    `json:"coinObjectCount"`
	TotalBalance    string `json:"totalBalance"`
}

type SUICoinMetadata struct {
	Decimals uint8  `json:"decimals"`
	Name     string `json:"name"`
	Symbol   string `json:"symbol"`
	IconUrl  string `json:"iconUrl"`
}

type SuiNFTObjectResponse struct {
	Data struct {
		ObjectID string `json:"objectId"`
		Version  string `json:"version"`
		Digest   string `json:"digest"`
		Type     string `json:"type"`
		Owner    struct {
			AddressOwner string `json:"AddressOwner"`
		} `json:"owner"`
		PreviousTransaction string `json:"previousTransaction"`
		StorageRebate       string `json:"storageRebate"`
		Display             struct {
			Data struct {
				Collection  string `json:"collection"`
				Creator     string `json:"creator"`
				Description string `json:"description"`
				ImageURL    string `json:"image_url"`
				ImgURL      string `json:"img_url"`
				Name        string `json:"name"`
				ProjectURL  string `json:"project_url"`
			} `json:"data"`
			Error interface{} `json:"error"`
		} `json:"display"`
		Content struct {
			DataType          string `json:"dataType"`
			Type              string `json:"type"`
			HasPublicTransfer bool   `json:"hasPublicTransfer"`
			Fields            struct {
				ID struct {
					ID string `json:"id"`
				} `json:"id"`
				ImageURL string `json:"image_url"`
				ImgURL   string `json:"img_url"`
				Name     string `json:"name"`
				URL      string `json:"url"`
			} `json:"fields"`
		} `json:"content"`
		Bcs struct {
			DataType          string `json:"dataType"`
			Type              string `json:"type"`
			HasPublicTransfer bool   `json:"hasPublicTransfer"`
			Version           int    `json:"version"`
			BcsBytes          string `json:"bcsBytes"`
		} `json:"bcs"`
	} `json:"data"`
}
