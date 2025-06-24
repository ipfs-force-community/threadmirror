package main

import (
	"context"

	"github.com/ipfs-force-community/threadmirror/i18n"
	"github.com/ipfs-force-community/threadmirror/internal/api/apifx"
	"github.com/ipfs-force-community/threadmirror/internal/bot/botfx"
	"github.com/ipfs-force-community/threadmirror/internal/config"
	"github.com/ipfs-force-community/threadmirror/internal/model"
	"github.com/ipfs-force-community/threadmirror/internal/service/servicefx"
	"github.com/ipfs-force-community/threadmirror/pkg/auth/authfx"
	"github.com/ipfs-force-community/threadmirror/pkg/database/sql"
	"github.com/ipfs-force-community/threadmirror/pkg/database/sql/sqlfx"
	"github.com/ipfs-force-community/threadmirror/pkg/i18n/i18nfx"
	"github.com/ipfs-force-community/threadmirror/pkg/log/logfx"
	"github.com/ipfs-force-community/threadmirror/pkg/util"
	"github.com/urfave/cli/v2"
	"go.uber.org/fx"
)

var ServerCommand = &cli.Command{
	Name:  "server",
	Usage: "Start the HTTP server",
	Flags: util.MergeSlices(
		config.GetServerCLIFlags(),
		config.GetDatabaseCLIFlags(),
		config.GetSupabaseCLIFlags(),
		config.GetBotCLIFlags(),
	),
	Action: func(c *cli.Context) error {
		serverConf := config.LoadServerConfigFromCLI(c)
		dbConf := config.LoadDatabaseConfigFromCLI(c)
		supabaseConf := config.LoadSupabaseConfigFromCLI(c)
		botConf := config.LoadBotConfigFromCLI(c)
		debug := serverConf.Debug

		fxApp := fx.New(
			// Provide the configuration
			fx.Supply(serverConf),
			fx.Supply(supabaseConf),
			fx.Supply(botConf),
			fx.Supply(&logfx.Config{
				Level:      c.String("log-level"),
				LogDevMode: debug,
			}),
			fx.Supply(&sqlfx.Config{
				Driver: dbConf.Driver,
				DSN:    dbConf.DSN,
			}),
			logfx.Module,
			sqlfx.Module,
			apifx.Module,
			servicefx.Module,
			i18nfx.Module(&i18n.LocaleFS),
			authfx.Module([]byte(supabaseConf.JWTKey)),
			botfx.Module,
			fx.Invoke(func(lc fx.Lifecycle, db *sql.DB) {
				if debug {
					lc.Append(fx.StartHook(migrateFn(db, dbConf.Driver)))
				}
			}),
		)
		fxApp.Run()
		return nil
	},
}

func migrateFn(db *sql.DB, dbDriver string) func(context.Context) error {
	return func(ctx context.Context) error {
		return db.Migrate(ctx, model.AllModels())
	}
}
