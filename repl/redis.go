package repl

import (
	"context"

	"github.com/redis/go-redis/v9"
)

var redisClient *redis.Client

func initClientPool() {
	redisClient = redis.NewClient(options.RedisOptions)
	if _, err := redisClient.Ping(context.Background()).Result(); err != nil {
		panic(err)
	}
}

func getRedisClient() (*redis.Client, error) {
	if _, err := redisClient.Ping(context.Background()).Result(); err != nil {
		return nil, err
	}
	return redisClient, nil
}
