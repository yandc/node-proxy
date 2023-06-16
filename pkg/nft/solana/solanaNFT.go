package solana

import (
	"encoding/json"
	"fmt"
	"gitlab.bixin.com/mili/node-proxy/internal/data/models"
	"gitlab.bixin.com/mili/node-proxy/pkg/nft"
	"gitlab.bixin.com/mili/node-proxy/pkg/nft/types"
	"time"
)

func GetSolanaNFTAsset(chain, tokenAddress, tokenId string) (models.NftList, error) {
	fmt.Println("GetSolanaNFTAsset start")
	asset, err := getAsset(tokenAddress)
	for i := 0; err != nil && i < 3 && (asset.Results.Img == "" && asset.Results.AnimationURL == ""); i++ {
		time.Sleep(500 * time.Microsecond)
		asset, err = getAsset(tokenAddress)
	}
	if err != nil {
		return models.NftList{}, err
	}
	network := nft.GetFullName(chain)
	var properties string
	if len(asset.Results.Attributes) > 0 {
		traitsInfo := make([]types.TraitsInfo, 0, len(asset.Results.Attributes))

		for _, attribute := range asset.Results.Attributes {
			traitsInfo = append(traitsInfo, types.TraitsInfo{
				TraitType: attribute.TraitType,
				Value:     attribute.Value,
			})
		}
		b, _ := json.Marshal(traitsInfo)
		properties = string(b)
	}
	return models.NftList{
		TokenId:               tokenId,
		TokenAddress:          tokenAddress,
		Description:           asset.Results.Content,
		ImageURL:              asset.Results.Img,
		ImageOriginalURL:      asset.Results.Img,
		Chain:                 chain,
		Rarity:                "",
		Network:               network,
		Properties:            properties,
		CollectionName:        asset.Results.CollectionTitle,
		CollectionSlug:        asset.Results.CollectionName,
		CollectionDescription: asset.Results.Content,
		CollectionImageURL:    asset.Results.Img,
		NftName:               asset.Results.Title,
		AnimationURL:          asset.Results.AnimationURL,
	}, nil
}

func getAsset(address string) (types.SolanaNFTData, error) {
	url := fmt.Sprintf("https://api-mainnet.magiceden.io/rpc/getNFTByMintAddress/%v?useRarity=true", address)
	out := types.SolanaNFTData{}
	result, err := nft.DoWebRequest(url)
	if err != nil {
		return out, err
	}
	if err = json.Unmarshal([]byte(result), &out); err != nil {
		return out, err
	}
	return out, nil
}
