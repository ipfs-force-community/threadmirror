package sqlfx

import (
	"log/slog"

	"github.com/ipfs-force-community/threadmirror/pkg/database/sql"
	"go.uber.org/fx"
)

type Config struct {
	Driver string
	DSN    string
	Debug  bool
}

var Module = fx.Module("database",
	fx.Provide(NewDB),
)

func NewDB(lc fx.Lifecycle, c *Config, logger *slog.Logger) (*sql.DB, error) {
	db, err := sql.NewWithDebug(c.Driver, c.DSN, logger, c.Debug)
	if err != nil {
		return nil, err
	}
	lc.Append(fx.StopHook(func() error {
		return db.Close()
	}))
	return db, nil
}
