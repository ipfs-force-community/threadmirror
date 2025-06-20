package i18n

import (
	"github.com/gin-gonic/gin"
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

const localizerKey = "i18n.localizer"

// Middleware 创建 i18n 中间件
func Middleware(i *I18nBundle) gin.HandlerFunc {
	return func(c *gin.Context) {
		lang := c.Query("lang")
		accept := c.GetHeader("Accept-Language")
		localizer := i.NewLocalizer(lang, accept)
		c.Set(localizerKey, localizer)
		c.Next()
	}
}

// GetLocalizer 从 gin.Context 获取 localizer
func GetLocalizer(c *gin.Context) *i18n.Localizer {
	if localizer, exists := c.Get(localizerKey); exists {
		if loc, ok := localizer.(*i18n.Localizer); ok {
			return loc
		}
	}
	// 如果没有找到，返回默认的英文 localizer
	// 这种情况不应该发生，如果正确使用了中间件
	return nil
}

// T 快捷函数用于翻译
func T(c *gin.Context, messageID string, templateData ...any) string {
	var td any = nil
	if len(templateData) > 0 {
		td = templateData[0]
	}

	return TWithConfig(c, &i18n.LocalizeConfig{
		MessageID:    messageID,
		TemplateData: td,
	})
}

func TWithConfig(c *gin.Context, config *i18n.LocalizeConfig) string {
	localizer := GetLocalizer(c)
	if localizer == nil {
		return config.MessageID
	}

	message, err := localizer.Localize(config)
	if err != nil {
		return config.MessageID
	}
	return message
}
