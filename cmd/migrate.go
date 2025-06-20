package main

import (
	"fmt"

	"github.com/ipfs-force-community/threadmirror/internal/config"
	"github.com/ipfs-force-community/threadmirror/internal/model"
	"github.com/ipfs-force-community/threadmirror/pkg/database/sql"
	"github.com/ipfs-force-community/threadmirror/pkg/log"
	"github.com/samber/lo"
	"github.com/urfave/cli/v2"
)

var MigrateCommand = &cli.Command{
	Name:  "migrate",
	Usage: "Run database migrations",
	Flags: config.GetDatabaseCLIFlags(),
	Action: func(c *cli.Context) error {
		// Load database configuration from CLI context
		dbConfig := config.LoadDatabaseConfigFromCLI(c)
		logger := lo.Must(log.New(c.String("log-level"), c.Bool("debug")))
		defer logger.Close() // nolint:errcheck

		logger.Info(
			"Starting database migration...",
			"driver",
			dbConfig.Driver,
			"dsn",
			dbConfig.DSN,
		)

		// Create database connection
		db, err := sql.New(dbConfig.Driver, dbConfig.DSN, logger.Logger)
		if err != nil {
			return fmt.Errorf("failed to connect to database: %w", err)
		}
		defer db.Close() // nolint:errcheck

		logger.Info("Connected to database successfully")

		// Run auto migration for all models and database functions
		logger.Info("Running auto migration...")
		if err := db.Migrate(c.Context, model.AllModels()); err != nil {
			return fmt.Errorf("failed to run migration: %w", err)
		}

		logger.Info("Database migration completed successfully")
		return nil
	},
}
