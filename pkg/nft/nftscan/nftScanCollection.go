package nftscan

import (
	"errors"
	"fmt"
	"gitlab.bixin.com/mili/node-proxy/internal/data/models"
	"gitlab.bixin.com/mili/node-proxy/pkg/nft/types"
	utils2 "gitlab.bixin.com/mili/node-proxy/pkg/utils"
	"strings"
)

func GetNFTScanCollection(chain, address string) (models.NftCollection, error) {
	collection, err := getCollection(chain, address)
	if err != nil {
		return models.NftCollection{}, err
	}
	if collection == nil {
		return models.NftCollection{}, errors.New("get collection is null")
	}
	return nftScanCollection2Models(chain, collection), nil
}

func getCollection(chain, address string) (*types.NFTScanCollection, error) {
	var url string
	if value, ok := chainURLMap[chain]; ok {
		url = fmt.Sprintf("%s/collections/%s", value, address)
	}
	if url == "" {
		return nil, errors.New("dont chain:" + chain)
	}
	out := &types.NFTScanCollection{}
	err := utils2.HttpsGet(url, nil, map[string]string{"x-api-key": "mETQXdrOHsTsA3qexgfd4s8t"}, out)
	if err != nil {
		return nil, err
	}
	if out.Msg != "" {
		return nil, errors.New(out.Msg)
	}
	return out, nil
}

func nftScanCollection2Models(chain string, collection *types.NFTScanCollection) models.NftCollection {
	return models.NftCollection{
		Name:        collection.Data.Name,
		Slug:        collection.Data.Symbol,
		ImageURL:    collection.Data.LogoURL,
		Chain:       chain,
		Description: collection.Data.Description,
		Address:     strings.ToLower(collection.Data.ContractAddress),
		TokenType:   strings.ToUpper(collection.Data.ErcType),
	}
}
