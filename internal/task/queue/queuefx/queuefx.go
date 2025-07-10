package queuefx

import (
	"context"

	"github.com/chromedp/chromedp"
	internalqueue "github.com/ipfs-force-community/threadmirror/internal/task/queue"
	"github.com/ipfs-force-community/threadmirror/pkg/jobq"
	"go.uber.org/fx"
)

type (
	chromedpContext  = internalqueue.ChromedpContext
	chromedpCancelFn context.CancelFunc
)

var Module = fx.Module("queue",
	fx.Provide(func() (chromedpContext, chromedpCancelFn) {
		allocCtx, cancelFn := chromedp.NewExecAllocator(context.Background(),
			chromedp.NoFirstRun,
			chromedp.NoDefaultBrowserCheck,
			chromedp.NoSandbox,
			chromedp.Flag("no-sandbox", true),
			chromedp.Flag("headless", true),
			chromedp.Flag("disable-default-apps", true),
			chromedp.Flag("disable-extensions", true),
			chromedp.Flag("hide-scrollbars", true),
		)
		chromedpCtx, cancelFn2 := chromedp.NewContext(allocCtx)
		return chromedpContext(chromedpCtx), chromedpCancelFn(func() {
			cancelFn2()
			cancelFn()
		})
	}),
	fx.Invoke(func(lc fx.Lifecycle, cancelFn chromedpCancelFn) {
		lc.Append(fx.StopHook(cancelFn))
	}),
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
