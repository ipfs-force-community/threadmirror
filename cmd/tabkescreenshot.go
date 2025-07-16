package main

import (
	"context"
	"errors"
	"os"

	"github.com/chromedp/chromedp"
	"github.com/ipfs-force-community/threadmirror/pkg/util"
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
		&cli.Float64Flag{
			Name:  "scale",
			Usage: "the scale of the screenshot",
			Value: 2,
		},
	},
	Action: func(c *cli.Context) error {
		opts := util.DefaultScreenshotOptions()
		opts.Scale = c.Float64("scale")
		opts.AdditionalOptions = chromedp.DefaultExecAllocatorOptions[:]

		buf, err := util.TakeScreenshotFromURL(context.Background(), c.String("url"), opts)
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
