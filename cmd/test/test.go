package main

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-redis/redis"
	v13 "gitlab.bixin.com/mili/node-proxy/api/commRPC/v1"
	v12 "gitlab.bixin.com/mili/node-proxy/api/nft/v1"
	pb "gitlab.bixin.com/mili/node-proxy/api/platform/v1"
	v1 "gitlab.bixin.com/mili/node-proxy/api/tokenlist/v1"
	"gitlab.bixin.com/mili/node-proxy/internal/conf"
	"gitlab.bixin.com/mili/node-proxy/internal/data"
	"gitlab.bixin.com/mili/node-proxy/internal/data/models"
	"gitlab.bixin.com/mili/node-proxy/pkg/chainlist"
	"gitlab.bixin.com/mili/node-proxy/pkg/nft"
	"gitlab.bixin.com/mili/node-proxy/pkg/nft/collection"
	"gitlab.bixin.com/mili/node-proxy/pkg/nft/list"
	"gitlab.bixin.com/mili/node-proxy/pkg/platform"
	"gitlab.bixin.com/mili/node-proxy/pkg/platform/types"
	"gitlab.bixin.com/mili/node-proxy/pkg/token-list/tokenlist"
	"gitlab.bixin.com/mili/node-proxy/pkg/utils"
	"google.golang.org/grpc"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"
)

// go build -ldflags "-X main.Version=x.y.z"
var (
	// Name is the name of the compiled software.
	Name string
	// Version is the version of the compiled software.
	Version string
	// flagconf is the config flag.
	flagconf string
	testFunc string
	cgIds    string
	chainIds []string
	id, _    = os.Hostname()
	db       *gorm.DB
	client   *redis.Client
	logger   log.Logger
	bc       conf.Bootstrap
)

func init() {

	flag.StringVar(&flagconf, "conf", "./../../configs", "config path, eg: -conf config.yaml")
	flag.StringVar(&testFunc, "name", "price", "test func name")
	flag.StringVar(&cgIds, "cgIds", "", "coingecko id list(comma separation)")

}

func Init() {
	logger = log.With(log.NewStdLogger(os.Stdout),
		"ts", log.DefaultTimestamp,
		"caller", log.DefaultCaller,
		"service.id", id,
		"service.name", Name,
		"service.version", Version,
		"trace.id", tracing.TraceID(),
		"span.id", tracing.SpanID(),
	)
	c := config.New(
		config.WithSource(
			file.NewSource(flagconf),
		),
	)
	defer c.Close()

	if err := c.Load(); err != nil {
		panic(err)
	}

	if err := c.Scan(&bc); err != nil {
		panic(err)
	}
	db = data.NewDB(bc.Data, logger)
	client = data.NewRedis(bc.Data)
	data.NewMarketClient(bc.TokenList)
	utils.SetRedisClient(client)
	fmt.Println("map==", bc.ChainData.TxHashErr)
}

func main() {
	flag.Parse()
	fmt.Println("func name", testFunc)
	Init()
	chainIds = strings.Split(cgIds, ",")
	switch testFunc {
	case "price":
		testGetPrice()
	case "balance":
		testGetBalance()
	case "updateTokenList":
		testAutoUpdateTokenList()
	case "tokenInfo":
		testGetTokenInfo()
	case "chainDecimals":
		testUpdateDecimalsByChain()
	case "tokenListCDN":
		testUpLoadTokenList()
	case "buildRequest":
		testBuildRequest()
	case "refreshLogo":
		testRefreshLogoURI()
	case "refreshDir":
		testRefreshDirs()
	case "EVMDecimals":
		testUpdateEVMDecimals()
	case "uploadImage":
		testUpLoadLocalImage()
	case "gasEstimate":
		testGetGasEstimate()
	case "top20List":
		testGetTop20TokenList()
	case "chainTokenList":
		testUpdateChainList()
	case "nftList":
		testCreateNFTList()
	case "nftInfo":
		testGetNFTInfo()
	case "nftCollection":
		testCreateNFTCollection()
	case "commRPC":
		testCommRPC()
	case "tokenTop20":
		testInsertTokenTop20()
	case "tokenPrice":
		testAutoUpdateTokenPrice()
	case "deleteToken":
		testDeleteTokenList()
	case "updateTokenPrice":
		TestUpdatePriceByChain()
	case "abi":
		TestGetContractAbi()
	case "migrateABI":
		TestResetContractABI()
	case "parseABI":
		TestParseDataByABI()
	case "cgList":
		TestCreateCGList()
	case "getPriceKey":
		TestInitGetPriceKey()
	case "logoByAddress":
		testRefreshLogoURIByAddress()
	case "parseABIData":
		TestParseContractABI()
	case "uploadChainList":
		TestUploadChainList()
	case "parseTokenInfo":
		TestParseTokenInfo()
	default:
		TestGetContractAbi()
	}
	fmt.Println("test main end")
}

func testRefreshLogoURIByAddress() {
	tokenlist.InitTokenList(bc.TokenList, db, client, logger)
	tokenlist.RefreshLogoURIByAddress("ethereum", []string{"0x64bc2ca1be492be7185faa2c8835d9b824c8a194"})
	//tokenlist.RefreshLogoURI([]string{"ethereum", "huobi-token", "okex-chain", "binance-smart-chain", "polygon-pos", "fantom",
	//	"avalanche", "cronos", "arbitrum-one", "klay-token", "aurora", "optimistic-ethereum",
	//	"oasis", "tron", "xdai", "solana", "starcoin", "ethereum-classic", "aptos", "nervos", "osmosis",
	//	"bitcoin-cash", "harmony-shard-0", "ronin", "arbitrum-nova"})
}

func testRefreshLogoURI() {
	tokenlist.InitTokenList(bc.TokenList, db, client, logger)
	fmt.Println("cgIds=", cgIds)
	//ids := strings.Split(cgIds, ",")
	if len(chainIds) == 0 {
		fmt.Println("ids length is nil.")
		return
	}
	tokenlist.RefreshLogoURI(chainIds)
	//tokenlist.RefreshLogoURI([]string{"ethereum", "huobi-token", "okex-chain", "binance-smart-chain", "polygon-pos", "fantom",
	//	"avalanche", "cronos", "arbitrum-one", "klay-token", "aurora", "optimistic-ethereum",
	//	"oasis", "tron", "xdai", "solana", "starcoin", "ethereum-classic", "aptos", "nervos", "osmosis",
	//	"bitcoin-cash", "harmony-shard-0", "ronin", "arbitrum-nova"})
}

func testAutoUpdateTokenList() {
	tokenlist.InitTokenList(bc.TokenList, db, client, logger)
	//tokenlist.AutoUpdatePrice()
	tokenlist.UpdateTokenListByMarket()
}

func testCommRPC() {
	conn, err := grpc.Dial("127.0.0.1:9001", grpc.WithInsecure())
	if err != nil {
		fmt.Println("error:", err)
	}
	defer conn.Close()
	p := v13.NewCommRPCClient(conn)
	//
	req := new(v13.ExecNodeProxyRPCRequest)
	req.Id = 1
	params := utils.GetPriceV2Req{
		CoinName: []string{"ethereum"},
		CoinAddress: []string{"aptos_0x5e156f1207d0ebfa19a9eeff00d62a282278fb8719f4fab3a586a0a2c0fffbea::coin::T",
			"aptos_0x8d87a65ba30e09357fa2edea2c80dbac296e5dec2b18287113500b902942929d::celer_coin_manager::UsdcCoin",
			"aptos_0xec42a352cc65eca17a9fa85d0fc602295897ed6b8b8af6a6c79ef490eb8f9eba::amm_swap::PoolLiquidityCoin<0x1::aptos_coin::AptosCoin, 0x5e156f1207d0ebfa19a9eeff00d62a282278fb8719f4fab3a586a0a2c0fffbea::coin::T>"},
		Currency: "CNY",
	}
	b, _ := json.Marshal(params)
	fmt.Println("params ==", string(b))
	req.Params = string(b)
	req.Method = "GetPriceV2"
	//req.Chain = "ETH"
	//req.Address = "0xa06ef134313C13e03B8682B0616147607B4E375E"
	//req.TokenAddress = "0xdAC17F958D2ee523a2206206994597C13D831ec7"
	//req.Decimals = "6"
	resp, err := p.ExecNodeProxyRPC(context.Background(), req)
	if err != nil {
		fmt.Println("get balacne error", err)
	}
	fmt.Println("result:", resp)
}

func testGetBalance() {
	conn, err := grpc.Dial("127.0.0.1:9001", grpc.WithInsecure())
	if err != nil {
		fmt.Println("error:", err)
	}
	defer conn.Close()
	p := pb.NewPlatformClient(conn)
	req := new(pb.GetBalanceRequest)
	req.Chain = "ETH"
	req.Address = "0xa06ef134313C13e03B8682B0616147607B4E375E"
	req.TokenAddress = "0xdAC17F958D2ee523a2206206994597C13D831ec7"
	req.Decimals = "6"
	resp, err := p.GetBalance(context.Background(), req)
	if err != nil {
		fmt.Println("get balacne error", err)
	}
	fmt.Println("result:", resp.Balance)
}

func testBuildRequest() {
	conn, err := grpc.Dial("127.0.0.1:9001", grpc.WithInsecure())
	if err != nil {
		fmt.Println("error:", err)
	}
	defer conn.Close()
	p := pb.NewPlatformClient(conn)
	reqs := make([]*pb.BuildWasmRequestRequest, 0, 4)
	params := []interface{}{"0x3b43ac14079565246aeed15da656809eddcc79ab"}
	b, _ := json.Marshal(params)
	reqs = append(reqs, &pb.BuildWasmRequestRequest{
		Chain:        "MYSTEN",
		NodeRpc:      "https://gateway.devnet.sui.io:443",
		FunctionName: "sui_getObjectsOwnedByAddress",
		Params:       string(b),
	})
	for _, req := range reqs {
		result, err := p.BuildWasmRequest(context.Background(), req)
		if err != nil {

		}
		fmt.Println(result)
		resp, err := ExecHttps(result.Url, result.Method, result.Body, result.Head)
		if err != nil {

		}
		fmt.Println("resp===", resp)
		coinTypeMap := map[string]string{
			"coinType": "0x2::coin::Coin<0x2::sui::SUI>",
		}
		c, _ := json.Marshal(coinTypeMap)
		list, err := p.AnalysisWasmResponse(context.Background(), &pb.AnalysisWasmResponseRequest{
			Chain:        "MYSTEN",
			FunctionName: "objectId",
			Params:       string(c),
			Response:     resp,
		})
		var objectList []string
		json.Unmarshal([]byte(list.Data), &objectList)
		fmt.Println("list====", objectList, len(objectList))
	}

}

func ExecHttps(url, method, body string, heads map[string]string) (string, error) {
	req, err := http.NewRequest(method, url, strings.NewReader(body))
	if err != nil {
		return "", err
	}
	for key, value := range heads {
		req.Header.Set(key, value)
	}
	client := http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(respBody), nil
}

func testGetTokenInfo() {
	conn, err := grpc.Dial("127.0.0.1:9001", grpc.WithInsecure())
	if err != nil {
		fmt.Println("error:", err)
	}
	defer conn.Close()
	c := v1.NewTokenlistClient(conn)
	data := []*v1.GetTokenInfoReq_Data{
		{Chain: "ETH", Address: "0x31903E333809897eE57Af57567f4377a1a78756c"},
		{Chain: "ETH", Address: "0x0000000DE40dfa9B17854cBC7869D80f9F98D823"},
		{Chain: "Aptos", Address: "0xf22bede237a07e121b56d91a491eb7bcdfd1f5907926a9e58338f964a01b17fa::asset::WETH"},
	}
	req := &v1.GetTokenInfoReq{
		Data: data,
	}
	resp, err := c.GetTokenInfo(context.Background(), req)
	if err != nil {
		fmt.Println("error:", err)
	}
	fmt.Println(resp)
}

func testGetPrice() {
	conn, err := grpc.Dial("127.0.0.1:9001", grpc.WithInsecure())
	if err != nil {
		fmt.Println("error:", err)
	}
	defer conn.Close()
	c := v1.NewTokenlistClient(conn)
	reqs := make([]*v1.PriceReq, 0, 10)

	//coin name
	reqs = append(reqs, &v1.PriceReq{
		CoinNames: "dogecoin,ethereum,ethereum,matic-network,nervos-network,klay-token,fantom,aptos,tron,casper-network,ethereum-classic,oec-token,bitcoin,Ethereum,crypto-com-chain,xdai,litecoin,starcoin,solana,huobi-token,avalanche-2,cosmos,binancecoin",
		Currency:  "USD",
	})
	//coin name and coin address
	reqs = append(reqs, &v1.PriceReq{
		CoinNames:     "ethereum",
		Currency:      "USD",
		CoinAddresses: "ethereum_0x31903E333809897eE57Af57567f4377a1a78756c,ethereum_0xdac17f958d2ee523a2206206994597c13d831ec7,ethereum_0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48",
	})

	//coin name and coin address
	reqs = append(reqs, &v1.PriceReq{
		CoinNames:     "ethereum",
		Currency:      "USD",
		CoinAddresses: "ethereum_0x31903E333809897eE57Af57567f4377a1a78756c",
	})

	//coin name and coin address
	reqs = append(reqs, &v1.PriceReq{
		CoinNames:     "ethereum,huobi-token",
		Currency:      "USD",
		CoinAddresses: "ethereum_0x31903E333809897eE57Af57567f4377a1a78756c,ethereum_0x31903e333809897ee57af57567f4377a1a78756c",
	})

	reqs = append(reqs, &v1.PriceReq{
		CoinNames:     "ethereum,huobi-token",
		Currency:      "USD",
		CoinAddresses: "ethereum_0x1a986F1659e11E2AE7CC6543F307bAE5cDe1C761,ethereum_0x1a986f1659e11e2ae7cc6543f307bae5cde1c761",
	})

	reqs = append(reqs, &v1.PriceReq{
		CoinNames:     "tron",
		Currency:      "usd",
		CoinAddresses: "tron_TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t",
	})
	for _, req := range reqs {
		resp, err := c.GetPrice(context.Background(), req)
		if err != nil {
			fmt.Println("get price error:", err)
		}
		fmt.Println("resp==", resp)
	}
}

func testUpdateEVMDecimals() {
	tokenlist.InitTokenList(bc.TokenList, db, client, logger)
	if len(chainIds) == 0 {
		fmt.Println("chain is nil.")
		return
	}
	//chains := []string{"solana"}
	for _, chain := range chainIds {
		tokenlist.UpdateEVMDecimasl(chain)
	}

}

func testUpdateDecimalsByChain() {
	tokenlist.InitTokenList(bc.TokenList, db, client, logger)
	if len(chainIds) == 0 {
		fmt.Println("chain is nil.")
		return
	}
	//chains := []string{"solana"}
	for _, chain := range chainIds {
		tokenlist.UpdateDecimalsByChain(chain)
	}

}

func testUpLoadTokenList() {
	utils.InitConfig(bc)
	tokenlist.InitTokenList(bc.TokenList, db, client, logger)
	fmt.Println("cgIds=", cgIds)
	//ids := strings.Split(cgIds, ",")
	if len(chainIds) == 0 {
		fmt.Println("ids length is nil.")
		return
	}
	tokenlist.UpLoadJsonToCDN(chainIds)
	//tokenlist.UpLoadJsonToCDN([]string{"ethereum", "huobi-token", "okex-chain", "binance-smart-chain", "polygon-pos", "fantom",
	//	"avalanche", "cronos", "arbitrum-one", "klay-token", "aurora", "optimistic-ethereum",
	//	"oasis", "tron", "xdai", "solana", "starcoin", "ethereum-classic", "aptos", "nervos", "osmosis",
	//	"bitcoin-cash", "harmony-shard-0", "ronin", "arbitrum-nova"})
}

func testRefreshDirs() {
	tokenlist.InitTokenList(bc.TokenList, db, client, logger)
	tokenlist.RefreshCDNDirs()
}

func testUpLoadLocalImage() {
	tokenlist.InitTokenList(bc.TokenList, db, client, logger)
	images := []string{"images/nervos/nervos_0xcce6d0ac83d2491f8b8bd3875f3577b9e77f08a0396cd2e9274f339eb76e08a4.svg"}
	for _, image := range images {
		tokenlist.UpLoadLocalImages(image)
	}
}

func testGetGasEstimate() {
	conn, err := grpc.Dial("127.0.0.1:9001", grpc.WithInsecure())
	if err != nil {
		fmt.Println("error:", err)
	}
	defer conn.Close()
	p := pb.NewPlatformClient(conn)
	req := new(pb.GetGasEstimateRequest)
	req.Chain = "Fantom"
	a := map[string]string{
		"gas_price": "345825200000",
	}
	b, _ := json.Marshal(a)
	req.GasInfo = string(b)
	fmt.Println("req", req)
	resp, err := p.GetGasEstimate(context.Background(), req)
	if err != nil {
		fmt.Println("get balacne error", err)
	}
	fmt.Println("result:", resp)
}

func testUpdateChainList() {
	tokenlist.InitTokenList(bc.TokenList, db, client, logger)
	if len(chainIds) == 0 {
		fmt.Println("testUpdateChainList id is nil")
	}
	for _, chain := range chainIds {
		tokenlist.UpdateChainToken(chain)
	}

}

func testGetTop20TokenList() {
	tokenlist.InitTokenList(bc.TokenList, db, client, logger)
	result, err := tokenlist.GetTopNTokenList("ETH", 20)
	if err != nil {
		fmt.Println("error=", err)
	}
	fmt.Println("result===", result)
}

func testCreateNFTList() {
	db := data.NewDB(bc.Data, logger)
	nft.InitNFT(db, logger, bc.NftList)
	list.CreateNFTList("ETHGoerliTEST")
}

func testCreateNFTCollection() {
	nft.InitNFT(db, logger, bc.NftList)
	collection.CreateCollectionList()
}

func testGetNFTInfo() {
	conn, err := grpc.Dial("127.0.0.1:9001", grpc.WithInsecure())
	if err != nil {
		fmt.Println("error:", err)
	}
	defer conn.Close()
	p := v12.NewNftClient(conn)

	req := new(v12.GetNftInfoRequest)
	req.Chain = "BSC"
	tokenInfo := []*v12.GetNftInfoRequest_NftInfo{
		{TokenAddress: "0x85f0e02cb992aa1f9f47112f815f519ef1a59e2d", TokenId: "10005032506"},
	}
	req.NftInfo = tokenInfo
	resp, err := p.GetNftInfo(context.Background(), req)
	if err != nil {
		fmt.Println("get balacne error", err)
	}
	fmt.Println("result:", resp)
}

func testInsertTokenTop20() {
	tokenlist.InitTokenList(bc.TokenList, db, client, logger)
	tokenlist.UpdateTokenTop20()
}

func testAutoUpdateTokenPrice() {
	tokenlist.InitTokenList(bc.TokenList, db, client, logger)
	tokenlist.AutoUpdatePrice()
}

func testDeleteTokenList() {
	tokenlist.InitTokenList(bc.TokenList, db, client, logger)
	tokenlist.DeleteJsonCDN()
}

func TestUpdatePriceByChain() {
	tokenlist.InitTokenList(bc.TokenList, db, client, logger)
	if len(chainIds) == 0 {
		fmt.Println("testUpdateChainList id is nil")
		return
	}
	tokenlist.UpdatePriceByChains(chainIds, []string{"ethereum"})
}

func TestGetContractAbi() {
	platform.InitPlatform(bc.Platform, logger, client)
	utils.InitConfig(bc)
	ret, err := platform.GetContractABI("TRX", "TDgrSuii9e7HLfY1DhEBxkcFa3vrLgS3Gx", "3805550f")
	if err != nil {
		fmt.Println("error==", err)
	}
	fmt.Println("ret==", ret)
}

func TestResetContractABI() {
	keys, _ := client.Keys("contract_abi:*").Result()
	for _, key := range keys {
		if !strings.HasPrefix(key, "contract_abi:methodId:") {
			cacheData, err := client.Get(key).Result()
			if err != nil {
				fmt.Println("redis get error:", err)
			}
			if cacheData != "" {
				if strings.Contains(key, "Aptos") || strings.Contains(key, "AptosTEST") {
					var aptosCacheData map[string]interface{}
					if err := json.Unmarshal([]byte(cacheData), &aptosCacheData); err != nil {
						fmt.Println("unmarshal error:", err)
					}
					for fullName, ef := range aptosCacheData {
						aptosKey := fmt.Sprintf("contract_abi:methodId:%v", fullName)
						aptosData, _ := client.Get(aptosKey).Result()
						if aptosData == "" || aptosData == "[]" {
							aptosDataList := make([]interface{}, 0, 1)
							aptosDataList = append(aptosDataList, ef)
							aptosDataListRedis, _ := jsonEncode(aptosDataList)
							err = client.Set(aptosKey, aptosDataListRedis, -1).Err()
							if err != nil {
								fmt.Println("set method redis error:", err)
							}
						}
					}
				} else {
					var evmCacheData []interface{}
					if err := json.Unmarshal([]byte(cacheData), &evmCacheData); err != nil {
						fmt.Println("json unmarshal error:", err)
					}
					for _, value := range evmCacheData {
						var abiMethod types.ABIMethod
						valueByte, _ := json.Marshal(value)
						if err := json.Unmarshal(valueByte, &abiMethod); err != nil {
							fmt.Println("json unmarshal error:", err)
						}
						if strings.ToLower(abiMethod.Type) == "function" {
							method := abiMethod.Name + "("
							for i, m := range abiMethod.Inputs {
								method += m.Type
								if i != len(abiMethod.Inputs)-1 {
									method += ","
								}
							}
							method += ")"
							ret := crypto.Keccak256([]byte(method))
							methodIdkey := fmt.Sprintf("contract_abi:methodId:%v", hex.EncodeToString(ret)[:8])
							methodRedisData, _ := client.Get(methodIdkey).Result()
							if methodRedisData == "" || methodRedisData == "[]" {
								methodABIList := make([]interface{}, 0, 1)
								methodABIList = append(methodABIList, value)
								methodABIListRedis, _ := json.Marshal(methodABIList)
								err = client.Set(methodIdkey, string(methodABIListRedis), -1).Err()
								if err != nil {
									fmt.Println("set method redis error:", err)
								}
							}
						}
					}
				}
			}
		}
	}
}

func jsonEncode(source interface{}) (string, error) {
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

func TestParseDataByABI() {
	platform.InitPlatform(bc.Platform, logger, client)
	utils.InitConfig(bc)
	ret := platform.ParseDataByABI("ETH", "0xdac17f958d2ee523a2206206994597c13d831ec7", "0xa9059cbb000000000000000000000000aaf75fe0b77b6c3d3de14554becf25d55414b19b0000000000000000000000000000000000000000000000000000000008d568f3")
	fmt.Println("ret==", ret)
}

func TestCreateCGList() {
	tokenlist.InitTokenList(bc.TokenList, db, client, logger)
	tokenlist.CreateCoinGeckoList()
}

func TestInitGetPriceKey() {
	chainlist.InitChainList(db, client, logger)
	blockChainLists, err := chainlist.GetAllBlockChain()
	if err != nil {
		fmt.Println("error=", err)
		return
	}
	fmt.Println("blockChainLists=", len(blockChainLists))
	for index, blockChain := range blockChainLists {
		if blockChain.GetPriceKey == "" {
			var tempCoinGeckoList models.CoinGeckoList
			db.Where("symbol = ?", strings.ToLower(blockChain.CurrencySymbol)).First(&tempCoinGeckoList)
			if tempCoinGeckoList.CgId != "" {
				blockChain.GetPriceKey = tempCoinGeckoList.CgId
				blockChainLists[index] = blockChain
				db.Model(&models.BlockChain{}).Where("id = ?", blockChain.ID).Update("get_price_key", tempCoinGeckoList.CgId)
			}
		}
	}
}

func TestParseContractABI() {
	parseKey := `"type":"Function"`
	newStr := `"type":"function"`
	keys, _ := client.Keys("contract_abi:methodId:*").Result()
	for _, key := range keys {
		//if key == "contract_abi:methodId:3805550f" {
		cacheData, err := client.Get(key).Result()
		if err != nil {
			fmt.Println("redis get error:", err)
			continue
		}
		if strings.Contains(cacheData, parseKey) {
			//fmt.Println("key==", key, "zql====", cacheData)
			newCacheData := strings.Replace(cacheData, parseKey, newStr, -1)
			if err := client.Set(key, newCacheData, -1).Err(); err != nil {
				fmt.Println("redis set error:", err)
			}
		}
		//}

	}
}

func TestUploadChainList() {
	//platform.InitPlatform(bc.Platform, logger, client)
	utils.InitConfig(bc)
	chainlist.InitChainList(db, client, logger)
	chainlist.UpLoadChainList2CDN()
}

func TestParseTokenInfo() {
	keys, _ := client.Keys("tokenlist:tokeninfo:*").Result()
	fmt.Println("keys length:", len(keys))
	dbLimit := 5000
	dbTokenInfoList := make([]*models.TokenInfo, 0, dbLimit)
	for _, key := range keys {
		tokenInfo, err := client.Get(key).Result()
		if err != nil {
			fmt.Println("redis get error:", err)
		}
		if tokenInfo != "" {
			var dbTokenInfo *models.TokenInfo
			json.Unmarshal([]byte(tokenInfo), &dbTokenInfo)
			dbTokenInfoList = append(dbTokenInfoList, dbTokenInfo)
		}
		if len(dbTokenInfoList) > dbLimit {
			if err := db.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "address"}, {Name: "chain"}},
				DoNothing: true,
			}).Create(&dbTokenInfoList).Error; err != nil {
				fmt.Println("TestParseTokenInfo create db error:", err)
			}
			dbTokenInfoList = make([]*models.TokenInfo, 0, dbLimit)
		}

	}
	if len(dbTokenInfoList) > 0 {
		if err := db.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "address"}, {Name: "chain"}},
			DoNothing: true,
		}).Create(&dbTokenInfoList).Error; err != nil {
			fmt.Println("TestParseTokenInfo create db error:", err)
		}
	}

}
