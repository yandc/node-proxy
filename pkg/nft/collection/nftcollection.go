package collection

import (
	"encoding/json"
	"fmt"
	v1 "gitlab.bixin.com/mili/node-proxy/api/nft/v1"
	"gitlab.bixin.com/mili/node-proxy/internal/data/models"
	"gitlab.bixin.com/mili/node-proxy/pkg/nft"
	"gitlab.bixin.com/mili/node-proxy/pkg/nft/types"
	"gitlab.bixin.com/mili/node-proxy/pkg/utils"
	"gorm.io/gorm/clause"
	"strings"
	"time"
)

func CreateCollectionList() {
	offset := 0
	collectionsList, err := GetCollections(offset)
	if err != nil {
		nft.GetNFTLog().Error("get collections list error:", err)
	}
	CollectionsType2Models(collectionsList)
	//for len(collectionsList.Collections) >= nft.MAXLISTLIMIT {
	//	offset++
	//	collectionsList,err = GetCollections(offset)
	//	if err != nil {
	//		nft.GetNFTLog().Error("get collections list error:",err)
	//	}
	//	CollectionsType2Models(collectionsList)
	//}
}

//GetCollections https://api.opensea.io/api/v1/collections
func GetCollections(offset int) (*types.CollectionList, error) {
	url := nft.BASEURL + "collections"
	params := map[string]string{
		"offset": fmt.Sprintf("%d", offset),
		"limit":  fmt.Sprintf("%d", nft.MAXLISTLIMIT),
	}
	out := &types.CollectionList{}
	headers := map[string]string{
		"X-API-KEY": nft.OPENSEA_KEY,
		"accept":    "application/json",
	}
	err := utils.HttpsGet(url, params, headers, out)
	if err != nil {
		for i := 0; err != nil && i < 3; i++ {
			time.Sleep(1 * time.Second)
			err = utils.HttpsGet(url, params, headers, out)
		}
	}
	return out, err
}

func CollectionsType2Models(source *types.CollectionList) {
	collectionModels := make([]models.NftCollection, len(source.Collections))
	for i := 0; i < len(source.Collections); i++ {
		collection := source.Collections[i]
		collectionModels[i] = models.NftCollection{
			Name:        collection.Name,
			Slug:        collection.Slug,
			ImageURL:    collection.ImageURL,
			Chain:       "ETH",
			Description: collection.Description,
			Address:     "",
		}
	}
	nft.GetNFTDb().Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(&collectionModels)
}

func GetNFTCollectionInfo(chain, address string) (*v1.GetNftCollectionInfoReply_Data, error) {
	var collectionInfo models.NftCollection
	tempAddress := address
	if strings.HasPrefix(address, "0x") && chain != "STC" {
		tempAddress = strings.ToLower(address)
	}
	err := nft.GetNFTDb().Where("chain = ? and address = ?", chain, tempAddress).First(&collectionInfo).Error
	if err != nil {
		return nil, err
	}
	if collectionInfo.ID > 0 {
		return &v1.GetNftCollectionInfoReply_Data{
			Chain:       chain,
			Address:     address,
			Name:        collectionInfo.Name,
			Slug:        collectionInfo.Slug,
			Description: collectionInfo.Description,
			ImageURL:    collectionInfo.ImageURL,
		}, nil
	}

	//select nft list
	var nftList models.NftList
	err = nft.GetNFTDb().Where("chain = ? and address = ?", chain, tempAddress).First(&nftList).Error
	if err != nil {
		return nil, err
	}
	if nftList.ID > 0 {
		return &v1.GetNftCollectionInfoReply_Data{
			Chain:       chain,
			Address:     address,
			Name:        nftList.CollectionName,
			Slug:        nftList.CollectionSlug,
			Description: nftList.CollectionDescription,
			ImageURL:    nftList.CollectionImageURL,
		}, nil
	}

	//get info to chain

	openSeaCollection, err := GetOpenSeaCollection(chain, address)
	if err != nil {
		return nil, err
	}
	OpenSeaCollection2Models(openSeaCollection)
	return &v1.GetNftCollectionInfoReply_Data{
		Chain:       chain,
		Address:     address,
		Name:        openSeaCollection.Name,
		Slug:        openSeaCollection.Collection.Slug,
		Description: openSeaCollection.Description,
		ImageURL:    openSeaCollection.ImageURL,
	}, nil
}

func GetOpenSeaCollection(chain, address string) (types.OpenSeaCollection, error) {
	var url string
	var result types.OpenSeaCollection
	//	url := "https://api.opensea.io/api/v1/asset_contract/0x06012c8cf97bead5deae237070f9587f8e7a266d"
	if strings.HasSuffix(chain, "TEST") {
		url = nft.TESTBASEURL + "asset_contract/" + fmt.Sprintf("%s?format=json", address)
	} else {
		url = nft.BASEURL + "asset_contract/" + fmt.Sprintf("%s?format=json", address)
	}
	resp, err := nft.DoOpenSeaRequest(url)
	if err != nil {
		return result, err
	}
	err = json.Unmarshal([]byte(resp), &result)
	return result, err
}

func OpenSeaCollection2Models(collection types.OpenSeaCollection) {
	collectionModel := models.NftCollection{
		Name:        collection.Name,
		Slug:        collection.Collection.Slug,
		ImageURL:    collection.ImageURL,
		Chain:       "ETH",
		Description: collection.Description,
		Address:     strings.ToLower(collection.Address),
		TokenType:   strings.ToUpper(collection.SchemaName),
	}
	nft.GetNFTDb().Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(&collectionModel)
}
