package casper

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
	"gitlab.bixin.com/mili/node-proxy/pkg/platform/utils"
	utils2 "gitlab.bixin.com/mili/node-proxy/pkg/utils"
)

const (
	JSONRPC  = "2.0"
	JSONID   = 1
	DECIMALS = 9
)

type platform struct {
	rpcURL []string
	log    *log.Helper
	chain  string
}

func (p *platform) GetBalance(ctx context.Context, address, tokenAddress, decimals string) (string, error) {
	return "", nil
}

func getBlock(nodeURL string) (*types.CasperBlock, error) {
	method := "chain_get_block"
	out := &types.CasperBlock{}
	err := callContext(nodeURL, method, out, nil)
	if err != nil {
		return nil, err
	}
	return out, err
}

func getAccount(nodeURL string, params interface{}) (*types.CasperAccount, error) {
	method := "state_get_account_info"
	out := &types.CasperAccount{}
	err := callContext(nodeURL, method, out, params)
	if err != nil {
		return nil, err
	}
	return out, err
}

func (p *platform) BuildWasmRequest(ctx context.Context, nodeRpc, functionName, params string) (*v1.BuildWasmRequestReply, error) {
	var reqParams interface{}
	tempParams, _ := hex.DecodeString(params)
	json.Unmarshal(tempParams, &reqParams)
	var body string
	var err error
	result := &v1.BuildWasmRequestReply{
		Method: "POST",
		Url:    nodeRpc,
		Head: map[string]string{
			"Content-Type": "application/json",
		},
		Body: body,
	}
	switch functionName {
	case types.RESPONSE_HEIGHT:
		block, err := getBlock(nodeRpc)
		if err != nil {
			body = "-1"
		} else {
			body = fmt.Sprintf("%v", block.Block.Header.Height)
		}
	case types.RESPONSE_MAINPURSE:
		account, err := getAccount(nodeRpc, reqParams)
		if err == nil {
			body = account.Account.MainPurse
		}
	case types.RESPONSE_STATEROOTHASH:
		block, err := getBlock(nodeRpc)
		if err == nil {
			body = block.Block.Header.StateRootHash
		}
	case types.RESPONSE_BALANCE:
		out := &types.CasperBalance{}
		err = callContext(nodeRpc, "state_get_balance", out, reqParams)
		if err == nil {
			if out != nil && out.BalanceValue != "" {
				body = utils.UpdateDecimals(out.BalanceValue, DECIMALS)
			}
		}
	case types.RESPONSE_TXHASH:
		out := &types.CasperTxResponse{}
		err = callContext(nodeRpc, "account_put_deploy", out, reqParams)
		if err == nil {
			body = out.DeployHash
		}
	}
	if err != nil {
		result.Method = err.Error()
	}
	result.Body = body
	return result, nil
}

func callContext(url, method string, out interface{}, params interface{}) error {
	var response types.Response
	rpcRequest := types.CasperRequest{
		Id:      JSONID,
		Method:  method,
		Params:  params,
		JsonRPC: JSONRPC,
	}
	resp, err := utils2.HttpsParamsPost(url, rpcRequest)
	if err != nil {
		return err
	}
	if err := json.Unmarshal([]byte(resp), &response); err != nil {
		return err
	}
	if response.Error != nil {
		return errors.New(response.Error.Message)
	}
	if len(response.Result) == 0 {
		return errors.New("call context result error")
	}
	err = json.Unmarshal(response.Result, &out)
	if err != nil {
		return err
	}
	return nil
}

var responseFunc = map[string]types.AnalysisResponseType{
	types.RESPONSE_BALANCE:       analysisBalance,
	types.RESPONSE_TXHASH:        analysisTxHash,
	types.RESPONSE_TXPARAMS:      analysisTxParams,
	types.RESPONSE_HEIGHT:        analysisTxHeight,
	types.RESPONSE_STATEROOTHASH: analysisStateRootHash,
	types.RESPONSE_MAINPURSE:     analysisMainPurse,
}

func analysisMainPurse(params string, result json.RawMessage) (interface{}, error) {
	var account types.CasperAccount
	if err := json.Unmarshal(result, &account); err != nil {
		return nil, err
	}
	return account.Account.MainPurse, nil
}

func analysisStateRootHash(params string, result json.RawMessage) (interface{}, error) {
	var block types.CasperBlock
	if err := json.Unmarshal(result, &block); err != nil {
		return nil, err
	}
	return block.Block.Header.StateRootHash, nil
}

func analysisTxHeight(params string, result json.RawMessage) (interface{}, error) {
	var block types.CasperBlock
	if err := json.Unmarshal(result, &block); err != nil {
		return nil, err
	}
	return block.Block.Header.Height, nil
}

func analysisTxParams(params string, result json.RawMessage) (interface{}, error) {
	return nil, nil
}

func analysisTxHash(params string, result json.RawMessage) (interface{}, error) {
	var txResponse types.CasperTxResponse
	if err := json.Unmarshal(result, &txResponse); err != nil {
		return nil, err
	}
	return txResponse.DeployHash, nil
}

func analysisBalance(params string, result json.RawMessage) (interface{}, error) {
	var balance types.CasperBalance
	if err := json.Unmarshal(result, &balance); err != nil {
		return nil, err
	}
	return utils.UpdateDecimals(balance.BalanceValue, 9), nil
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

func (p *platform) GetRpcURL() []string {
	return p.rpcURL
}

func (p *platform) GetTokenType(token string) (*v12.GetTokenInfoResp_Data, error) {
	return nil, nil
}

func NewCasperPlatform(chain string, rpcURL []string, logger log.Logger) types.Platform {
	log := log.NewHelper(log.With(logger, "module", "platform/sui"))
	return &platform{rpcURL: rpcURL, log: log, chain: chain}
}

func (p *platform) IsContractAddress(address string) (bool, error) {
	return false, nil
}

func (p *platform) GetERCType(token string) string {
	return ""
}
