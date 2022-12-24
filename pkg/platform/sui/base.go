package sui

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
	"sort"
	"strconv"
	"strings"
)

const (
	JSONRPC        = "2.0"
	JSONID         = 1
	RESULT_SUCCESS = "Exists"
	NATIVE_TYPE    = "0x2::coin::Coin<0x2::sui::SUI>"
	TXFAILED       = "failed"
	TXSUCCESS      = "success"
)

type platform struct {
	rpcURL []string
	log    *log.Helper
	chain  string
}

func NewSuiPlatform(chain string, rpcURL []string, logger log.Logger) types.Platform {
	log := log.NewHelper(log.With(logger, "module", "platform/sui"))
	return &platform{rpcURL: rpcURL, log: log, chain: chain}
}

func (p *platform) GetBalance(ctx context.Context, address, tokenAddress, decimals string) (string, error) {
	return "", nil
}

func (p *platform) BuildWasmRequest(ctx context.Context, nodeRpc, functionName, params string) (*v1.BuildWasmRequestReply, error) {
	var reqParams []interface{}
	tempParams, _ := hex.DecodeString(params)
	json.Unmarshal(tempParams, &reqParams)
	request := types.Request{
		ID:      JSONID,
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
	types.RESPONSE_BALANCE:    analysisBalance,
	types.RESPONSE_OBJECTID:   analysisObjectIds,
	types.RESPONSE_TXHASH:     analysisTxHash,
	types.RESPONSE_TXPARAMS:   analysisTxParams,
	types.RESPONSE_OBJECTREAD: analysisObjectRead,
	types.RESPONSE_TXSTATUS:   analysisTxStatus,
	types.RESPONSE_HEIGHT:     analysisTxHeight,
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

func analysisTxHash(params string, result json.RawMessage) (interface{}, error) {
	var txInfo types.SuiTransactionResponse
	if err := json.Unmarshal(result, &txInfo); err != nil {
		return "", err
	}
	return txInfo.EffectsCert.Certificate.TransactionDigest, nil
}

func analysisObjectIds(params string, result json.RawMessage) (interface{}, error) {
	var objectInfo []types.SuiObjectInfo
	if err := json.Unmarshal(result, &objectInfo); err != nil {
		return nil, err
	}
	var paramsMap map[string]string
	if err := json.Unmarshal([]byte(params), &paramsMap); err != nil {
		return nil, err
	}
	coinType := NATIVE_TYPE
	if value, ok := paramsMap["coinType"]; ok {
		coinType = value
	}
	objectIds := make([]string, 0, len(objectInfo))
	for _, object := range objectInfo {
		if object.Type == coinType {
			objectIds = append(objectIds, object.ObjectID)
		}
	}
	return objectIds, nil
}

func analysisBalance(params string, result json.RawMessage) (interface{}, error) {
	var objectRead types.SuiObjectRead
	if err := json.Unmarshal(result, &objectRead); err != nil {
		return nil, err
	}
	return strconv.Atoi(objectRead.Details.Data.Fields.Balance)
}

func analysisObjectRead(params string, result json.RawMessage) (interface{}, error) {
	var objectRead types.SuiObjectRead
	if err := json.Unmarshal(result, &objectRead); err != nil {
		return nil, err
	}
	return objectRead, nil
}

func analysisTxStatus(params string, result json.RawMessage) (interface{}, error) {
	var objectRead types.SuiTransactionEffects
	if err := json.Unmarshal(result, &objectRead); err != nil {
		return nil, err
	}
	if strings.Contains(objectRead.Effects.Status.Status, "fai") {
		return TXFAILED, nil
	}
	return TXSUCCESS, nil
}

func analysisTxParams(params string, result json.RawMessage) (interface{}, error) {
	var paramsMap map[string]string
	if err := json.Unmarshal([]byte(params), &paramsMap); err != nil {
		return nil, err
	}
	var amount = 0
	if value, ok := paramsMap["amount"]; ok {
		amount, _ = strconv.Atoi(value)
	}
	var tempResult []string
	if err := json.Unmarshal(result, &tempResult); err != nil {
		return nil, err
	}
	objectReads := make([]types.SuiObjectRead, 0, len(tempResult))
	for _, r := range tempResult {
		var tempRead types.SuiObjectRead
		if err := json.Unmarshal([]byte(r), &tempRead); err != nil {
			return nil, err
		}
		objectReads = append(objectReads, tempRead)
	}
	gasLimit := 60
	balance := 0
	sort.Slice(objectReads, func(i, j int) bool {
		balanceI, _ := strconv.Atoi(objectReads[i].Details.Data.Fields.Balance)
		balanceJ, _ := strconv.Atoi(objectReads[j].Details.Data.Fields.Balance)
		return balanceI > balanceJ
	})
	suiObjects := make([]interface{}, 0, len(objectReads))
	for _, objectRead := range objectReads {
		filedBalance, err := strconv.Atoi(objectRead.Details.Data.Fields.Balance)
		if err != nil {
			return nil, err
		}
		suiObjects = append(suiObjects, map[string]interface{}{
			"objectId": objectRead.Details.Reference.ObjectID,
			"seqNo":    fmt.Sprintf("%v", objectRead.Details.Reference.Version),
			"digest":   objectRead.Details.Reference.Digest,
			"balance":  fmt.Sprintf("%v", filedBalance),
		})
		balance += filedBalance
		if amount > balance+gasLimit {
			gasLimit += 60
		}
	}
	if amount == 0 {
		gasLimit += 60 * len(objectReads)
	}
	if len(objectReads) == 0 || amount > balance+gasLimit {
		return nil, errors.New("insufficiency of balance")
	}
	return map[string]interface{}{
		"gasPrice":   "1",
		"suiObjects": suiObjects,
		"gasLimit":   strconv.Itoa(int(float32(gasLimit) * 1.2)),
	}, nil
}

func (p *platform) GetRpcURL() []string {
	return p.rpcURL
}

func (p *platform) GetTokenType(token string) (*v12.GetTokenInfoResp_Data, error) {
	return nil, nil
}
