package types

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
)

type Platform interface {
	GetBalance(ctx context.Context, address, tokenAddress, decimals string) (string, error)
	GetRpcURL() []string
}

type BtcClient interface {
	GetBalance(address string) (string, error)
}

type TronClient interface {
	GetBalance(address string) (string, error)
	GetTokenBalance(address string, tokenAddress string, decimals int) (string, error)
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
