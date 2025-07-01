package xscraperfx

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/ipfs-force-community/threadmirror/internal/service"
	"github.com/ipfs-force-community/threadmirror/pkg/xscraper"
	"go.uber.org/fx"
)

type Config struct {
	Username string
	Password string
	Email    string
}

var Module = fx.Module("xscraper",
	fx.Provide(func(c *Config, botCookieService *service.BotCookieService, logger *slog.Logger) *xscraper.XScraper {
		// Create login options for xscraper
		loginOpts := xscraper.LoginOptions{
			LoadCookies: func(ctx context.Context) ([]*http.Cookie, error) {
				return botCookieService.LoadCookies(ctx, c.Email, c.Username)
			},
			SaveCookies: func(ctx context.Context, cookies []*http.Cookie) error {
				return botCookieService.SaveCookies(ctx, c.Email, c.Username, cookies)
			},
			Username: c.Username,
			Password: c.Password,
			Email:    c.Email,
		}
		return xscraper.New(loginOpts, logger)
	}),
)
