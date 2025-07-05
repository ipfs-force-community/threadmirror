package i18n

import (
	"embed"
	"io/fs"
	"path/filepath"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/pelletier/go-toml/v2"
	"golang.org/x/text/language"
)

type I18nBundle struct {
	bundle *i18n.Bundle
}

func NewI18nBundle(localeFS *embed.FS) (*I18nBundle, error) {
	bundle := i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)

	err := fs.WalkDir(localeFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() && filepath.Ext(path) == ".toml" {
			_, err := bundle.LoadMessageFileFS(localeFS, path)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &I18nBundle{
		bundle: bundle,
	}, nil
}

// NewLocalizer 根据语言偏好创建本地化器
func (i *I18nBundle) NewLocalizer(languages ...string) *i18n.Localizer {
	return i18n.NewLocalizer(i.bundle, languages...)
}
