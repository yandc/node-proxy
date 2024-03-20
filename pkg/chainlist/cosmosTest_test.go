package chainlist

import (
	"fmt"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"os"
	"strings"
	"testing"
)

func TestGetTestCosmosChainList(t *testing.T) {
	logger := log.With(log.NewStdLogger(os.Stdout),
		"ts", log.DefaultTimestamp,
		"caller", log.DefaultCaller,

		"trace.id", tracing.TraceID(),
		"span.id", tracing.SpanID(),
	)
	GetTestCosmosChainList(log.NewHelper(logger), nil)
}

func TestCheckCosmosChainId(t *testing.T) {
	a := "entrypointtestnet"
	b := strings.Replace(a, "testnet", "", -1) + "TEST"
	//b := strings.Split(a, "testnet")[0] + "TEST"
	fmt.Println("b=", b)
}
