package bitcoin

import (
	"errors"
	"gitlab.bixin.com/mili/node-proxy/pkg/platform/types"
	"gitlab.bixin.com/mili/node-proxy/pkg/platform/utils"
)

type ChainClient struct {
	url string
}

func (b ChainClient) GetBalance( address string) (string, error) {
	u, err := buildURL(b.url+"/api/v2/get_address_balance/BTC/"+address, nil)
	if err != nil {
		return "", err
	}
	var chainBalance types.ChainBalance
	err = getResponse(u, &chainBalance)
	if err != nil {
		return "", err
	}
	if chainBalance.Status == "fail" {
		return "", errors.New("request blockchain failed")
	}
	btcValue := utils.Clean(chainBalance.Data.ConfirmedBalance)
	return btcValue, nil
}
