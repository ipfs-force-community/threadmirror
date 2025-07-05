package sqlfx

import (
	"log/slog"

	"github.com/ipfs-force-community/threadmirror/pkg/database/sql"
	"go.uber.org/fx"
)

type Config struct {
	Driver string
	DSN    string
}

var Module = fx.Module("database",
	fx.Provide(NewDB),
)

func NewDB(lc fx.Lifecycle, c *Config, logger *slog.Logger) (*sql.DB, error) {
	db, err := sql.New(c.Driver, c.DSN, logger)
	if err != nil {
		return nil, err
	}
	lc.Append(fx.StopHook(func() error {
		return db.Close()
	}))
	return db, nil
}
