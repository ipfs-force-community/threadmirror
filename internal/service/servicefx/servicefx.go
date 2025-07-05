package servicefx

import (
	"github.com/ipfs-force-community/threadmirror/internal/repo/sqlrepo"
	"github.com/ipfs-force-community/threadmirror/internal/service"
	"go.uber.org/fx"
)

var Module = fx.Module("service",
	fx.Provide(fx.Annotate(sqlrepo.NewMentionRepo, fx.As(new(service.MentionRepoInterface)))),
	fx.Provide(fx.Annotate(sqlrepo.NewProcessedMarkRepo, fx.As(new(service.ProcessedMarkRepoInterface)))),
	fx.Provide(fx.Annotate(sqlrepo.NewBotCookieRepo, fx.As(new(service.BotCookieRepoInterface)))),
	fx.Provide(fx.Annotate(sqlrepo.NewThreadRepo, fx.As(new(service.ThreadRepoInterface)))),
	fx.Provide(service.NewMentionService),
	fx.Provide(service.NewProcessedMarkService),
	fx.Provide(service.NewBotCookieService),
	fx.Provide(service.NewThreadService),
)
