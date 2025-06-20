package i18n

import "embed"

//go:embed locale.*.toml
var LocaleFS embed.FS
