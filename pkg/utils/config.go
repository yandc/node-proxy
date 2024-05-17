package utils

import (
	"gitlab.bixin.com/mili/node-proxy/internal/conf"
)

var bc conf.Bootstrap

func InitConfig(c conf.Bootstrap) {
	bc = c
}

func GetBlockExplorerApiURL(chain string) string {
	if value, ok := bc.BlockExplorerApi[chain]; ok {
		return value
	}
	return ""
}

func GetDefaultAbiList(chainType string) string {
	if value, ok := bc.DefaultAbiList[chainType]; ok {
		return value
	}
	return ""
}

func GetGasOracleURL(key string) string {
	if value, ok := bc.GasOracle.Url[key]; ok {
		return value
	}
	return ""
}

func GetGasOracleConfig() []*conf.GasOracleInfoOracleConf {
	return bc.GasOracle.OracleConfig
}

func GetTokenListChains() []string {
	return bc.TokenList.Chains
}

func GetAWSConfig() []*conf.TokenList_AWS {
	return bc.TokenList.Aws
}

func GetQiNiuConfig() *conf.TokenList_QiNiu {
	return bc.TokenList.Qiniu
}

func GetTokenListConfig() *conf.TokenList {
	return bc.TokenList
}
