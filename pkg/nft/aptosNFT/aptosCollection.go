package aptosNFT

import (
	"errors"
	"gitlab.bixin.com/mili/node-proxy/internal/data/models"
	"gitlab.bixin.com/mili/node-proxy/pkg/nft/types"
	"gitlab.bixin.com/mili/node-proxy/pkg/utils"
	"strings"
	"time"
)

// GetAptosCollection address is creator::collection::property_version
func GetAptosCollection(chain, address string) (models.NftCollection, error) {
	var nftModel models.NftCollection
	collectionInfo := strings.Split(address, "::")
	if len(collectionInfo) < 2 {
		return models.NftCollection{}, errors.New("address  invalid")
	}
	creator, collection := collectionInfo[0], collectionInfo[1]
	collData, err := getCollectionData(chain, collection, creator)
	if err != nil {
		return models.NftCollection{}, err
	}
	if len(collData.Data.CurrentCollectionDatas) == 0 {
		return nftModel, errors.New("not find collection")
	}
	nftModel = models.NftCollection{
		Name:        collection,
		Address:     address,
		ImageURL:    collData.Data.CurrentCollectionDatas[0].MetadataURI,
		Description: collData.Data.CurrentCollectionDatas[0].Description,
		TokenType:   APTOS_TYPE,
	}
	//parse collection
	collectionSourceData, _ := parseNFTJson(collData.Data.CurrentCollectionDatas[0].MetadataURI)
	if collectionSourceData.Name != "" {
		nftModel.ImageURL = collectionSourceData.Image
		nftModel.Description = collectionSourceData.Description
	}
	return nftModel, nil
}

func getCollectionData(chain, collectionName, creatorAddress string) (*types.AptosCollectionData, error) {
	var reqParams = types.AptosNFTReq{
		OperationName: "CollectionData",
		Variables: map[string]string{
			"collection_name": collectionName,
			"creator_address": creatorAddress,
		},
		Query: "query CollectionData($collection_name: String, $creator_address: String) {\n  " +
			"current_collection_datas(\n    where: {collection_name: {_eq: $collection_name}, creator_address: " +
			"{_eq: $creator_address}}) {\n    collection_name\n    description\n    creator_address\n    " +
			"metadata_uri\n    maximum\n    __typename\n  }\n}",
	}
	url := APTOS_MAIN
	if strings.HasSuffix(chain, "TEST") {
		url = APTOS_TEST
	}
	out := &types.AptosCollectionData{}
	err := utils.CommHttpsForm(url, "POST", nil, nil, reqParams, out)
	if err != nil {
		for i := 0; err != nil && i < 3; i++ {
			time.Sleep(time.Duration(i) * time.Second)
			err = utils.CommHttpsForm(url, "POST", nil, nil, reqParams, out)
		}
	}
	return out, err
}
