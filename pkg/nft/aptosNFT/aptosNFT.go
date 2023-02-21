package aptosNFT

import (
	"encoding/json"
	"errors"
	"fmt"
	"gitlab.bixin.com/mili/node-proxy/internal/data/models"
	"gitlab.bixin.com/mili/node-proxy/pkg/nft"
	"gitlab.bixin.com/mili/node-proxy/pkg/nft/types"
	"gitlab.bixin.com/mili/node-proxy/pkg/utils"
	"strings"
	"sync"
	"time"
)

const (
	APTOS_MAIN = "https://wqb9q2zgw7i7-mainnet.hasura.app/v1/graphql"
	APTOS_TEST = "https://knmpjhsurbz8-testnet.hasura.app/v1/graphql"
	APTOS_TYPE = "AptosNFT"
	TOPAZ_RUL  = "https://www.topaz.so/api"
)

// GetAptosNFTAsset tokenAddress is creator::collection::property_version,tokenId is nftName
func GetAptosNFTAsset(chain, tokenAddress, tokenId string) (models.NftList, error) {
	var nftModel models.NftList
	tokenAddressInfo := strings.Split(tokenAddress, "::")
	if len(tokenAddressInfo) < 2 {
		return nftModel, errors.New("tokenId invalid")
	}
	name := tokenId
	creator, collection, property_version := tokenAddressInfo[0], tokenAddressInfo[1], tokenAddressInfo[2]
	var wg sync.WaitGroup
	var nftData *types.AptosNFTData
	var collectionData *types.AptosCollectionData
	var collectionURL string
	var resultErr error
	wg.Add(3)
	go func() {
		defer wg.Done()
		var err error
		//get nft info
		nftData, err = getNFTDataByCollection(chain, collection, name, creator)
		if err != nil {
			resultErr = err
		}
	}()

	go func() {
		//get collection
		defer wg.Done()
		var err error
		collectionData, err = getCollectionData(chain, collection, creator)
		if err != nil {
			resultErr = err
		}
	}()
	go func() {
		defer wg.Done()
		topazTokenId := fmt.Sprintf("%s::%s::%s::%s", creator, collection, name, property_version)
		slug := getTopazNFTData(topazTokenId)
		if slug != "" {
			collectionURL = getTopazCollectionBySlug(slug)
		}
	}()
	wg.Wait()
	if resultErr != nil {
		return models.NftList{}, resultErr
	}
	if len(nftData.Data.CurrentTokenDatas) == 0 || len(collectionData.Data.CurrentCollectionDatas) == 0 {
		return models.NftList{}, errors.New("not find nft")
	}
	nftModel = models.NftList{
		TokenAddress:          tokenAddress,
		TokenId:               nftData.Data.CurrentTokenDatas[0].TokenDataIDHash,
		TokenType:             APTOS_TYPE,
		Chain:                 chain,
		Network:               nft.GetFullName(chain),
		Description:           nftData.Data.CurrentTokenDatas[0].Description,
		ImageURL:              nftData.Data.CurrentTokenDatas[0].MetadataURI,
		ImageOriginalURL:      nftData.Data.CurrentTokenDatas[0].MetadataURI,
		NftName:               name,
		CollectionName:        collection,
		CollectionImageURL:    collectionData.Data.CurrentCollectionDatas[0].MetadataURI,
		CollectionDescription: collectionData.Data.CurrentCollectionDatas[0].Description,
	}
	wg.Add(2)
	//parse collection
	go func() {
		defer wg.Done()
		collectionSourceData, _ := parseNFTJson(collectionData.Data.CurrentCollectionDatas[0].MetadataURI)
		if collectionSourceData.Name != "" {
			nftModel.CollectionImageURL = collectionSourceData.Image
			nftModel.CollectionDescription = collectionSourceData.Description
		} else if collectionURL != "" {
			nftModel.CollectionImageURL = collectionURL
		}
	}()

	//parse nft source data
	go func() {
		defer wg.Done()
		nftSourceData, _ := parseNFTJson(nftData.Data.CurrentTokenDatas[0].MetadataURI)
		if nftSourceData.Name != "" {
			if len(nftSourceData.Attributes) > 0 {
				b, _ := json.Marshal(nftSourceData.Attributes)
				nftModel.Properties = string(b)
			}
			nftModel.ImageURL = nftSourceData.Image
			nftModel.ImageOriginalURL = nftSourceData.Image
			nftModel.Description = nftSourceData.Description
		}
	}()
	wg.Wait()
	return nftModel, nil
}

func getNFTDataByCollection(chain, collectionName, nftName, creatorAddress string) (*types.AptosNFTData, error) {
	var reqParams = types.AptosNFTReq{
		OperationName: "TokenData",
		Variables: map[string]string{
			"collection_name": collectionName,
			"creator_address": creatorAddress,
			"name":            nftName,
		},
		Query: "query TokenData($collection_name: String, $name: String, $creator_address: String) {\n  " +
			"current_token_datas(where: {collection_name: {_eq: $collection_name}, name: {_eq: $name}, " +
			"creator_address: {_eq: $creator_address}}) {\n    token_data_id_hash\n    name\n    collection_name\n    " +
			"description\n    creator_address\n    default_properties\n    largest_property_version\n    maximum\n   " +
			" metadata_uri\n    payee_address\n    royalty_points_denominator\n    royalty_points_numerator\n    " +
			"supply\n    __typename\n  }\n}",
	}
	url := APTOS_MAIN
	if strings.HasSuffix(chain, "TEST") {
		url = APTOS_TEST
	}
	out := &types.AptosNFTData{}
	err := utils.CommHttpsForm(url, "POST", nil, nil, reqParams, out)
	if err != nil {
		for i := 0; err != nil && i < 3; i++ {
			time.Sleep(time.Duration(i) * time.Second)
			err = utils.CommHttpsForm(url, "POST", nil, nil, reqParams, out)
		}
	}
	return out, err
}

func getTopazCollectionBySlug(slug string) string {
	url := fmt.Sprintf("%s/collection?slug=%s", TOPAZ_RUL, slug)
	out := &types.AptosTopazCollection{}
	err := utils.HttpsGet(url, nil, nil, out)
	if err != nil {
		return ""
	}
	return out.Data.Collection.LogoURI
}

func getTopazNFTData(tokenId string) string {
	reqBody := map[string]string{
		"token_id": tokenId,
	}
	url := fmt.Sprintf("%s/token-data", TOPAZ_RUL)
	out := &types.AptosTopazNFTData{}
	err := utils.CommHttpsForm(url, "POST", nil, nil, reqBody, out)
	if err != nil {
		return ""
	}
	return out.Data.CollectionSlug
}

func parseNFTJson(uri string) (*types.AptosNFTSourceData, error) {
	if uri != "" && !strings.HasPrefix(uri, "https") {
		uri = strings.Replace(uri, "ipfs://", nft.GetIPFS(), 1)
	}
	out := &types.AptosNFTSourceData{}
	err := utils.HttpsGet(uri, nil, nil, out)
	if err != nil {
		body, err := nft.DoWebRequest(uri)
		if err != nil {
			return out, err
		}
		if err := json.Unmarshal([]byte(body), &out); err != nil {
			return out, err
		}
	}
	return out, err
}
