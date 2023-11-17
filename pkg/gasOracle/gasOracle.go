package gasOracle

import (
	"encoding/json"
	"fmt"
	"github.com/go-kratos/kratos/v2/log"
	"gitlab.bixin.com/mili/node-proxy/pkg/utils"
	"strconv"
	"time"
)

const (
	REDIS_KEY_DATA      = "data"
	REDIS_KEY_TIMESTAMP = "timestamp"
)

type config struct {
	log *log.Helper
}

var gasOracleConfig config

func InitGasOracle(logger log.Logger) {
	log := log.NewHelper(log.With(logger, "module", "gasOracle/InitGasOracle"))
	gasOracleConfig.log = log
}

func GetGasOracle(key string, cacheTime int64) string {
	proxyURL := utils.GetGasOracleURL(key)
	if proxyURL == "" {
		return ""
	}
	cacheKey := fmt.Sprintf("subscribe:proxy:%v", key)
	data, isUpdate, err := getRedisData(cacheKey, cacheTime)
	if err != nil {
		gasOracleConfig.log.Error("get redis data error:", err)
		return ""
	}
	if data == "" {
		resultData, err := getByKey(key)
		if err != nil {
			gasOracleConfig.log.Error("getByKey error:", err)
		}
		return resultData
	}
	if isUpdate {
		go func() {
			//update gas oracle
			getByKey(key)
		}()
	}
	return data
}

func getByKey(key string) (string, error) {
	proxyURL := utils.GetGasOracleURL(key)
	if proxyURL == "" {
		return "", nil
	}
	dataByte, err := utils.HttpsGetByte(proxyURL, nil, nil)
	if err != nil {
		gasOracleConfig.log.Error("get bytesError:", err)
		return "", err
	}
	var resultData *utils.GasOracleRes
	switch key {
	case "feeOracleBTC", "feeOracleLTC", "feeOracleDOGE":
		var tempData map[string]interface{}
		if err := json.Unmarshal(dataByte, &tempData); err != nil {
			fmt.Println("zql===1")
			return "", err
		}
		resultData = &utils.GasOracleRes{
			SafeGasPrice:    parseMapData(tempData, "low_fee_per_kb"),
			ProposeGasPrice: parseMapData(tempData, "medium_fee_per_kb"),
			FastGasPrice:    parseMapData(tempData, "high_fee_per_kb"),
			SuggestBaseFee:  "",
		}
	case "gasOracleOkex":
		var tempData utils.GasOracleOkex
		if err := json.Unmarshal(dataByte, &tempData); err != nil {
			return "", err
		}
		resultData = &utils.GasOracleRes{
			SafeGasPrice:    parseMapData(tempData.Data, "slow"),
			ProposeGasPrice: parseMapData(tempData.Data, "average"),
			FastGasPrice:    parseMapData(tempData.Data, "fast"),
			SuggestBaseFee:  "",
		}
	case "gasOracleETH", "gasOracleHeco", "gasOracleBsc", "gasOraclePolygon",
		"gasOracleFantom", "gasOracleAvalanche":
		var tempData utils.GasOracleResult
		if err := json.Unmarshal(dataByte, &tempData); err != nil {
			return "", err
		}
		resultData = &utils.GasOracleRes{
			SafeGasPrice:    parseMapData(tempData.Result, "SafeGasPrice"),
			ProposeGasPrice: parseMapData(tempData.Result, "ProposeGasPrice"),
			FastGasPrice:    parseMapData(tempData.Result, "FastGasPrice"),
			SuggestBaseFee:  parseMapData(tempData.Result, "suggestBaseFee"),
		}
	case "gasOraclexDai":
		var tempData utils.GasOracleResult
		if err := json.Unmarshal(dataByte, &tempData); err != nil {
			return "", err
		}
		resultData = &utils.GasOracleRes{
			SafeGasPrice:    parseMapData(tempData.Result, "standardgaspricegwei"),
			ProposeGasPrice: parseMapData(tempData.Result, "fastgaspricegwei"),
			FastGasPrice:    parseMapData(tempData.Result, "rapidgaspricegwei"),
			SuggestBaseFee:  "",
		}
	}
	if resultData != nil {
		resultByte, err := json.Marshal(resultData)
		if err != nil {
			return "", err
		}
		cacheKey := fmt.Sprintf("subscribe:proxy:%v", key)
		if err := setRedisData(cacheKey, string(resultByte)); err != nil {
			return "", err
		}
		return string(resultByte), nil
	}
	return "", nil

}

func parseMapData(data map[string]interface{}, key string) string {
	if value, ok := data[key]; ok {
		if valueInt, okInt := value.(int); okInt && valueInt == 0 {
			return ""
		}
		return fmt.Sprintf("%v", value)
	}
	return ""
}

func getRedisData(key string, cacheTime int64) (string, bool, error) {
	result, err := utils.GetRedisClient().HGetAll(key).Result()
	if err != nil || len(result) == 0 {
		return "", true, err
	}
	flag := true
	data := result[REDIS_KEY_DATA]
	val := result[REDIS_KEY_TIMESTAMP]
	timestamp, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return data, flag, err
	}
	if time.Now().Unix()-timestamp < cacheTime {
		flag = false
	}
	return data, flag, nil
}

func setRedisData(key, data string) error {
	fields := map[string]interface{}{
		REDIS_KEY_DATA:      data,
		REDIS_KEY_TIMESTAMP: time.Now().Unix(),
	}
	return utils.GetRedisClient().HMSet(key, fields).Err()
}

func AutoUpdateGasOracle() {
	//"定期刷新 blockChain Gas"
	handlerGasOracle()
	transactionPlan := time.NewTicker(1 * time.Minute)
	//uc.repo.AutoUpdateTokenList(ctx)
	for true {
		select {
		case <-transactionPlan.C:
			go handlerGasOracle()
		}
	}
}

func handlerGasOracle() {
	gasOracleConfig.log.Infof("定期刷新 blockChain Gas")
	gasOracleConf := utils.GetGasOracleConfig()
	if gasOracleConf == nil {
		return
	}
	for _, conf := range gasOracleConf {
		GetGasOracle(conf.GetKey(), conf.GetCacheTime())
	}
}
