package redisfx

import (
	"go.uber.org/fx"

	redisv9 "github.com/redis/go-redis/v9"
)

// RedisConfig holds Redis configuration for Redis Client
// 避免依赖 internal/config
// 可由上层自行转换
type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

// Module 提供给 fx 使用的 Redis Client 构造器
var Module = fx.Provide(
	func(cfg *RedisConfig) *redisv9.Client {
		return redisv9.NewClient(&redisv9.Options{
			Addr:     cfg.Addr,
			Password: cfg.Password,
			DB:       cfg.DB,
		})
	},
)
