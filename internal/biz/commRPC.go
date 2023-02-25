package biz

import (
	"context"
	"encoding/json"
	"github.com/go-kratos/kratos/v2/log"
	v1 "gitlab.bixin.com/mili/node-proxy/api/commRPC/v1"
	"gitlab.bixin.com/mili/node-proxy/pkg/utils"
	"reflect"
	"strings"
)

type CommRPCRepo interface {
	GetPriceV2(ctx context.Context, coinName, coinAddress []string, currency string) (map[string]map[string]string, error)
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
	respone := ss[0].Interface()
	ret, _ := json.Marshal(respone)
	return &v1.ExecNodeProxyRPCReply{
		Result: string(ret),
		Ok:     true,
	}, nil
}

func (uc *CommRPCUsecase) GetPriceV2(ctx context.Context, req *utils.GetPriceV2Req) (map[string]map[string]string, error) {
	return uc.repo.GetPriceV2(ctx, req.CoinName, req.CoinAddress, req.Currency)
}
