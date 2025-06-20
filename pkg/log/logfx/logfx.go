package logfx

import (
	"log/slog"

	"github.com/ipfs-force-community/threadmirror/pkg/log"
	"go.uber.org/fx"
)

type Config struct {
	Level      string
	LogDevMode bool
}

var Module = fx.Module("log",
	fx.Provide(NewLogger),
	fx.Provide(func(logger *log.Logger) *slog.Logger {
		return logger.Logger
	}),
)

func NewLogger(lc fx.Lifecycle, c *Config) (*log.Logger, error) {
	logger, err := log.New(c.Level, c.LogDevMode)
	if err != nil {
		return nil, err
	}
	lc.Append(fx.StopHook(func() error {
		return logger.Close()
	}))
	return logger, nil
}
