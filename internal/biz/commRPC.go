package biz

import (
	"context"
	"encoding/json"
	"reflect"
	"strings"

	"github.com/go-kratos/kratos/v2/log"
	v1 "gitlab.bixin.com/mili/node-proxy/api/commRPC/v1"
	"gitlab.bixin.com/mili/node-proxy/pkg/platform/types"
	"gitlab.bixin.com/mili/node-proxy/pkg/utils"
)

type CommRPCRepo interface {
	GetPriceV2(ctx context.Context, coinName, coinAddress []string, currency string) (map[string]map[string]string, error)
	GetContractABI(ctx context.Context, chain, contract, methodId string) (interface{}, error)
	ParseDataByABI(ctx context.Context, chain, contract, data string) *types.ParseDataResponse
	GetPretreatmentAmount(ctx context.Context, chain, from, to, data, value string) map[string][]interface{}
	IsContractAddress(ctx context.Context, chain, address string) (bool, error)
	GetGasConstants(ctx context.Context) map[string]interface{}
	GetChainDataConfig(ctx context.Context) map[string]interface{}
	GetGasOracle(ctx context.Context, key string, cacheTime int64) string
}

type CommRPCUsecase struct {
	repo CommRPCRepo
	log  *log.Helper
}

// NewCommRPCUsecase new a comm rpc usecase.
func NewCommRPCUsecase(repo CommRPCRepo, logger log.Logger) *CommRPCUsecase {
	return &CommRPCUsecase{repo: repo, log: log.NewHelper(logger)}
}

func (uc *CommRPCUsecase) ExecNodeProxyRPC(ctx context.Context, req *v1.ExecNodeProxyRPCRequest) (*v1.ExecNodeProxyRPCReply, error) {
	valueUC := reflect.TypeOf(uc)
	mv, ok := valueUC.MethodByName(req.Method)
	if !ok {
		return &v1.ExecNodeProxyRPCReply{
			Ok:     false,
			ErrMsg: "not support " + req.Method,
		}, nil
	}
	args := make([]reflect.Value, 0)
	args = append(args, reflect.ValueOf(uc))
	args = append(args, reflect.ValueOf(ctx))
	if len(req.Params) > 0 {
		u := mv.Type.NumIn()
		paseJson := reflect.New(mv.Type.In(u - 1).Elem())
		reqKey := strings.ReplaceAll(utils.ListToString(req.Params), "\\", "")

		jsonErr := json.Unmarshal([]byte(reqKey), paseJson.Interface())
		if jsonErr == nil {
			args = append(args, reflect.ValueOf(paseJson.Interface()))
		} else {
			return &v1.ExecNodeProxyRPCReply{
				Ok:     false,
				ErrMsg: "param error ",
			}, jsonErr
		}
	}

	ss := mv.Func.Call(args)

	// Error handling.
	if len(ss) > 1 {
		if err, ok := ss[1].Interface().(error); ok && err != nil {
			return &v1.ExecNodeProxyRPCReply{
				Ok:     false,
				ErrMsg: err.Error(),
			}, nil
		}
	}
	response := ss[0].Interface()
	var result string
	if value, ok := response.(string); ok {
		result = value
	} else {
		ret, _ := json.Marshal(response)
		result = string(ret)
	}
	if result == "null" {
		result = ""
	}
	if req.Method == "GetContractABI" && strings.Contains(result, "Function") {
		result = strings.Replace(result, "Function", "function", -1)
	}
	return &v1.ExecNodeProxyRPCReply{
		Result: result,
		Ok:     true,
	}, nil
}

func (uc *CommRPCUsecase) GetPriceV2(ctx context.Context, req *utils.GetPriceV2Req) (map[string]map[string]string, error) {
	return uc.repo.GetPriceV2(ctx, req.CoinName, req.CoinAddress, req.Currency)
}

func (uc *CommRPCUsecase) GetContractABI(ctx context.Context, req *utils.GetABIReq) (interface{}, error) {
	return uc.repo.GetContractABI(ctx, req.Chain, req.Contract, req.MethodId)
}

func (uc *CommRPCUsecase) ParseDataByABI(ctx context.Context, req *utils.ParseDataByABIReq) *types.ParseDataResponse {
	return uc.repo.ParseDataByABI(ctx, req.Chain, req.Contract, req.Data)
}

func (uc *CommRPCUsecase) GetPretreatment(ctx context.Context, req *utils.PretreatmentReq) map[string][]interface{} {
	return uc.repo.GetPretreatmentAmount(ctx, req.Chain, req.From, req.To, req.Data, req.Value)
}

func (uc *CommRPCUsecase) IsContractAddress(ctx context.Context, req *utils.IsContractReq) (bool, error) {
	return uc.repo.IsContractAddress(ctx, req.Chain, req.Address)
}

func (uc *CommRPCUsecase) GetGasConstants(ctx context.Context, req *utils.GasDefaultsReq) map[string]interface{} {
	return uc.repo.GetGasConstants(ctx)
}

func (uc *CommRPCUsecase) GetChainDataConfig(ctx context.Context, req *utils.GasDefaultsReq) map[string]interface{} {
	return uc.repo.GetChainDataConfig(ctx)
}

func (uc *CommRPCUsecase) GetGasOracle(ctx context.Context, req *utils.GasOracleReq) string {
	return uc.repo.GetGasOracle(ctx, req.Key, req.CacheTime)
}
