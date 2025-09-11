package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ipfs-force-community/threadmirror/internal/config"
	"github.com/ipfs-force-community/threadmirror/internal/service"
	"github.com/ipfs-force-community/threadmirror/pkg/database/sql"
	"github.com/ipfs-force-community/threadmirror/pkg/log"
	"github.com/ipfs-force-community/threadmirror/pkg/xscraper"
	"github.com/urfave/cli/v2"
)

var TweetCommand = &cli.Command{
	Name:  "tweet",
	Flags: config.GetDatabaseCLIFlags(),
	Subcommands: []*cli.Command{
		{
			Name: "get-thread",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "id",
					Required: true,
				},
			},
			Action: func(c *cli.Context) error {
				scraper, err := getScraper(c)
				if err != nil {
					return err
				}

				tweets, err := xscraper.GetCompleteThread(c.Context, scraper, c.String("id"), 0)
				if err != nil {
					return err
				}

				json, err := json.Marshal(tweets)
				if err != nil {
					return err
				}
				fmt.Println(string(json))
				return nil
			},
		},
		{
			Name: "get-tweet-detail",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "id",
					Required: true,
				},
			},
			Action: func(c *cli.Context) error {
				scraper, err := getScraper(c)
				if err != nil {
					return err
				}

				tweets, err := scraper.GetTweetDetail(c.Context, c.String("id"))
				if err != nil {
					return err
				}

				json, err := json.Marshal(tweets)
				if err != nil {
					return err
				}
				fmt.Println(string(json))
				return nil
			},
		},
		{
			Name: "get-tweet-result-by-rest-id",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "id",
					Required: true,
				},
			},
			Action: func(c *cli.Context) error {
				scraper, err := getScraper(c)
				if err != nil {
					return err
				}

				tweet, err := scraper.GetTweetResultByRestId(c.Context, c.String("id"))
				if err != nil {
					return err
				}

				json, err := json.Marshal(tweet)
				if err != nil {
					return err
				}
				fmt.Println(string(json))
				return nil
			},
		},
	},
}

func getScraper(c *cli.Context) (*xscraper.XScraper, error) {
	logger, err := log.New(c.String("log-level"), c.Bool("debug"))
	if err != nil {
		return nil, err
	}

	dbConf := config.LoadDatabaseConfigFromCLI(c)
	db, err := sql.New(dbConf.Driver, dbConf.DSN, logger.Logger)
	if err != nil {
		return nil, err
	}
	defer db.Close() // nolint:errcheck

	botCookieService := service.NewBotCookieService(db)
	// TODO: Implement GetLatestBotCookie method or use ListBotCookies
	botCookies, _, err := botCookieService.ListBotCookies(c.Context, 1, 0)
	if err != nil || len(botCookies) == 0 {
		return nil, fmt.Errorf("no bot cookies found: %w", err)
	}
	botCookie := botCookies[0]

	return xscraper.New(xscraper.LoginOptions{
		LoadCookies: func(ctx context.Context) ([]*http.Cookie, error) {
			var cookies []*http.Cookie
			err := json.Unmarshal(botCookie.CookiesData, &cookies)
			if err != nil {
				return nil, err
			}
			return cookies, nil
		},
		SaveCookies: func(ctx context.Context, cookies []*http.Cookie) error {
			panic("unreachable")
		},
		Username: botCookie.Username,
		Email:    botCookie.Email,
	}, logger.Logger)
}
