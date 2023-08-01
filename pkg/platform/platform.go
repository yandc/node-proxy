package platform

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-redis/redis"
	v1 "gitlab.bixin.com/mili/node-proxy/api/platform/v1"
	v12 "gitlab.bixin.com/mili/node-proxy/api/tokenlist/v1"
	"gitlab.bixin.com/mili/node-proxy/internal/conf"
	"gitlab.bixin.com/mili/node-proxy/pkg/chainlist"
	"gitlab.bixin.com/mili/node-proxy/pkg/platform/aptos"
	"gitlab.bixin.com/mili/node-proxy/pkg/platform/bitcoin"
	"gitlab.bixin.com/mili/node-proxy/pkg/platform/casper"
	"gitlab.bixin.com/mili/node-proxy/pkg/platform/ethereum"
	"gitlab.bixin.com/mili/node-proxy/pkg/platform/solana"
	"gitlab.bixin.com/mili/node-proxy/pkg/platform/starcoin"
	"gitlab.bixin.com/mili/node-proxy/pkg/platform/sui"
	"gitlab.bixin.com/mili/node-proxy/pkg/platform/tron"
	"gitlab.bixin.com/mili/node-proxy/pkg/platform/types"
	"gitlab.bixin.com/mili/node-proxy/pkg/platform/utils"
	utils2 "gitlab.bixin.com/mili/node-proxy/pkg/utils"
	"math"
	"math/big"
	"strings"
	"time"
)

const (
	STC               = "STC"
	BTC               = "BTC"
	EVM               = "EVM"
	TVM               = "TVM"
	SUI               = "SUI"
	APTOS             = "APTOS"
	SOL               = "SOL"
	CSPR              = "CSPR"
	REDIS_ESTIME_KEY  = "platform:estime:"
	SETAPPROVALFORALL = "setApprovalForAll"
	APPROVENFT        = "approveNFT"
)

type TypeAndRpc struct {
	Type    string
	RpcURL  []string
	ChainId uint64
}

type config struct {
	log         *log.Helper
	logger      log.Logger
	chainInfo   map[string]TypeAndRpc
	redisClient *redis.Client
}

var c config

func InitPlatform(conf []*conf.Platform, logger log.Logger, client *redis.Client) {
	log := log.NewHelper(log.With(logger, "module", "platform/InitPlatform"))
	tempMap := make(map[string]TypeAndRpc, len(conf))
	for _, chainInfo := range conf {
		tempMap[chainInfo.Chain] = TypeAndRpc{
			Type:    chainInfo.Type,
			RpcURL:  chainInfo.RpcURL,
			ChainId: chainInfo.ChainId,
		}
	}
	c = config{
		log:         log,
		logger:      logger,
		chainInfo:   tempMap,
		redisClient: client,
	}
}

func GetBalance(ctx context.Context, chain, address, tokenAddress, decimals string) (string, error) {
	platform := newPlatform(chain)
	if platform == nil {
		c.log.Error("get platform is nil")
		return "", errors.New("platform is nil")
	}

	return platform.GetBalance(ctx, address, tokenAddress, decimals)
}

func BuildWasmRequest(ctx context.Context, chain, nodeRpc, functionName, params string) (*v1.BuildWasmRequestReply, error) {
	platform := newPlatform(chain)
	if platform == nil {
		c.log.Error("get platform is nil")
		return nil, errors.New("platform is nil")
	}
	result, err := platform.BuildWasmRequest(ctx, nodeRpc, functionName, params)
	if err != nil {
		c.log.Error("BuildWasmRequest Error:", err)
	}
	return result, err
}

func AnalysisWasmResponse(ctx context.Context, chain, functionName, params, response string) (string, error) {
	platform := newPlatform(chain)
	if platform == nil {
		c.log.Error("get platform is nil")
		return "", errors.New("platform is nil")
	}
	return platform.AnalysisWasmResponse(ctx, functionName, params, response)
}

func GetPlatformTokenInfo(chain, token string) (*v12.GetTokenInfoResp_Data, error) {
	platform := newPlatform(chain)
	if platform == nil {
		c.log.Error("get platform is nil")
		return nil, errors.New("platform is nil")
	}
	result, err := platform.GetTokenType(token)
	if err != nil {
		return result, err
	}
	if result == nil {
		return nil, errors.New("dont get token:" + token + " info")
	}
	return result, nil
}

func newPlatform(chain string) types.Platform {
	if value, ok := c.chainInfo[chain]; ok {
		switch value.Type {
		case EVM:
			return ethereum.NewEVMPlatform(chain, value.RpcURL, c.logger)
		case STC:
			return starcoin.NewSTCPlatform(chain, value.RpcURL, c.logger)
		case BTC:
			return bitcoin.NewBTCPlatform(chain, value.RpcURL, c.logger)
		case TVM:
			return tron.NewTronPlatform(chain, value.RpcURL, c.logger)
		case SUI:
			return sui.NewSuiPlatform(chain, value.RpcURL, c.logger)
		case APTOS:
			return aptos.NewAptosPlatform(chain, value.RpcURL, c.logger)
		case SOL:
			return solana.NewSolanaPlatform(chain, value.RpcURL, c.logger)
		case CSPR:
			return casper.NewCasperPlatform(chain, value.RpcURL, c.logger)
		}
	} else if strings.HasPrefix(strings.ToLower(chain), "evm") { //支持自定义EVM
		url := getRpcUrl(chain)
		return ethereum.NewEVMPlatform(chain, url, c.logger)
	}

	return nil
}

func getRpcUrl(chain string) []string {
	_, chainId, found := strings.Cut(strings.ToLower(chain), "evm")
	if !found {
		c.log.Error("unsupported evm chain")
		return nil
	}
	nodeUrls, err := chainlist.FindChainNodeUrlList(chainId)
	if err != nil {
		c.log.Error("get chain node url list error", "err", err)
		return nil
	}

	rpcUrls := make([]string, len(nodeUrls))
	for i, nodeUrl := range nodeUrls {
		rpcUrls[i] = nodeUrl.Url
	}

	return rpcUrls
}

func GetSUINFTInfo(chain, objectId string) (*types.SuiNFTObjectResponse, error) {
	platform := newPlatform(chain)
	if platform == nil {
		c.log.Error("get platform is nil")
		return nil, errors.New("platform is nil")
	}
	suiPlatform := sui.Platform2SUIPlatform(platform)
	if platform == nil {
		c.log.Error("sui platform is nil")
		return nil, errors.New("sui platform is nil")
	}
	return suiPlatform.GetNFTObject(objectId)
}

// GetGasEstimateTime returns the estimated time, in seconds, for a transaction to be confirmed on the blockchain.
func GetGasEstimateTime(chain string, gasInfo string) (string, error) {
	switch chain {
	case "ETH":
		return getETHGasEstimate(gasInfo)
	case "BSC", "Fantom", "Polygon", "Avalanche":
		return getEVMGasEstimate(chain, gasInfo)
	}
	return "", nil
}

func getEVMGasEstimate(chain string, gasInfo string) (string, error) {
	var url string
	var blockInterval int
	switch chain {
	case "BSC":
		url = "https://gbsc.blockscan.com/gasapi.ashx?apikey=key&method=pendingpooltxgweidata"
		blockInterval = 3
	case "Fantom":
		url = "https://gftm.blockscan.com/gasapi.ashx?apikey=key&method=pendingpooltxgweidata"
		blockInterval = 2
	case "Polygon":
		url = "https://gpoly.blockscan.com/gasapi.ashx?apikey=key&method=pendingpooltxgweidata"
		blockInterval = 2
	case "Avalanche":
		url = "https://gavax.blockscan.com/gasapi.ashx?apikey=key&method=pendingpooltxgweidata"
		blockInterval = 2
	default:
		return "", nil
	}
	var gasMap map[string]string
	if err := json.Unmarshal([]byte(gasInfo), &gasMap); err != nil {
		return "", err
	}
	gasPrice := gasMap["gas_price"]
	tempGasPrice, flag := new(big.Float).SetString(gasPrice)
	if !flag {
		return "", errors.New("float set string error:" + gasPrice)
	}
	gasPriceGWei, _ := new(big.Float).Quo(tempGasPrice, big.NewFloat(1000000000)).Float64()
	out := &types.EVMGasEstimate{}
	redisKey := REDIS_ESTIME_KEY + chain
	esTimeData, updateFlag, _ := utils.GetESTimeRedisValueByKey(c.redisClient, redisKey)
	if esTimeData != "" {
		if err := json.Unmarshal([]byte(esTimeData), &out.Result); err != nil {
			return "", err
		}
	}
	if esTimeData == "" || updateFlag {
		err := utils.HttpsGetForm(url, nil, out)
		if err != nil {
			return "", err
		}
		if out.Status != "1" {
			return "", errors.New(out.Message)
		}
		resultByte, _ := json.Marshal(out.Result)
		err = utils.SetESTimeRedisKey(c.redisClient, redisKey, string(resultByte))
		if err != nil {
			c.log.Error("set estime redis error:", err)
		}
	}

	var data [][]float64
	if err := json.Unmarshal([]byte(out.Result.Data), &data); err != nil {
		return "", err
	}
	txSum := 0
	for i := 0; i < len(data); i++ {
		if data[i][0] > gasPriceGWei {
			txSum += int(data[i][1])
		} else {
			break
		}
	}
	block := int(math.Ceil(float64(txSum) / float64(out.Result.Avgtxnsperblock)))
	t := block * blockInterval
	if t == 0 {
		t = blockInterval
	}
	return fmt.Sprintf("%v", t), nil
}

func getETHGasEstimate(gasInfo string) (string, error) {
	var gasMap map[string]string
	if err := json.Unmarshal([]byte(gasInfo), &gasMap); err != nil {
		return "", err
	}
	url := "https://api.etherscan.io/api"
	gasPrice := gasMap["gas_price"]
	params := map[string]string{
		"module":   "gastracker",
		"action":   "gasestimate",
		"gasprice": gasPrice,
		"apikey":   "CT5GUMRVZMMB94IZ34SNWSI5MEBPBXPPIK",
	}
	out := &types.ETHGasEstimate{}
	err := utils.HttpsGetForm(url, params, out)
	if err != nil {
		return "", err
	}
	if out.Status != "1" {
		return "", errors.New(out.Result)
	}
	return out.Result, nil
}

func GetERCType(chain, token string) string {
	platform := newPlatform(chain)
	if platform == nil {
		c.log.Error("get platform is nil")
		return ""
	}
	evmPlatform := ethereum.Platform2EVMPlatform(platform)
	if platform == nil {
		c.log.Error("sui platform is nil")
		return ""
	}
	return evmPlatform.GetERCType(token)
}

func ParseDataByABI(chain, contract, data string) *types.ParseDataResponse {
	if chain == "Aptos" || chain == "AptosTEST" {
		return nil
	}
	if data[:2] == "0x" {
		data = data[2:]
	}
	abiJson, err := GetContractABI(chain, contract, data[:8])
	if err != nil {
		c.log.Error("get contract abi error:", err)
		return nil
	}
	abiByte, _ := json.Marshal(abiJson)
	contractABI, err := abi.JSON(strings.NewReader(string(abiByte)))
	if err != nil {
		c.log.Error("abi json error:", err)
		return nil
	}
	methodSigData, _ := hex.DecodeString(data[:8])
	inputsSigData, _ := hex.DecodeString(data[8:])
	method, err := contractABI.MethodById(methodSigData)
	if err != nil {
		c.log.Error("method by id error:", err)
		return nil
	}
	inputsMap := make(map[string]interface{})
	if err := method.Inputs.UnpackIntoMap(inputsMap, inputsSigData); err != nil {
		c.log.Error("method UnpackIntoMap error:", err)
		return nil
	}
	abi := types.AbiDecodeResult{
		MethodName: method.Name,
		InputArgs:  inputsMap,
		Selector:   method.Sig,
	}
	transactionType := "contract"
	if abi.MethodName == "transfer" || abi.MethodName == "approve" {
		transactionType = abi.MethodName
	} else if abi.MethodName == SETAPPROVALFORALL {
		transactionType = APPROVENFT
	} else {
		transactionType = "contract"
	}
	return &types.ParseDataResponse{
		TransactionType: transactionType,
		DesData:         abi,
	}
}

func GetContractABI(chain, contract, methodId string) (interface{}, error) {
	//if strings.HasPrefix(strings.ToLower(chain), "evm") {
	//	//自定义链
	//	return utils2.GetDefaultAbiList("EVM"), nil
	//}
	//if methodId[:2] == "0x" {
	//	methodId = methodId[2:]
	//}
	var abiResultList []interface{}
	methodKey := fmt.Sprintf("contract_abi:methodId:%v", methodId)
	abiResult, _ := utils2.GetRedisClient().Get(methodKey).Result()
	if abiResult != "" && abiResult != "[]" {
		json.Unmarshal([]byte(abiResult), &abiResultList)
		return abiResultList, nil
	}
	//
	chainConfig := c.chainInfo[chain]
	if chainConfig.Type == "" {
		return abiResultList, nil
	}
	contractAddress := contract
	if chainConfig.Type == "EVM" || chainConfig.Type == "APTOS" {
		contractAddress = strings.ToLower(contract)
	}
	cacheKey := fmt.Sprintf("contract_abi:%v:%v", chain, contractAddress)
	cacheData, err := utils2.GetRedisClient().Get(cacheKey).Result()
	fmt.Println("cacheData==", cacheData)
	if cacheData != "" {
		//var cacheDataList []interface{}
		//json.Unmarshal([]byte(cacheData), &cacheDataList)
		return abiResultList, nil
	}
	//get default abi
	//defaultAbiList := utils2.GetDefaultAbiList(chainConfig.Type)
	abiURL := utils2.GetBlockExplorerApiURL(chain)
	if abiURL == "" {
		return abiResultList, nil
	}
	headers := make(map[string]string)
	if chain == "OEC" {
		headers["X-Apikey"] = "LWIzMWUtNDU0Ny05Mjk5LWI2ZDA3Yjc2MzFhYmEyYzkwM2NjfDI3OTg5NTk1NDc0ODQ4MTM="
	}
	var out map[string]interface{}
	abiURL = fmt.Sprintf(abiURL, contract)
	fmt.Println("abiURL=", abiURL)
	body, err := utils2.HttpsGetByte(abiURL, nil, headers)
	if err != nil {
		return abiResultList, err
	}
	if chain != "Aptos" && chain != "AptosTEST" && chain != "zkSync" {
		json.Unmarshal(body, &out)
	}
	var rawAbiList []interface{}
	if chain == "Aptos" || chain == "AptosTEST" {
		var aptosOut types.AptosABI
		json.Unmarshal(body, &aptosOut)
		if len(aptosOut) == 0 {
			return abiResultList, err
		}
		//abi := aptosOut[0]
		abiMap := make(map[string]interface{})
		for _, abi := range aptosOut {
			for _, ef := range abi.Abi.ExposedFunctions {
				if ef.IsEntry {
					fullName := fmt.Sprintf("%v::%v::%v", abi.Abi.Address, abi.Abi.Name, ef.Name)
					abiMap[fullName] = ef
					//set redis
					aptosKey := fmt.Sprintf("contract_abi:methodId:%v", fullName)
					aptosData, _ := utils2.GetRedisClient().Get(aptosKey).Result()
					if aptosData == "" || aptosData == "[]" {
						aptosDataList := make([]interface{}, 0, 1)
						//efStr, _ := JsonEncode(ef)
						aptosDataList = append(aptosDataList, ef)
						aptosDataListRedis, _ := JsonEncode(aptosDataList)
						err = utils2.GetRedisClient().Set(aptosKey, aptosDataListRedis, -1).Err()
						if err != nil {
							c.log.Error("set method redis error:", err)
						}
						if aptosKey == methodKey {
							abiResultList = aptosDataList
						}
					}
				}
			}
		}

		if len(abiMap) > 0 {
			abiByte, _ := json.Marshal(abiMap)
			if err := utils2.GetRedisClient().Set(cacheKey, string(abiByte), -1).Err(); err != nil {
				fmt.Println("set aptos redis error:", err)
			}
		}
		return abiResultList, nil
	} else if chain == "OEC" {
		var tempMap map[string]interface{}
		b, _ := json.Marshal(out["methodId"])
		json.Unmarshal(b, &tempMap)
		dec := json.NewDecoder(strings.NewReader(tempMap["contractAbi"].(string)))
		if err := dec.Decode(&rawAbiList); err != nil {
			c.log.Error("decode error:", err)
		}
	} else if chain == "Klaytn" {
		var tempABI types.KlaytnABI
		b, _ := json.Marshal(out["result"])
		json.Unmarshal(b, &tempABI)
		dec := json.NewDecoder(strings.NewReader(tempABI.MatchedContract.ContractAbi))
		if err := dec.Decode(&rawAbiList); err != nil {
			c.log.Error("decode error:", err)
		}

	} else if chain == "TRX" || chain == "TRXTEST" {
		var tempMap map[string]interface{}
		b, _ := json.Marshal(out["data"])
		json.Unmarshal(b, &tempMap)
		var tempABI map[string][]interface{}
		if value, ok := tempMap["abi"]; ok {
			json.Unmarshal([]byte(value.(string)), &tempABI)
		}
		rawAbiList = tempABI["entrys"]
	} else if chain == "Ronin" {
		var tempABI types.RoninABI
		b, _ := json.Marshal(out["result"])
		json.Unmarshal(b, &tempABI)
		rawAbiList = tempABI.Output.ABI
	} else if chain == "Conflux" {
		var tempMap map[string]interface{}
		b, _ := json.Marshal(out["result"])
		json.Unmarshal(b, &tempMap)
		if value, ok := tempMap["abi"]; ok {
			dec := json.NewDecoder(strings.NewReader(value.(string)))
			if err := dec.Decode(&rawAbiList); err != nil {
				c.log.Error("decode error:", err)
			}
		}
	} else if chain == "zkSync" {
		var zkSyncResp types.ZkSyncABIInfo
		json.Unmarshal(body, &zkSyncResp)
		rawAbiList = zkSyncResp.Info.VerificationInfo.Artifacts.Abi
	} else {
		dec := json.NewDecoder(strings.NewReader(out["result"].(string)))
		if err := dec.Decode(&rawAbiList); err != nil {
			c.log.Error("decode error:", err)
		}
	}
	//if len(rawAbiList) == 0 {
	//	//by method id
	//	if methodId[:2] == "0x" {
	//		methodId = methodId[2:]
	//	}
	//	key := fmt.Sprintf("contract_abi:methodId:%v", methodId[:8])
	//	abiResult, _ := utils2.GetRedisClient().Get(key).Result()
	//	if abiResult != "" {
	//		var abiResultList []interface{}
	//		json.Unmarshal([]byte(abiResult), &abiResultList)
	//		return abiResultList, nil
	//	}
	//	defaultAbiListRedis, _ := json.Marshal(defaultAbiList)
	//	utils2.GetRedisClient().Set(cacheKey, string(defaultAbiListRedis), 24*3600)
	//	return defaultAbiList, nil
	//}
	if len(rawAbiList) > 0 {
		var result []interface{}
		for _, value := range rawAbiList {
			b, _ := json.Marshal(value)
			var abiMethod types.ABIMethod
			json.Unmarshal(b, &abiMethod)
			if strings.ToLower(abiMethod.Type) == "function" {
				result = append(result, string(b))
				method := abiMethod.Name + "("
				for i, m := range abiMethod.Inputs {
					method += m.Type
					if i != len(abiMethod.Inputs)-1 {
						method += ","
					}
				}
				method += ")"
				ret := crypto.Keccak256([]byte(method))
				key := fmt.Sprintf("contract_abi:methodId:%v", hex.EncodeToString(ret)[:8])
				methodRedisData, _ := utils2.GetRedisClient().Get(key).Result()
				if methodRedisData == "" || methodRedisData == "[]" {
					methodABIList := make([]interface{}, 0, 1)
					methodABIList = append(methodABIList, value)
					methodABIListRedis, _ := json.Marshal(methodABIList)
					err = utils2.GetRedisClient().Set(key, string(methodABIListRedis), -1).Err()
					if err != nil {
						c.log.Error("set method redis error:", err)
					}
					if key == methodKey {
						abiResultList = methodABIList
					}
				}
				//else {
				//	//var methodABIList []string
				//	json.Unmarshal([]byte(methodRedisData), &methodABIList)
				//	isRetry := true
				//	for _, methodABI := range methodABIList {
				//		methodABIByte, _ := json.Marshal(methodABI)
				//		if string(methodABIByte) == string(b) {
				//			isRetry = false
				//		}
				//	}
				//	if isRetry {
				//		fmt.Println("key===", key)
				//		methodABIList = append(methodABIList, value)
				//	}
				//
				//}

			}
		}
		if len(result) > 0 {
			fmt.Println("abiList=", result)
			redisData, _ := json.Marshal(result)
			err = utils2.GetRedisClient().Set(cacheKey, string(redisData), -1).Err()
			if err != nil {
				c.log.Error("set redis error:", err)
			}
		}
	}

	return abiResultList, nil
}

func JsonEncode(source interface{}) (string, error) {
	bytesBuffer := &bytes.Buffer{}
	encoder := json.NewEncoder(bytesBuffer)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(source)
	if err != nil {
		return "", err
	}

	jsons := string(bytesBuffer.Bytes())
	tsjsons := strings.TrimSuffix(jsons, "\n")
	return tsjsons, nil
}

func GetPretreatmentAmount(chain, from, to, data, value string) map[string][]interface{} {
	transactionInfo := map[string]interface{}{
		"from": from,
		"to":   to,
		"data": data,
		//"value": value,
	}
	if value != "" {
		transactionInfo["value"] = value
	}
	chainId := c.chainInfo[chain].ChainId
	pretreatmentParams := map[string]interface{}{
		"args": map[string]interface{}{
			"method":      "eth_sendTransaction",
			"transaction": transactionInfo,
		},
		"chainId":  fmt.Sprintf("0x%x", chainId),
		"signer":   from,
		"hostname": "pancake.web3app.vip",
		"id":       time.Now().String(),
	}

	url := "https://api.pocketnode.app/request"
	out := &types.PretreatmentResponse{}
	if err := utils2.CommHttpsForm(url, "POST", nil, nil, pretreatmentParams, out); err != nil {
		c.log.Error("GetPretreatmentAmount error:", err)
		return nil
	}
	result := make(map[string][]interface{})
	for _, assetChanges := range out.AssetChanges {
		var actionInfo []string
		var key string
		if strings.Contains(assetChanges.Action, "-") {
			action := strings.Split(assetChanges.Action, "-")[1]
			if strings.Contains(action, " ") {
				actionInfo = strings.Split(strings.Split(assetChanges.Action, "-")[1], " ")
			}
			key = "subAmount"
		} else if strings.Contains(assetChanges.Action, "+") {
			key = "addAmount"
			if strings.Contains(strings.Split(assetChanges.Action, "+")[1], " ") {
				actionInfo = strings.Split(strings.Split(assetChanges.Action, "+")[1], " ")
			}
		} else {
			continue
		}
		amountInfo := make(map[string]interface{})
		if len(actionInfo) > 1 {
			amountInfo["amount"] = actionInfo[0]
			amountInfo["symbol"] = actionInfo[1]
		} else if len(actionInfo) == 1 {
			amountInfo["amount"] = actionInfo[0]
			amountInfo["symbol"] = actionInfo[0]
		}
		amountInfo["name"] = assetChanges.Metadata.Name
		if strings.Contains(assetChanges.Metadata.URL, "token/") {
			amountInfo["token"] = strings.Split(assetChanges.Metadata.URL, "token/")[1]
		}
		result[key] = append(result[key], amountInfo)
	}
	return result
}

func IsContractAddress(chain, address string) (bool, error) {
	platform := newPlatform(chain)
	if platform == nil {
		c.log.Error("get platform is nil")
		return false, errors.New("platform is nil")
	}
	return platform.IsContractAddress(address)
}
