package sui

import (
	"gitlab.bixin.com/mili/node-proxy/internal/data/models"
	"gitlab.bixin.com/mili/node-proxy/pkg/platform"
	"time"
)

func GetSUINFTAsset(chain, tokenAddress, tokenId string) (models.NftList, error) {
	objectInfo, err := platform.GetSUINFTInfo(chain, tokenId)
	for i := 0; err != nil && i < 3; i++ {
		time.Sleep(500 * time.Microsecond)
		objectInfo, err = platform.GetSUINFTInfo(chain, tokenId)
	}
	if err != nil {
		return models.NftList{}, err
	}
	return models.NftList{
		TokenId:               tokenId,
		TokenAddress:          tokenAddress,
		Description:           objectInfo.Data.Display.Data.Description,
		ImageURL:              objectInfo.Data.Display.Data.ImageURL,
		ImageOriginalURL:      objectInfo.Data.Display.Data.ImageURL,
		Chain:                 chain,
		Network:               chain,
		CollectionName:        objectInfo.Data.Display.Data.Name,
		CollectionDescription: objectInfo.Data.Display.Data.Description,
		CollectionImageURL:    objectInfo.Data.Display.Data.ImageURL,
		NftName:               objectInfo.Data.Display.Data.Name,
		TokenType:             "SuiNFT",
	}, nil
}
