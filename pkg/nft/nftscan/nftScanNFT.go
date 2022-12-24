package nftscan

import (
	"encoding/json"
	"errors"
	"fmt"
	"gitlab.bixin.com/mili/node-proxy/internal/data/models"
	"gitlab.bixin.com/mili/node-proxy/pkg/nft"
	"gitlab.bixin.com/mili/node-proxy/pkg/nft/types"
	utils2 "gitlab.bixin.com/mili/node-proxy/pkg/utils"
	"strings"
	"sync"
)

var chainURLMap = map[string]string{
	"Arbitrm":     "https://arbitrumapi.nftscan.com/api/v2",
	"ArbitrmTEST": "https://arbitrumapi.nftscan.com/api/v2",

	"BSC":     "https://bnbapi.nftscan.com/api/v2",
	"BSCTEST": "https://bnbapi.nftscan.com/api/v2",

	"Polygon":     "https://polygonapi..nftscan.com/api/v2",
	"PolygonTEST": "https://polygonapi..nftscan.com/api/v2",
}

func GetNFTScanNFTAsset(chain, tokenAddress, tokenId string) (models.NftList, error) {
	var wg sync.WaitGroup
	var asset *types.NFTScanAsset
	var collection models.NftCollection
	var err error
	wg.Add(2)
	go func() {
		defer wg.Done()
		asset, err = getNFTScanAsset(chain, tokenAddress, tokenId)
	}()
	go func() {
		defer wg.Done()
		collection, err = GetNFTScanCollection(chain, tokenAddress)
	}()
	wg.Wait()
	if err != nil {
		return models.NftList{}, err
	}
	return nftScan2Models(chain, asset, collection), nil
}

func getNFTScanAsset(chain, tokenAddress, tokenId string) (*types.NFTScanAsset, error) {
	var url string
	if value, ok := chainURLMap[chain]; ok {
		url = fmt.Sprintf("%s/assets/%s/%s", value, tokenAddress, tokenId)
	}
	if url == "" {
		return nil, errors.New("dont chain:" + chain)
	}
	out := &types.NFTScanAsset{}
	err := utils2.HttpsGet(url, map[string]string{"show_attribute": "true"},
		map[string]string{"x-api-key": "mETQXdrOHsTsA3qexgfd4s8t"}, out)
	if err != nil {
		return nil, err
	}
	if out == nil {
		return nil, errors.New("get asset is null")
	}
	if out.Msg != "" {
		return nil, errors.New(out.Msg)
	}
	return out, err
}

func nftScan2Models(chain string, asset *types.NFTScanAsset, collection models.NftCollection) models.NftList {
	var mateData types.NFTScanNFTAssetMateData
	var properties string
	if err := json.Unmarshal([]byte(asset.Data.MetadataJSON), &mateData); err != nil {

	}
	if len(asset.Data.Attributes) > 0 {
		tempAttributes := make([]types.TraitsInfo, len(asset.Data.Attributes))
		for i := 0; i < len(asset.Data.Attributes); i++ {
			tempAttributes[i] = types.TraitsInfo{
				TraitType:  asset.Data.Attributes[i].AttributeName,
				Value:      asset.Data.Attributes[i].AttributeValue,
				TraitCount: asset.Data.Attributes[i].Percentage,
			}
		}
		b, _ := json.Marshal(tempAttributes)
		properties = string(b)
	}
	network := nft.GetFullName(chain)
	return models.NftList{
		TokenId:               asset.Data.TokenID,
		TokenAddress:          asset.Data.ContractAddress,
		Name:                  asset.Data.ContractName,
		Symbol:                collection.Slug,
		TokenType:             strings.ToUpper(asset.Data.ErcType),
		Description:           mateData.Description,
		ImageURL:              mateData.Image,
		Chain:                 chain,
		Network:               network,
		Properties:            properties,
		NftName:               mateData.Name,
		AnimationURL:          mateData.AnimationUrl,
		CollectionName:        collection.Name,
		CollectionDescription: collection.Description,
		CollectionImageURL:    collection.ImageURL,
	}
}
