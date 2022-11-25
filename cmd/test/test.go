package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-redis/redis"
	v12 "gitlab.bixin.com/mili/node-proxy/api/nft/v1"
	pb "gitlab.bixin.com/mili/node-proxy/api/platform/v1"
	v1 "gitlab.bixin.com/mili/node-proxy/api/tokenlist/v1"
	"gitlab.bixin.com/mili/node-proxy/internal/conf"
	"gitlab.bixin.com/mili/node-proxy/internal/data"
	"gitlab.bixin.com/mili/node-proxy/pkg/nft"
	"gitlab.bixin.com/mili/node-proxy/pkg/nft/collection"
	"gitlab.bixin.com/mili/node-proxy/pkg/nft/list"
	"gitlab.bixin.com/mili/node-proxy/pkg/token-list/tokenlist"
	"google.golang.org/grpc"
	"gorm.io/gorm"
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
	id, _    = os.Hostname()
	db       *gorm.DB
	client   *redis.Client
	logger   log.Logger
	bc       conf.Bootstrap
)

func init() {

	flag.StringVar(&flagconf, "conf", "../../configs", "config path, eg: -conf config.yaml")
	flag.StringVar(&testFunc, "name", "price", "test func name")

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
}

func main() {
	flag.Parse()
	fmt.Println("func name", testFunc)
	Init()
	switch testFunc {
	case "price":
		testGetPrice()
	case "balance":
		testGetBalance()
	case "tokenList":
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
	}
	fmt.Println("test main end")
}

func testRefreshLogoURI() {
	tokenlist.InitTokenList(bc.TokenList, db, client, logger)
	tokenlist.RefreshLogoURI("nervos")
}

func testAutoUpdateTokenList() {
	tokenlist.AutoUpdateTokenList(false, false, true)
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
		CoinNames: "ethereum",
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
	chains := []string{"solana"}
	for _, chain := range chains {
		tokenlist.UpdateEVMDecimasl(chain)
	}

}

func testUpdateDecimalsByChain() {
	chains := []string{"solana"}
	for _, chain := range chains {
		tokenlist.UpdateDecimalsByChain(chain)
	}

}

func testUpLoadTokenList() {
	tokenlist.InitTokenList(bc.TokenList, db, client, logger)
	tokenlist.UpLoadJsonToCDN([]string{"arbitrum-one"})
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
	req.Chain = "ETH"
	a := map[string]string{
		"gas_price": "8000000000",
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
	tokenlist.UpdateChainToken("nervos")
}

func testGetTop20TokenList() {
	tokenlist.InitTokenList(bc.TokenList, db, client, logger)
	result, err := tokenlist.GetTop20TokenList("ETH")
	if err != nil {
		fmt.Println("error=", err)
	}
	fmt.Println("result===", result)
}

func testCreateNFTList() {
	logger := log.With(log.NewStdLogger(os.Stdout),
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

	var bc conf.Bootstrap
	if err := c.Scan(&bc); err != nil {
		panic(err)
	}
	db := data.NewDB(bc.Data, logger)
	nft.InitNFT(db, logger, bc.NftList)
	list.CreateNFTList("ETHGoerliTEST")
}

func testCreateNFTCollection() {
	logger := log.With(log.NewStdLogger(os.Stdout),
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

	var bc conf.Bootstrap
	if err := c.Scan(&bc); err != nil {
		panic(err)
	}
	db := data.NewDB(bc.Data, logger)
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
	req.Chain = "ETH"
	tokenInfo := []*v12.GetNftInfoRequest_NftInfo{
		{TokenAddress: "0x495f947276749ce646f68ac8c248420045cb7b5e", TokenId: "7913402202769379533690164279743878593095549349620263384589938601384149516289"},
		{TokenAddress: "0xc8ff927b56d617ea04976f5d5f77383cf72712d3", TokenId: "907"},
		{TokenAddress: "0xc8ff927b56d617ea04976f5d5f77383cf72712d3", TokenId: "908"},
	}
	req.NftInfo = tokenInfo
	resp, err := p.GetNftInfo(context.Background(), req)
	if err != nil {
		fmt.Println("get balacne error", err)
	}
	fmt.Println("result:", resp)
}
