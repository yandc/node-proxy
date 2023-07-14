package utils

var handler2ChainMap = map[string]string{
	"ethereum":     "ETH",
	"heco":         "HECO",
	"okex":         "OEC",
	"bsc":          "BSC",
	"polygon":      "Polygon",
	"fantom":       "Fantom",
	"avalanche":    "Avalanche",
	"cronos":       "Cronos",
	"arbitrum":     "Arbitrum",
	"klaytn":       "Klaytn",
	"aurora":       "Aurora",
	"optimism":     "Optimism",
	"oasis":        "Oasis",
	"tron":         "TRX",
	"xDai":         "xDai",
	"ETC":          "ETC",
	"solana":       "Solana",
	"aptos":        "Aptos",
	"starcoin":     "STC",
	"nervos":       "Nervos",
	"cosmos":       "Cosmos",
	"smartbch":     "SmartBCH",
	"osmosis":      "Osmosis",
	"harmony":      "Harmony",
	"ronin":        "Ronin",
	"arbitrumnova": "ArbitrumNova",
	"conflux":      "Conflux",
	"zksync":       "zkSync",
	"sui":          "SUI",
	"evm210425":    "evm210425",
}

var chain2HandlerMap = map[string]string{
	"ETH":          "ethereum",
	"HECO":         "heco",
	"OEC":          "okex",
	"BSC":          "bsc",
	"Polygon":      "polygon",
	"Fantom":       "fantom",
	"Avalanche":    "avalanche",
	"Cronos":       "cronos",
	"Arbitrum":     "arbitrum",
	"Klaytn":       "klaytn",
	"Aurora":       "aurora",
	"Optimism":     "optimism",
	"Oasis":        "oasis",
	"TRX":          "tron",
	"xDai":         "xDai",
	"ETC":          "ETC",
	"Solana":       "solana",
	"Aptos":        "aptos",
	"STC":          "starcoin",
	"Nervos":       "nervos",
	"Cosmos":       "cosmos",
	"SmartBCH":     "smartbch",
	"Osmosis":      "osmosis",
	"Harmony":      "harmony",
	"Ronin":        "ronin",
	"ArbitrumNova": "arbitrumnova",
	"Conflux":      "conflux",
	"zkSync":       "zksync",
	"SUI":          "sui",
	"evm210425":    "evm210425",
}

func GetChainByHandler(handler string) string {
	if value, ok := handler2ChainMap[handler]; ok {
		return value
	}
	return handler
}

func GetHandlerByChain(chain string) string {
	if value, ok := chain2HandlerMap[chain]; ok {
		return value
	}
	return chain
}
