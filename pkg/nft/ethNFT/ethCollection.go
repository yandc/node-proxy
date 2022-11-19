package ethNFT

import (
	"encoding/json"
	"fmt"
	"gitlab.bixin.com/mili/node-proxy/internal/data/models"
	"gitlab.bixin.com/mili/node-proxy/pkg/nft/types"
	"strings"
)

func GetETHCollection(chain, address string) (models.NftCollection, error) {
	collection, err := getOpenSeaCollection(chain, address)
	if err != nil {
		return models.NftCollection{}, err
	}
	return openSeaCollection2Models(chain, collection), nil
}

func getOpenSeaCollection(chain, address string) (types.OpenSeaCollection, error) {
	var url string
	var result types.OpenSeaCollection
	//	url := "https://api.opensea.io/api/v1/asset_contract/0x06012c8cf97bead5deae237070f9587f8e7a266d"
	if strings.HasSuffix(chain, "TEST") {
		url = ETH_TEST + "asset_contract/" + fmt.Sprintf("%s?format=json", address)
	} else {
		url = ETH_MAIN + "asset_contract/" + fmt.Sprintf("%s?format=json", address)
	}
	resp, err := doOpenSeaRequest(url)
	if err != nil {
		return result, err
	}
	err = json.Unmarshal([]byte(resp), &result)
	return result, err
}

func openSeaCollection2Models(chain string, collection types.OpenSeaCollection) models.NftCollection {

	return models.NftCollection{
		Name:        collection.Name,
		Slug:        collection.Collection.Slug,
		ImageURL:    collection.ImageURL,
		Chain:       chain,
		Description: collection.Description,
		Address:     strings.ToLower(collection.Address),
		TokenType:   strings.ToUpper(collection.SchemaName),
	}
	//nft.GetNFTDb().Clauses(clause.OnConflict{
	//	UpdateAll: true,
	//}).Create(&collectionModel)
}
