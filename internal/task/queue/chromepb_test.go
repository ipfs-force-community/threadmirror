package queue

import (
	"context"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/stretchr/testify/require"
)

func TestScreenshot(t *testing.T) {
	allocCtx, cancelFn1 := chromedp.NewExecAllocator(context.Background(),
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-default-apps", true),
		chromedp.Flag("disable-extensions", true),
		chromedp.Flag("hide-scrollbars", true),
		chromedp.DisableGPU,
	)
	defer cancelFn1()

	chromedpCtx, cancelFn2 := chromedp.NewContext(allocCtx)
	defer cancelFn2()

	var buf []byte
	err := chromedp.Run(chromedpCtx,
		chromedp.EmulateViewport(485, 0),
		chromedp.Navigate("https://www.google.com"),
		chromedp.Sleep(1*time.Second),
		chromedp.FullScreenshot(&buf, 100),
	)
	require.NoError(t, err)

	require.NotEmpty(t, buf)
}
