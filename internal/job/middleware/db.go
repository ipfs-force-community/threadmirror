package middleware

import (
	"context"

	"github.com/ipfs-force-community/threadmirror/pkg/database/sql"
	"github.com/ipfs-force-community/threadmirror/pkg/jobq"
)

func DBInjector(db *sql.DB) jobq.JobMiddleware {
	return func(handler jobq.JobHandler) jobq.JobHandler {
		return jobq.JobHandlerFunc(func(ctx context.Context, job *jobq.Job) error {
			ctx = sql.WithDBToContext(ctx, db)
			return handler.HandleJob(ctx, job)
		})
	}
}
