package main

import (
	"context"
	"log/slog"

	"github.com/urfave/cli/v2"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"

	"github.com/ipfs-force-community/threadmirror/i18n"
	"github.com/ipfs-force-community/threadmirror/internal/api/apifx"
	"github.com/ipfs-force-community/threadmirror/internal/config"
	"github.com/ipfs-force-community/threadmirror/internal/service/servicefx"
	"github.com/ipfs-force-community/threadmirror/pkg/auth/authfx"
	"github.com/ipfs-force-community/threadmirror/pkg/database/redis"
	"github.com/ipfs-force-community/threadmirror/pkg/database/redis/redisfx"
	"github.com/ipfs-force-community/threadmirror/pkg/database/sql/sqlfx"
	"github.com/ipfs-force-community/threadmirror/pkg/i18n/i18nfx"
	"github.com/ipfs-force-community/threadmirror/pkg/ipfs/ipfsfx"
	"github.com/ipfs-force-community/threadmirror/pkg/llm/llmfx"
	"github.com/ipfs-force-community/threadmirror/pkg/log/logfx"
	"github.com/ipfs-force-community/threadmirror/pkg/util"
	"github.com/ipfs-force-community/threadmirror/pkg/xscraper/xscraperfx"
)

var ServerCommand = &cli.Command{
	Name:  "server",
	Usage: "Start the HTTP server",
	Flags: util.MergeSlices(
		config.GetCommonCLIFlags(),
		config.GetServerCLIFlags(),
		config.GetDatabaseCLIFlags(),
		config.GetRedisCLIFlags(),
		config.GetAuth0CLIFlags(),
		config.GetLLMCLIFlags(),
		config.GetIPFSCLIFlags(),
	),
	Action: func(c *cli.Context) error {
		commonConfig := config.LoadCommonConfigFromCLI(c)
		serverConf := config.LoadServerConfigFromCLI(c)
		dbConf := config.LoadDatabaseConfigFromCLI(c)
		redisConf := config.LoadRedisConfigFromCLI(c)
		debug := c.Bool("debug")
		auth0Conf := config.LoadAuth0ConfigFromCLI(c)
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
			fx.Supply(serverConf),
			fx.Supply(&redis.RedisConfig{
				Addr:     redisConf.Addr,
				Password: redisConf.Password,
				DB:       redisConf.DB,
			}),
			fx.Supply(&logfx.Config{
				Level:      c.String("log-level"),
				LogDevMode: debug,
			}),
			fx.Supply(&sqlfx.Config{
				Driver: dbConf.Driver,
				DSN:    dbConf.DSN,
			}),
			fx.Supply(llmConf),
			fx.Supply(ipfsConf),
			logfx.Module,
			sqlfx.Module,
			redisfx.Module,
			apifx.Module,
			servicefx.Module,
			llmfx.Module,
			ipfsfx.Module,
			xscraperfx.Module,
			i18nfx.Module(&i18n.LocaleFS),
			authfx.ModuleAuth0(auth0Conf),
			fx.WithLogger(func(logger *slog.Logger) fxevent.Logger {
				return &fxevent.SlogLogger{Logger: logger}
			}),
		)
		fxApp.Run()
		return nil
	},
}
