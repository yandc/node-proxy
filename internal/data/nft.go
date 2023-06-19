package data

import (
	"context"
	"github.com/go-kratos/kratos/v2/log"
	v12 "gitlab.bixin.com/mili/node-proxy/api/nft-marketplace/v1"
	v1 "gitlab.bixin.com/mili/node-proxy/api/nft/v1"
	"gitlab.bixin.com/mili/node-proxy/internal/biz"
	"gitlab.bixin.com/mili/node-proxy/internal/conf"
	"gitlab.bixin.com/mili/node-proxy/pkg/nft"
	"gitlab.bixin.com/mili/node-proxy/pkg/nft/collection"
	"gitlab.bixin.com/mili/node-proxy/pkg/nft/list"
	"gitlab.bixin.com/mili/node-proxy/pkg/utils"
	"gorm.io/gorm"
	"strings"
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

	if strings.ToLower(chain) == "solana" ||
		strings.ToLower(chain) == "eth" ||
		strings.ToLower(chain) == "avalanche" ||
		strings.ToLower(chain) == "bsc" ||
		strings.ToLower(chain) == "polygon" ||
		strings.ToLower(chain) == "optimism" ||
		strings.ToLower(chain) == "klaytn" ||
		strings.ToLower(chain) == "arbitrum" {
		var nftInfos []*v1.GetNftReply_NftInfoResp

		for _, tokenInfo := range tokenInfos {

			info, err := utils.GetNFTApiClient().Info(context.Background(), &v12.InfoApiReq{
				Address: tokenInfo.TokenAddress,
				TokenId: tokenInfo.TokenId,
				Chain:   chain,
			})
			if err != nil {
				continue
			}

			nftInfos = append(nftInfos, &v1.GetNftReply_NftInfoResp{
				TokenId:               info.Data.TokenId,
				TokenAddress:          info.Data.Address,
				Name:                  info.Data.Name,
				TokenType:             info.Data.Type,
				ImageURL:              info.Data.Image,
				ImageOriginalURL:      info.Data.Image,
				Description:           info.Data.Description,
				Chain:                 chain,
				CollectionName:        info.Data.CollectionName,
				Network:               chain,
				Properties:            info.Data.Traits,
				CollectionDescription: info.Data.CollectionDescription,
				NftName:               info.Data.Name,
				CollectionImageURL:    info.Data.CollectionImage,
				AnimationURL:          info.Data.Animation,
			})
		}

		return &v1.GetNftReply{
			Data: nftInfos,
			Ok:   true,
		}, nil
	}

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

func (r *nftListRepo) GetNftCollectionInfo(ctx context.Context, chain, address string) (*v1.GetNftCollectionInfoReply, error) {
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
		Data:   collectionInfo,
	}, nil
}

func (r *nftListRepo) AutoUpdateNFTInfo(ctx context.Context) {
	r.log.WithContext(ctx).Infof("AutoUpdateNFTInfo")
	list.AutoUpdateNFTInfo()
}
