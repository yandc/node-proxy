package utils

import (
	"context"
	"fmt"
	v1 "gitlab.bixin.com/mili/node-proxy/api/market/v1"
	nftmarketplacev1 "gitlab.bixin.com/mili/node-proxy/api/nft-marketplace/v1"
	nftmarketplacev2 "gitlab.bixin.com/mili/node-proxy/api/nft-marketplace/v2"
	"sort"
	"time"
)

var marketClient v1.MarketClient
var nftApiClient nftmarketplacev1.NFTApiClient
var collectionApiClient nftmarketplacev2.CollectionApiClient

func SetMarketClient(cli v1.MarketClient) {
	marketClient = cli
}
func GetMarketClient() v1.MarketClient {
	return marketClient
}

func SetNFTApiClient(cli nftmarketplacev1.NFTApiClient) {
	nftApiClient = cli
}

func GetNFTApiClient() nftmarketplacev1.NFTApiClient {
	return nftApiClient
}

func SetCollectionApiClient(cli nftmarketplacev2.CollectionApiClient) {
	collectionApiClient = cli
}

func GetCollectionApiClient() nftmarketplacev2.CollectionApiClient {
	return collectionApiClient
}

func GetPriceByMarket(coinIds []string) (map[string]map[string]float64, error) {
	result := make(map[string]map[string]float64, len(coinIds))
	pageSize := 100
	endIndex := 0
	for i := 0; i < len(coinIds); i += pageSize {
		if i+pageSize > len(coinIds) {
			endIndex = len(coinIds)
		} else {
			endIndex = i + pageSize
		}
		reply, err := marketClient.DescribeCoinsByFields(
			context.Background(), &v1.DescribeCoinsByFieldsRequest{
				EventId: fmt.Sprintf("%d", time.Now().Unix()),
				CoinIDs: coinIds[i:endIndex],
				Fields:  []string{"price"},
			})
		if err != nil {
			return result, err
		}
		for _, coin := range reply.Coins {
			if coin.Price != nil {
				result[coin.CoinID] = map[string]float64{
					"cny": coin.Price.Cny,
					"usd": coin.Price.Usd,
				}
			}
		}
	}
	return result, nil
}

func GetPriceByMarketToken(tokenAddress []*v1.Tokens) (map[string]map[string]float64, error) {
	result := make(map[string]map[string]float64, len(tokenAddress))
	pageSize := 100
	endIndex := 0
	for i := 0; i < len(tokenAddress); i += pageSize {
		if i+pageSize > len(tokenAddress) {
			endIndex = len(tokenAddress)
		} else {
			endIndex = i + pageSize
		}
		reply, err := marketClient.DescribePriceByCoinAddress(
			context.Background(), &v1.DescribePriceByCoinAddressRequest{
				EventId: fmt.Sprintf("%d", time.Now().Unix()),
				Tokens:  tokenAddress[i:endIndex],
			})
		if err != nil {
			return result, err
		}
		for _, coin := range reply.Tokens {
			if coin.Price != nil {
				handler := GetHandlerByChain(coin.Chain)
				result[handler+"_"+coin.Address] = map[string]float64{
					"cny": coin.Price.Cny,
					"usd": coin.Price.Usd,
				}
			}
		}
	}
	return result, nil
}

func GetCoinInfoByMarket(chain string) ([]*v1.DescribeCoinsByChainReply_Coin, error) {
	pageSize := 300
	page := 1
	result := make([]*v1.DescribeCoinsByChainReply_Coin, 0, pageSize)
	for ; ; page++ {
		reply, err := marketClient.DescribeCoinsByChain(context.Background(), &v1.DescribeCoinsByChainRequest{
			EventId:  fmt.Sprintf("%d", time.Now().Unix()),
			Chain:    chain,
			Page:     int32(page),
			PageSize: int32(pageSize),
		})
		if err != nil {
			return nil, err
		}
		result = append(result, reply.Coins...)
		if len(reply.Coins) < pageSize {
			return result, nil
		}
	}
	return nil, nil
}

func GetTopNByMarkets(ids []string, topN int) []*v1.DescribeCexCoinsReply_Coin {
	markets := make([]*v1.DescribeCexCoinsReply_Coin, 0, len(ids)+2)
	pageSize := 500
	endIndex := 0
	for i := 0; i < len(ids); i += pageSize {
		if i+pageSize > len(ids) {
			endIndex = len(ids)
		} else {
			endIndex = i + pageSize
		}

		reply, err := marketClient.DescribeCexCoins(context.Background(), &v1.DescribeCexCoinsRequest{
			EventId:   fmt.Sprintf("%d", time.Now().Unix()),
			CoinIDs:   ids[i:endIndex],
			Currency:  "cny",
			PageSize:  int32(topN),
			Page:      1,
			SortField: v1.DescribeCexCoinsRequest_MarketCap,
		})

		if err != nil {
			fmt.Errorf("get cg markets error:%v", err)
			continue
		}

		markets = append(markets, reply.Coins...)
	}
	//sort markets
	sort.Slice(markets, func(i, j int) bool {
		return markets[i].Rank <= markets[j].Rank
	})
	//index := 0
	//for ; index < len(markets) && markets[index].Rank == 0; index++ {
	//}
	//markets = append(markets[index:], markets[:index]...)
	if len(markets) > topN {
		markets = markets[:topN]
	}
	return markets
}
