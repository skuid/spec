package redis

import (
	"os"

	"gopkg.in/redis.v6"
)

func init() {
	redisHost := os.Getenv("REDIS_HOST")
}

//RedisClient is a *redis.Client that can be set to the same client you've created
var RedisClient *redis.Client

// NewStandardRedisClient generates a preconfigured redis client according to our spec.
func NewStandardRedisClient(options *redis.Options) *redis.Client {
	options.Addr = redisHost + ":6379"
	return redis.NewClient(options)
}
