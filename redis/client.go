package redis

import (
	"os"

	"fmt"

	"github.com/go-redis/redis"
)

//RedisClient is a *redis.Client that can be set to the same client you've created
var RedisClient *redis.Client

// NewStandardRedisClient generates a preconfigured redis client according to our spec.
func NewStandardRedisClient(redisHostEnvironmentVariableName string) (client *redis.Client, err error) {
	if redisHostEnvironmentVariableName == "" {
		redisHostEnvironmentVariableName = "REDIS_HOST"
	}

	redisHost, present := os.LookupEnv("REDIS_HOST")
	if !present {
		err = fmt.Errorf("environment variable %s must be set in order to retrieve the redis hostname", redisHostEnvironmentVariableName)
		return
	}

	client = redis.NewClient(&redis.Options{
		Addr:     redisHost + ":6379",
		Password: "",
		DB:       0,
	})
	return
}
