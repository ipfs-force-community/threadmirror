package servicefx

import (
	"go.uber.org/fx"

	"github.com/ipfs-force-community/threadmirror/internal/repo/sqlrepo"
	"github.com/ipfs-force-community/threadmirror/internal/service"
)

var Module = fx.Module("servicefx",
	// Repositories
	fx.Provide(fx.Annotate(sqlrepo.NewMentionRepo, fx.As(new(service.MentionRepoInterface)))),
	fx.Provide(fx.Annotate(sqlrepo.NewProcessedMarkRepo, fx.As(new(service.ProcessedMarkRepoInterface)))),
	fx.Provide(fx.Annotate(sqlrepo.NewBotCookieRepo, fx.As(new(service.BotCookieRepoInterface)))),
	fx.Provide(fx.Annotate(sqlrepo.NewThreadRepo, fx.As(new(service.ThreadRepoInterface)))),

	// Services
	fx.Provide(service.NewMentionService),
	fx.Provide(service.NewProcessedMarkService),
	fx.Provide(service.NewBotCookieService),
	fx.Provide(service.NewThreadService),
	fx.Provide(service.NewTranslationService),
)
