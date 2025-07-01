package apifx

import (
	"github.com/ipfs-force-community/threadmirror/internal/api"
	v1 "github.com/ipfs-force-community/threadmirror/internal/api/v1"
	"go.uber.org/fx"
)

var Module = fx.Module("api",
	fx.Provide(
		api.NewServer,
	),
	fx.Provide(v1.NewV1Handler),
	fx.Invoke(func(lc fx.Lifecycle, server *api.Server) {
		lc.Append(fx.StartStopHook(server.Start, server.Stop))
	}),
)
