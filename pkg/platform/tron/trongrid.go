package tron

import (
	"gitlab.bixin.com/mili/node-proxy/pkg/platform/types"
	"gitlab.bixin.com/mili/node-proxy/pkg/platform/utils"
	"math/big"
	"net/http"
)

type TronGridClient struct {
	url string
}

func (t TronGridClient) GetBalance(address string) (string, error) {
	url := t.url + "/wallet/getaccount"
	reqBody := types.TronBalanceReq{
		Address: address,
		Visible: true,
	}
	out := &types.TronBalance{}
	err := utils.HttpsForm(url, http.MethodPost, nil, reqBody, out)
	if err != nil {
		return "", err
	}
	balance := utils.BigIntString(new(big.Int).SetInt64(out.Balance), 6)
	return balance, nil
}

func (t TronGridClient) GetTokenBalance(address string, tokenAddress string, decimals int) (string, error) {
	url := t.url + "/wallet/triggerconstantcontract"
	addrB := Base58ToHex(address)
	parameter := "0000000000000000000000000000000000000000000000000000000000000000"[len(addrB):] + addrB
	out := &types.TronTokenBalanceRes{}
	reqBody := types.TronTokenBalanceReq{
		OwnerAddress:     address,
		ContractAddress:  tokenAddress,
		FunctionSelector: "balanceOf(address)",
		Parameter:        parameter,
		Visible:          true,
	}
	err := utils.HttpsForm(url, http.MethodPost, nil, reqBody, out)
	if err != nil {
		return "0", err
	}
	tokenBalance := "0"
	if len(out.ConstantResult) > 0 {
		banInt, b := new(big.Int).SetString(out.ConstantResult[0], 16)
		if b {
			tokenBalance = utils.BigIntString(banInt, decimals)
		}
	}
	return tokenBalance, err
}
