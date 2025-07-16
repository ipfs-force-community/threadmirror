package util

import (
	"context"
	"encoding/base64"
	"time"

	"github.com/chromedp/chromedp"
)

// ScreenshotOptions holds configuration options for taking screenshots
type ScreenshotOptions struct {
	// ViewportWidth sets the viewport width (default: 485)
	ViewportWidth int
	// ViewportHeight sets the viewport height (default: 0 for auto)
	ViewportHeight int
	// Scale sets the device scale factor (default: 3)
	Scale float64
	// WaitDuration sets how long to wait before taking screenshot (default: 1 second)
	WaitDuration time.Duration
	// Quality sets the image quality for screenshot (default: 100)
	Quality int
	// AdditionalOptions allows adding extra chromedp allocator options
	AdditionalOptions []chromedp.ExecAllocatorOption
}

// DefaultScreenshotOptions returns default options for screenshots
func DefaultScreenshotOptions() *ScreenshotOptions {
	return &ScreenshotOptions{
		ViewportWidth:  485,
		ViewportHeight: 0,
		Scale:          3,
		WaitDuration:   1 * time.Second,
		Quality:        100,
	}
}

// TakeScreenshotFromURL takes a screenshot of a webpage from URL
func TakeScreenshotFromURL(ctx context.Context, url string, opts *ScreenshotOptions) ([]byte, error) {
	if opts == nil {
		opts = DefaultScreenshotOptions()
	}

	return takeScreenshot(ctx, opts, func() chromedp.Action {
		return chromedp.Navigate(url)
	})
}

// TakeScreenshotFromHTML takes a screenshot of rendered HTML content
func TakeScreenshotFromHTML(ctx context.Context, html string, opts *ScreenshotOptions) ([]byte, error) {
	if opts == nil {
		opts = DefaultScreenshotOptions()
	}

	dataURL := "data:text/html;base64," + base64.StdEncoding.EncodeToString([]byte(html))
	return takeScreenshot(ctx, opts, func() chromedp.Action {
		return chromedp.Navigate(dataURL)
	})
}

// takeScreenshot is the internal function that handles the chromedp setup and screenshot taking
func takeScreenshot(ctx context.Context, opts *ScreenshotOptions, navigationAction func() chromedp.Action) ([]byte, error) {
	// Setup default chromedp options
	allOptions := []chromedp.ExecAllocatorOption{
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
		chromedp.NoSandbox,
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-default-apps", true),
		chromedp.Flag("disable-extensions", true),
		chromedp.Flag("hide-scrollbars", true),
	}

	// Add any additional options
	if opts.AdditionalOptions != nil {
		allOptions = append(allOptions, opts.AdditionalOptions...)
	}

	// Create context with allocator
	allocCtx, cancelAlloc := chromedp.NewExecAllocator(ctx, allOptions...)
	defer cancelAlloc()

	chromedpCtx, cancelChromedp := chromedp.NewContext(allocCtx)
	defer cancelChromedp()

	// Prepare actions
	actions := []chromedp.Action{
		chromedp.EmulateViewport(int64(opts.ViewportWidth), int64(opts.ViewportHeight), chromedp.EmulateScale(opts.Scale)),
		navigationAction(),
		chromedp.Sleep(opts.WaitDuration),
	}

	var buf []byte
	actions = append(actions, chromedp.FullScreenshot(&buf, opts.Quality))

	// Execute all actions
	err := chromedp.Run(chromedpCtx, actions...)
	if err != nil {
		return nil, err
	}

	return buf, nil
}
