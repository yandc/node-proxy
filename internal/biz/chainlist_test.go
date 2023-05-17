package biz

import (
	"context"
	"github.com/ethereum/go-ethereum/ethclient"
	"testing"
	"time"
)

func TestConnection(t *testing.T) {
	client, err := ethclient.Dial("https://eth.llamarpc.com")
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second*10)
	defer cancelFunc()

	chainID, err := client.ChainID(ctx)
	if err != nil {

		t.Fatal(err)
	}

	t.Log(chainID.Uint64())
}
