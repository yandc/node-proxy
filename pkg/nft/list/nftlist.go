package list

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	v1 "gitlab.bixin.com/mili/node-proxy/api/nft/v1"
	v12 "gitlab.bixin.com/mili/node-proxy/api/platform/v1"
	"gitlab.bixin.com/mili/node-proxy/internal/data/models"
	"gitlab.bixin.com/mili/node-proxy/pkg/nft"
	"gitlab.bixin.com/mili/node-proxy/pkg/nft/aptosNFT"
	"gitlab.bixin.com/mili/node-proxy/pkg/nft/ethNFT"
	"gitlab.bixin.com/mili/node-proxy/pkg/nft/opensea"
	"gitlab.bixin.com/mili/node-proxy/pkg/nft/types"
	"gitlab.bixin.com/mili/node-proxy/pkg/utils"
	"gorm.io/gorm/clause"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

func CreateNFTList(chain string) {
	assetList, err := GetNFTList(chain, "")
	if err != nil {
		nft.GetNFTLog().Error("get asset list error:", err)
	}
	i := 1
	AssetType2Models(chain, assetList.Assets)
	for len(assetList.Assets) >= nft.MAXLISTLIMIT {
		cursor := assetList.Next
		assetList, err = GetNFTList(chain, cursor)
		if err != nil {
			nft.GetNFTLog().Error("get asset list error:", err)
		}
		AssetType2Models(chain, assetList.Assets)
		i++
	}
}

// GetNFTList get nft list
func GetNFTList(chain, cursor string) (*types.AssetList, error) {
	//https://api.opensea.io/api/v1/assets
	var url string
	var headers map[string]string
	if strings.HasSuffix(chain, "TEST") {
		url = nft.TESTBASEURL + "assets"
	} else {
		url = nft.BASEURL + "assets"
		headers = map[string]string{
			"X-API-KEY": nft.OPENSEA_KEY,
			"accept":    "application/json",
		}
	}
	//url := nft.BASEURL + "assets"
	params := map[string]string{
		"format": "json",
		"limit":  fmt.Sprintf("%d", nft.MAXLISTLIMIT),
	}
	if cursor != "" {
		params["cursor"] = cursor
	}
	out := &types.AssetList{}
	err := utils.HttpsGet(url, params, headers, out)
	if err != nil {
		for i := 0; err != nil && i < 3; i++ {
			time.Sleep(1 * time.Second)
			err = utils.HttpsGet(url, params, headers, out)
		}
	}
	return out, err
}
func AssetType2Models(chain string, assets []types.Asset) {
	assetModel := make([]models.NftList, len(assets))
	for i := 0; i < len(assets); i++ {
		var properties string
		if len(assets[i].Traits) > 0 {
			b, _ := json.Marshal(assets[i].Traits)
			properties = string(b)
		}
		assetModel[i] = models.NftList{
			TokenId:               assets[i].TokenID,
			TokenAddress:          strings.ToLower(assets[i].AssetContract.Address),
			Name:                  assets[i].AssetContract.Name,
			Symbol:                assets[i].AssetContract.Symbol,
			TokenType:             assets[i].AssetContract.SchemaName,
			Description:           assets[i].Description,
			ImageURL:              assets[i].ImageURL,
			ImageOriginalURL:      assets[i].ImageOriginalURL,
			Chain:                 chain,
			Rarity:                "",
			Network:               "Ethereum",
			Properties:            properties,
			CollectionName:        assets[i].Collection.Name,
			CollectionSlug:        assets[i].Collection.Slug,
			CollectionDescription: assets[i].Collection.Description,
			CollectionImageURL:    assets[i].Collection.ImageURL,
			NftName:               assets[i].Name,
			AnimationURL:          assets[i].AnimationURL,
		}
	}
	if len(assetModel) > 0 {
		result := nft.GetNFTDb().Clauses(clause.OnConflict{
			UpdateAll: true,
		}).Create(&assetModel)
		if result.Error != nil {
			nft.GetNFTLog().Error("create db error:", result.Error)
		}
	}

}

// GetNFTInfo get nft info
func GetNFTInfo(chain string, tokenInfos []*v1.GetNftInfoRequest_NftInfo) ([]*v1.GetNftReply_NftInfoResp, error) {
	params := make([][]interface{}, 0, len(tokenInfos))
	addressMap := make(map[string]string, len(tokenInfos))
	for _, tokenInfo := range tokenInfos {
		address := tokenInfo.TokenAddress
		if strings.HasPrefix(tokenInfo.TokenAddress, "0x") && !strings.Contains(chain, "Aptos") {
			address = strings.ToLower(tokenInfo.TokenAddress)
		}
		params = append(params, []interface{}{address, tokenInfo.TokenId})
		addressMap[address+"_"+tokenInfo.TokenId] = tokenInfo.TokenAddress
	}
	var nftListModes []models.NftList
	err := nft.GetNFTDb().Where("(chain = ? and (token_address,token_id) IN ?) or (chain = 'Aptos' and (token_address,nft_name) IN ? )", chain, params, params).Find(&nftListModes).Error
	if err != nil {
		nft.GetNFTLog().Error("get db nft list error:", err)
		return nil, err
	}
	result := make([]*v1.GetNftReply_NftInfoResp, 0, len(tokenInfos)+2)
	for _, nftList := range nftListModes {
		address := nftList.TokenAddress
		tempId := nftList.TokenId
		if chain == "Aptos" {
			if value, ok := addressMap[nftList.TokenAddress+"_"+nftList.NftName]; ok {
				address = value
				delete(addressMap, nftList.TokenAddress+"_"+nftList.NftName)
			}
		}
		if value, ok := addressMap[nftList.TokenAddress+"_"+nftList.TokenId]; ok {
			address = value
			delete(addressMap, nftList.TokenAddress+"_"+tempId)
		}
		if nftList.ImageURL != "" && !strings.HasPrefix(nftList.ImageURL, "https") {
			nftList.ImageURL = strings.Replace(nftList.ImageURL, "ipfs://", nft.GetIPFS(), 1)
		}
		if nftList.ImageOriginalURL != "" &&
			!strings.HasPrefix(nftList.ImageOriginalURL, "https") {
			nftList.ImageOriginalURL = strings.Replace(nftList.ImageOriginalURL, "ipfs://", nft.GetIPFS(), 1)
		}

		if nftList.CollectionImageURL != "" && !strings.HasPrefix(nftList.CollectionImageURL, "https") {
			nftList.CollectionImageURL = strings.Replace(nftList.CollectionImageURL, "ipfs://", nft.GetIPFS(), 1)
		}
		if nftList.AnimationURL != "" && !strings.HasPrefix(nftList.AnimationURL, "https") {
			nftList.AnimationURL = strings.Replace(nftList.AnimationURL, "ipfs://", nft.GetIPFS(), 1)
		}
		result = append(result, &v1.GetNftReply_NftInfoResp{
			TokenId:               nftList.TokenId,
			TokenType:             nftList.TokenType,
			TokenAddress:          address,
			Name:                  nftList.Name,
			Symbol:                nftList.Symbol,
			ImageURL:              nftList.ImageURL,
			ImageOriginalURL:      nftList.ImageOriginalURL,
			Description:           nftList.Description,
			Chain:                 chain,
			CollectionName:        nftList.CollectionName,
			CollectionSlug:        nftList.CollectionSlug,
			Rarity:                nftList.Rarity,
			Network:               nftList.Network,
			Properties:            nftList.Properties,
			CollectionDescription: nftList.CollectionDescription,
			NftName:               nftList.NftName,
			CollectionImageURL:    nftList.CollectionImageURL,
			AnimationURL:          nftList.AnimationURL,
		})
	}

	if len(addressMap) > 0 {
		var wg sync.WaitGroup
		var resultErr error
		for key, value := range addressMap {
			wg.Add(1)
			go func(addressInfo, oldAddress string) {
				defer wg.Done()
				addressAndId := strings.Split(addressInfo, "_")
				if len(addressAndId) > 1 {
					address, tokenId := addressAndId[0], addressAndId[1]
					nftListModel, err := GetNFTListModel(chain, address, tokenId)
					if err != nil {
						nft.GetNFTLog().Error("get nft list model error:", err)
						resultErr = err
						return
					}
					if nftListModel.TokenId != "" {
						dbresult := nft.GetNFTDb().Clauses(clause.OnConflict{
							UpdateAll: true,
						}).Create(&nftListModel)
						if dbresult.Error != nil {
							nft.GetNFTLog().Error("create db error:", dbresult.Error)
							resultErr = err
							return
						}

						//nftListModel.TokenType = strings.ToUpper(nftListModel.TokenType)
						if nftListModel.ImageURL != "" && !strings.HasPrefix(nftListModel.ImageURL, "https") {
							nftListModel.ImageURL = strings.Replace(nftListModel.ImageURL, "ipfs://", nft.GetIPFS(), 1)
						}
						if nftListModel.ImageOriginalURL != "" &&
							!strings.HasPrefix(nftListModel.ImageOriginalURL, "https") {
							nftListModel.ImageOriginalURL = strings.Replace(nftListModel.ImageOriginalURL, "ipfs://", nft.GetIPFS(), 1)
						}

						if nftListModel.CollectionImageURL != "" && !strings.HasPrefix(nftListModel.CollectionImageURL, "https") {
							nftListModel.CollectionImageURL = strings.Replace(nftListModel.CollectionImageURL, "ipfs://", nft.GetIPFS(), 1)
						}
						if nftListModel.AnimationURL != "" && !strings.HasPrefix(nftListModel.AnimationURL, "https") {
							nftListModel.AnimationURL = strings.Replace(nftListModel.AnimationURL, "ipfs://", nft.GetIPFS(), 1)
						}
						if oldAddress == "" {
							oldAddress = nftListModel.TokenAddress
						}
						result = append(result, &v1.GetNftReply_NftInfoResp{
							TokenId:               nftListModel.TokenId,
							TokenType:             nftListModel.TokenType,
							TokenAddress:          oldAddress,
							Name:                  nftListModel.Name,
							Symbol:                nftListModel.Symbol,
							ImageURL:              nftListModel.ImageURL,
							ImageOriginalURL:      nftListModel.ImageOriginalURL,
							Description:           nftListModel.Description,
							Chain:                 chain,
							CollectionName:        nftListModel.CollectionName,
							CollectionSlug:        nftListModel.CollectionSlug,
							Rarity:                nftListModel.Rarity,
							Network:               nftListModel.Network,
							Properties:            nftListModel.Properties,
							CollectionDescription: nftListModel.CollectionDescription,
							NftName:               nftListModel.NftName,
							CollectionImageURL:    nftListModel.CollectionImageURL,
							AnimationURL:          nftListModel.AnimationURL,
						})
					}

				}
			}(key, value)
			wg.Wait()
		}
		if resultErr != nil {
			return nil, resultErr
		}
	}
	return result, nil
}

func GetNFTListModel(chain, tokenAddress, tokenId string) (models.NftList, error) {
	fullName := nft.GetFullName(chain)
	switch fullName {
	case "Ethereum":
		return ethNFT.GetETHNFTAsset(chain, tokenAddress, tokenId)
	case "Aptos":
		return aptosNFT.GetAptosNFTAsset(chain, tokenAddress, tokenId)
	case "Arbitrum", "BSC", "Polygon":
		return opensea.GetOpenSeaNFTAsset(chain, tokenAddress, tokenId)

	}
	return models.NftList{}, nil
}

func GetSingeNFTByNFTPort(tokenAddress, tokenId string) (*types.NFTPortAssetInfo, error) {
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

func NFTPort2Models(chain string, asset types.NFTPortAssetInfo) models.NftList {
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

func BuildNFTInfoFunc(chain, params string) (*v12.BuildWasmRequestReply, error) {
	var tokenInfos [][]interface{}
	tempParams, _ := hex.DecodeString(params)
	if err := json.Unmarshal(tempParams, &tokenInfos); err != nil {

	}
	var nftListModes []models.NftList
	err := nft.GetNFTDb().Where("chain = ? and (token_address,token_id) IN ?", chain, tokenInfos).Find(&nftListModes).Error
	if err != nil {
		nft.GetNFTLog().Error("get db nft list error:", err)
		return nil, err
	}
	for index, nftList := range nftListModes {
		if nftList.ImageURL != "" && !strings.HasPrefix(nftList.ImageURL, "https") {
			nftList.ImageURL = strings.Replace(nftList.ImageURL, "ipfs://", nft.GetIPFS(), 1)
		}
		if nftList.ImageOriginalURL != "" &&
			!strings.HasPrefix(nftList.ImageOriginalURL, "https") {
			nftList.ImageOriginalURL = strings.Replace(nftList.ImageOriginalURL, "ipfs://", nft.GetIPFS(), 1)
		}

		if nftList.CollectionImageURL != "" && !strings.HasPrefix(nftList.CollectionImageURL, "https") {
			nftList.CollectionImageURL = strings.Replace(nftList.CollectionImageURL, "ipfs://", nft.GetIPFS(), 1)
		}
		if nftList.AnimationURL != "" && !strings.HasPrefix(nftList.AnimationURL, "https") {
			nftList.AnimationURL = strings.Replace(nftList.AnimationURL, "ipfs://", nft.GetIPFS(), 1)
		}
		nftListModes[index] = nftList
	}
	body, err := json.Marshal(nftListModes)
	if err != nil {
		nft.GetNFTLog().Error("json marshal error:", err)
	}
	return &v12.BuildWasmRequestReply{
		Body: string(body),
	}, nil
}

func AnalysisNFTResponse(chain, params, response string) (string, error) {
	var result models.NftList
	var paramsMap map[string]string
	json.Unmarshal([]byte(params), &paramsMap)
	var oldAddress string
	if value, ok := paramsMap["oldAddress"]; ok {
		oldAddress = value
	}
	switch chain {
	case "ETH":
		var assets types.Asset
		if err := json.Unmarshal([]byte(response), &assets); err != nil {
			return "", err
		}
		result = Assets2ModesList(chain, assets)
	}
	if oldAddress != "" {
		result.TokenAddress = oldAddress
	}
	b, _ := json.Marshal(result)
	return string(b), nil
}

func Assets2ModesList(chain string, asset types.Asset) models.NftList {
	fullName := nft.GetFullName(chain)

	var properties, ratiry string
	if len(asset.Traits) > 0 {
		b, _ := json.Marshal(asset.Traits)
		properties = string(b)
	}
	if asset.RarityData != nil {
		b, _ := json.Marshal(asset.RarityData)
		ratiry = string(b)
	}
	tempModel := models.NftList{
		TokenId:               asset.TokenID,
		TokenAddress:          strings.ToLower(asset.AssetContract.Address),
		Name:                  asset.AssetContract.Name,
		Symbol:                asset.AssetContract.Symbol,
		TokenType:             strings.ToUpper(asset.AssetContract.SchemaName),
		Description:           asset.Description,
		ImageURL:              asset.ImageURL,
		ImageOriginalURL:      asset.ImageOriginalURL,
		Chain:                 chain,
		Rarity:                ratiry,
		Network:               fullName,
		Properties:            properties,
		CollectionName:        asset.Collection.Name,
		CollectionSlug:        asset.Collection.Slug,
		CollectionDescription: asset.Collection.Description,
		CollectionImageURL:    asset.Collection.ImageURL,
		NftName:               asset.Name,
		AnimationURL:          asset.AnimationURL,
	}

	dbresult := nft.GetNFTDb().Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(&tempModel)
	if dbresult.Error != nil {
		nft.GetNFTLog().Error("create db error:", dbresult.Error)
	}

	if tempModel.ImageURL != "" && !strings.HasPrefix(tempModel.ImageURL, "https") {
		tempModel.ImageURL = strings.Replace(tempModel.ImageURL, "ipfs://", nft.GetIPFS(), 1)
	}
	if tempModel.ImageOriginalURL != "" &&
		!strings.HasPrefix(tempModel.ImageOriginalURL, "https") {
		tempModel.ImageOriginalURL = strings.Replace(tempModel.ImageOriginalURL, "ipfs://", nft.GetIPFS(), 1)
	}

	if tempModel.CollectionImageURL != "" && !strings.HasPrefix(tempModel.CollectionImageURL, "https") {
		tempModel.CollectionImageURL = strings.Replace(tempModel.CollectionImageURL, "ipfs://", nft.GetIPFS(), 1)
	}
	if tempModel.AnimationURL != "" && !strings.HasPrefix(tempModel.AnimationURL, "https") {
		tempModel.AnimationURL = strings.Replace(tempModel.AnimationURL, "ipfs://", nft.GetIPFS(), 1)
	}

	return tempModel
}

func AutoUpdateNFTInfo() {
	nft.GetNFTLog().Info("AutoUpdateNFTInfo start")
	var nftListModes []models.NftList
	err := nft.GetNFTDb().Where("image_url = ? AND animation_url= ? AND refresh_count < ?", "", "", nft.GetRefreshCount()).Find(&nftListModes).Error
	if err != nil {
		nft.GetNFTLog().Error("get db nft list error:", err)
		return
	}
	if len(nftListModes) == 0 {
		return
	}
	var wg sync.WaitGroup
	var count uint64
	for _, nftList := range nftListModes {
		wg.Add(1)
		go func(model models.NftList) {
			defer wg.Done()
			tokenId := model.TokenId
			if strings.Contains(model.Chain, "Aptos") {
				tokenId = model.NftName
			}
			updateModel, err := GetNFTListModel(model.Chain, model.TokenAddress, tokenId)
			for i := 0; i < 3 && (err != nil || (updateModel.ImageURL == "" && updateModel.AnimationURL == "")); i++ {
				time.Sleep(time.Duration(i) * time.Second)
				updateModel, err = GetNFTListModel(model.Chain, model.TokenAddress, tokenId)
			}
			if err != nil {
				nft.GetNFTLog().Error("get nft list model error:", err, model.Chain, model.TokenAddress, tokenId)
				return
			}
			tempModel := model
			if updateModel.ImageURL != "" || updateModel.AnimationURL != "" {
				tempModel = updateModel
				atomic.AddUint64(&count, 1)
			}
			tempModel.ID = model.ID
			tempModel.RefreshCount = model.RefreshCount + 1
			err = nft.GetNFTDb().Save(tempModel).Error
			if err != nil {
				nft.GetNFTLog().Error("update nft info error:", err, model.Chain, model.TokenAddress, tokenId)
				return
			}
		}(nftList)
	}
	wg.Wait()
	nft.GetNFTLog().Info("AutoUpdateNFTInfo end", len(nftListModes), count)
}
