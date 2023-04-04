package utils

import "github.com/go-redis/redis"

var redisClient *redis.Client

func SetRedisClient(client *redis.Client) {
	redisClient = client
}

func GetRedisClient() *redis.Client {
	return redisClient
}
