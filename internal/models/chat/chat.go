package chat

import (
	"context"
	"fmt"
	"strings"

	"github.com/Tencent/WeKnowRust/internal/models/utils/ollama"
	"github.com/Tencent/WeKnowRust/internal/runtime"
	"github.com/Tencent/WeKnowRust/internal/types"
)

// ChatOptions defines options for chat generation
type ChatOptions struct {
	Temperature         float64 `json:"temperature"`           // Temperature
	TopP                float64 `json:"top_p"`                 // Top P
	Seed                int     `json:"seed"`                  // Random seed
	MaxTokens           int     `json:"max_tokens"`            // Max tokens
	MaxCompletionTokens int     `json:"max_completion_tokens"` // Max completion tokens
	FrequencyPenalty    float64 `json:"frequency_penalty"`     // Frequency penalty
	PresencePenalty     float64 `json:"presence_penalty"`      // Presence penalty
	Thinking            *bool   `json:"thinking"`              // Whether to enable "thinking"
}

// Message represents a chat message
type Message struct {
	Role    string `json:"role"`    // Role: system | user | assistant
	Content string `json:"content"` // Message content
}

// Chat defines the chat interface
type Chat interface {
	// Chat performs non-streaming chat
	Chat(ctx context.Context, messages []Message, opts *ChatOptions) (*types.ChatResponse, error)

	// ChatStream performs streaming chat
	ChatStream(ctx context.Context, messages []Message, opts *ChatOptions) (<-chan types.StreamResponse, error)

	// GetModelName returns the model name
	GetModelName() string

	// GetModelID returns the model ID
	GetModelID() string
}

type ChatConfig struct {
	Source    types.ModelSource
	BaseURL   string
	ModelName string
	APIKey    string
	ModelID   string
}

// NewChat creates a chat client based on the model source
func NewChat(config *ChatConfig) (Chat, error) {
	var chat Chat
	var err error
	switch strings.ToLower(string(config.Source)) {
	case string(types.ModelSourceLocal):
		runtime.GetContainer().Invoke(func(ollamaService *ollama.OllamaService) {
			chat, err = NewOllamaChat(config, ollamaService)
		})
		if err != nil {
			return nil, err
		}
		return chat, nil
	case string(types.ModelSourceRemote):
		return NewRemoteAPIChat(config)
	default:
		return nil, fmt.Errorf("unsupported chat model source: %s", config.Source)
	}
}
