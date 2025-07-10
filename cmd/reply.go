package main

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/samber/lo"
	"github.com/urfave/cli/v2"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"

	"github.com/ipfs-force-community/threadmirror/internal/config"
	"github.com/ipfs-force-community/threadmirror/internal/task/queue"
	"github.com/ipfs-force-community/threadmirror/pkg/database/redis"
	"github.com/ipfs-force-community/threadmirror/pkg/database/redis/redisfx"
	"github.com/ipfs-force-community/threadmirror/pkg/jobq"
	"github.com/ipfs-force-community/threadmirror/pkg/jobq/jobqfx"
	"github.com/ipfs-force-community/threadmirror/pkg/log/logfx"
	"github.com/ipfs-force-community/threadmirror/pkg/util"
)

var ReplyCommand = &cli.Command{
	Name:  "reply",
	Usage: "Reply to a tweet",
	Flags: util.MergeSlices(
		config.GetRedisCLIFlags(),
		[]cli.Flag{
			&cli.StringFlag{
				Name:    "mention-id",
				Usage:   "The ID of the mention to reply to",
				EnvVars: []string{"MENTION_ID"},
			},
		},
	),
	Action: func(c *cli.Context) error {
		redisConf := config.LoadRedisConfigFromCLI(c)

		fxApp := fx.New(
			// Provide the configuration
			fx.Supply(&redis.RedisConfig{
				Addr:     redisConf.Addr,
				Password: redisConf.Password,
				DB:       redisConf.DB,
			}),
			fx.Supply(&logfx.Config{
				Level:      c.String("log-level"),
				LogDevMode: false,
			}),
			logfx.Module,
			redisfx.Module,
			jobqfx.ModuleClient,
			fx.WithLogger(func(logger *slog.Logger) fxevent.Logger {
				return &fxevent.SlogLogger{Logger: logger}
			}),
			fx.Invoke(func(lc fx.Lifecycle, jobqClient jobq.JobQueueClient) {
				lc.Append(fx.StartHook(func(ctx context.Context) error {
					jobID, err := jobqClient.Enqueue(ctx, lo.Must(queue.NewReplyTweetJob(c.String("mention-id"))))
					if err != nil {
						return fmt.Errorf("enqueue reply tweet job: %w", err)
					}
					fmt.Println("Enqueued reply tweet job with ID:", jobID)
					return nil
				}))
			}),
		)
		fxApp.Run()
		return nil
	},
}
