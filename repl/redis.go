package repl

import (
	"github.com/go-redis/redis"
)

var redisClient *redis.Client

func initClientPool() {
	redisClient = redis.NewClient(options.RedisOptions)
	if _, err := redisClient.Ping().Result(); err != nil {
		panic(err)
	}
}

func getRedisClient() (*redis.Client, error) {
	if _, err := redisClient.Ping().Result(); err != nil {
		return nil, err
	}
	return redisClient, nil
}
