package redis

import (
	"os"

	"gopkg.in/go-redis/cache.v6"
	"gopkg.in/redis.v6"
	"gopkg.in/vmihailenco/msgpack.v2"
)

func init() {
	redisHost := os.Getenv("REDIS_HOST")
}

// NewStandardRedisClient generates a preconfigured redis client according to our spec.
// Accepts an *redis.Options object, and overrides the Addr field to use `$REDIS_HOST:6379` instead
func NewStandardRedisClient(options *redis.Options) *redis.Client {
	options.Addr = redisHost + ":6379"
	return redis.NewClient(options)
}

// NewStandardRedisCache generates a preconfigured redis cache according to our spec, using msgpack for serialization format.
// Accepts an *redis.Options object, and overrides the Addr field to use `$REDIS_HOST:6379` instead
func NewStandardRedisCache(options *redis.Options) codec *cache.Codec {
	return &cache.Codec{
		Redis: NewStandardRedisClient(options),
		Marshal: msgpack.Marshal
		Unmarshal: msgpack.Unmarshal,
	}
}