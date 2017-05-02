package redis

import (
	"gopkg.in/go-redis/cache.v6"
	"gopkg.in/vmihailenco/msgpack.v2"
)

// RedisCache is a *cache.Codec that can be set to the same cache you've created
var RedisCache *cache.Codec

// NewStandardRedisCache generates a preconfigured redis cache according to our spec, using msgpack for serialization format.
func NewStandardRedisCache(redisHostEnvironmentVariableName string) (codec *cache.Codec, err error) {
	client, err := NewStandardRedisClient(redisHostEnvironmentVariableName)
	if err != nil {
		return
	}
	codec = &cache.Codec{
		Redis: client,
		Marshal: msgpack.Marshal
		Unmarshal: msgpack.Unmarshal,
	}
	return
}
