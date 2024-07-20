package cache

import (
	"github.com/go-redis/redis"
)

type cache struct {
	redis *redis.Client
}

func New(redis *redis.Client) Cache {
	return &cache{
		redis: redis,
	}
}

type Cache interface {
}
