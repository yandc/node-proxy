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
	pb "gitlab.bixin.com/mili/node-proxy/api/platform/v1"
	v1 "gitlab.bixin.com/mili/node-proxy/api/tokenlist/v1"
	"gitlab.bixin.com/mili/node-proxy/internal/conf"
	"gitlab.bixin.com/mili/node-proxy/internal/data"
	"gitlab.bixin.com/mili/node-proxy/pkg/token-list/tokenlist"
	"google.golang.org/grpc"
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
)

func init() {
	flag.StringVar(&flagconf, "conf", "../../configs", "config path, eg: -conf config.yaml")
	flag.StringVar(&testFunc, "name", "price", "test func name")
}

func main() {
	flag.Parse()
	fmt.Println("func name", testFunc)
	switch testFunc {
	case "price":
		testGetPrice()
	case "balance":
		testGetBalance()
	case "tokenList":
		testAutoUpdateTokenList()
	case "tokenInfo":
		testGetTokenInfo()
	case "tronDecimals":
		testUpdateTronDecimals()
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
	}
	fmt.Println("test main end")
	//testGetPrice()
	//testGetBalance()
	//tokenlist.AutoUpdateTokenList()
}

func testRefreshLogoURI() {
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
	client := data.NewRedis(bc.Data)
	tokenlist.InitTokenList(bc.TokenList, db, client, logger)
	tokenlist.RefreshLogoURI()
}

func testAutoUpdateTokenList() {
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
	client := data.NewRedis(bc.Data)
	tokenlist.InitTokenList(bc.TokenList, db, client, logger)
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
		{Chain: "HECO", Address: "0x0298c2b32eae4da002a15f36fdf7615bea3da047"},
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
	client := data.NewRedis(bc.Data)
	tokenlist.InitTokenList(bc.TokenList, db, client, logger)
	chains := []string{"ethereum-classic", "xdai"}
	for _, chain := range chains {
		tokenlist.UpdateEVMDecimasl(chain)
	}

}

func testUpdateTronDecimals() {
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
	client := data.NewRedis(bc.Data)
	tokenlist.InitTokenList(bc.TokenList, db, client, logger)
	tokenlist.UpdateTronDecimals()
}

func testUpLoadTokenList() {
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
	client := data.NewRedis(bc.Data)
	tokenlist.InitTokenList(bc.TokenList, db, client, logger)
	tokenlist.UpLoadJsonToCDN([]string{"xdai", "ethereum-classic"})
}

func testUpLoadLocalImage() {
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
	client := data.NewRedis(bc.Data)
	tokenlist.InitTokenList(bc.TokenList, db, client, logger)
	images := []string{"images/xdai/xdai_0xf929b6ce804b06a4ce92f5ea3b13fb1141c82368.png"}
	for _, image := range images {
		tokenlist.UpLoadLocalImages(image)
	}

}
