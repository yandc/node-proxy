package ton

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-kratos/kratos/v2/log"
	v1 "gitlab.bixin.com/mili/node-proxy/api/platform/v1"
	v12 "gitlab.bixin.com/mili/node-proxy/api/tokenlist/v1"
	"gitlab.bixin.com/mili/node-proxy/pkg/platform/types"
	utils2 "gitlab.bixin.com/mili/node-proxy/pkg/platform/utils"
	"gitlab.bixin.com/mili/node-proxy/pkg/utils"
	"math"
	"strings"
)

type platform struct {
	rpcURL []string
	log    *log.Helper
	chain  string
}

const TON_DECIMALS = 9

func (p *platform) GetBalance(ctx context.Context, address, tokenAddress, decimals string) (string, error) {
	return "", nil
}

func (p *platform) BuildWasmRequest(ctx context.Context, nodeRpc, functionName, params string) (*v1.BuildWasmRequestReply, error) {
	tempParams, _ := hex.DecodeString(params)
	var body string
	var err error
	result := &v1.BuildWasmRequestReply{
		Method: "POST",
		Url:    nodeRpc,
		Head: map[string]string{
			"Content-Type": "application/json",
		},
		Body: body,
	}
	switch functionName {
	case types.RESPONSE_BALANCE:
		balance, balanceErr := GetBalanceByAddress(nodeRpc, tempParams)
		if balanceErr != nil {
			err = balanceErr
		}
		body = balance
	case types.RESPONSE_TXHASH:
		txHash, txErr := SendRawTransaction(nodeRpc, tempParams)
		if txErr != nil {
			err = txErr
		}
		body = txHash
	case types.RESPONSE_JETTON_ADDRESS:
		jettonAddress, addreErr := GetJettonAddress(nodeRpc, tempParams)
		if addreErr != nil {
			err = addreErr
		}
		body = jettonAddress
	case types.RESPONSE_NFT_ADDRESS:
		nftAddress, addreErr := GetNFTItemAddress(nodeRpc, tempParams)
		if addreErr != nil {
			err = addreErr
		}
		body = nftAddress
	case types.RESPONSE_DRY_RUN:
		gasFee, gasFeeErr := GetEstimateFee(nodeRpc, tempParams)
		if gasFeeErr != nil {
			err = gasFeeErr
		}
		body = gasFee
	case types.RESPONSE_TOKEN_BALANCE:
		tokenBalance, balanceErr := GetAllBalance(nodeRpc, tempParams)
		if balanceErr != nil {
			err = balanceErr
		}
		balanceByte, _ := json.Marshal(tokenBalance)
		body = string(balanceByte)
	case types.RESPONSE_NONCE:
		nonce, nonceErr := GetNonce(nodeRpc, tempParams)
		if nonceErr != nil {
			err = nonceErr
		}
		body = nonce
	case types.RESPONSE_TOKEN_INFO:
		tokenInfo, infoErr := p.GetTokenType(string(tempParams))
		if infoErr != nil {
			err = infoErr
		}
		if tokenInfo != nil {
			infoByte, _ := json.Marshal(tokenInfo)
			body = string(infoByte)
		}
	}
	if err != nil {
		result.Method = err.Error()
	}
	result.Body = body
	return result, nil
}

func (p *platform) AnalysisWasmResponse(ctx context.Context, functionName, params, response string) (string, error) {
	return "", nil
}

func (p *platform) GetRpcURL() []string {
	return p.rpcURL
}

func (p *platform) GetTokenType(token string) (*v12.GetTokenInfoResp_Data, error) {
	url := fmt.Sprintf("https://api.ton.cat/v2/contracts/jetton/%v", token)

	out := &types.TonTokenInfo{}
	if err := utils.HttpsGet(url, nil, nil, out); err != nil {
		return nil, err
	}
	if out.Message != "" {
		return nil, errors.New(out.Message)
	}
	return &v12.GetTokenInfoResp_Data{
		Chain:    p.chain,
		Address:  token,
		Symbol:   strings.ToUpper(out.Jetton.Metadata.Symbol),
		Name:     out.Jetton.Metadata.Name,
		Decimals: uint32(out.Jetton.Metadata.Decimals),
		LogoURI:  out.Jetton.Metadata.Image.Original,
	}, nil
}

func (p *platform) IsContractAddress(address string) (bool, error) {
	return false, nil
}

func (p *platform) GetERCType(token string) string {
	return ""
}

func NewTonPlatform(chain string, rpcURL []string, logger log.Logger) types.Platform {
	log := log.NewHelper(log.With(logger, "module", "platform/ton"))
	return &platform{rpcURL: rpcURL, log: log, chain: chain}
}

func GetBalanceByAddress(nodeURL string, reqParams []byte) (string, error) {
	url := fmt.Sprintf("%v/api/v3/account", nodeURL)
	var params map[string]string
	json.Unmarshal(reqParams, &params)
	out := &types.TonAccountResp{}
	headers := map[string]string{}
	err := utils.HttpsGet(url, params, headers, out)
	if err != nil {
		return "0", err
	}
	return utils2.UpdateDecimals(out.Balance, TON_DECIMALS), nil
}

func SendRawTransaction(nodeURL string, reqParams []byte) (string, error) {
	url := fmt.Sprintf("%v/api/v3/message", nodeURL)
	out := &types.TonSendTxHash{}
	var params map[string]string
	json.Unmarshal(reqParams, &params)
	headers := map[string]string{}
	err := utils.CommHttpsForm(url, "POST", nil, headers, params, out)
	if err != nil {
		return "", err
	}
	if out.Error != "" {
		return "", errors.New(out.Error)
	}
	return out.MessageHash, nil
}

func GetJettonAddress(nodeURL string, reqParams []byte) (string, error) {
	url := fmt.Sprintf("%v/api/v3/jetton/wallets", nodeURL)
	var params map[string]string
	json.Unmarshal(reqParams, &params)
	out := &types.TonJettonList{}
	headers := map[string]string{}
	err := utils.HttpsGet(url, params, headers, out)
	if err != nil {
		return "", err
	}
	if len(out.JettonWallets) > 0 {
		jettonAddress, err := ParseAccount(out.JettonWallets[0].Address)
		return jettonAddress, err
	}
	return "", errors.New("dont get jetton wallets")
}

func GetNFTItemAddress(nodeURL string, reqParams []byte) (string, error) {

	url := fmt.Sprintf("%v/api/v3/nft/items", nodeURL)
	var params map[string]string
	json.Unmarshal(reqParams, &params)
	out := &types.TonNFTList{}
	headers := map[string]string{}
	err := utils.HttpsGet(url, params, headers, out)
	if err != nil {
		return "", err
	}
	if len(out.NftItems) > 0 {
		return ParseAccount(out.NftItems[0].Address)
	}
	return "", errors.New("dont get nft items")
}

func GetEstimateFee(nodeURL string, reqParams []byte) (string, error) {
	url := fmt.Sprintf("%v/api/v3/estimateFee", nodeURL)
	out := &types.TonEstimateFeeResp{}
	var reqBody interface{}
	json.Unmarshal(reqParams, &reqBody)
	headers := map[string]string{}
	err := utils.CommHttpsForm(url, "POST", nil, headers, reqBody, out)
	if err != nil {
		return "", err
	}
	if out.Error != "" {
		return "", errors.New(out.Error)
	}
	gasFee := out.SourceFees.GasFee + out.SourceFees.FwdFee + out.SourceFees.InFwdFee + out.SourceFees.StorageFee
	gasFee = int(math.Ceil(float64(gasFee) * 1.2))
	//gasLimit = utils.UpdateDecimals(fmt.Sprintf("%v",gasLimit),TON_DECIMALS)
	return fmt.Sprintf("%v", gasFee), nil
}

func GetAllBalance(nodeURL string, reqParams []byte) (map[string]string, error) {
	result := make(map[string]string)
	index := 0
	limit := 128
	ownerAddress := string(reqParams)
	balanceMap, err := GetJettonWallets(nodeURL, ownerAddress, index, limit)
	if err != nil {
		return balanceMap, err
	}
	for token, balance := range balanceMap {
		result[token] = balance
	}
	for len(balanceMap) >= limit {
		index++
		balanceMap, err = GetJettonWallets(nodeURL, ownerAddress, index, limit)
		if err != nil {
			return result, err
		}
		for token, balance := range balanceMap {
			result[token] = balance
		}
	}
	return result, nil
}

func GetJettonWallets(nodeURL, ownerAddress string, index, limit int) (map[string]string, error) {

	url := fmt.Sprintf("%v/api/v3/jetton/wallets", nodeURL)
	offset := limit * index
	params := map[string]string{
		//"address":       tokenAddress,
		"owner_address": ownerAddress,
		"offset":        fmt.Sprintf("%d", offset),
		"limit":         fmt.Sprintf("%d", limit),
	}
	out := &types.TonJettonList{}
	err := utils.HttpsGet(url, params, nil, out)
	if err != nil {
		return nil, err
	}
	result := make(map[string]string)
	for _, jetton := range out.JettonWallets {
		address, err := ParseAccount(jetton.Jetton)
		if err != nil {
			address = jetton.Jetton
		}
		result[address] = jetton.Balance
	}
	return result, err
}

func GetNonce(nodeURL string, reqParams []byte) (string, error) {
	url := fmt.Sprintf("%v/api/v3/wallet", nodeURL)
	var params map[string]string
	json.Unmarshal(reqParams, &params)
	out := &types.TonWalletInfo{}
	headers := map[string]string{}
	err := utils.HttpsGet(url, params, headers, out)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%d", out.SeqNo), nil
}
