package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"

	"github.com/ipfs-force-community/threadmirror/internal/model"
	"github.com/ipfs-force-community/threadmirror/pkg/llm"
	"github.com/ipfs-force-community/threadmirror/pkg/util"
	"github.com/ipfs-force-community/threadmirror/pkg/xscraper"
	"github.com/tmc/langchaingo/llms"
)

// TranslationRepoInterface defines the interface for translation repo operations
type TranslationRepoInterface interface {
	Create(ctx context.Context, translation *model.Translation) error
	GetByID(ctx context.Context, id string) (*model.Translation, error)
	GetByThreadAndLanguages(ctx context.Context, threadID, sourceLanguage, targetLanguage string) (*model.Translation, error)
	ListByThreadID(ctx context.Context, threadID string, limit, offset int) ([]*model.Translation, int64, error)
	Update(ctx context.Context, translation *model.Translation) error
}

// TranslationService handles thread translation operations
type TranslationService struct {
	translationRepo TranslationRepoInterface
	llm        llm.Model
	logger     *slog.Logger
}

// NewTranslationService creates a new translation service
func NewTranslationService(
	translationRepo TranslationRepoInterface,
	llm llm.Model,
	logger *slog.Logger,
) *TranslationService {
	return &TranslationService{
		threadRepo: threadRepo,
		llm:        llm,
		logger:     logger.With("service", "translation"),
	}
}

// TranslateThread translates a complete thread to target language
func (s *TranslationService) TranslateThread(ctx context.Context, threadID, targetLanguage string, tweets []*xscraper.Tweet) (*model.Translation, error) {
	if len(tweets) == 0 {
		return nil, fmt.Errorf("no tweets provided")
	}

	// Extract tweet texts using the simple utility
	tweetTexts := util.ExtractTweetTexts(tweets)
	if len(tweetTexts) == 0 {
		return nil, fmt.Errorf("no translatable content found")
	}

	// Detect source language from first tweet
	sourceLanguage := s.detectLanguage(tweetTexts[0].DisplayableText)

	// Check if translation already exists
	existing, err := s..GetThreadTranslation(ctx, threadID, sourceLanguage, targetLanguage)
	if err != nil {
		return nil, fmt.Errorf("check existing translation: %w", err)
	}
	if existing != nil {
		return existing, nil
	}

	// Translate all tweet texts
	translatedTweetTexts, err := s.translateTweetTexts(ctx, tweetTexts, sourceLanguage, targetLanguage)
	if err != nil {
		return nil, fmt.Errorf("translate tweet texts: %w", err)
	}

	// Convert to JSON for storage
	tweetTextsJSON, err := json.Marshal(translatedTweetTexts)
	if err != nil {
		return nil, fmt.Errorf("marshal tweet texts to JSON: %w", err)
	}

	// Create translation record
	translation := &model.Translation{
		ID:             uuid.New().String(),
		ThreadID:       threadID,
		SourceLanguage: sourceLanguage,
		TargetLanguage: targetLanguage,
		TweetTexts:     datatypes.JSON(tweetTextsJSON),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// Save to database
	err = s.threadRepo.CreateTranslation(ctx, translation)
	if err != nil {
		return nil, fmt.Errorf("save translation: %w", err)
	}

	s.logger.Info("translation completed",
		"thread_id", threadID,
		"source_language", sourceLanguage,
		"target_language", targetLanguage,
		"tweets_count", len(translatedTweetTexts),
	)

	return translation, nil
}

// GetTranslation retrieves a translation by ID
func (s *TranslationService) GetTranslation(ctx context.Context, translationID string) (*model.Translation, error) {
	return s.translationRepo.GetByID(ctx, translationID)
}

// ListTranslations lists all translations for a thread
func (s *TranslationService) ListTranslations(ctx context.Context, threadID string, limit, offset int) ([]*model.Translation, int64, error) {
	return s.threadRepo.ListTranslations(ctx, threadID, limit, offset)
}

// GetThreadTranslation gets translation for specific language pair
func (s *TranslationService) GetThreadTranslation(ctx context.Context, threadID, sourceLanguage, targetLanguage string) (*model.Translation, error) {
	return s.threadRepo.GetThreadTranslation(ctx, threadID, sourceLanguage, targetLanguage)
}

// GetTweetTextsFromTranslation extracts TweetText slice from Translation
func (s *TranslationService) GetTweetTextsFromTranslation(translation *model.Translation) ([]model.TweetText, error) {
	if translation == nil {
		return nil, nil
	}

	var tweetTexts []model.TweetText
	if err := json.Unmarshal(translation.TweetTexts, &tweetTexts); err != nil {
		return nil, fmt.Errorf("unmarshal tweet texts from JSON: %w", err)
	}

	return tweetTexts, nil
}

// translateTweetTexts translates an array of TweetText using LLM
func (s *TranslationService) translateTweetTexts(ctx context.Context, tweetTexts []model.TweetText, sourceLanguage, targetLanguage string) ([]model.TweetText, error) {
	result := make([]model.TweetText, len(tweetTexts))

	// Process in batches of 5 tweets
	batchSize := 5
	for i := 0; i < len(tweetTexts); i += batchSize {
		end := i + batchSize
		if end > len(tweetTexts) {
			end = len(tweetTexts)
		}

		batch := tweetTexts[i:end]
		translatedBatch, err := s.translateBatch(ctx, batch, sourceLanguage, targetLanguage)
		if err != nil {
			return nil, fmt.Errorf("translate batch: %w", err)
		}

		copy(result[i:end], translatedBatch)
	}

	return result, nil
}

// translateBatch translates a batch of tweet texts
func (s *TranslationService) translateBatch(ctx context.Context, batch []model.TweetText, sourceLanguage, targetLanguage string) ([]model.TweetText, error) {
	if len(batch) == 0 {
		return nil, nil
	}

	// Collect texts to translate
	texts := make([]string, len(batch))
	for i, tweetText := range batch {
		texts[i] = tweetText.DisplayableText
	}

	// Create translation prompt
	prompt := s.buildTranslationPrompt(texts, sourceLanguage, targetLanguage)

	// Call LLM using GenerateContent
	messages := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextPart(prompt),
			},
		},
	}

	response, err := s.llm.GenerateContent(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("llm generate content: %w", err)
	}

	// Extract text from response
	responseText := ""
	if len(response.Choices) > 0 && len(response.Choices[0].Content) > 0 {
		responseText = response.Choices[0].Content
	}

	// Parse response
	translatedTexts := s.parseTranslationResponse(responseText, len(texts))

	// Build result
	result := make([]model.TweetText, len(batch))
	for i, tweetText := range batch {
		translatedText := ""
		if i < len(translatedTexts) {
			translatedText = translatedTexts[i]
		}

		result[i] = model.TweetText{
			TweetID:         tweetText.TweetID,
			DisplayableText: tweetText.DisplayableText,
			TranslatedText:  translatedText,
		}
	}

	return result, nil
}

// buildTranslationPrompt creates a translation prompt for LLM
func (s *TranslationService) buildTranslationPrompt(texts []string, sourceLanguage, targetLanguage string) string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("Translate the following social media posts from %s to %s. ", sourceLanguage, targetLanguage))
	builder.WriteString("Keep the tone natural and conversational. ")
	builder.WriteString("Preserve line breaks and formatting. ")
	builder.WriteString("Return only the translations, one per line, in the same order:\n\n")

	for i, text := range texts {
		builder.WriteString(fmt.Sprintf("%d. %s\n", i+1, text))
	}

	return builder.String()
}

// parseTranslationResponse parses LLM response into individual translations
func (s *TranslationService) parseTranslationResponse(response string, expectedCount int) []string {
	lines := strings.Split(strings.TrimSpace(response), "\n")
	result := make([]string, 0, expectedCount)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Remove numbering if present (e.g., "1. " or "1) ")
		if len(line) > 3 && (line[1] == '.' || line[1] == ')') && line[2] == ' ' {
			line = line[3:]
		}

		result = append(result, strings.TrimSpace(line))

		if len(result) >= expectedCount {
			break
		}
	}

	// Pad with empty strings if needed
	for len(result) < expectedCount {
		result = append(result, "")
	}

	return result
}

// detectLanguage detects the language of a text (simple implementation)
func (s *TranslationService) detectLanguage(text string) string {
	text = strings.TrimSpace(text)
	if text == "" {
		return "en"
	}

	// Simple heuristic-based language detection
	runes := []rune(text)

	// Count character types
	var cjkCount, latinCount, arabicCount int

	for _, r := range runes {
		switch {
		case (r >= 0x4E00 && r <= 0x9FFF) || // CJK Unified Ideographs
			(r >= 0x3400 && r <= 0x4DBF) || // CJK Extension A
			(r >= 0x3040 && r <= 0x309F) || // Hiragana
			(r >= 0x30A0 && r <= 0x30FF) || // Katakana
			(r >= 0xAC00 && r <= 0xD7AF): // Hangul
			cjkCount++
		case (r >= 0x0041 && r <= 0x005A) || // Latin uppercase
			(r >= 0x0061 && r <= 0x007A) || // Latin lowercase
			(r >= 0x00C0 && r <= 0x017F): // Latin extended
			latinCount++
		case (r >= 0x0600 && r <= 0x06FF) || // Arabic
			(r >= 0x0750 && r <= 0x077F): // Arabic supplement
			arabicCount++
		}
	}

	total := cjkCount + latinCount + arabicCount
	if total == 0 {
		return "en"
	}

	// Determine primary language
	if cjkCount > total/3 {
		// Further detect between Chinese, Japanese, Korean
		if strings.ContainsAny(text, "の、を、が、は、に、で、と、から、まで") {
			return "ja"
		}
		if strings.ContainsAny(text, "이、가、을、를、에서、으로") {
			return "ko"
		}
		return "zh"
	}

	if arabicCount > total/3 {
		return "ar"
	}

	return "en" // Default to English for Latin script
}
