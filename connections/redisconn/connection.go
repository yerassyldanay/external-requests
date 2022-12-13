package redisconn

import (
	"fmt"

	"github.com/go-redis/redis/v8"
)

func NewRedisConnection(host string, port int32, logicalDatabase int32) (*redis.Client, error) {
	redisFeedUrl, err := redis.ParseURL(fmt.Sprintf("redis://%s:%d/%d", host, port, logicalDatabase))
	if err != nil {
		return nil, fmt.Errorf("failed to parse redis url. err: %w", err)
	}

	inMemoryStorage := redis.NewClient(redisFeedUrl)
	return inMemoryStorage, err
}
