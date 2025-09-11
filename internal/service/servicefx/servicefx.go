package servicefx

import (
	"github.com/ipfs-force-community/threadmirror/internal/service"
	"go.uber.org/fx"
)

var Module = fx.Module("service",
	fx.Provide(service.NewMentionService),
	fx.Provide(service.NewProcessedMarkService),
	fx.Provide(service.NewBotCookieService),
	fx.Provide(service.NewThreadService),
)
