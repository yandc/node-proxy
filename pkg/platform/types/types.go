package types

import (
	"context"
	"encoding/json"
	"fmt"
	v1 "gitlab.bixin.com/mili/node-proxy/api/platform/v1"
	v12 "gitlab.bixin.com/mili/node-proxy/api/tokenlist/v1"
	"math/big"
)

const (
	RESPONSE_BALANCE              = "balance"
	RESPONSE_SUI_BALANCE          = "suiBalance"
	RESPONSE_TOKEN_BALANCE        = "tokenBalance"
	RESPONSE_TXHASH               = "txHash"
	RESPONSE_TXDATA               = "txData"
	RESPONSE_OBJECTID             = "objectId"
	RESPONSE_TXPARAMS             = "txParams"
	RESPONSE_OBJECTREAD           = "objectRead"
	RESPONSE_TXSTATUS             = "txStatus"
	RESPONSE_HEIGHT               = "height"
	RESPONSE_RECENT_BLOCK_HASH    = "recentBlockHash"
	RESPONSE_TOKEN_ACTIVE         = "tokenActive"
	RESPONSE_ADDRESS_ACTIVE       = "addressActive"
	RESPONSE_RENT                 = "rent"
	RESPONSE_TOKEN_INFO           = "tokenInfo"
	RESPONSE_ACCOUNTS             = "accounts"
	NFTINFO                       = "nftInfo"
	RESPONSE_STATEROOTHASH        = "stateRootHash"
	RESPONSE_MAINPURSE            = "mainPurse"
	RESPONSE_GAS_PRICE            = "gasPrice"
	RESPONSE_DRY_RUN              = "dryRun"
	RESPONSE_BATCH_OBJECTID       = "batchObjectId"
	RESPONSE_DRY_RUN_PRETREATMENT = "dryRunPretreatment"

	BUILD_HEIGHT   = "height"
	BUILD_ACCOUNTS = "accounts"
	BUILD_BALANCE  = "balance"
	BUILD_TX       = "transaction"
	IS_COCHAIN     = "isCochain"
)

type AnalysisResponseType func(params string, result json.RawMessage) (interface{}, error)

type Platform interface {
	GetBalance(ctx context.Context, address, tokenAddress, decimals string) (string, error)
	BuildWasmRequest(ctx context.Context, nodeRpc, functionName, params string) (*v1.BuildWasmRequestReply, error)
	AnalysisWasmResponse(ctx context.Context, functionName, params, response string) (string, error)
	GetRpcURL() []string
	GetTokenType(token string) (*v12.GetTokenInfoResp_Data, error)
	IsContractAddress(address string) (bool, error)
	GetERCType(token string) string
}

type BtcClient interface {
	GetBalance(address string) (string, error)
}

type TronClient interface {
	GetBalance(address string) (string, error)
	GetTokenBalance(address string, tokenAddress string, decimals int) (string, error)
	IsContractAddress(address string) (bool, error)
}

type STCBalance struct {
	Raw  string `json:"raw"`
	JSON struct {
		Token struct {
			Value int64 `json:"value"`
		} `json:"token"`
	} `json:"json"`
}

//STCRequest is a jsonrpc request
type STCRequest struct {
	ID      int           `json:"id"`
	Jsonrpc string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
}

// STCResponse is a jsonrpc response
type STCResponse struct {
	ID     uint64          `json:"id"`
	Result json.RawMessage `json:"result"`
	Error  *ErrorObject    `json:"error,omitempty"`
}

// ErrorObject is a jsonrpc error
type ErrorObject struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Error implements error interface
func (e *ErrorObject) Error() string {
	data, err := json.Marshal(e)
	if err != nil {
		return fmt.Sprintf("jsonrpc.internal marshal error: %v", err)
	}
	return string(data)
}

type BlockCypherBalance struct {
	Balance big.Int `json:"balance"`
}

type BlockStreamBalance struct {
	Address    string `json:"address"`
	ChainStats struct {
		FundedTxoCount int     `json:"funded_txo_count"`
		FundedTxoSum   big.Int `json:"funded_txo_sum"`
		SpentTxoCount  int     `json:"spent_txo_count"`
		SpentTxoSum    big.Int `json:"spent_txo_sum"`
		TxCount        int     `json:"tx_count"`
	} `json:"chain_stats"`
	MempoolStats struct {
		FundedTxoCount int     `json:"funded_txo_count"`
		FundedTxoSum   big.Int `json:"funded_txo_sum"`
		SpentTxoCount  int     `json:"spent_txo_count"`
		SpentTxoSum    big.Int `json:"spent_txo_sum"`
		TxCount        int     `json:"tx_count"`
	} `json:"mempool_stats"`
}

type ChainBalance struct {
	Status string `json:"status"`
	Data   struct {
		Network            string      `json:"network"`
		Address            string      `json:"address"`
		ConfirmedBalance   string      `json:"confirmed_balance"` //"0.19900000"
		UnconfirmedBalance interface{} `json:"unconfirmed_balance"`
	} `json:"data"`
}

type TronBalance struct {
	Balance int64 `json:"balance"`
}

type TronBalanceReq struct {
	Address string `json:"address"`
	Visible bool   `json:"visible"`
}

type TronscanBalances struct {
	Balances           []TScanBalance `json:"balances"`
	Trc20TokenBalances []TScanBalance `json:"trc20token_balances"`
}
type TScanBalance struct {
	TokenId   string `json:"tokenId"`
	TokenName string `json:"tokenName"`
	Balance   string `json:"balance"`
}

type TronTokenBalanceRes struct {
	ConstantResult []string `json:"constant_result"`
}

type TronTokenBalanceReq struct {
	OwnerAddress     string `json:"owner_address"`
	ContractAddress  string `json:"contract_address"`
	FunctionSelector string `json:"function_selector"`
	Parameter        string `json:"parameter"`
	Visible          bool   `json:"visible"`
}

//Request is a jsonrpc request
type Request struct {
	ID      int           `json:"id"`
	Jsonrpc string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
}

// Response is a jsonrpc response
type Response struct {
	ID     uint64          `json:"id"`
	Result json.RawMessage `json:"result"`
	Error  *ErrorObject    `json:"error,omitempty"`
}

//type AnalysisResponse struct {
//	Ok      bool     `json:"ok"`
//	Message string   `json:"message"`
//	Data    Response `json:"data"`
//}

type STCListResource struct {
	Resources map[string]STCResource `json:"resources"`
}

type STCResource struct {
	Json struct {
		ScalingFactor int `json:"scaling_factor"`
	} `json:"json"`
}

type STCDecimal struct {
	ScalingFactor int `json:"scaling_factor"`
}

type ABIMethod struct {
	Type   string `json:"type"`
	Name   string `json:"name"`
	Inputs []struct {
		Type string `json:"type"`
	} `json:"inputs"`
}

type AptosExposedFunctions struct {
	Name              string        `json:"name"`
	Visibility        string        `json:"visibility"`
	IsEntry           bool          `json:"is_entry"`
	IsView            bool          `json:"is_view"`
	GenericTypeParams []interface{} `json:"generic_type_params"`
	Params            []string      `json:"params"`
	Return            []interface{} `json:"return"`
}

type AptosABI []struct {
	Abi struct {
		Address          string                  `json:"address"`
		Name             string                  `json:"name"`
		ExposedFunctions []AptosExposedFunctions `json:"exposed_functions"`
	} `json:"abi"`
}

type KlaytnABI struct {
	MatchedContract struct {
		ContractAbi string `json:"contractAbi"`
	} `json:"matchedContract"`
}

type RoninABI struct {
	Output struct {
		ABI []interface{} `json:"abi"`
	} `json:"output"`
}

type AbiDecodeResult struct {
	MethodName string                 `json:"method_name"`
	InputArgs  map[string]interface{} `json:"input_args"`
	Selector   string                 `json:"selector"`
}

type ParseDataResponse struct {
	TransactionType string          `json:"transaction_type"`
	DesData         AbiDecodeResult `json:"des_data"`
}

type PretreatmentResponse struct {
	ID       string `json:"id"`
	ChainID  string `json:"chainId"`
	Signer   string `json:"signer"`
	Hostname string `json:"hostname"`
	Type     string `json:"type"`
	To       struct {
		Description  string      `json:"description"`
		Address      string      `json:"address"`
		EtherscanURL string      `json:"etherscanUrl"`
		Info         interface{} `json:"info"`
	} `json:"to"`
	AssetChanges []struct {
		Action   string `json:"action"`
		Color    string `json:"color"`
		Metadata struct {
			Icon          string      `json:"icon"`
			URL           string      `json:"url"`
			Verified      bool        `json:"verified"`
			Name          string      `json:"name"`
			SecondaryLine interface{} `json:"secondaryLine"`
		} `json:"metadata"`
	} `json:"assetChanges"`
}

type ZkSyncABIInfo struct {
	Info struct {
		VerificationInfo struct {
			Artifacts struct {
				Abi []interface{} `json:"abi"`
			} `json:"artifacts"`
		} `json:"verificationInfo"`
	} `json:"info"`
}
