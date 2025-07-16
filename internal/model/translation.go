package model

import (
	"time"

	"gorm.io/datatypes"
)

// TweetText represents the translation unit for a single tweet
type TweetText struct {
	TweetID         string `json:"tweet_id"`         // Original tweet ID
	DisplayableText string `json:"displayable_text"` // Original displayable text (no entities)
	TranslatedText  string `json:"translated_text"`  // Translated text
}

// Translation represents a complete thread translation
type Translation struct {
	ID             string         `gorm:"primaryKey;type:varchar(36)" json:"id"`
	ThreadID       string         `gorm:"not null;type:varchar(36);index" json:"thread_id"`
	SourceLanguage string         `gorm:"not null;type:varchar(10)" json:"source_language"`
	TargetLanguage string         `gorm:"not null;type:varchar(10)" json:"target_language"`
	TweetTexts     datatypes.JSON `gorm:"type:json" json:"tweet_texts"` // JSON array of TweetText
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
}

func (Translation) TableName() string {
	return "translations"
}

// LanguageConfig holds configuration for supported languages
type LanguageConfig struct {
	Code        string `json:"code"`         // ISO 639-1 language code
	Name        string `json:"name"`         // Language display name
	FontFamily  string `json:"font_family"`  // Preferred font family for this language
	TextDir     string `json:"text_dir"`     // Text direction: ltr or rtl
	IsSupported bool   `json:"is_supported"` // Whether translation is supported
}

// GetSupportedLanguages returns list of supported languages for translation
func GetSupportedLanguages() []LanguageConfig {
	return []LanguageConfig{
		{Code: "zh", Name: "中文", FontFamily: "PingFang SC, Microsoft YaHei, sans-serif", TextDir: "ltr", IsSupported: true},
		{Code: "en", Name: "English", FontFamily: "Arial, sans-serif", TextDir: "ltr", IsSupported: true},
		{Code: "ja", Name: "日本語", FontFamily: "Hiragino Sans, Yu Gothic, sans-serif", TextDir: "ltr", IsSupported: true},
		{Code: "ko", Name: "한국어", FontFamily: "Malgun Gothic, Apple SD Gothic Neo, sans-serif", TextDir: "ltr", IsSupported: true},
		{Code: "es", Name: "Español", FontFamily: "Arial, sans-serif", TextDir: "ltr", IsSupported: true},
		{Code: "fr", Name: "Français", FontFamily: "Arial, sans-serif", TextDir: "ltr", IsSupported: true},
		{Code: "de", Name: "Deutsch", FontFamily: "Arial, sans-serif", TextDir: "ltr", IsSupported: true},
		{Code: "it", Name: "Italiano", FontFamily: "Arial, sans-serif", TextDir: "ltr", IsSupported: true},
		{Code: "ar", Name: "العربية", FontFamily: "Tahoma, Arial, sans-serif", TextDir: "rtl", IsSupported: true},
	}
}
