package jobfx

import (
	"context"

	"github.com/chromedp/chromedp"
	internaljob "github.com/ipfs-force-community/threadmirror/internal/job"
	"github.com/ipfs-force-community/threadmirror/internal/job/middleware"
	"github.com/ipfs-force-community/threadmirror/pkg/database/sql"
	"github.com/ipfs-force-community/threadmirror/pkg/jobq"
	"github.com/ipfs-force-community/threadmirror/pkg/jobq/jobqfx"
	"go.uber.org/fx"
)

type (
	chromedpContext  = internaljob.ChromedpContext
	chromedpCancelFn context.CancelFunc
)

var Module = fx.Module("job",
	jobqfx.Module,
	fx.Provide(func() (chromedpContext, chromedpCancelFn) {
		allocCtx, cancelFn := chromedp.NewExecAllocator(context.Background(),
			chromedp.NoFirstRun,
			chromedp.NoDefaultBrowserCheck,
			chromedp.NoSandbox,
			chromedp.Flag("no-sandbox", true),
			chromedp.Flag("headless", true),
			chromedp.Flag("disable-default-apps", true),
			chromedp.Flag("disable-extensions", true),
			chromedp.DisableGPU,
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
	fx.Provide(internaljob.NewMentionHandler),
	fx.Provide(internaljob.NewReplyTweetHandler),
	// Register lifecycle hooks for proper startup/shutdown
	fx.Invoke(registerJobLifecycle),
)

// registerJobLifecycle sets up proper startup and shutdown hooks for job processing
func registerJobLifecycle(lc fx.Lifecycle, registry jobq.JobHandlerRegistry, mentionHandler *internaljob.MentionHandler, replyHandler *internaljob.ReplyTweetHandler, db *sql.DB) {
	dbInjector := middleware.DBInjector(db)
	lc.Append(fx.StartHook(func(ctx context.Context) error {
		registry.RegisterHandler(internaljob.TypeProcessMention, dbInjector(mentionHandler))
		registry.RegisterHandler(internaljob.TypeReplyTweet, dbInjector(replyHandler))
		return nil
	}))
}
