package data

import (
	"context"
	"fmt"
	"github.com/go-kratos/kratos/v2/log"
	v12 "gitlab.bixin.com/mili/node-proxy/api/nft-marketplace/v1"
	v2 "gitlab.bixin.com/mili/node-proxy/api/nft-marketplace/v2"
	v1 "gitlab.bixin.com/mili/node-proxy/api/nft/v1"
	"gitlab.bixin.com/mili/node-proxy/internal/biz"
	"gitlab.bixin.com/mili/node-proxy/internal/conf"
	"gitlab.bixin.com/mili/node-proxy/internal/data/models"
	"gitlab.bixin.com/mili/node-proxy/pkg/lark"
	"gitlab.bixin.com/mili/node-proxy/pkg/nft"
	"gitlab.bixin.com/mili/node-proxy/pkg/nft/list"
	"gitlab.bixin.com/mili/node-proxy/pkg/platform"
	"gitlab.bixin.com/mili/node-proxy/pkg/platform/ethereum"
	"gitlab.bixin.com/mili/node-proxy/pkg/platform/tron"
	"gitlab.bixin.com/mili/node-proxy/pkg/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type nftListRepo struct {
	log *log.Helper
	db  *gorm.DB
}

// NewNFTRepo .
func NewNFTRepo(db *gorm.DB, logger log.Logger, nftList *conf.NFTList) biz.NFTRepo {
	nft.InitNFT(db, logger, nftList)
	return &nftListRepo{
		log: log.NewHelper(logger),
		db:  db,
	}
}

func (r *nftListRepo) GetNFTInfo(ctx context.Context, chain string, tokenInfos []*v1.GetNftInfoRequest_NftInfo) (*v1.GetNftReply, error) {
	r.log.WithContext(ctx).Infof("GetNFTInfo", chain, tokenInfos)

	var nftInfos []*v1.GetNftReply_NftInfoResp

	for _, tokenInfo := range tokenInfos {

		info, err := utils.GetNFTApiClient().Info(context.Background(), &v12.InfoApiReq{
			Address: tokenInfo.TokenAddress,
			TokenId: tokenInfo.TokenId,
			Chain:   chain,
		})
		if err != nil || info == nil || info.Data == nil {
			if err != nil {
				//alarmMsg := fmt.Sprintf("请注意：%s链查询NFT信息失败，tokenAddress:%s,tokenID:%s\n错误消息：%s",
				//	chain, tokenInfo.TokenAddress, tokenInfo.TokenId, err)
				//alarmOpts := lark.WithMsgLevel("FATAL")
				//alarmOpts = lark.WithAlarmChannel("nft-marketplace")
				//lark.LarkClient.NotifyLark(alarmMsg, alarmOpts)
				tokenInfoLark := &models.NodeProxyLark{
					Chain:   chain,
					Address: tokenInfo.TokenAddress,
					ErrMsg:  err.Error(),
					TokenId: tokenInfo.TokenId,
					ErrType: models.ERR_TYPE_NFT,
				}
				r.db.Clauses(clause.OnConflict{
					Columns:   []clause.Column{{Name: "address"}, {Name: "chain"}, {Name: "token_id"}},
					UpdateAll: true,
				}).Create(tokenInfoLark)
			}
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

func (r *nftListRepo) GetNftCollectionInfo(ctx context.Context, chain, address string) (*v1.GetNftCollectionInfoReply, error) {
	r.log.WithContext(ctx).Infof("NetNftCollectionInfo", chain, address)
	ercType := platform.GetERCType(chain, address)
	if ercType == ethereum.ERC20 || ercType == tron.TRC20 {
		collectionInfo := &v1.GetNftCollectionInfoReply_Data{
			Chain:   chain,
			Address: address,
		}
		return &v1.GetNftCollectionInfoReply{
			Data: collectionInfo,
			Ok:   true,
		}, nil
	}

	info, err := utils.GetCollectionApiClient().Info(context.Background(), &v2.InfoApiReq{
		Address: address,
		Chain:   chain,
	})
	if err != nil || info == nil || info.Data == nil {
		if err != nil {
			alarmMsg := fmt.Sprintf("请注意：%s链查询NFT集合信息失败，tokenAddress:%s\n错误消息：%s", chain, address, err)
			alarmOpts := lark.WithMsgLevel("FATAL")
			alarmOpts = lark.WithAlarmChannel("nft-marketplace")
			lark.LarkClient.NotifyLark(alarmMsg, alarmOpts)
		}
		return nil, err
	}

	collectionInfo := &v1.GetNftCollectionInfoReply_Data{
		Chain:   chain,
		Address: address,
		Name:    info.Data.Name,
		//Slug:        info.Data.slug,
		ImageURL:    info.Data.Logo,
		Description: info.Data.Description,
		TokenType:   info.Data.ErcType,
	}

	return &v1.GetNftCollectionInfoReply{
		Data: collectionInfo,
		Ok:   true,
	}, nil
}

func (r *nftListRepo) AutoUpdateNFTInfo(ctx context.Context) {
	r.log.WithContext(ctx).Infof("AutoUpdateNFTInfo")
	list.AutoUpdateNFTInfo()
}
