package tron

import (
	"gitlab.bixin.com/mili/node-proxy/pkg/platform/types"
	"gitlab.bixin.com/mili/node-proxy/pkg/platform/utils"
	"math/big"
)

type TronScanClient struct {
	url string
}

func (t TronScanClient) GetBalance(address string) (string, error) {
	url := t.url + "/api/account?address=" + address
	out := &types.TronscanBalances{}
	err := utils.HttpsGetForm(url, nil, out)
	if err != nil {
		return "", err
	}
	trc10Balances := out.Balances
	if len(trc10Balances) == 0 {
		return "", err
	}

	balance := ""
	for _, trc10Balance := range trc10Balances {
		if trc10Balance.TokenId == "_" && trc10Balance.TokenName == "trx" {
			balance = trc10Balance.Balance
			break
		}
	}
	balanceInt, b := new(big.Int).SetString(balance, 10)
	if b {
		balance = utils.BigIntString(balanceInt, 6)
	}
	return balance, nil
}

func (t TronScanClient) GetTokenBalance(address string, tokenAddress string, decimals int) (string, error) {
	url := t.url + "/api/account?address=" + address
	out := &types.TronscanBalances{}
	err := utils.HttpsGetForm(url, nil, out)
	if err != nil {
		return "0", err
	}
	trc20TokenBalances := out.Trc20TokenBalances
	if len(trc20TokenBalances) == 0 {
		return "0", err
	}

	balance := "0"
	for _, trc20TokenBalance := range trc20TokenBalances {
		if trc20TokenBalance.TokenId == tokenAddress {
			balance = trc20TokenBalance.Balance
			break
		}
	}
	balanceInt, b := new(big.Int).SetString(balance, 10)
	if b {
		balance = utils.BigIntString(balanceInt, decimals)
	}
	return balance, nil
}
