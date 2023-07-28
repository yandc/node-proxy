package utils

import (
	"gitlab.bixin.com/mili/node-proxy/internal/conf"
)

var bc conf.Bootstrap

func InitConfig(c conf.Bootstrap) {
	bc = c
	//fmt.Println("abi==", bc.BlockExplorerApi)
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
