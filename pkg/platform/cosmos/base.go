package cosmos

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-kratos/kratos/v2/log"
	v1 "gitlab.bixin.com/mili/node-proxy/api/platform/v1"
	v12 "gitlab.bixin.com/mili/node-proxy/api/tokenlist/v1"
	"gitlab.bixin.com/mili/node-proxy/pkg/platform/types"
	"strconv"
	"strings"
)

type platform struct {
	rpcURL []string
	log    *log.Helper
	chain  string
}

func (p *platform) GetBalance(ctx context.Context, address, tokenAddress, decimals string) (string, error) {
	return "", nil
}

func (p *platform) BuildWasmRequest(ctx context.Context, nodeRpc, functionName, params string) (*v1.BuildWasmRequestReply, error) {
	var url, method, body string
	switch functionName {
	case types.BUILD_HEIGHT:
		url = fmt.Sprintf("%s/blocks/latest", nodeRpc)
		method = "GET"
	case types.BUILD_ACCOUNTS:
		url = fmt.Sprintf("%s/cosmos/auth/v1beta1/accounts/%s", nodeRpc, params)
		method = "GET"
	case types.BUILD_BALANCE:
		url = fmt.Sprintf("%s/cosmos/bank/v1beta1/balances/%s", nodeRpc, params)
		method = "GET"
	case types.BUILD_TX:
		url = fmt.Sprintf("%s/cosmos/tx/v1beta1/txs", nodeRpc)
		method = "POST"
		body = params
	}
	return &v1.BuildWasmRequestReply{
		Method: method,
		Url:    url,
		Head: map[string]string{
			"Content-Type": "application/json",
		},
		Body: body,
	}, nil
}

type cosmosAnalysisResponseType func(params string, result string) (interface{}, error)

var responseFunc = map[string]cosmosAnalysisResponseType{
	types.RESPONSE_BALANCE:  analysisBalance,
	types.RESPONSE_TXHASH:   analysisTxHash,
	types.RESPONSE_ACCOUNTS: analysisAccounts,
	types.RESPONSE_HEIGHT:   analysisTxHeight,
	types.RESPONSE_TXPARAMS: analysisTxParams,
}

func (p *platform) AnalysisWasmResponse(ctx context.Context, functionName, params, response string) (string, error) {

	if doFunc, ok := responseFunc[functionName]; ok {
		result, err := doFunc(params, response)
		if err != nil {
			return "", err
		}
		if value, strOk := result.(string); strOk {
			return value, nil
		}
		b, _ := json.Marshal(result)
		return string(b), nil
	}
	return "", errors.New("no func to  analysis response")
}

func analysisTxHeight(params string, result string) (interface{}, error) {
	var resp types.CosmosBlockHeight
	var height int
	if err := json.Unmarshal([]byte(result), &resp); err != nil {
		return height, err
	}
	height, _ = strconv.Atoi(resp.Block.Header.Height)
	return height, nil
}

func analysisBalance(params string, result string) (interface{}, error) {
	var resp types.CosmosBalances
	if err := json.Unmarshal([]byte(result), &resp); err != nil {
		return nil, err
	}
	if resp.Code != 0 || resp.Message != "" {
		return nil, errors.New(resp.Message)
	}
	balances := make(map[string]string)
	for _, balance := range resp.Balances {
		balances[strings.ToLower(balance.Denom)] = balance.Amount
	}
	return balances, nil
}

func analysisTxHash(params string, result string) (interface{}, error) {
	var resp types.CosmosTxHash
	if err := json.Unmarshal([]byte(result), &resp); err != nil {
		return nil, err
	}
	if resp.Code != 0 || resp.Message != "" {
		return "", errors.New(resp.Message)
	}
	if resp.TxResponse.Height == "0" && resp.TxResponse.RawLog != "" {
		return "", errors.New(resp.TxResponse.RawLog)
	}
	return resp.TxResponse.Txhash, nil
}

func analysisAccounts(params string, result string) (interface{}, error) {
	var resp types.CosmosAccount
	if err := json.Unmarshal([]byte(result), &resp); err != nil {
		return nil, err
	}
	if resp.Code != 0 || resp.Message != "" {
		return "", errors.New(resp.Message)
	}
	return map[string]string{
		"accountNumber": resp.Account.AccountNumber,
		"sequence":      resp.Account.Sequence,
	}, nil
}

func analysisTxParams(params string, result string) (interface{}, error) {
	var paramsMap map[string]string
	if err := json.Unmarshal([]byte(params), &paramsMap); err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"accountNumber": paramsMap["accountNumber"],
		"nonce":         paramsMap["sequence"],
		"lowFee":        "0.0025",
		"mediumFee":     "0.0025",
		"highFee":       "0.025",
		"gasLimit":      "100000",
	}, nil
}

func (p *platform) GetRpcURL() []string {
	return p.rpcURL
}

func (p *platform) GetTokenType(token string) (*v12.GetTokenInfoResp_Data, error) {
	return nil, nil
}

func NewCosmosPlatform(chain string, rpcURL []string, logger log.Logger) types.Platform {
	log := log.NewHelper(log.With(logger, "module", "platform/sui"))
	return &platform{rpcURL: rpcURL, log: log, chain: chain}
}

func (p *platform) IsContractAddress(address string) (bool, error) {
	return false, nil
}
