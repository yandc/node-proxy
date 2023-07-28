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
	"gitlab.bixin.com/mili/node-proxy/pkg/platform/utils"
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

func Platform2SUIPlatform(p types.Platform) *platform {
	suiPlatform, ok := p.(*platform)
	if ok {
		return suiPlatform
	}
	return nil
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
	types.RESPONSE_BALANCE:       analysisBalance,
	types.RESPONSE_OBJECTID:      analysisObjectIds,
	types.RESPONSE_TXHASH:        analysisTxHash,
	types.RESPONSE_TXPARAMS:      analysisTxParams,
	types.RESPONSE_OBJECTREAD:    analysisObjectRead,
	types.RESPONSE_TXSTATUS:      analysisTxStatus,
	types.RESPONSE_HEIGHT:        analysisTxHeight,
	types.RESPONSE_TOKEN_BALANCE: analysisTokenBalance,
	types.RESPONSE_TOKEN_INFO:    analysisTokenInfo,
	types.RESPONSE_GAS_PRICE:     analysisGasPrice,
	types.RESPONSE_TXDATA:        analysisTxData,
}

func (p *platform) AnalysisWasmResponse(ctx context.Context, functionName, params, response string) (string, error) {
	var result interface{}
	var err error
	if functionName == types.RESPONSE_SUI_BALANCE {
		result, err = analysisBalanceV2(response)
	} else if doFunc, ok := responseFunc[functionName]; ok {
		var resp types.Response
		json.Unmarshal([]byte(response), &resp)
		if resp.Error != nil {
			return "", resp.Error
		}
		result, err = doFunc(params, resp.Result)
	} else {
		return "", errors.New("no func to  analysis response")
	}
	if err != nil {
		return "", err
	}
	if value, strOk := result.(string); strOk {
		return value, nil
	}
	b, _ := json.Marshal(result)
	return string(b), nil
}

func analysisTxHeight(params string, result json.RawMessage) (interface{}, error) {
	var height string
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
	return txInfo.Effects.TransactionDigest, nil
}

func analysisTxData(params string, result json.RawMessage) (interface{}, error) {
	var txInfo types.SuiTransactionResponse
	if err := json.Unmarshal(result, &txInfo); err != nil {
		return "", err
	}
	resp, _ := json.Marshal(txInfo)
	return string(resp), nil
}

func analysisObjectIds(params string, result json.RawMessage) (interface{}, error) {
	var objectInfo types.SuiObjectInfo
	if err := json.Unmarshal(result, &objectInfo); err != nil {
		return nil, err
	}
	//var paramsMap map[string]string
	//if err := json.Unmarshal([]byte(params), &paramsMap); err != nil {
	//	return nil, err
	//}
	//coinType := NATIVE_TYPE
	//if value, ok := paramsMap["coinType"]; ok {
	//	coinType = value
	//}
	objectIds := make([]string, 0, len(objectInfo.Data))
	for _, info := range objectInfo.Data {
		//if info.Type == coinType {
		objectIds = append(objectIds, info.Data.ObjectID)
		//}
	}
	return objectIds, nil
}

func analysisBalance(params string, result json.RawMessage) (interface{}, error) {
	var out types.SUIBalance
	if err := json.Unmarshal(result, &out); err != nil {
		return nil, err
	}
	return out.TotalBalance, nil
}

func analysisTokenBalance(params string, result json.RawMessage) (interface{}, error) {
	var out []types.SUIBalance
	if err := json.Unmarshal(result, &out); err != nil {
		return nil, err
	}
	tokenBalances := make(map[string]string)
	for _, info := range out {
		tokenBalances[info.CoinType] = info.TotalBalance
	}
	return tokenBalances, nil
}

func analysisTokenInfo(params string, result json.RawMessage) (interface{}, error) {
	var out types.SUICoinMetadata
	if err := json.Unmarshal(result, &out); err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"decimals": fmt.Sprintf("%v", out.Decimals),
		"symbol":   strings.ToUpper(out.Symbol),
		"name":     out.Name,
	}, nil
}

func analysisBalanceV2(response string) (interface{}, error) {
	balance := 0
	var resps []types.Response
	if err := json.Unmarshal([]byte(response), &resps); err != nil {
		return 0, err
	}
	for _, resp := range resps {
		if resp.Error != nil {
			return 0, errors.New(resp.Error.Message)
		}
		var objectRead types.SuiObjectRead
		if err := json.Unmarshal(resp.Result, &objectRead); err != nil {
			return 0, err
		}
		filedBalance, _ := strconv.Atoi(objectRead.Details.Data.Fields.Balance)
		balance += filedBalance

	}
	return balance, nil
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
	//var amount = 0
	//if value, ok := paramsMap["amount"]; ok {
	//	amount, _ = strconv.Atoi(value)
	//}
	var gasPrice = "1000"
	if value, ok := paramsMap["gasPrice"]; ok {
		gasPrice = value
	}
	coinType := ""
	if value, ok := paramsMap["coinType"]; ok {
		coinType = value
	}
	coinKey := ""
	if value, ok := paramsMap["coinKey"]; ok {
		coinKey = value
	}
	tokenId := ""
	if value, ok := paramsMap["tokenId"]; ok {
		tokenId = value
	}
	var objectReads []types.SuiObjectResponse
	if err := json.Unmarshal(result, &objectReads); err != nil {
		return nil, err
	}
	gasLimit := 3000
	sort.Slice(objectReads, func(i, j int) bool {
		balanceI, _ := strconv.Atoi(objectReads[i].Data.Content.Fields.Balance)
		balanceJ, _ := strconv.Atoi(objectReads[j].Data.Content.Fields.Balance)
		return balanceI > balanceJ
	})
	suiObjects := make([]interface{}, 0, len(objectReads))
	for _, objectRead := range objectReads {
		if objectRead.Data.Type == NATIVE_TYPE {
			if objectRead.Data.Content.Fields.Balance != "" && objectRead.Data.Content.Fields.Balance != "0" {
				filedBalance, _ := strconv.Atoi(objectRead.Data.Content.Fields.Balance)
				suiObjects = append(suiObjects, map[string]interface{}{
					"objectId": objectRead.Data.ObjectID,
					"seqNo":    fmt.Sprintf("%v", objectRead.Data.Version),
					"digest":   objectRead.Data.Digest,
					"balance":  fmt.Sprintf("%v", filedBalance),
				})
			}
		} else if (coinKey == "nft" && objectRead.Data.ObjectID == tokenId) ||
			(coinKey != "nft" && (objectRead.Data.Type == coinType ||
				objectRead.Data.Type == fmt.Sprintf("0x2::coin::Coin<%v>", coinType))) {
			filedBalance, _ := strconv.Atoi(objectRead.Data.Content.Fields.Balance)
			suiObjects = append(suiObjects, map[string]interface{}{
				"objectId": objectRead.Data.ObjectID,
				"seqNo":    fmt.Sprintf("%v", objectRead.Data.Version),
				"digest":   objectRead.Data.Digest,
				"balance":  fmt.Sprintf("%v", filedBalance),
				coinKey:    coinType,
			})
		}

	}
	if len(objectReads) == 0 {
		return nil, errors.New("insufficiency of balance")
	}
	return map[string]interface{}{
		"gasPrice":   gasPrice,
		"suiObjects": suiObjects,
		"gasLimit":   strconv.Itoa(int(float32(gasLimit) * 1.2)),
	}, nil
}

func analysisGasPrice(params string, result json.RawMessage) (interface{}, error) {
	var gasPrice string
	if err := json.Unmarshal(result, &gasPrice); err != nil {
		return nil, err
	}
	return gasPrice, nil
}

func (p *platform) GetRpcURL() []string {
	return p.rpcURL
}

func (p *platform) GetTokenType(token string) (*v12.GetTokenInfoResp_Data, error) {
	var resultErr error
	for i := 0; i < len(p.rpcURL); i++ {
		ret, err := getWASMCoinMetadata(p.chain, p.rpcURL[i], token)
		if err != nil {
			resultErr = err
			continue
		}
		return ret, nil
	}
	if resultErr != nil && strings.Contains(token, "::") {
		tokenInfo := strings.Split(token, "::")
		if len(tokenInfo) >= 3 {
			return &v12.GetTokenInfoResp_Data{
				Chain:    p.chain,
				Address:  token,
				Decimals: uint32(0),
				Symbol:   tokenInfo[2],
				Name:     tokenInfo[2],
			}, nil
		}

	}
	return nil, resultErr
}

func (p *platform) GetNFTObject(objectId string) (*types.SuiNFTObjectResponse, error) {
	var resultErr error
	for i := 0; i < len(p.rpcURL); i++ {
		method := "sui_getObject"
		params := []interface{}{objectId, map[string]bool{"showType": true,
			"showOwner":               true,
			"showPreviousTransaction": true,
			"showDisplay":             true,
			"showContent":             true,
			"showBcs":                 true,
			"showStorageRebate":       true}}
		out := &types.SuiNFTObjectResponse{}
		err := call(p.rpcURL[i], JSONID, method, out, params)
		if err != nil {
			resultErr = err
			continue
		}
		return out, nil
	}

	return nil, resultErr
}

func getWASMCoinMetadata(chain, url, coinType string) (*v12.GetTokenInfoResp_Data, error) {
	method := "suix_getCoinMetadata"
	out := &types.SUICoinMetadata{}
	params := []interface{}{coinType}
	err := call(url, JSONID, method, &out, params)
	if err != nil {
		return nil, err
	}
	return &v12.GetTokenInfoResp_Data{
		Chain:    chain,
		Address:  coinType,
		Decimals: uint32(out.Decimals),
		Symbol:   out.Symbol,
		Name:     out.Name,
		LogoURI:  out.IconUrl,
	}, nil
}

func call(url string, id int, method string, out interface{}, params []interface{}) error {
	var resp types.Response
	err := utils.HttpsPost(url, id, method, JSONRPC, &resp, params)
	if err != nil {
		return err
	}
	if resp.Error != nil {
		return errors.New(resp.Error.Message)
	}
	return json.Unmarshal(resp.Result, &out)
}

func (p *platform) IsContractAddress(address string) (bool, error) {
	var resultErr error
	for i := 0; i < len(p.rpcURL); i++ {
		method := "sui_getObject"
		out := &types.SuiObjectRead{}
		params := []interface{}{address}
		err := call(p.rpcURL[i], JSONID, method, &out, params)
		if err != nil {
			resultErr = err
			continue
		}
		if out.Status != RESULT_SUCCESS {
			resultErr = errors.New(out.Status)
			continue
		}
		return out.Error.Code != "notExists", nil
	}
	return false, resultErr
}
