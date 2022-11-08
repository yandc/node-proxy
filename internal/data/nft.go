package data

import (
	"context"
	"github.com/go-kratos/kratos/v2/log"
	v1 "gitlab.bixin.com/mili/node-proxy/api/nft/v1"
	"gitlab.bixin.com/mili/node-proxy/internal/biz"
	"gitlab.bixin.com/mili/node-proxy/internal/conf"
	"gitlab.bixin.com/mili/node-proxy/pkg/nft"
	"gitlab.bixin.com/mili/node-proxy/pkg/nft/collection"
	"gitlab.bixin.com/mili/node-proxy/pkg/nft/list"
	"gorm.io/gorm"
)

type nftListRepo struct {
	log *log.Helper
}

// NewNFTRepo .
func NewNFTRepo(db *gorm.DB, logger log.Logger, nftList *conf.NFTList) biz.NFTRepo {
	nft.InitNFT(db, logger, nftList)
	return &nftListRepo{
		log: log.NewHelper(logger),
	}
}

func (r *nftListRepo) GetNFTInfo(ctx context.Context, chain string, tokenInfos []*v1.GetNftInfoRequest_NftInfo) (*v1.GetNftReply, error) {
	r.log.WithContext(ctx).Infof("GetNFTInfo", chain, tokenInfos)
	nftInfo, err := list.GetNFTInfo(chain, tokenInfos)
	ok := true
	errMsg := ""
	if err != nil {
		ok = false
		errMsg = err.Error()
	}
	return &v1.GetNftReply{
		Data:   nftInfo,
		Ok:     ok,
		ErrMsg: errMsg,
	}, nil
}

func (r *nftListRepo) NetNftCollectionInfo(ctx context.Context, chain, address string) (*v1.GetNftCollectionInfoReply, error) {
	r.log.WithContext(ctx).Infof("NetNftCollectionInfo", chain, address)
	collectionInfo, err := collection.GetNFTCollectionInfo(chain, address)
	ok := true
	errMsg := ""
	if err != nil {
		ok = false
		errMsg = err.Error()
	}
	return &v1.GetNftCollectionInfoReply{
		Ok:     ok,
		ErrMsg: errMsg,
		Data: &v1.GetNftCollectionInfoReply_Data{
			Chain:       collectionInfo.Chain,
			Address:     collectionInfo.Address,
			Name:        collectionInfo.Name,
			Slug:        collectionInfo.Slug,
			Description: collectionInfo.Description,
			ImageURL:    collectionInfo.ImageURL,
		},
	}, nil
}
