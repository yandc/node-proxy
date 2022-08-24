package main

import (
	"context"
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
	"os"
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
	}
	fmt.Println("test main end")
	//testGetPrice()
	//testGetBalance()
	//tokenlist.AutoUpdateTokenList()
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
	reqs := make([]*v1.PriceReq, 5)
	//coin name
	reqs[0] = &v1.PriceReq{
		CoinNames: "ethereum",
		Currency:  "USD",
	}
	//coin name and coin address
	reqs[1] = &v1.PriceReq{
		CoinNames:     "ethereum",
		Currency:      "USD",
		CoinAddresses: "ethereum_0x31903E333809897eE57Af57567f4377a1a78756c,ethereum_0xdac17f958d2ee523a2206206994597c13d831ec7,ethereum_0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48",
	}

	//coin name and coin address
	reqs[2] = &v1.PriceReq{
		CoinNames:     "ethereum",
		Currency:      "USD",
		CoinAddresses: "ethereum_0x31903E333809897eE57Af57567f4377a1a78756c",
	}

	//coin name and coin address
	reqs[3] = &v1.PriceReq{
		CoinNames:     "ethereum,huobi-token",
		Currency:      "USD",
		CoinAddresses: "ethereum_0x31903E333809897eE57Af57567f4377a1a78756c,ethereum_0x31903e333809897ee57af57567f4377a1a78756c",
	}

	reqs[4] = &v1.PriceReq{
		CoinNames:     "ethereum,huobi-token",
		Currency:      "USD",
		CoinAddresses: "ethereum_0x1a986F1659e11E2AE7CC6543F307bAE5cDe1C761,ethereum_0x1a986f1659e11e2ae7cc6543f307bae5cde1c761",
	}
	for _, req := range reqs {
		resp, err := c.GetPrice(context.Background(), req)
		if err != nil {
			fmt.Println("get price error:", err)
		}
		fmt.Println("resp==", resp)
	}

}
