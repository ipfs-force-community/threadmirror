package botfx

import (
	"log/slog"

	"github.com/ipfs-force-community/threadmirror/internal/bot"
	"github.com/ipfs-force-community/threadmirror/internal/config"
	"github.com/ipfs-force-community/threadmirror/internal/service"
	"go.uber.org/fx"
)

// Module provides the bot dependency injection module
var Module = fx.Module("bot",
	fx.Provide(provideTwitterBot),
	fx.Invoke(registerBotLifecycle),
)

// provideTwitterBot provides a TwitterBot instance by extracting config fields
func provideTwitterBot(
	botConfig *config.BotConfig,
	processedMentionService *service.ProcessedMentionService,
	botCookieService *service.BotCookieService,
	logger *slog.Logger,
) *bot.TwitterBot {
	return bot.NewTwitterBot(
		botConfig.Username,
		botConfig.Password,
		botConfig.Email,
		botConfig.CheckInterval,
		botConfig.MaxMentionsCheck,
		processedMentionService,
		botCookieService,
		logger,
	)
}

// registerBotLifecycle registers the bot's start and stop hooks with fx
func registerBotLifecycle(lc fx.Lifecycle, bot *bot.TwitterBot) {
	lc.Append(fx.StartStopHook(bot.Start, bot.Stop))
}
