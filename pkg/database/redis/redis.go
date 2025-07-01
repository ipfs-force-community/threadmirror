package redis

import (
	"github.com/ipfs-force-community/threadmirror/internal/config"
	"github.com/redis/go-redis/v9"
)

// Client wraps go-redis Client for unified usage
// 统一 Redis 客户端封装
// 直接暴露 *redis.Client 以便调用原生方法

type Client struct {
	*redis.Client
}

// NewClient creates a new Redis client from config.RedisConfig
func NewClient(cfg *config.RedisConfig) *Client {
	opt := &redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	}
	return &Client{redis.NewClient(opt)}
}
