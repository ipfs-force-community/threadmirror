package servicefx

import (
	"github.com/ipfs-force-community/threadmirror/internal/repo/sqlrepo"
	"github.com/ipfs-force-community/threadmirror/internal/service"
	"go.uber.org/fx"
)

var Module = fx.Module("service",
	fx.Provide(fx.Annotate(sqlrepo.NewUserRepo, fx.As(new(service.UserRepoInterface)))),
	fx.Provide(fx.Annotate(sqlrepo.NewPostRepo, fx.As(new(service.PostRepoInterface)))),
	fx.Provide(fx.Annotate(sqlrepo.NewProcessedMentionRepo, fx.As(new(service.ProcessedMentionRepoInterface)))),
	fx.Provide(fx.Annotate(sqlrepo.NewBotCookieRepo, fx.As(new(service.BotCookieRepoInterface)))),
	fx.Provide(service.NewUserService),
	fx.Provide(service.NewPostService),
	fx.Provide(service.NewProcessedMentionService),
	fx.Provide(service.NewBotCookieService),
)
