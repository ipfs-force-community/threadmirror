package xscraperfx

import (
	"fmt"
	"log/slog"

	"github.com/ipfs-force-community/threadmirror/pkg/xscraper"
	"go.uber.org/fx"
)

var Module = fx.Module("xscraper",
	fx.Provide(func(o xscraper.LoginOptions, logger *slog.Logger) (*xscraper.XScraper, error) {
		fmt.Println(o)
		return xscraper.New(o, logger)
	}),
)
