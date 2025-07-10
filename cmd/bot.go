package main

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/urfave/cli/v2"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"

	"github.com/ipfs-force-community/threadmirror/i18n"
	"github.com/ipfs-force-community/threadmirror/internal/config"
	"github.com/ipfs-force-community/threadmirror/internal/service"
	"github.com/ipfs-force-community/threadmirror/internal/service/servicefx"
	"github.com/ipfs-force-community/threadmirror/internal/task/cron/cronfx"
	"github.com/ipfs-force-community/threadmirror/internal/task/queue/queuefx"
	"github.com/ipfs-force-community/threadmirror/pkg/database/redis"
	"github.com/ipfs-force-community/threadmirror/pkg/database/redis/redisfx"
	"github.com/ipfs-force-community/threadmirror/pkg/database/sql/sqlfx"
	"github.com/ipfs-force-community/threadmirror/pkg/i18n/i18nfx"
	"github.com/ipfs-force-community/threadmirror/pkg/ipfs/ipfsfx"
	"github.com/ipfs-force-community/threadmirror/pkg/jobq/jobqfx"
	"github.com/ipfs-force-community/threadmirror/pkg/llm/llmfx"
	"github.com/ipfs-force-community/threadmirror/pkg/log/logfx"
	"github.com/ipfs-force-community/threadmirror/pkg/util"
	"github.com/ipfs-force-community/threadmirror/pkg/xscraper"
	"github.com/ipfs-force-community/threadmirror/pkg/xscraper/xscraperfx"
)

var BotCommand = &cli.Command{
	Name:  "bot",
	Usage: "Start the bot (now runs as cron tasks)",
	Flags: util.MergeSlices(
		config.GetCommonCLIFlags(),
		config.GetDatabaseCLIFlags(),
		config.GetRedisCLIFlags(),
		config.GetBotCLIFlags(),
		config.GetCronCLIFlags(),
		config.GetLLMCLIFlags(),
		config.GetIPFSCLIFlags(),
	),
	Action: func(c *cli.Context) error {
		commonConfig := config.LoadCommonConfigFromCLI(c)
		dbConf := config.LoadDatabaseConfigFromCLI(c)
		redisConf := config.LoadRedisConfigFromCLI(c)
		botConf := config.LoadBotConfigFromCLI(c)
		cronConf := config.LoadCronConfigFromCLI(c)
		llmConf := config.LoadLLMConfigFromCLI(c)
		ipfsConf := config.LoadIPFSConfigFromCLI(c)

		fxApp := fx.New(
			// Provide the configuration
			fx.Supply(commonConfig),
			fx.Supply(&redis.RedisConfig{
				Addr:     redisConf.Addr,
				Password: redisConf.Password,
				DB:       redisConf.DB,
			}),
			fx.Supply(llmConf),
			fx.Supply(ipfsConf),
			fx.Supply(botConf),
			fx.Supply(cronConf),
			fx.Supply(&logfx.Config{
				Level:      c.String("log-level"),
				LogDevMode: commonConfig.Debug,
			}),
			fx.Supply(&sqlfx.Config{
				Driver: dbConf.Driver,
				DSN:    dbConf.DSN,
			}),
			logfx.Module,
			sqlfx.Module,
			redisfx.Module,
			servicefx.Module,
			jobqfx.ModuleClient,
			jobqfx.ModuleServer,
			queuefx.Module,
			cronfx.Module,
			fx.Provide(func(s *service.BotCookieService) []xscraper.LoginOptions {
				loginOpts := make([]xscraper.LoginOptions, len(botConf.Credentials))
				for i, cred := range botConf.Credentials {
					credCopy := cred // capture loop variable
					loginOpts[i] = xscraper.LoginOptions{
						Username:          credCopy.Username,
						Password:          credCopy.Password,
						Email:             credCopy.Email,
						APIKey:            credCopy.APIKey,
						APIKeySecret:      credCopy.APIKeySecret,
						AccessToken:       credCopy.AccessToken,
						AccessTokenSecret: credCopy.AccessTokenSecret,
						LoadCookies: func(ctx context.Context) ([]*http.Cookie, error) {
							return s.LoadCookies(ctx, credCopy.Email, credCopy.Username)
						},
						SaveCookies: func(ctx context.Context, cookies []*http.Cookie) error {
							return s.SaveCookies(ctx, credCopy.Email, credCopy.Username, cookies)
						},
					}
				}
				return loginOpts
			}),
			llmfx.Module,
			ipfsfx.Module,
			xscraperfx.Module,
			i18nfx.Module(&i18n.LocaleFS),
			fx.WithLogger(func(logger *slog.Logger) fxevent.Logger {
				return &fxevent.SlogLogger{Logger: logger}
			}),
		)
		fxApp.Run()
		return nil
	},
}
