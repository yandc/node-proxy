package ethereum

import (
	"context"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/metachris/eth-go-bindings/erc165"
	"github.com/metachris/eth-go-bindings/erc20"
	"github.com/metachris/eth-go-bindings/erc721"
	v1 "gitlab.bixin.com/mili/node-proxy/api/platform/v1"
	v12 "gitlab.bixin.com/mili/node-proxy/api/tokenlist/v1"
	"gitlab.bixin.com/mili/node-proxy/pkg/platform/types"
	"gitlab.bixin.com/mili/node-proxy/pkg/platform/utils"
	"math/big"
	"strconv"
	"strings"
)

const (
	ERC20   = "ERC20"
	ERC721  = "ERC721"
	ERC1155 = "ERC1155"
)

type platform struct {
	rpcURL []string
	log    *log.Helper
	chain  string
}

func Platform2EVMPlatform(p types.Platform) *platform {
	evmPlatform, ok := p.(*platform)
	if ok {
		return evmPlatform
	}
	return nil
}

func NewEVMPlatform(chain string, rpcURL []string, logger log.Logger) types.Platform {
	log := log.NewHelper(log.With(logger, "module", "platform/ethereum"))
	return &platform{rpcURL: rpcURL, log: log, chain: chain}
}

func (p *platform) GetERCType(token string) string {
	if strings.HasPrefix(p.chain, "Ronin") && strings.HasPrefix(token, "ronin:") {
		token = strings.Replace(token, "ronin:", "0x", -1)
	}
	tokenAddress := common.HexToAddress(token)
	for i := 0; i < len(p.rpcURL); i++ {
		client, err := ethclient.Dial(p.rpcURL[i])
		erc20Token, err := erc20.NewErc20(tokenAddress, client)
		if err != nil {
			continue
		}
		_, err = erc20Token.Decimals(nil)
		if err == nil {
			return ERC20
		}
		erc721Token, err := erc721.NewErc721(tokenAddress, client)
		if err != nil {
			continue
		}
		checkMap := map[string][4]byte{
			ERC721:  erc165.InterfaceIdErc721,
			ERC1155: erc165.InterfaceIdErc1155,
		}
		for key, value := range checkMap {
			result, err := erc721Token.SupportsInterface(nil, value)
			if err != nil {
				continue
			}
			if result {
				return key
			}
		}
	}
	return ""

}

func (p *platform) GetTokenType(token string) (*v12.GetTokenInfoResp_Data, error) {
	sourceToken := token
	if strings.HasPrefix(p.chain, "Ronin") && strings.HasPrefix(token, "ronin:") {
		token = strings.Replace(token, "ronin:", "0x", -1)
	}
	tokenAddress := common.HexToAddress(token)
	var resultErr error
	for i := 0; i < len(p.rpcURL); i++ {
		client, err := ethclient.Dial(p.rpcURL[i])
		erc20Token, err := erc20.NewErc20(tokenAddress, client)
		if err != nil {
			resultErr = err
			continue
		}
		decimals, err := erc20Token.Decimals(nil)
		if err != nil {
			resultErr = err
			continue
		}
		if err == nil {
			symbol, err := erc20Token.Symbol(nil)
			if err != nil {
				symbol, err = erc20Token.Name(nil)
				if err != nil {
					resultErr = err
					continue
				}
			}
			return &v12.GetTokenInfoResp_Data{
				Chain:    p.chain,
				Address:  sourceToken,
				Symbol:   strings.ToUpper(symbol),
				Decimals: uint32(decimals),
			}, nil
		}
	}
	return nil, resultErr
}

func (p *platform) GetBalance(ctx context.Context, address, tokenAddress, decimals string) (string, error) {
	account := common.HexToAddress(address)
	if tokenAddress == "" || address == tokenAddress {
		for i := 0; i < len(p.rpcURL); i++ {
			client, err := ethclient.Dial(p.rpcURL[i])
			if err != nil {
				p.log.Error("new client error:", err)
				continue
			}
			balance, err := client.BalanceAt(context.Background(), account, nil)
			if err != nil {
				p.log.Error("get balance  error:", err)
				continue
			}
			d := 18
			if decimals != "" {
				d, _ = strconv.Atoi(decimals)
			}
			return utils.BigIntString(balance, d), nil
		}
	}

	//get token address
	if address != tokenAddress && tokenAddress != "" && decimals != "" {
		decimalsInt, _ := strconv.Atoi(decimals)
		tokenMap := map[string]int{tokenAddress: decimalsInt}
		for i := 0; i < len(p.rpcURL); i++ {
			tokenBalance, err := batchTokenBalance(p.rpcURL[i], address, tokenMap)
			if err != nil {
				p.log.Error("get token balance error:", err)
				continue
			}
			return tokenBalance[tokenAddress], nil
		}
	}

	return "0", nil
}

func (p *platform) BuildWasmRequest(ctx context.Context, nodeRpc, functionName, params string) (*v1.BuildWasmRequestReply, error) {

	return nil, nil
}
func (p *platform) AnalysisWasmResponse(ctx context.Context, functionName, params, response string) (string, error) {
	return "", nil
}

func batchTokenBalance(url, address string, tokenMap map[string]int) (map[string]string, error) {
	result := make(map[string]string)
	destAddress := common.HexToAddress(address)
	balanceFun := []byte("balanceOf(address)")
	hash := crypto.NewKeccakState()
	hash.Write(balanceFun)
	methodID := hash.Sum(nil)[:4]
	rpcClient, _ := rpc.DialHTTP(url)
	var tokenAddrs []string
	var be []rpc.BatchElem
	for token, _ := range tokenMap {
		var data []byte
		data = append(data, methodID...)
		tokenAddress := common.HexToAddress(token)
		data = append(data, common.LeftPadBytes(destAddress.Bytes(), 32)...)
		callMsg := map[string]interface{}{
			"from": destAddress,
			"to":   tokenAddress,
			"data": hexutil.Bytes(data),
		}
		be = append(be, rpc.BatchElem{
			Method: "eth_call",
			Args:   []interface{}{callMsg, "latest"},
			Result: new(string),
		})
		tokenAddrs = append(tokenAddrs, token)
	}
	err := rpcClient.BatchCall(be)
	if err != nil {
		return result, err
	}
	for index, b := range be {
		token := tokenAddrs[index]
		hexAmount := b.Result.(*string)
		bi := new(big.Int)
		bi.SetBytes(common.FromHex(*hexAmount))
		var balance string
		if tokenMap[token] == 0 {
			balance = bi.String()
		} else {
			balance = utils.BigIntString(bi, tokenMap[token])
		}
		result[token] = balance
	}
	return result, nil
}

func (p *platform) GetRpcURL() []string {
	return p.rpcURL
}

func (p *platform) IsContractAddress(address string) (bool, error) {
	var resultErr error
	for i := 0; i < len(p.rpcURL); i++ {
		client, err := ethclient.Dial(p.rpcURL[i])
		codeAt, err := client.CodeAt(context.Background(), common.HexToAddress(address), nil)
		if err != nil {
			resultErr = err
			continue
		}
		return len(codeAt) > 0, nil
	}
	return false, resultErr
}
