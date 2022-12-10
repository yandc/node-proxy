package types

import "time"

type CasperBlock struct {
	APIVersion string `json:"api_version"`
	Block      struct {
		Hash   string `json:"hash"`
		Header struct {
			ParentHash      string      `json:"parent_hash"`
			StateRootHash   string      `json:"state_root_hash"`
			BodyHash        string      `json:"body_hash"`
			RandomBit       bool        `json:"random_bit"`
			AccumulatedSeed string      `json:"accumulated_seed"`
			EraEnd          interface{} `json:"era_end"`
			Timestamp       time.Time   `json:"timestamp"`
			EraID           int         `json:"era_id"`
			Height          int         `json:"height"`
			ProtocolVersion string      `json:"protocol_version"`
		} `json:"header"`
		Body struct {
			Proposer       string        `json:"proposer"`
			DeployHashes   []interface{} `json:"deploy_hashes"`
			TransferHashes []interface{} `json:"transfer_hashes"`
		} `json:"body"`
		Proofs []struct {
			PublicKey string `json:"public_key"`
			Signature string `json:"signature"`
		} `json:"proofs"`
	} `json:"block"`
}

type CasperRequest struct {
	JsonRPC string      `json:"jsonrpc"`
	Id      uint64      `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
}

type CasperAccount struct {
	APIVersion string `json:"api_version"`
	Account    struct {
		AccountHash    string        `json:"account_hash"`
		NamedKeys      []interface{} `json:"named_keys"`
		MainPurse      string        `json:"main_purse"`
		AssociatedKeys []struct {
			AccountHash string `json:"account_hash"`
			Weight      int    `json:"weight"`
		} `json:"associated_keys"`
		ActionThresholds struct {
			Deployment    int `json:"deployment"`
			KeyManagement int `json:"key_management"`
		} `json:"action_thresholds"`
	} `json:"account"`
	MerkleProof string `json:"merkle_proof"`
}

type CasperBalance struct {
	BalanceValue string `json:"balance_value"`
}

type CasperTxResponse struct {
	APIVersion string `json:"api_version"`
	DeployHash string `json:"deploy_hash"`
}
