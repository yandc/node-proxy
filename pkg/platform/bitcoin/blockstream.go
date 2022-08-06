package bitcoin

import (
	"gitlab.bixin.com/mili/node-proxy/pkg/platform/types"
	"gitlab.bixin.com/mili/node-proxy/pkg/platform/utils"
)

type BlockStreamClient struct {
	url string
}

func (b BlockStreamClient) GetBalance(address string) (string, error) {
	u, err := buildURL(b.url+"/api/address/"+address, nil)
	if err != nil {
		return "", err
	}
	var addr types.BlockStreamBalance
	err = getResponse(u, &addr)
	if err != nil {
		return "",err
	}
	//余额=收入的数量-支出的数量=xxx BTC
	balance := addr.ChainStats.FundedTxoSum.Sub(&addr.ChainStats.FundedTxoSum, &addr.ChainStats.SpentTxoSum)
	btcValue := utils.BigIntString(balance, 8)
	return btcValue, nil
}
