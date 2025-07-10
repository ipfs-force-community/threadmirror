package main

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/urfave/cli/v2"
)

var TakeScreenshotCommand = &cli.Command{
	Name:  "take-screenshot",
	Usage: "take a screenshot of the current tab",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "url",
			Usage: "the URL to screenshot",
			Value: "https://www.google.com",
		},
		&cli.PathFlag{
			Name:  "output",
			Usage: "the path to save the screenshot",
			Value: "screenshot.png",
		},
	},
	Action: func(c *cli.Context) error {
		allocCtx, cancelFn1 := chromedp.NewExecAllocator(context.Background(),
			chromedp.NoFirstRun,
			chromedp.NoDefaultBrowserCheck,
			chromedp.NoSandbox,
			chromedp.Flag("no-sandbox", true),
			chromedp.Flag("headless", true),
			chromedp.Flag("disable-default-apps", true),
			chromedp.Flag("disable-extensions", true),
			chromedp.Flag("hide-scrollbars", true),
		)
		defer cancelFn1()

		chromedpCtx, cancelFn2 := chromedp.NewContext(allocCtx)
		defer cancelFn2()

		var buf []byte
		err := chromedp.Run(chromedpCtx,
			chromedp.EmulateViewport(485, 0),
			chromedp.Navigate(c.String("url")),
			chromedp.Sleep(1*time.Second),
			chromedp.FullScreenshot(&buf, 100),
		)
		if err != nil {
			return err
		}

		if len(buf) == 0 {
			return errors.New("screenshot is empty")
		}

		if err := os.WriteFile(c.Path("output"), buf, 0644); err != nil {
			return err
		}

		return nil
	},
}
