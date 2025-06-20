package sql

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/ipfs-force-community/threadmirror/pkg/log/gormlog"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// DB wraps the GORM database connection
type DB struct {
	*gorm.DB
}

// New creates a new database connection
func New(driver string, dsn string, logger *slog.Logger) (*DB, error) {
	// Configure GORM logger
	gormConfig := &gorm.Config{
		Logger: gormlog.New(logger),
	}

	var dia gorm.Dialector
	switch driver {
	case "postgres":
		dia = postgres.Open(dsn)
	case "sqlite":
		dia = sqlite.Open(dsn)
	default:
		return nil, fmt.Errorf("invalid driver: %s", driver)
	}

	db, err := gorm.Open(dia, gormConfig)
	if err != nil {
		return nil, fmt.Errorf("connect to database: %w", err)
	}

	return &DB{DB: db}, nil
}

// Migrate runs database migration for all models
func (d *DB) Migrate(ctx context.Context, models []any, rawSQL ...string) error {
	if err := d.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(models...); err != nil {
			return fmt.Errorf("migrate database: %w", err)
		}
		for _, sql := range rawSQL {
			if err := tx.Exec(sql).Error; err != nil {
				return fmt.Errorf("execute raw SQL: %w", err)
			}
		}
		return nil
	}); err != nil {
		return fmt.Errorf("migrate database: %w", err)
	}
	return nil
}

// Close closes the database connection
func (d *DB) Close() error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// Health checks the database connection health
func (d *DB) Health() error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}
