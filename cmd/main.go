package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/urfave/cli/v2"
)

func main() {
	if ocDotenv := os.Getenv("TM_DOT_ENV"); ocDotenv != "" {
		_ = godotenv.Load(ocDotenv)
	} else {
		_ = godotenv.Load()
	}

	app := &cli.App{
		Name:  "threadmirror",
		Usage: "A backend service for Threadmirror application",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "debug",
				Value:   false,
				EnvVars: []string{"TM_DEBUG"},
				Usage:   "Enable debug mode",
			},
			&cli.StringFlag{
				Name:    "log-level",
				Aliases: []string{"loglevel"},
				Value:   "info",
				EnvVars: []string{"TM_LOG_LEVEL"},
				Usage:   "Set log level",
			},
		},
		Commands: []*cli.Command{
			ServerCommand,
			MigrateCommand,
			BotCommand,
			ReplyCommand,
			TakeScreenshotCommand,
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
