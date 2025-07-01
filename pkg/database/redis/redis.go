package redis

import (
	"github.com/redis/go-redis/v9"
)

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}
type Client struct {
	*redis.Client
}

// NewClient creates a new Redis client from config.RedisConfig
func NewClient(cfg *RedisConfig) *Client {
	opt := &redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	}
	return &Client{redis.NewClient(opt)}
}
