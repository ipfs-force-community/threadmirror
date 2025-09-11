package sql

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/ipfs-force-community/threadmirror/internal/sqlc_generated"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/tracelog"
)

// DB wraps the standard database connection and sqlc queries
type DB struct {
	pool    *pgxpool.Pool
	queries *sqlc_generated.Queries
	logger  *slog.Logger
}

// Context key for storing pgx.Tx
type contextTxKey struct{}

var txKey = contextTxKey{}

// SQLLogger implements pgx tracelog.Logger for SQL logging
type SQLLogger struct {
	logger *slog.Logger
}

// Log implements the tracelog.Logger interface
func (l *SQLLogger) Log(
	ctx context.Context,
	level tracelog.LogLevel,
	msg string,
	data map[string]any,
) {
	// Extract SQL and arguments from the data
	if sql, ok := data["sql"]; ok {
		// Clean up the SQL for better readability
		sqlStr := fmt.Sprintf("%v", sql)
		sqlStr = strings.ReplaceAll(sqlStr, "\n", " ")
		sqlStr = strings.ReplaceAll(sqlStr, "\t", " ")
		for strings.Contains(sqlStr, "  ") {
			sqlStr = strings.ReplaceAll(sqlStr, "  ", " ")
		}
		sqlStr = strings.TrimSpace(sqlStr)

		attrs := []slog.Attr{
			slog.String("sql", sqlStr),
		}

		if args, ok := data["args"]; ok {
			attrs = append(attrs, slog.Any("args", args))
		}

		if duration, ok := data["time"]; ok {
			attrs = append(attrs, slog.Any("duration", duration))
		}

		if err, ok := data["err"]; ok && err != nil {
			l.logger.LogAttrs(ctx, slog.LevelError, msg, attrs...)
		} else {
			l.logger.LogAttrs(ctx, slog.LevelDebug, msg, attrs...)
		}
	} else {
		// Fallback for other log messages
		l.logger.LogAttrs(ctx, slog.LevelDebug, msg, slog.Any("data", data))
	}
}

// New creates a new database connection using pgx
func New(driver string, dsn string, logger *slog.Logger) (*DB, error) {
	return NewWithDebug(driver, dsn, logger, false)
}

// NewWithDebug creates a new database connection with optional debug mode
func NewWithDebug(driver string, dsn string, logger *slog.Logger, debug bool) (*DB, error) {
	if driver != "postgres" {
		return nil, fmt.Errorf("only postgres driver is supported, got: %s", driver)
	}

	// Parse connection config
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse dsn: %w", err)
	}

	// Add SQL logging in debug mode
	if debug {
		sqlLogger := &SQLLogger{logger: logger}
		config.ConnConfig.Tracer = &tracelog.TraceLog{
			Logger:   sqlLogger,
			LogLevel: tracelog.LogLevelDebug,
		}
		logger.Info("SQL logging enabled (debug mode)")
	}

	// Create connection pool
	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("create connection pool: %w", err)
	}

	// Create sqlc queries
	queries := sqlc_generated.New(pool)

	return &DB{
		pool:    pool,
		queries: queries,
		logger:  logger,
	}, nil
}

// Queries returns the sqlc queries instance
func (d *DB) Queries() *sqlc_generated.Queries {
	return d.queries
}

// QueriesFromContext returns queries bound to a transaction found in context if present, otherwise base queries
func (d *DB) QueriesFromContext(ctx context.Context) *sqlc_generated.Queries {
	if tx, ok := d.TxFromContext(ctx); ok {
		return d.queries.WithTx(tx)
	}
	return d.queries
}

// ContextWithTx stores a pgx.Tx into the context
func (d *DB) ContextWithTx(ctx context.Context, tx pgx.Tx) context.Context {
	return context.WithValue(ctx, txKey, tx)
}

// TxFromContext retrieves a pgx.Tx from context
func (d *DB) TxFromContext(ctx context.Context) (pgx.Tx, bool) {
	if ctx == nil {
		return nil, false
	}
	tx, ok := ctx.Value(txKey).(pgx.Tx)
	return tx, ok
}

// BeginTx starts a transaction and returns the transaction and a context carrying it
func (d *DB) BeginTx(ctx context.Context) (pgx.Tx, context.Context, error) {
	tx, err := d.pool.Begin(ctx)
	if err != nil {
		return nil, ctx, err
	}
	return tx, d.ContextWithTx(ctx, tx), nil
}

// RunInTx executes fn within a transaction boundary. If a transaction already exists in ctx, it reuses it.
func (d *DB) RunInTx(ctx context.Context, fn func(ctx context.Context) error) (err error) {
	if _, ok := d.TxFromContext(ctx); ok {
		return fn(ctx)
	}
	tx, ctxWithTx, err := d.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctxWithTx)
		}
	}()
	if err = fn(ctxWithTx); err != nil {
		return err
	}
	if err = tx.Commit(ctxWithTx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}

// Pool returns the connection pool for direct use if needed
func (d *DB) Pool() *pgxpool.Pool {
	return d.pool
}

// WithTx returns queries wrapped in a transaction
func (d *DB) WithTx(ctx context.Context, fn func(*sqlc_generated.Queries) error) error {
	// Deprecated in favor of RunInTx + QueriesFromContext. Kept for compatibility.
	tx, err := d.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			if err := tx.Rollback(ctx); err != nil {
				d.logger.Error("failed to rollback transaction", "error", err)
			}
		}
	}()

	queries := d.queries.WithTx(tx)
	if err := fn(queries); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

// Migrate runs database migration using raw SQL
func (d *DB) Migrate(ctx context.Context, models []any, rawSQL ...string) error {
	d.logger.Info("SQLC architecture uses declarative schema files in supabase/schemas/")
	d.logger.Info("Please use Supabase CLI to generate and apply migrations")

	// Execute any provided raw SQL for compatibility
	for _, sql := range rawSQL {
		if sql != "" {
			_, err := d.pool.Exec(ctx, sql)
			if err != nil {
				return fmt.Errorf("execute raw SQL: %w", err)
			}
		}
	}
	return nil
}

// Close closes the database connection
func (d *DB) Close() error {
	if d.pool != nil {
		d.pool.Close()
	}
	return nil
}

// Health checks the database connection health
func (d *DB) Health() error {
	return d.pool.Ping(context.Background())
}
