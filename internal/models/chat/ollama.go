package chat

import (
	"context"
	"fmt"

	"github.com/Tencent/WeKnowRust/internal/logger"
	"github.com/Tencent/WeKnowRust/internal/models/utils/ollama"
	"github.com/Tencent/WeKnowRust/internal/types"
	ollamaapi "github.com/ollama/ollama/api"
)

// OllamaChat implements chat based on Ollama
type OllamaChat struct {
	modelName     string
	modelID       string
	ollamaService *ollama.OllamaService
}

// NewOllamaChat creates an Ollama chat instance
func NewOllamaChat(config *ChatConfig, ollamaService *ollama.OllamaService) (*OllamaChat, error) {
	return &OllamaChat{
		modelName:     config.ModelName,
		modelID:       config.ModelID,
		ollamaService: ollamaService,
	}, nil
}

// convertMessages converts messages to Ollama API format
func (c *OllamaChat) convertMessages(messages []Message) []ollamaapi.Message {
	ollamaMessages := make([]ollamaapi.Message, len(messages))
	for i, msg := range messages {
		ollamaMessages[i] = ollamaapi.Message{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}
	return ollamaMessages
}

// buildChatRequest builds chat request parameters
func (c *OllamaChat) buildChatRequest(messages []Message, opts *ChatOptions, isStream bool) *ollamaapi.ChatRequest {
	// Set streaming flag
	streamFlag := isStream

	// Build request
	chatReq := &ollamaapi.ChatRequest{
		Model:    c.modelName,
		Messages: c.convertMessages(messages),
		Stream:   &streamFlag,
		Options:  make(map[string]interface{}),
	}

	// Add optional parameters
	if opts != nil {
		if opts.Temperature > 0 {
			chatReq.Options["temperature"] = opts.Temperature
		}
		if opts.TopP > 0 {
			chatReq.Options["top_p"] = opts.TopP
		}
		if opts.MaxTokens > 0 {
			chatReq.Options["num_predict"] = opts.MaxTokens
		}
		if opts.Thinking != nil {
			chatReq.Think = &ollamaapi.ThinkValue{
				Value: *opts.Thinking,
			}
		}
	}

	return chatReq
}

// Chat performs non-stream chat
func (c *OllamaChat) Chat(ctx context.Context, messages []Message, opts *ChatOptions) (*types.ChatResponse, error) {
	// Ensure model is available
	if err := c.ensureModelAvailable(ctx); err != nil {
		return nil, err
	}

	// Build request
	chatReq := c.buildChatRequest(messages, opts, false)

	// Log request
	logger.GetLogger(ctx).Infof("Send chat request to model %s", c.modelName)

	var responseContent string
	var promptTokens, completionTokens int

	// Send request via Ollama client
	err := c.ollamaService.Chat(ctx, chatReq, func(resp ollamaapi.ChatResponse) error {
		responseContent = resp.Message.Content

		// Get token counts
		if resp.EvalCount > 0 {
			promptTokens = resp.PromptEvalCount
			completionTokens = resp.EvalCount - promptTokens
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("chat request failed: %w", err)
	}

	// Build response
	return &types.ChatResponse{
		Content: responseContent,
		Usage: struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		}{
			PromptTokens:     promptTokens,
			CompletionTokens: completionTokens,
			TotalTokens:      promptTokens + completionTokens,
		},
	}, nil
}

// ChatStream performs streaming chat
func (c *OllamaChat) ChatStream(
	ctx context.Context,
	messages []Message,
	opts *ChatOptions,
) (<-chan types.StreamResponse, error) {
	// Ensure model is available
	if err := c.ensureModelAvailable(ctx); err != nil {
		return nil, err
	}

	// Build request
	chatReq := c.buildChatRequest(messages, opts, true)

	// Log request
	logger.GetLogger(ctx).Infof("Send streaming chat request to model %s", c.modelName)

	// Create stream response channel
	streamChan := make(chan types.StreamResponse)

	// Start goroutine to handle streaming response
	go func() {
		defer close(streamChan)

		err := c.ollamaService.Chat(ctx, chatReq, func(resp ollamaapi.ChatResponse) error {
			if resp.Message.Content != "" {
				streamChan <- types.StreamResponse{
					ResponseType: types.ResponseTypeAnswer,
					Content:      resp.Message.Content,
					Done:         false,
				}
			}

			if resp.Done {
				streamChan <- types.StreamResponse{
					ResponseType: types.ResponseTypeAnswer,
					Done:         true,
				}
			}

			return nil
		})
		if err != nil {
			logger.GetLogger(ctx).Errorf("streaming chat request failed: %v", err)
			// Send end signal on error
			streamChan <- types.StreamResponse{
				ResponseType: types.ResponseTypeAnswer,
				Done:         true,
			}
		}
	}()

	return streamChan, nil
}

// ensureModelAvailable ensures the model is available
func (c *OllamaChat) ensureModelAvailable(ctx context.Context) error {
	logger.GetLogger(ctx).Infof("Ensure model %s is available", c.modelName)
	return c.ollamaService.EnsureModelAvailable(ctx, c.modelName)
}

// GetModelName returns the model name
func (c *OllamaChat) GetModelName() string {
	return c.modelName
}

// GetModelID returns the model ID
func (c *OllamaChat) GetModelID() string {
	return c.modelID
}
