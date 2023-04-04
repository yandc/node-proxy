package solana

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-kratos/kratos/v2/log"
	v1 "gitlab.bixin.com/mili/node-proxy/api/platform/v1"
	v12 "gitlab.bixin.com/mili/node-proxy/api/tokenlist/v1"
	"gitlab.bixin.com/mili/node-proxy/pkg/platform/types"
	utils2 "gitlab.bixin.com/mili/node-proxy/pkg/platform/utils"
	"gitlab.bixin.com/mili/node-proxy/pkg/utils"
	"strconv"
)

const (
	SOLANA_DECIMALS = 9
	ID              = 1
	JSONRPC         = "2.0"
)

type platform struct {
	rpcURL []string
	log    *log.Helper
	chain  string
}

func NewSolanaPlatform(chain string, rpcURL []string, logger log.Logger) types.Platform {
	log := log.NewHelper(log.With(logger, "module", "platform/solana"))
	return &platform{rpcURL: rpcURL, log: log, chain: chain}
}

func (p *platform) GetRpcURL() []string {
	return p.rpcURL
}

func (p *platform) BuildWasmRequest(ctx context.Context, nodeRpc, functionName, params string) (*v1.BuildWasmRequestReply, error) {
	var reqParams []interface{}
	tempParams, _ := hex.DecodeString(params)
	json.Unmarshal(tempParams, &reqParams)
	request := types.Request{
		ID:      ID,
		Jsonrpc: JSONRPC,
		Method:  functionName,
		Params:  reqParams,
	}
	body, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}
	return &v1.BuildWasmRequestReply{
		Method: "POST",
		Url:    nodeRpc,
		Head: map[string]string{
			"Content-Type": "application/json",
		},
		Body: string(body),
	}, nil
}

var responseFunc = map[string]types.AnalysisResponseType{
	types.RESPONSE_BALANCE:           analysisBalance,
	types.RESPONSE_TXHASH:            analysisTxHash,
	types.RESPONSE_TXPARAMS:          analysisTxParams,
	types.RESPONSE_HEIGHT:            analysisTxHeight,
	types.RESPONSE_TOKEN_BALANCE:     analysisTokenBalance,
	types.RESPONSE_RECENT_BLOCK_HASH: analysisRecentBlockHash,
	types.RESPONSE_TOKEN_ACTIVE:      analysisTokenActive,
	types.RESPONSE_RENT:              analysisRent,
	types.RESPONSE_ADDRESS_ACTIVE:    analysisAddressActive,
	types.RESPONSE_TOKEN_INFO:        analysisTokenInfo,
}

func (p *platform) AnalysisWasmResponse(ctx context.Context, functionName, params, response string) (string, error) {
	var resp types.Response
	json.Unmarshal([]byte(response), &resp)

	if resp.Error != nil {
		return "", resp.Error
	}
	if doFunc, ok := responseFunc[functionName]; ok {
		result, err := doFunc(params, resp.Result)
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

func analysisTxHeight(params string, result json.RawMessage) (interface{}, error) {
	var height int
	if err := json.Unmarshal(result, &height); err != nil {
		return height, err
	}
	return height, nil
}

func analysisBalance(params string, result json.RawMessage) (interface{}, error) {
	var balanceInfo types.SolanaBalance
	if err := json.Unmarshal(result, &balanceInfo); err != nil {
		return nil, err
	}
	balance := fmt.Sprintf("%d", balanceInfo.Value)
	return utils2.UpdateDecimals(balance, SOLANA_DECIMALS), nil
}

func analysisTokenBalance(params string, result json.RawMessage) (interface{}, error) {
	var tokenAccounts types.SolanaTokenAccount
	if err := json.Unmarshal(result, &tokenAccounts); err != nil {
		return nil, err
	}
	if len(tokenAccounts.Value) > 0 {
		return tokenAccounts.Value[0].Account.Data.Parsed.Info.TokenAmount.UIAmountString, nil
	}
	return "0", nil
}

func analysisTxHash(params string, result json.RawMessage) (interface{}, error) {
	var txHash string
	if err := json.Unmarshal(result, &txHash); err != nil {
		return nil, err
	}
	return txHash, nil
}

func analysisRecentBlockHash(params string, result json.RawMessage) (interface{}, error) {
	var recentInfo types.SolanaRecentBlockHash
	if err := json.Unmarshal(result, &recentInfo); err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"block_hash": recentInfo.Value.Blockhash,
		"gas_price":  recentInfo.Value.FeeCalculator.LamportsPerSignature,
	}, nil
}

func analysisTokenInfo(params string, result json.RawMessage) (interface{}, error) {
	var tokenInfo types.SolanaTokenInfo
	if err := json.Unmarshal(result, &tokenInfo); err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"decimals": fmt.Sprintf("%v", tokenInfo.Value.Decimals),
	}, nil
}

func analysisTokenActive(params string, result json.RawMessage) (interface{}, error) {
	var tokenAccounts types.SolanaTokenAccount
	if err := json.Unmarshal(result, &tokenAccounts); err != nil {
		return nil, err
	}
	if len(tokenAccounts.Value) == 0 {
		return "false", nil
	}
	return "true", nil
}

func analysisRent(params string, result json.RawMessage) (interface{}, error) {
	var rent int
	if err := json.Unmarshal(result, &rent); err != nil {
		return rent, err
	}
	return rent, nil
}

func analysisAddressActive(params string, result json.RawMessage) (interface{}, error) {
	var addressInfo types.SolanaAccountInfo
	if err := json.Unmarshal(result, &addressInfo); err != nil {
		return nil, err
	}
	if addressInfo.Value.Owner != "" {
		return "true", nil
	}
	return "false", nil
}

func analysisTxParams(params string, result json.RawMessage) (interface{}, error) {
	var paramsMap map[string]string
	if err := json.Unmarshal([]byte(params), &paramsMap); err != nil {
		return nil, err
	}
	txParams := make(map[string]interface{})
	recentBlockHash := paramsMap["recent_block_hash"]
	gasPrice := paramsMap["gas_price"]
	rentBalance := paramsMap["rent_balance"]
	if tokenActive, ok := paramsMap["token_active"]; ok {
		gasPriceInt, _ := strconv.Atoi(gasPrice)
		if tokenActive == "false" {
			gasPriceInt += 2039280
		}
		txParams["is_token_active"] = tokenActive
	} else {
		toBalance := paramsMap["to_balance"]
		fromBalance := paramsMap["from_balance"]
		addressIsActive := paramsMap["address_is_active"]
		rentFlat, _ := strconv.ParseFloat(rentBalance, 64)
		toFlat, _ := strconv.ParseFloat(toBalance, 64)
		fromFlat, _ := strconv.ParseFloat(fromBalance, 64)
		gasPriceInfo := utils2.UpdateDecimals(gasPrice, SOLANA_DECIMALS)
		gasPriceFlag, _ := strconv.ParseFloat(gasPriceInfo, 64)
		maxAmount := fromFlat - rentFlat - gasPriceFlag
		minAmount := rentFlat - toFlat
		if minAmount < 0 {
			minAmount = 0
		}
		if addressIsActive == "false" {
			minAmount = rentFlat
		}
		maxAmount, _ = strconv.ParseFloat(fmt.Sprintf("%.9f", maxAmount), 64)
		minAmount, _ = strconv.ParseFloat(fmt.Sprintf("%.9f", minAmount), 64)
		txParams["max_amount"] = maxAmount
		txParams["min_amount"] = minAmount
	}
	txParams["recent_blockhash"] = recentBlockHash
	txParams["gasPrice"] = gasPrice
	txParams["rent"] = rentBalance
	return txParams, nil
}

func (p *platform) GetTokenType(token string) (*v12.GetTokenInfoResp_Data, error) {
	method := "getTokenSupply"
	params := []interface{}{token}
	out := &types.SolanaTokenInfo{}
	var resultErr error
	for _, url := range p.rpcURL {
		err := utils.JsonHttpsPost(url, ID, method, JSONRPC, out, params)
		if err != nil {
			resultErr = err
			continue
		}
		return &v12.GetTokenInfoResp_Data{
			Chain:    p.chain,
			Address:  token,
			Decimals: uint32(out.Value.Decimals),
		}, nil
	}
	return nil, resultErr
}

func (p *platform) GetBalance(ctx context.Context, address, tokenAddress, decimals string) (string, error) {
	//get native balance
	if tokenAddress == "" || address == tokenAddress {
		for _, url := range p.rpcURL {
			method := "getBalance"
			params := []interface{}{address}
			out := &types.SolanaBalance{}
			err := utils.JsonHttpsPost(url, ID, method, JSONRPC, out, params)
			if err != nil {
				return "", err
			}
			balance := fmt.Sprintf("%d", out.Value)
			return utils2.UpdateDecimals(balance, SOLANA_DECIMALS), nil
		}
	}

	//get token balance
	if address != tokenAddress && tokenAddress != "" {
		for _, url := range p.rpcURL {
			method := "getTokenAccountsByOwner"
			params := []interface{}{address, map[string]string{"mint": tokenAddress}, map[string]string{"encoding": "jsonParsed"}}
			out := &types.SolanaTokenAccount{}
			err := utils.JsonHttpsPost(url, ID, method, JSONRPC, out, params)
			if err != nil {
				return "", err
			}
			if len(out.Value) > 0 {
				return out.Value[0].Account.Data.Parsed.Info.TokenAmount.UIAmountString, nil
			}
			return "0", err

		}
	}
	return "0", nil
}
