package redisfx

import (
	"go.uber.org/fx"

	"github.com/ipfs-force-community/threadmirror/pkg/database/redis"
)

var Module = fx.Provide(redis.NewClient)
