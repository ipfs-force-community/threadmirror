package llmfx

import (
	"github.com/ipfs-force-community/threadmirror/pkg/llm"
	"go.uber.org/fx"
)

type Config struct {
	OpenAIAPIKey  string
	OpenAIModel   string
	OpenAIBaseURL string
}

// Module provides the fx module for llm
var Module = fx.Module("llm",
	fx.Provide(NewLLM),
)

// NewLLM creates a new LLM instance
func NewLLM(config *Config) (llm.Model, error) {
	return llm.NewOpenAI(config.OpenAIBaseURL, config.OpenAIAPIKey, config.OpenAIModel)
}
