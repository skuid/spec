package redis

import (
	"gopkg.in/go-redis/cache.v6"
	"gopkg.in/redis.v6"
	"gopkg.in/vmihailenco/msgpack.v2"
)

// NewStandardRedisCache generates a preconfigured redis cache according to our spec, using msgpack for serialization format.
func NewStandardRedisCache(options *redis.Options) codec *cache.Codec {
	return &cache.Codec{
		Redis: NewStandardRedisClient(options),
		Marshal: msgpack.Marshal
		Unmarshal: msgpack.Unmarshal,
	}
}
