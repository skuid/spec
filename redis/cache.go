package redis

import (
	"github.com/go-redis/cache"
	"gopkg.in/vmihailenco/msgpack.v2"
)

// RedisCache is a *cache.Codec that can be set to the same cache you've created
var RedisCache *cache.Codec

// NewStandardRedisCache generates a preconfigured redis cache according to our spec.
func NewStandardRedisCache(redisHostEnvironmentVariableName string) (codec *cache.Codec, err error) {
	client, err := NewStandardRedisClient(redisHostEnvironmentVariableName)
	if err != nil {
		return
	}
	codec = &cache.Codec{
		Redis: client,
		Marshal: func(v interface{}) ([]byte, error) {
			return msgpack.Marshal(v)
		},
		Unmarshal: func(b []byte, v interface{}) error {
			return msgpack.Unmarshal(b, v)
		},
	}
	return
}
