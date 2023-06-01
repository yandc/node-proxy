package service

import (
	"context"
	"errors"
	"gitlab.bixin.com/mili/node-proxy/internal/biz"
	"time"

	pb "gitlab.bixin.com/mili/node-proxy/api/platform/v1"
)

type PlatformService struct {
	pb.UnimplementedPlatformServer
	pc *biz.PlatformUseCase
}

func NewPlatformService(pc *biz.PlatformUseCase) *PlatformService {
	return &PlatformService{pc: pc}
}

func (s *PlatformService) GetBalance(ctx context.Context, req *pb.GetBalanceRequest) (*pb.GetBalanceReply, error) {
	// 设置接口 3s 超时
	subctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	balance, err := s.pc.GetBalance(subctx, req.Chain, req.Address, req.TokenAddress, req.Decimals)
	return &pb.GetBalanceReply{Balance: balance}, err
}

func (s *PlatformService) BuildWasmRequest(ctx context.Context, req *pb.BuildWasmRequestRequest) (
	*pb.BuildWasmRequestReply, error) {
	// 设置接口 3s 超时
	subctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	data := make(chan *pb.BuildWasmRequestReply, 1)
	var resultErr error
	go func() {
		//time.Sleep(5 * time.Second)
		result, err := s.pc.BuildWasmRequest(subctx, req.Chain, req.NodeRpc, req.FunctionName, req.Params)
		data <- result
		resultErr = err
	}()
	select {
	case ret := <-data:
		return ret, resultErr
	case <-subctx.Done():
		return nil, errors.New("timeOut")
	}
	//return s.pc.BuildWasmRequest(subctx, req.Chain, req.NodeRpc, req.FunctionName, req.Params)
}

func (s *PlatformService) AnalysisWasmResponse(ctx context.Context, req *pb.AnalysisWasmResponseRequest) (
	*pb.AnalysisWasmResponseReply, error) {
	// 设置接口 3s 超时
	subctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	data, err := s.pc.AnalysisWasmResponse(subctx, req.Chain, req.FunctionName, req.Params, req.Response)
	ok := true
	errMsg := ""
	if err != nil {
		ok = false
		errMsg = err.Error()
	}

	return &pb.AnalysisWasmResponseReply{
		Data:   data,
		Ok:     ok,
		ErrMsg: errMsg,
	}, nil
}

func (s *PlatformService) GetGasEstimate(ctx context.Context, req *pb.GetGasEstimateRequest) (*pb.GetGasEstimateReply,
	error) {
	// 设置接口 3s 超时
	subctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	data, err := s.pc.GetGasEstimate(subctx, req.Chain, req.GasInfo)
	ok := true
	errMsg := ""
	if err != nil {
		ok = false
		errMsg = err.Error()
	}
	return &pb.GetGasEstimateReply{
		Ok:     ok,
		Data:   data,
		ErrMsg: errMsg,
	}, nil
}
