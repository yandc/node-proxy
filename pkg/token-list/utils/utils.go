package utils

import (
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis"
	v12 "gitlab.bixin.com/mili/node-proxy/api/tokenlist/v1"
	"gitlab.bixin.com/mili/node-proxy/pkg/utils"
	"io/ioutil"
	"math/big"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rpc"
	"gitlab.bixin.com/mili/node-proxy/pkg/token-list/types"
)

const (
	REDIS_PRICE_PRICE         = "price"
	REDIS_PRICE_TIMESTAMP     = "timestamp"
	REDIS_TOKENLIST_TOKENLIST = "tokenlist"
	REDIS_TOKENLIST_TIMESTAMP = "timestamp"
	REDIS_PRICE_INTERVAL      = 60
	REDIS_TOKENLIST_INTERVAL  = 86400 //24H
	STARCOIN_CHAIN            = "starcoin"
	APTOS_CHAIN               = "aptos"
	COSMOS_CHAIN              = "Cosmos"
	REDIS_TOKENLIST_TOP20     = "tokenTop20"
)

var platformMap = map[string]string{
	"aurora":      "aurora",
	"aurora-near": "aurora",

	"binance-smart-chain": "binance-smart-chain",
	"bnb":                 "binance-smart-chain",

	"bitgert":       "bitgert",
	"bitrise-token": "bitgert",

	"boba":         "boba",
	"boba-network": "boba",

	"conflux":         "conflux",
	"conflux-network": "conflux",

	"cube":         "cube",
	"cube-network": "cube",

	"elrond":      "elrond",
	"elrond-egld": "elrond",

	"fuse":         "fuse",
	"fuse-network": "fuse",

	"fusion-network": "fusion-network",
	"fusion":         "fusion-network",

	"harmony-shard-0": "harmony-shard-0",
	"harmony":         "harmony-shard-0",

	"klay-token": "klay-token",
	"klaytn":     "klay-token",

	"kucoin-community-chain": "kucoin-community-chain",
	"kucoin-token":           "kucoin-community-chain",

	"meter":            "meter",
	"meter-governance": "meter",

	"metis-andromeda": "metis-andromeda",
	"metisdao":        "metis-andromeda",

	"oasis":         "oasis",
	"oasis-network": "oasis",

	"okex-chain": "okex-chain",
	"okt":        "okex-chain",

	"polkadot":     "polkadot",
	"polkadot-new": "polkadot",

	"polygon-pos": "polygon-pos",
	"polygon":     "polygon-pos",

	"shiden network": "shiden network",
	"shiden-network": "shiden network",

	"terra":      "terra",
	"terra-luna": "terra",
}

var CMCNameChainMap = map[string]string{
	"Milkomeda":               "milkomeda-cardano",
	"BNB Beacon Chain (BEP2)": "binancecoin",
}

var CGCNameChainMap = map[string]string{
	"defi-kingdoms-blockchain": "avalanche",
}

var handlerNameMap = map[string]string{
	"ethereum":  "ethereum",
	"heco":      "huobi-token",
	"okex":      "okex-chain",
	"bsc":       "binance-smart-chain",
	"polygon":   "polygon-pos",
	"fantom":    "fantom",
	"avalanche": "avalanche",
	"cronos":    "cronos",
	"arbitrum":  "arbitrum-one",
	"klaytn":    "klay-token",
	"aurora":    "aurora",
	"optimism":  "optimistic-ethereum",
	"oasis":     "oasis",
	"tron":      "tron",
	"xDai":      "xdai",
	"ETC":       "ethereum-classic",
	"solana":    "solana",
	"aptos":     "aptos",
	"nervos":    "nervos",
	"cosmos":    "cosmos",
}

var chainNameMap = map[string]string{
	"ETH":       "ethereum",
	"HECO":      "huobi-token",
	"OEC":       "okex-chain",
	"BSC":       "binance-smart-chain",
	"Polygon":   "polygon-pos",
	"Fantom":    "fantom",
	"Avalanche": "avalanche",
	"Cronos":    "cronos",
	"Arbitrum":  "arbitrum-one",
	"Klaytn":    "klay-token",
	"Aurora":    "aurora",
	"Optimism":  "optimistic-ethereum",
	"Oasis":     "oasis",
	"TRX":       "tron",
	"STC":       "starcoin",
	"xDai":      "xdai",
	"ETC":       "ethereum-classic",
	"Solana":    "solana",
	"Aptos":     "aptos",
	"Nervos":    "nervos",
	"Cosmos":    "cosmos",

	"ETHTEST":       "ethereum",
	"HECOTEST":      "huobi-token",
	"OECTEST":       "okex-chain",
	"BSCTEST":       "binance-smart-chain",
	"PolygonTEST":   "polygon-pos",
	"FantomTEST":    "fantom",
	"AvalancheTEST": "avalanche",
	"CronosTEST":    "cronos",
	"arbitrum":      "arbitrum-one",
	"KlaytnTEST":    "klay-token",
	"AuroraTEST":    "aurora",
	"OptimismTEST":  "optimistic-ethereum",
	"OasisTEST":     "oasis",
	"TRXTEST":       "tron",
	"STCTEST":       "starcoin",
	"xDaiTEST":      "xdai",
	"ETCTEST":       "ethereum-classic",
	"SolanaTEST":    "solana",
	"AptosTEST":     "aptos",
	"NervosTEST":    "nervos",
	"CosmosTEST":    "cosmos",
}

var db2Chain = map[string]string{
	"ethereum":            "ETH",
	"huobi-token":         "HECO",
	"okex-chain":          "OEC",
	"binance-smart-chain": "BSC",
	"polygon-pos":         "Polygon",
	"fantom":              "Fantom",
	"avalanche":           "Avalanche",
	"cronos":              "Cronos",
	"arbitrum-one":        "Arbitrum",
	"klay-token":          "Klaytn",
	"aurora":              "Aurora",
	"optimistic-ethereum": "Optimism",
	"oasis":               "Oasis",
	"tron":                "TRX",
	"starcoin":            "STC",
	"xdai":                "xDai",
	"ethereum-classic":    "ETC",
	"solana":              "Solana",
	"aptos":               "Aptos",
	"nervos":              "Nervos",
	//"cosmos":              "Cosmos",
}

var TokenFileMap = map[string][]string{
	"ethereum":            {"https://api.coinmarketcap.com/data-api/v3/uniswap/all.json", "https://bxhp.243096.com/mili/tokens/eth.json"},
	"binance-smart-chain": {"https://tokens.pancakeswap.finance/coingecko.json", "https://bxhp.243096.com/mili/tokens/bsc.json"},
	"huobi-token":         {"https://bxhp.243096.com/mili/tokens/heco.json"},
	"okex-chain":          {"https://static.kswap.finance/tokenlist/kswap-hosted-list.json", "https://bxhp.243096.com/mili/tokens/okex.json"},
	"fantom":              {"https://tokens.coingecko.com/fantom/all.json"},
	"polygon-pos":         {"https://bxhp.243096.com/mili/tokens/matic.json", "https://tokens.coingecko.com/polygon-pos/all.json"},
	"avalanche":           {"https://raw.githubusercontent.com/pangolindex/tokenlists/main/pangolin.tokenlist.json", "https://tokens.coingecko.com/avalanche/all.json"},
	"cronos":              {"https://swap.crodex.app/tokens.json", "https://tokens.coingecko.com/cronos/all.json"},
	"arbitrum-one":        {"https://tokens.coingecko.com/arbitrum-one/all.json"},
	"aurora":              {"https://tokens.coingecko.com/aurora/all.json"},
	"optimistic-ethereum": {"https://tokens.coingecko.com/optimistic-ethereum/all.json"},
	"oasis":               {"https://tokens.coingecko.com/oasis/all.json"},
	"tron":                {"https://list.justswap.link/justswap.json"},
	"starcoin":            {"https://bxhp.243096.com/mili/tokens/stc.json"},
}

var chainURLMap = map[string]string{
	"ethereum":            "https://mainnet.infura.io/v3/9aa3d95b3bc440fa88ea12eaa4456161",
	"polygon-pos":         "https://rpc.ankr.com/polygon",
	"oasis":               "https://emerald.oasis.dev",
	"binance-smart-chain": "https://rpc.ankr.com/bsc",
	"okex-chain":          "https://okchain.mytokenpocket.vip/",
	"avalanche":           "https://rpc.ankr.com/avalanche",
	"fantom":              "https://rpc.ankr.com/fantom",
	"klay-token":          "https://klaytn05.fandom.finance/",
	"huobi-token":         "https://http-mainnet-node.defibox.com",
	"cronos":              "https://rpc.artemisone.org/cronos",
	"xdai":                "https://rpc.ankr.com/gnosis",
	"ethereum-classic":    "https://etc.mytokenpocket.vip",
}

var OtherTokenFileMap = map[string][]string{
	"klay-token": {"https://s.klayswap.com/data/klayswap/tokens.json"},
}

func GetPlatform(chain string) string {
	if value, ok := platformMap[chain]; ok {
		return value
	}
	return chain
}

func ParseTokenListFile() map[string][]types.TokenInfo {
	result := make(map[string][]types.TokenInfo, len(TokenFileMap)+len(OtherTokenFileMap))
	for chain, urls := range TokenFileMap {
		tokenInfos := make([]types.TokenInfo, 0, 260)
		for _, url := range urls {
			out := &types.Token{}
			HttpsGetForm(url, nil, out)
			for _, t := range out.Tokens {
				if strings.HasPrefix(t.Address, "0x") && chain != STARCOIN_CHAIN {
					t.Address = strings.ToLower(t.Address)
				}
				tokenInfos = append(tokenInfos, t)
			}
		}
		result[chain] = tokenInfos
	}

	for chain, urls := range OtherTokenFileMap {
		switch chain {
		case "klay-token":
			tokenInfo := PraseKlaytnFile(urls)
			result[chain] = tokenInfo
		}
	}
	return result
}

func PraseKlaytnFile(urls []string) []types.TokenInfo {
	result := make([]types.TokenInfo, 0, 260)
	for _, url := range urls {
		out := map[string]types.KlaytnTokenInfo{}
		err := HttpsGetForm(url, nil, &out)
		if err != nil {
			fmt.Println("error:", err)
		}
		for _, t := range out {
			if strings.HasPrefix(t.Address, "0x") {
				t.Address = strings.ToLower(t.Address)
			}
			result = append(result, types.TokenInfo{
				ChainId:  t.Id,
				Name:     t.NameEn,
				Symbol:   t.Symbol,
				Address:  t.Address,
				Decimals: t.Decimals,
				LogoURI:  t.Icon,
			})
		}
	}
	return result
}

//key:chain;value []tokenAddress
func GetDecimalsByMap(noDecimals map[string][]string) map[string]int {
	result := make(map[string]int)
	pageSize := 500
	for chain, tokenAddress := range noDecimals {
		var decimals map[string]int
		if chain == "tron" {
			decimals = GetTronBatchDecimals(chain, tokenAddress)
		} else if chain == "solana" {
			decimals = GetSolanaBatchDecimals(chain, tokenAddress)
		} else {
			endIndex := 0
			for i := 0; i < len(tokenAddress); i += pageSize {
				if i+pageSize > len(tokenAddress) {
					endIndex = len(tokenAddress)
				} else {
					endIndex = i + pageSize
				}
				decimals = GetBatchDecimals(chain, tokenAddress[i:endIndex])
			}
			for key, decimal := range decimals {
				if _, ok := result[key]; !ok {
					result[key] = decimal
				}
			}
		}
	}
	return result
}

func GetBatchDecimals(chain string, tokens []string) map[string]int {
	result := make(map[string]int)
	balanceFun := []byte("decimals()")
	hash := crypto.NewKeccakState()
	hash.Write(balanceFun)
	methodID := hash.Sum(nil)[:4]
	var url string
	if value, ok := chainURLMap[chain]; ok {
		url = value
	} else {
		return result
	}
	rpcClient, _ := rpc.DialHTTP(url)
	var tokenAddrs []string
	var be []rpc.BatchElem
	for _, token := range tokens {
		var data []byte
		data = append(data, methodID...)
		tokenAddress := common.HexToAddress(token)

		callMsg := map[string]interface{}{
			"to":   tokenAddress,
			"data": hexutil.Bytes(data),
		}
		be = append(be, rpc.BatchElem{
			Method: "eth_call",
			Args:   []interface{}{callMsg, "latest"},
			Result: new(string),
		})
		tokenAddrs = append(tokenAddrs, token)
	}
	fmt.Println("GetBatchDecimals:", chain, len(tokens))
	rpcClient.BatchCall(be)
	for index, b := range be {
		hexAmount := b.Result.(*string)
		bi := new(big.Int)
		bi.SetBytes(common.FromHex(*hexAmount))
		result[chain+":"+tokenAddrs[index]] = int(bi.Int64())
	}
	return result
}

func GetSolanaBatchDecimals(chain string, tokens []string) map[string]int {
	result := make(map[string]int, len(tokens))
	for _, token := range tokens {
		decimal, err := GetSolanaDecimal(token)
		if err != nil && strings.Contains(err.Error(), "Too many requests") {
			time.Sleep(1 * time.Minute)
			for i := 0; err != nil && strings.Contains(err.Error(), "Too many requests") && i < 3; i++ {
				decimal, err = GetSolanaDecimal(token)
				time.Sleep(1 * time.Minute)
			}
		}
		if err != nil {
			continue
		}
		result[chain+":"+token] = decimal
	}
	return result
}

func GetTronBatchDecimals(chain string, tokens []string) map[string]int {
	result := make(map[string]int, len(tokens))
	for _, token := range tokens {
		decimal, err := GetTronDecimals(token)
		if err != nil {
			continue
		}
		result[chain+":"+token] = decimal
	}
	return result
}

func GetSolanaDecimal(token string) (int, error) {
	var err error
	//url := "https://api.mainnet-beta.solana.com"
	urls := []string{"https://solana-api.projectserum.com", "https://api.mainnet-beta.solana.com",
		"https://ssc-dao.genesysgo.net"}
	for _, url := range urls {
		method := "getTokenSupply"
		params := []interface{}{token}
		out := &types.SolanaTokenInfo{}
		err = utils.JsonHttpsPost(url, 1, method, "2.0", out, params)
		if err != nil {
			continue
		}
		return out.Value.Decimals, nil
	}
	return -1, err
}

func GetDecimalsByChain(chain, token string) (int, error) {
	switch chain {
	case "tron":
		if len(token) > 30 && len(token) < 40 {
			return GetTronDecimals(token)
		}
	case "solana":
		return GetSolanaDecimal(token)
	}
	return 0, nil
}

func GetTronDecimals(token string) (int, error) {
	url := "https://apilist.tronscan.org/api/contract"
	out := &types.TronTokenInfo{}
	params := map[string]string{
		"contract": token,
	}
	err := HttpsGetForm(url, params, out)
	if err != nil {
		return 0, err
	}
	if len(out.Data) == 0 {
		return 0, err
	}

	return out.Data[0].TokenInfo.TokenDecimal, nil
}

func HttpsGetForm(url string, params map[string]string, out interface{}) error {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	q := req.URL.Query()
	for k, v := range params {
		q.Add(k, v)
	}
	req.URL.RawQuery = q.Encode()
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if 200 != resp.StatusCode {
		return fmt.Errorf("%s", body)
	}
	if err := json.Unmarshal(body, &out); err != nil {
		return err
	}
	return nil
}

//var globalTransport *http.Transport
//func init() {
//	uu, _ := url.Parse("http://127.0.0.1:1087")
//	globalTransport = &http.Transport{
//		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
//		Proxy:           http.ProxyURL(uu),
//	}
//}

func GetChainNameByPlatform(handler string) string {
	if value, ok := handlerNameMap[handler]; ok {
		return value
	}
	return handler
}

func GetChainNameByChain(chain string) string {
	if value, ok := chainNameMap[chain]; ok {
		return value
	}
	return chain
}

func DownLoad(base string, url string) error {
	client := &http.Client{
		//Transport: utils.GlobalTransport,
	}
	v, err := client.Get(url)
	if err != nil {
		return err
	}
	defer v.Body.Close()
	content, err := ioutil.ReadAll(v.Body)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(base, content, 0666)
	if err != nil {
		return err
	}
	return nil
}

// 判断所给路径文件/文件夹是否存在
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	//isnotexist来判断，是不是不存在的错误
	if os.IsNotExist(err) { //如果返回的错误类型使用os.isNotExist()判断为true，说明文件或者文件夹不存在
		return false, nil
	}
	return false, err //如果有错误了，但是不是不存在的错误，所以把这个错误原封不动的返回
}

func ParseCoinAddress(coinAddress []string) map[string][]string {
	result := make(map[string][]string, len(coinAddress))
	for _, chainAddress := range coinAddress {
		if !strings.Contains(chainAddress, "_") {
			result[chainAddress] = append(result[chainAddress], chainAddress)
			continue
		}
		addressInfo := strings.Split(chainAddress, "_")
		chain := GetChainNameByPlatform(addressInfo[0])
		address := addressInfo[1]
		if strings.HasPrefix(address, "0x") && chain != STARCOIN_CHAIN {
			address = strings.ToLower(address)
		}
		key := fmt.Sprintf("%s_%s", chain, address)
		result[key] = append(result[key], chainAddress)
	}
	return result
}

// GetPriceRedisValueByKey get price,whether update
func GetPriceRedisValueByKey(redisClient *redis.Client, key string) (string, bool, error) {
	result, err := redisClient.HGetAll(key).Result()
	if err != nil || len(result) == 0 {
		return "", true, err
	}
	flag := true
	price := result[REDIS_PRICE_PRICE]
	val := result[REDIS_PRICE_TIMESTAMP]
	timestamp, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return price, flag, err
	}
	if time.Now().Unix()-timestamp < REDIS_PRICE_INTERVAL {
		flag = false
	}
	return price, flag, nil
}

func SetPriceRedisKey(redisClient *redis.Client, key, price string) error {
	fields := map[string]interface{}{
		REDIS_PRICE_PRICE:     price,
		REDIS_PRICE_TIMESTAMP: time.Now().Unix(),
	}
	return redisClient.HMSet(key, fields).Err()
}

// GetTokenListRedisValueByKey get token list,whether update
func GetTokenListRedisValueByKey(redisClient *redis.Client, key string) (string, bool, error) {
	result, err := redisClient.HGetAll(key).Result()
	if err != nil || len(result) == 0 {
		return "", true, err
	}
	flag := true
	tokenList := result[REDIS_TOKENLIST_TOKENLIST]
	val := result[REDIS_TOKENLIST_TIMESTAMP]
	timestamp, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return tokenList, flag, err
	}
	if time.Now().Unix()-timestamp < REDIS_TOKENLIST_INTERVAL {
		flag = false
	}
	return tokenList, flag, nil
}

func SetTokenListRedisKey(redisClient *redis.Client, key, tokenList string) error {
	fields := map[string]interface{}{
		REDIS_TOKENLIST_TOKENLIST: tokenList,
		REDIS_TOKENLIST_TIMESTAMP: time.Now().Unix(),
	}
	return redisClient.HMSet(key, fields).Err()
}

func GetChainByDBChain(dbChain string) string {
	if value, ok := db2Chain[dbChain]; ok {
		return value
	}
	return ""
}

func GetCDNTokenList(url string) map[string]types.TokenInfoVersion {
	var tokenListVersion []types.TokenInfoVersion
	err := HttpsGetForm(url, nil, &tokenListVersion)
	for i := 0; err != nil && i < 3; i++ {
		time.Sleep(1 * time.Second)
		err = HttpsGetForm(url, nil, &tokenListVersion)
	}
	if err != nil {
		fmt.Println("get cdn token list error:", err)
		return nil
	}
	result := make(map[string]types.TokenInfoVersion)
	for _, info := range tokenListVersion {
		result[info.Chain] = info
	}
	return result
}

func ReadTokenListVersion(fileName string) map[string]types.TokenInfoVersion {
	//fileName := "tokenlist.json"
	exist, _ := PathExists(fileName)
	if !exist {
		return map[string]types.TokenInfoVersion{}
	}
	var tokenListVersion []types.TokenInfoVersion
	file, _ := os.Open(fileName)
	// 关闭文件
	defer file.Close()
	// NewDecoder创建一个从file读取并解码json对象的*Decoder，解码器有自己的缓冲，并可能超前读取部分json数据。
	decoder := json.NewDecoder(file)
	//Decode从输入流读取下一个json编码值并保存在v指向的值里
	err := decoder.Decode(&tokenListVersion)
	if err != nil {
		return nil
	}
	fmt.Println(tokenListVersion)
	result := make(map[string]types.TokenInfoVersion)
	for _, info := range tokenListVersion {
		result[info.Chain] = info
	}
	return result
}

func WriteJsonToFile(fileName string, tokenVersions []types.TokenInfoVersion) error {
	listFile, _ := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0777)
	defer listFile.Close()
	encoder := json.NewEncoder(listFile)
	err := encoder.Encode(tokenVersions)
	return err
}

func GetRedisTokenInfo(redisClient *redis.Client, key string) *v12.GetTokenInfoResp_Data {
	result, err := redisClient.Get(key).Result()
	if err != nil {
		return nil
	}

	if result != "" {
		var tokenInfo *v12.GetTokenInfoResp_Data
		json.Unmarshal([]byte(result), &tokenInfo)
		return tokenInfo
	}
	return nil
}

func SetRedisTokenInfo(redisClient *redis.Client, key string, value *v12.GetTokenInfoResp_Data) error {
	b, _ := json.Marshal(value)
	return redisClient.Set(key, string(b), -1).Err()
}

// GetTokenTop20RedisValueByKey get token list top20,whether update
func GetTokenTop20RedisValueByKey(redisClient *redis.Client, key string) (string, bool, error) {
	result, err := redisClient.HGetAll(key).Result()
	if err != nil || len(result) == 0 {
		return "", true, err
	}
	flag := true
	tokenTop20 := result[REDIS_TOKENLIST_TOP20]
	val := result[REDIS_TOKENLIST_TIMESTAMP]
	timestamp, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return tokenTop20, flag, err
	}
	if time.Now().Unix()-timestamp < REDIS_TOKENLIST_INTERVAL {
		flag = false
	}
	return tokenTop20, flag, nil
}

func SetTokenTop20RedisKey(redisClient *redis.Client, key, tokenTop20 string) error {
	fields := map[string]interface{}{
		REDIS_TOKENLIST_TOP20:     tokenTop20,
		REDIS_TOKENLIST_TIMESTAMP: time.Now().Unix(),
	}
	return redisClient.HMSet(key, fields).Err()
}
