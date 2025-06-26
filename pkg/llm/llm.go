package llm

import (
	"fmt"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

type Model llms.Model

// NewOpenAI creates a new OpenAI LLM instance
func NewOpenAI(baseURL, apiKey, model string) (Model, error) {
	llm, err := openai.New(
		openai.WithToken(apiKey),
		openai.WithModel(model),
		openai.WithBaseURL(baseURL),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create OpenAI model: %w", err)
	}

	return llm, nil
}
