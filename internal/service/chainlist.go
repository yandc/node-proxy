package service

import (
	"context"
	"google.golang.org/protobuf/types/known/emptypb"
	"time"

	"gitlab.bixin.com/mili/node-proxy/api/chainlist/v1"
	"gitlab.bixin.com/mili/node-proxy/internal/biz"
)

type ChainListService struct {
	v1.UnimplementedChainListServer
	uc *biz.ChainListUsecase
}

func NewChainListService(uc *biz.ChainListUsecase) *ChainListService {
	return &ChainListService{uc: uc}
}

func (s *ChainListService) GetAllChainList(ctx context.Context, req *emptypb.Empty) (*v1.GetAllChainListResp, error) {
	// 设置接口 3s 超时
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	chainList, err := s.uc.GetAllChainList(ctx)
	return &v1.GetAllChainListResp{
		Data: chainList,
	}, err
}

func (s *ChainListService) GetChainList(ctx context.Context, req *v1.GetChainListReq) (*v1.GetChainListResp, error) {
	// 设置接口 3s 超时
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	chainList, err := s.uc.GetChainList(ctx, req.ChainIds)
	return &v1.GetChainListResp{
		Data: chainList,
	}, err
}

func (s *ChainListService) GetChainNodeList(ctx context.Context, req *v1.GetChainNodeListReq) (*v1.GetChainNodeListResp, error) {
	// 设置接口 3s 超时
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	chainNodeList, err := s.uc.GetChainNodeUrlList(ctx, req.ChainId)
	return &v1.GetChainNodeListResp{
		Data: chainNodeList,
	}, err
}

func (s *ChainListService) GetChainNodeInUsedList(ctx context.Context, req *emptypb.Empty) (*v1.GetChainNodeInUsedListResp, error) {
	// 设置接口 3s 超时
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	data, err := s.uc.GetChainNodeInUsedList(ctx)
	return &v1.GetChainNodeInUsedListResp{
		Data: data,
	}, err
}

func (s *ChainListService) UseChainNode(ctx context.Context, req *v1.UseChainNodeReq) (*emptypb.Empty, error) {
	// 设置接口 3s 超时
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	err := s.uc.UseChainNode(ctx, req.ChainId, req.Url, req.Source)
	return &emptypb.Empty{}, err
}
