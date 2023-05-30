package utils

import (
	"context"
	"fmt"
	v1 "gitlab.bixin.com/mili/node-proxy/api/market/v1"
	"time"
)

var marketClient v1.MarketClient

func SetMarketClient(cli v1.MarketClient) {
	marketClient = cli
}

func GetMarketClient() v1.MarketClient {
	return marketClient
}

func GetPriceByMarket(coinIds []string) (map[string]map[string]float64, error) {
	result := make(map[string]map[string]float64, len(coinIds))
	pageSize := 100
	endIndex := 0
	currencys := []string{"cny", "usd"}
	for i := 0; i < len(coinIds); i += pageSize {
		if i+pageSize > len(coinIds) {
			endIndex = len(coinIds)
		} else {
			endIndex = i + pageSize
		}
		for _, currency := range currencys {
			reply, err := marketClient.DescribeCexCoins(
				context.Background(), &v1.DescribeCexCoinsRequest{
					EventId:  fmt.Sprintf("%d", time.Now().Unix()),
					CoinIDs:  coinIds[i:endIndex],
					Currency: currency,
					PageSize: 100,
					Page:     1,
				})
			if err != nil {
				return result, err
			}
			for _, coin := range reply.Coins {
				if value, ok := result[coin.CoinID]; ok {
					value[currency] = coin.Price
					result[coin.CoinID] = value
				} else {
					result[coin.CoinID] = map[string]float64{
						currency: coin.Price,
					}
				}

			}
		}
	}
	return result, nil
}
