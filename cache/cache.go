package cache

import (
	"time"

	"github.com/go-redis/redis"
)

var (
	client redis.Cmdable
)

func newPool(address string, maxConnections int) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:         address,
		MaxRetries:   3,
		PoolSize:     maxConnections,
		MinIdleConns: 3,
		IdleTimeout:  240 * time.Second,
	})
}

// GetConnection returns a connection from the Redis pool
// which should have been created via Start on server startup
func GetConnection() *redis.Client {
	return client.(*redis.Client)
}

// SetConnection allows the user to explicitly set the client. This is useful
// for mocking mostly
func SetConnection(conn redis.Cmdable) {
	client = conn
}

// Start should be called once at server startup to initialize a pool
// for connections to Redis.
func Start(redisAddress string, maxConnections int) {
	client = newPool(redisAddress, maxConnections)
}

// Get returns the value of a single string key in cache
func Get(key string) (string, error) {

	get := client.Get(key)

	if get.Err() != nil {
		return "", get.Err()
	}

	return get.Val(), nil
}

// GetMap returns an object of all values stored in a cache value that is a hash map
func GetMap(key string) (map[string]string, error) {
	value, err := client.HGetAll(key).Result()

	if err != nil {
		return nil, err
	}

	return value, nil
}

// Set populates the value of a single string key in cache,
// and sets an expiration for the cache key (in seconds).
func Set(key string, value string, expirationSeconds time.Duration) (interface{}, error) {
	set := client.Set(key, value, expirationSeconds)

	if set.Err() != nil {
		return nil, set.Err()
	}

	return set.Val(), nil

}

// SetMap takes a map of key-value pairs and populates this in a hash-map cache value,
// and sets an expiration for the cache key (in seconds).
func SetMap(key string, obj map[string]string, expiration time.Duration) (interface{}, error) {

	imap := make(map[string]interface{})
	for k, v := range obj {
		imap[k] = v
	}

	pipe := client.Pipeline()

	pipe.HMSet(key, imap)
	pipe.Expire(key, expiration)

	return pipe.Exec()
}

func SetMapInterface(key string, obj map[string]interface{}, expiration time.Duration) (interface{}, error) {
	pipe := client.Pipeline()

	pipe.HMSet(key, obj)
	pipe.Expire(key, expiration)

	return pipe.Exec()
}
