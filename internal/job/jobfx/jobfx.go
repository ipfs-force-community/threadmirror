package jobfx

import (
	"context"

	internaljob "github.com/ipfs-force-community/threadmirror/internal/job"
	"github.com/ipfs-force-community/threadmirror/pkg/jobq"
	"github.com/ipfs-force-community/threadmirror/pkg/jobq/jobqfx"
	"go.uber.org/fx"
)

var Module = fx.Module("job",
	jobqfx.Module,
	// Domain services used by workers
	fx.Provide(internaljob.NewMentionHandler),
	// Register lifecycle hooks for proper startup/shutdown
	fx.Invoke(registerJobLifecycle),
)

// registerJobLifecycle sets up proper startup and shutdown hooks for job processing
func registerJobLifecycle(lc fx.Lifecycle, registry jobq.JobHandlerRegistry, mentionHandler *internaljob.MentionHandler) {
	lc.Append(fx.StartHook(func(ctx context.Context) error {
		registry.RegisterHandler(internaljob.TypeProcessMention, mentionHandler)
		return nil
	}))
}
