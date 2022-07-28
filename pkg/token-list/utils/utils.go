package utils

import (
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rpc"
	"io/ioutil"
	"math/big"
	"net/http"
	"node-proxy/pkg/token-list/types"
	"os"
	"strings"
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
			//if err != nil {
			//	fmt.Println("ParseTokenListFile error", err, chain, url)
			//}
			//fmt.Println("chain=", chain, "url=", url, "length:", len(out.Tokens))
			for _, t := range out.Tokens {
				if strings.HasPrefix(t.Address, "0x") {
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

		//fmt.Println("chain=", chain, "url=", url, "length:", len(out))
	}
	//fmt.Println("PraseKlaytnFile length:", len(result))
	return result
}

//key:chain;value []tokenAddress
func GetDecimalsByMap(noDecimals map[string][]string) map[string]int {
	result := make(map[string]int)
	pageSize := 500
	for chain, tokenAddress := range noDecimals {
		endIndex := 0
		for i := 0; i < len(tokenAddress); i += pageSize {
			if i+pageSize > len(tokenAddress) {
				endIndex = len(tokenAddress)
			} else {
				endIndex = i + pageSize
			}
			decimals := GetBatchDecimals(chain, tokenAddress[i:endIndex])
			for key, decimal := range decimals {
				if _, ok := result[key]; !ok {
					result[key] = decimal
				}
			}
		}
	}
	//fmt.Println("GetDecimalsByMap length:", len(result))
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
	//if err != nil {
	//	//log.Info("errorInfo", zap.Any("chain", chain), zap.Any("tokens", tokens))
	//	//fmt.Println("error:", err, chain)
	//}
	for index, b := range be {
		hexAmount := b.Result.(*string)
		bi := new(big.Int)
		bi.SetBytes(common.FromHex(*hexAmount))
		result[chain+":"+tokenAddrs[index]] = bi.Sign()
	}
	//if err != nil {
	//	//log.Info("errorInfo", zap.Any("result", result))
	//}
	return result
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
	//client := &http.Client{Transport: globalTransport}
	resp, err := http.DefaultClient.Do(req)
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
	v, err := http.Get(url)
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

func ParseCoinAddress(coinAddress []string) map[string]string {
	result := make(map[string]string, len(coinAddress))
	for _, chainAddress := range coinAddress {
		if !strings.Contains(chainAddress, "_") {
			result[chainAddress] = chainAddress
			continue
		}
		addressInfo := strings.Split(chainAddress, "_")
		chain := GetChainNameByPlatform(addressInfo[0])
		address := addressInfo[1]
		if strings.HasPrefix(address, "0x") {
			address = strings.ToLower(address)
		}
		key := fmt.Sprintf("%s_%s", chain, address)
		result[key] = chainAddress
	}
	return result
}
