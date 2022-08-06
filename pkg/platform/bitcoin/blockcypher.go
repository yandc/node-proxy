package bitcoin

import (
	"gitlab.bixin.com/mili/node-proxy/pkg/platform/types"
	"gitlab.bixin.com/mili/node-proxy/pkg/platform/utils"
)

type BlockCypherClient struct {
	url string
}

func (b BlockCypherClient) GetBalance(address string) (string, error) {
	u, err := buildURL(b.url+"/addrs/"+address+"/balance", nil)
	var balance types.BlockCypherBalance
	err = getResponse(u, &balance)
	if err != nil {
		return "", err
	}
	btcValue := utils.BigIntString(&balance.Balance, 8)
	return btcValue, nil
}
