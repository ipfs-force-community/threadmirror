package xscraperfx

import (
	"log/slog"

	"github.com/ipfs-force-community/threadmirror/pkg/xscraper"
	"go.uber.org/fx"
)

var Module = fx.Module("xscraper",
	fx.Provide(func(loginOpts []xscraper.LoginOptions, logger *slog.Logger) ([]*xscraper.XScraper, error) {
		scrapers := make([]*xscraper.XScraper, len(loginOpts))
		for i, loginOpt := range loginOpts {
			scraper, err := xscraper.New(loginOpt, logger)
			if err != nil {
				return nil, err
			}
			scrapers[i] = scraper
		}
		return scrapers, nil
	}),
)
