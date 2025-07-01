package servicefx

import (
	"github.com/ipfs-force-community/threadmirror/internal/repo/sqlrepo"
	"github.com/ipfs-force-community/threadmirror/internal/service"
	"go.uber.org/fx"
)

var Module = fx.Module("service",
	fx.Provide(fx.Annotate(sqlrepo.NewMentionRepo, fx.As(new(service.MentionRepoInterface)))),
	fx.Provide(fx.Annotate(sqlrepo.NewProcessedMentionRepo, fx.As(new(service.ProcessedMentionRepoInterface)))),
	fx.Provide(fx.Annotate(sqlrepo.NewBotCookieRepo, fx.As(new(service.BotCookieRepoInterface)))),
	fx.Provide(sqlrepo.NewThreadRepo),
	fx.Provide(service.NewMentionService),
	fx.Provide(service.NewProcessedMentionService),
	fx.Provide(service.NewBotCookieService),
)
