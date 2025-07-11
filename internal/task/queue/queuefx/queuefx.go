package queuefx

import (
	"context"

	internalqueue "github.com/ipfs-force-community/threadmirror/internal/task/queue"
	"github.com/ipfs-force-community/threadmirror/pkg/jobq"
	"go.uber.org/fx"
)

var Module = fx.Module("queue",
	// Domain services used by workers
	fx.Provide(internalqueue.NewMentionHandler),
	fx.Provide(internalqueue.NewReplyTweetHandler),
	fx.Provide(internalqueue.NewThreadScrapeHandler),
	// Register lifecycle hooks for proper startup/shutdown
	fx.Invoke(registerJobLifecycle),
)

// registerJobLifecycle sets up proper startup and shutdown hooks for job processing
func registerJobLifecycle(lc fx.Lifecycle, registry jobq.JobHandlerRegistry, mentionHandler *internalqueue.MentionHandler, replyHandler *internalqueue.ReplyTweetHandler, threadScrapeHandler *internalqueue.ThreadScrapeHandler) {
	lc.Append(fx.StartHook(func(ctx context.Context) error {
		registry.RegisterHandler(internalqueue.TypeProcessMention, mentionHandler)
		registry.RegisterHandler(internalqueue.TypeReplyTweet, replyHandler)
		registry.RegisterHandler(internalqueue.TypeThreadScrape, threadScrapeHandler)
		return nil
	}))
}
