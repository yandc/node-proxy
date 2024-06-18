package types

type TonTokenInfo struct {
	Jetton struct {
		Metadata struct {
			Name   string `json:"name"`
			Symbol string `json:"symbol"`
			Image  struct {
				Original string `json:"original"`
			} `json:"image"`
			Decimals int `json:"decimals"`
		} `json:"metadata"`
	} `json:"jetton"`
	Message string `json:"message"`
}

type TonBlocks struct {
	Blocks []TonBlockInfo `json:"blocks"`
}

type TonBlockInfo struct {
	Workchain              int    `json:"workchain"`
	Shard                  string `json:"shard"`
	Seqno                  uint64 `json:"seqno"`
	RootHash               string `json:"root_hash"`
	FileHash               string `json:"file_hash"`
	GlobalID               int    `json:"global_id"`
	Version                int    `json:"version"`
	AfterMerge             bool   `json:"after_merge"`
	BeforeSplit            bool   `json:"before_split"`
	AfterSplit             bool   `json:"after_split"`
	WantMerge              bool   `json:"want_merge"`
	WantSplit              bool   `json:"want_split"`
	KeyBlock               bool   `json:"key_block"`
	VertSeqnoIncr          bool   `json:"vert_seqno_incr"`
	Flags                  int    `json:"flags"`
	GenUtime               string `json:"gen_utime"`
	StartLt                string `json:"start_lt"`
	EndLt                  string `json:"end_lt"`
	ValidatorListHashShort int    `json:"validator_list_hash_short"`
	GenCatchainSeqno       int    `json:"gen_catchain_seqno"`
	MinRefMcSeqno          int    `json:"min_ref_mc_seqno"`
	PrevKeyBlockSeqno      int    `json:"prev_key_block_seqno"`
	VertSeqno              int    `json:"vert_seqno"`
	MasterRefSeqno         int    `json:"master_ref_seqno"`
	RandSeed               string `json:"rand_seed"`
	CreatedBy              string `json:"created_by"`
	TxCount                int    `json:"tx_count"`
	MasterchainBlockRef    struct {
		Workchain int    `json:"workchain"`
		Shard     string `json:"shard"`
		Seqno     int    `json:"seqno"`
	} `json:"masterchain_block_ref"`
	PrevBlocks []struct {
		Workchain int    `json:"workchain"`
		Shard     string `json:"shard"`
		Seqno     int    `json:"seqno"`
	} `json:"prev_blocks"`
}

type TonAccountResp struct {
	Balance             string      `json:"balance"`
	Code                string      `json:"code"`
	Data                string      `json:"data"`
	LastTransactionLt   string      `json:"last_transaction_lt"`
	LastTransactionHash string      `json:"last_transaction_hash"`
	FrozenHash          interface{} `json:"frozen_hash"`
	Status              string      `json:"status"`
}

type TonSendTxHash struct {
	MessageHash string `json:"message_hash"`
	Error       string `json:"error"`
}

type TonTxParams struct {
	FromAddress    string `json:"from_address"`
	ToAddress      string `json:"to_address"`
	Amount         string `json:"amount"`
	Memo           string `json:"memo"`
	TokenAddress   string `json:"token_address"`
	Type           string `json:"type"` //token/nft
	EstimateFeeReq string `json:"estimate_fee_req"`
}

type TonJettonList struct {
	JettonWallets []JettonWallet `json:"jetton_wallets"`
}

type JettonWallet struct {
	Address           string `json:"address"`
	Balance           string `json:"balance"`
	Owner             string `json:"owner"`
	Jetton            string `json:"jetton"`
	LastTransactionLt string `json:"last_transaction_lt"`
	CodeHash          string `json:"code_hash"`
	DataHash          string `json:"data_hash"`
}

type TonNFTList struct {
	NftItems []TonNftItem `json:"nft_items"`
}

type TonNftItem struct {
	Address           string `json:"address"`
	CollectionAddress string `json:"collection_address"`
	OwnerAddress      string `json:"owner_address"`
	Init              bool   `json:"init"`
	Index             string `json:"index"`
	LastTransactionLt string `json:"last_transaction_lt"`
	CodeHash          string `json:"code_hash"`
	DataHash          string `json:"data_hash"`
	Content           string `json:"content"`
	Collection        struct {
		Address           string `json:"address"`
		OwnerAddress      string `json:"owner_address"`
		LastTransactionLt string `json:"last_transaction_lt"`
		NextItemIndex     string `json:"next_item_index"`
		CollectionContent string `json:"collection_content"`
		CodeHash          string `json:"code_hash"`
		DataHash          string `json:"data_hash"`
	} `json:"collection"`
}

type TonEstimateFeeResp struct {
	SourceFees struct {
		InFwdFee   int `json:"in_fwd_fee"`
		StorageFee int `json:"storage_fee"`
		GasFee     int `json:"gas_fee"`
		FwdFee     int `json:"fwd_fee"`
	} `json:"source_fees"`
	DestinationFees []struct {
		InFwdFee   int `json:"in_fwd_fee"`
		StorageFee int `json:"storage_fee"`
		GasFee     int `json:"gas_fee"`
		FwdFee     int `json:"fwd_fee"`
	} `json:"destination_fees"`
	Error string `json:"error"`
}

type TonWalletInfo struct {
	Balance             string `json:"balance"`
	WalletType          string `json:"wallet_type"`
	SeqNo               int    `json:"seqno"`
	WalletID            int    `json:"wallet_id"`
	LastTransactionLt   string `json:"last_transaction_lt"`
	LastTransactionHash string `json:"last_transaction_hash"`
	Status              string `json:"status"`
}
