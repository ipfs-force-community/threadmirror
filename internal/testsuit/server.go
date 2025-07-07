package testsuit

import (
	"context"
	"fmt"
	"time"

	apifx "github.com/ipfs-force-community/threadmirror/internal/api/apifx"
	"github.com/ipfs-force-community/threadmirror/internal/config"
	servicefx "github.com/ipfs-force-community/threadmirror/internal/service/servicefx"
	sqlfx "github.com/ipfs-force-community/threadmirror/pkg/database/sql/sqlfx"
	logfx "github.com/ipfs-force-community/threadmirror/pkg/log/logfx"
	"go.uber.org/fx"
)

// NewTestServer creates a new fx.App for testing purposes.
// It allows overriding specific modules or providing additional dependencies.
func NewTestServer(ctx context.Context, opts ...fx.Option) (*fx.App, error) {
	// Default test configuration
	serverCfg := &config.ServerConfig{
		Addr: "localhost:8080", // Use Addr instead of Port
	}
	dbCfg := &config.DatabaseConfig{
		DSN: "sqlite://:memory:?_foreign_keys=on", // Use in-memory SQLite for tests
	}
	botCfg := &config.BotConfig{
		Credentials: []config.BotCredential{{
			Username: "testbot",
			Password: "testpass",
			Email:    "test@example.com",
		}},
		CheckInterval: 60 * time.Second, // Use CheckInterval
	}

	// Base modules for the test server
	baseModules := []fx.Option{
		fx.NopLogger,
		fx.Supply(serverCfg, dbCfg, botCfg), // Supply individual config structs
		logfx.Module,
		sqlfx.Module,
		servicefx.Module,
		apifx.Module,
		// i18n.Module 需要 localeFS, 测试可不注入或 mock
		// auth.Module 需要 JWTKey, 测试可不注入或 mock
	}

	// Combine base modules with provided options
	allOpts := append(baseModules, opts...)

	app := fx.New(allOpts...)

	// Start the app with a context that can be cancelled
	if err := app.Start(ctx); err != nil {
		return nil, fmt.Errorf("failed to start test server: %w", err)
	}

	return app, nil
}
