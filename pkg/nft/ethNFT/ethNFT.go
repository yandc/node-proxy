package ethNFT

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Danny-Dasilva/CycleTLS/cycletls"
	"github.com/RomainMichau/cloudscraper_go/cloudscraper"
	"gitlab.bixin.com/mili/node-proxy/internal/data/models"
	"gitlab.bixin.com/mili/node-proxy/pkg/nft"
	"gitlab.bixin.com/mili/node-proxy/pkg/nft/types"
	"gitlab.bixin.com/mili/node-proxy/pkg/utils"
	"strings"
	"time"
)

const (
	ETH_MAIN = "https://api.opensea.io/api/v1/"
	ETH_TEST = "https://testnets-api.opensea.io/api/v1/"
)

func GetETHNFTAsset(chain, contractAddress, tokenId string) (models.NftList, error) {
	asset, err := getOpenSeaAsset(chain, contractAddress, tokenId)
	var nftListModel models.NftList
	if asset.ID > 0 && err == nil {
		nftListModel = openSea2Models(chain, asset)
	} else {
		nftInfo, err := getSingeNFTByNFTPort(contractAddress, tokenId)
		if err != nil {
			nft.GetNFTLog().Error("get asset scan error:", err)
			return nftListModel, nil
		}
		nftListModel = nftPort2Models(chain, *nftInfo)
	}
	return nftListModel, nil
}

func getOpenSeaAsset(chain, contractAddress, tokenId string) (types.Asset, error) {
	var url string
	var result types.Asset
	if strings.HasSuffix(chain, "TEST") {
		url = ETH_TEST + "asset/" + fmt.Sprintf("%s/%s?format=json", contractAddress, tokenId)
	} else {
		url = ETH_MAIN + "asset/" + fmt.Sprintf("%s/%s?format=json", contractAddress, tokenId)
	}
	resp, err := doOpenSeaRequest(url)
	if err != nil {
		for i := 0; err != nil && i < 3; i++ {
			time.Sleep(time.Duration(i) * time.Second)
			resp, err = doOpenSeaRequest(url)
		}
	}
	if err != nil {
		return result, err
	}
	err = json.Unmarshal([]byte(resp), &result)
	return result, err
}

func openSea2Models(chain string, asset types.Asset) models.NftList {
	var properties, rarity string
	if len(asset.Traits) > 0 {
		b, _ := json.Marshal(asset.Traits)
		properties = string(b)
	}
	if asset.RarityData != nil {
		b, _ := json.Marshal(asset.Traits)
		rarity = string(b)
	}
	network := nft.GetFullName(chain)
	return models.NftList{
		TokenId:               asset.TokenID,
		TokenAddress:          strings.ToLower(asset.AssetContract.Address),
		Name:                  asset.AssetContract.Name,
		Symbol:                asset.AssetContract.Symbol,
		TokenType:             strings.ToUpper(asset.AssetContract.SchemaName),
		Description:           asset.Description,
		ImageURL:              asset.ImageURL,
		ImageOriginalURL:      asset.ImageOriginalURL,
		Chain:                 chain,
		Rarity:                rarity,
		Network:               network,
		Properties:            properties,
		CollectionName:        asset.Collection.Name,
		CollectionSlug:        asset.Collection.Slug,
		CollectionDescription: asset.Collection.Description,
		CollectionImageURL:    asset.Collection.ImageURL,
		NftName:               asset.Name,
		AnimationURL:          asset.AnimationURL,
	}
}

func getSingeNFTByNFTPort(tokenAddress, tokenId string) (*types.NFTPortAssetInfo, error) {
	url := fmt.Sprintf("https://api.nftport.xyz/v0/nfts/%s/%s", tokenAddress, tokenId)
	headers := map[string]string{
		"Authorization": "275c2680-9df1-42b8-aaca-8d50ee42c3bf",
	}
	params := map[string]string{
		"chain": "ethereum",
	}
	out := &types.NFTPortAssetInfo{}
	err := utils.HttpsGet(url, params, headers, out)
	if err != nil {
		for i := 0; err != nil && i < 3; i++ {
			time.Sleep(1 * time.Second)
			err = utils.HttpsGet(url, nil, headers, out)
		}
	}
	if err != nil {
		return nil, err
	}
	if out.Response != "OK" {
		return nil, errors.New(out.Error.Message)
	}
	return out, nil
}

func nftPort2Models(chain string, asset types.NFTPortAssetInfo) models.NftList {
	tempTokenIdName := fmt.Sprintf("#%s", asset.Nft.TokenID)
	if asset.Contract.Name == "" && strings.Contains(asset.Nft.Metadata.Name, tempTokenIdName) {
		asset.Contract.Name = strings.Trim(strings.Split(asset.Nft.Metadata.Name, tempTokenIdName)[0], " ")
	}
	if asset.Nft.Metadata.Name == "" {
		asset.Nft.Metadata.Name = tempTokenIdName
	}
	properties, _ := json.Marshal(asset.Nft.Metadata.Attributes)
	network := nft.GetFullName(chain)
	return models.NftList{
		TokenId:               asset.Nft.TokenID,
		TokenAddress:          strings.ToLower(asset.Nft.ContractAddress),
		TokenType:             strings.ToUpper(asset.Contract.Type),
		Description:           asset.Nft.Metadata.Description,
		ImageURL:              asset.Nft.Metadata.Image,
		Chain:                 chain,
		Rarity:                "",
		Network:               network,
		Properties:            string(properties),
		CollectionName:        asset.Contract.Name,
		CollectionDescription: asset.Contract.Metadata.Description,
		CollectionImageURL:    asset.Contract.Metadata.ThumbnailURL,
		NftName:               asset.Nft.Metadata.Name,
		AnimationURL:          asset.Nft.AnimationURL,
	}

}

func doOpenSeaRequest(url string) (string, error) {
	client, err := cloudscraper.Init(false, false)
	options := cycletls.Options{
		Headers: map[string]string{"User-Agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 12_2_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/98.0.4758.102 Safari/537.36",
			"Accept":       "application/json",
			"Content-Type": "application/json"},

		//Proxy:           "http://127.0.0.1:1087",
		Timeout:         10,
		DisableRedirect: true,
	}
	resp, err := client.Do(url, options, "GET")
	return resp.Body, err
}
