package main

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/urfave/cli/v2"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"

	"github.com/ipfs-force-community/threadmirror/i18n"
	"github.com/ipfs-force-community/threadmirror/internal/bot/botfx"
	"github.com/ipfs-force-community/threadmirror/internal/config"
	"github.com/ipfs-force-community/threadmirror/internal/job/jobfx"
	"github.com/ipfs-force-community/threadmirror/internal/service"
	"github.com/ipfs-force-community/threadmirror/internal/service/servicefx"
	"github.com/ipfs-force-community/threadmirror/pkg/database/redis"
	"github.com/ipfs-force-community/threadmirror/pkg/database/redis/redisfx"
	"github.com/ipfs-force-community/threadmirror/pkg/database/sql/sqlfx"
	"github.com/ipfs-force-community/threadmirror/pkg/i18n/i18nfx"
	"github.com/ipfs-force-community/threadmirror/pkg/ipfs/ipfsfx"
	"github.com/ipfs-force-community/threadmirror/pkg/llm/llmfx"
	"github.com/ipfs-force-community/threadmirror/pkg/log/logfx"
	"github.com/ipfs-force-community/threadmirror/pkg/util"
	"github.com/ipfs-force-community/threadmirror/pkg/xscraper"
	"github.com/ipfs-force-community/threadmirror/pkg/xscraper/xscraperfx"
)

var BotCommand = &cli.Command{
	Name:  "bot",
	Usage: "Start the bot",
	Flags: util.MergeSlices(
		config.GetCommonCLIFlags(),
		config.GetDatabaseCLIFlags(),
		config.GetRedisCLIFlags(),
		config.GetBotCLIFlags(),
		config.GetLLMCLIFlags(),
		config.GetIPFSCLIFlags(),
	),
	Action: func(c *cli.Context) error {
		commonConfig := config.LoadCommonConfigFromCLI(c)
		dbConf := config.LoadDatabaseConfigFromCLI(c)
		redisConf := config.LoadRedisConfigFromCLI(c)
		botConf := config.LoadBotConfigFromCLI(c)
		llmConf := config.LoadLLMConfigFromCLI(c)
		ipfsConf := config.LoadIPFSConfigFromCLI(c)

		baseContext, cancel := context.WithCancel(context.Background())

		fxApp := fx.New(
			fx.Supply(baseContext),
			fx.Invoke(func(ctx context.Context, lc fx.Lifecycle) {
				lc.Append(fx.StopHook(cancel))
			}),
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
			jobfx.Module,
			fx.Provide(func(s *service.BotCookieService) xscraper.LoginOptions {
				return xscraper.LoginOptions{
					Username:          botConf.Username,
					Password:          botConf.Password,
					Email:             botConf.Email,
					APIKey:            botConf.APIKey,
					APIKeySecret:      botConf.APIKeySecret,
					AccessToken:       botConf.AccessToken,
					AccessTokenSecret: botConf.AccessTokenSecret,
					LoadCookies: func(ctx context.Context) ([]*http.Cookie, error) {
						return s.LoadCookies(ctx, botConf.Email, botConf.Username)
					},
					SaveCookies: func(ctx context.Context, cookies []*http.Cookie) error {
						return s.SaveCookies(ctx, botConf.Email, botConf.Username, cookies)
					},
				}
			}),
			llmfx.Module,
			ipfsfx.Module,
			xscraperfx.Module,
			i18nfx.Module(&i18n.LocaleFS),
			botfx.Module(botConf.Enable),
			fx.WithLogger(func(logger *slog.Logger) fxevent.Logger {
				return &fxevent.SlogLogger{Logger: logger}
			}),
		)
		fxApp.Run()
		return nil
	},
}
