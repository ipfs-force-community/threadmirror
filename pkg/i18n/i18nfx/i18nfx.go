package i18nfx

import (
	"embed"

	"github.com/ipfs-force-community/threadmirror/pkg/i18n"
	"go.uber.org/fx"
)

func Module(localeFS *embed.FS) fx.Option {
	return fx.Module("i18n",
		fx.Provide(func() (*i18n.I18nBundle, error) {
			return i18n.NewI18nBundle(localeFS)
		}),
	)
}
