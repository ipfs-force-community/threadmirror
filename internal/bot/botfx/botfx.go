package botfx

import (
	"log/slog"

	"github.com/ipfs-force-community/threadmirror/internal/bot"
	"github.com/ipfs-force-community/threadmirror/internal/config"
	"github.com/ipfs-force-community/threadmirror/pkg/jobq"
	"github.com/ipfs-force-community/threadmirror/pkg/xscraper"
	"go.uber.org/fx"
)

// Module provides the bot dependency injection module
func Module(startBot bool) fx.Option {
	opts := []fx.Option{
		fx.Provide(provideTwitterBot),
	}

	if startBot {
		opts = append(opts, fx.Invoke(registerBotLifecycle))
	}

	return fx.Module("bot", opts...)
}

// provideTwitterBot provides a TwitterBot instance by extracting config fields
func provideTwitterBot(
	botConfig *config.BotConfig,
	scraper *xscraper.XScraper,
	jobQueueClient jobq.JobQueueClient,
	logger *slog.Logger,
) *bot.TwitterBot {
	return bot.NewTwitterBot(
		botConfig.Username,
		botConfig.Email,
		scraper,
		botConfig.CheckInterval,
		botConfig.MaxMentionsCheck,
		jobQueueClient,
		logger,
	)
}

// registerBotLifecycle registers the bot's start and stop hooks with fx
func registerBotLifecycle(lc fx.Lifecycle, bot *bot.TwitterBot) {
	lc.Append(fx.StartStopHook(bot.Start, bot.Stop))
}
