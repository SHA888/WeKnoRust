package chat

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/Tencent/WeKnowRust/internal/types"
	"github.com/sashabaranov/go-openai"
)

// RemoteAPIChat implements chat over a remote (OpenAI-compatible) API
type RemoteAPIChat struct {
	modelName string
	client    *openai.Client
	modelID   string
	baseURL   string
	apiKey    string
}

// QwenChatCompletionRequest is a custom request struct used for qwen models
type QwenChatCompletionRequest struct {
	openai.ChatCompletionRequest
	EnableThinking *bool `json:"enable_thinking,omitempty"` // qwen-specific field
}

// NewRemoteAPIChat creates a chat instance for remote API
func NewRemoteAPIChat(chatConfig *ChatConfig) (*RemoteAPIChat, error) {
	apiKey := chatConfig.APIKey
	config := openai.DefaultConfig(apiKey)
	if baseURL := chatConfig.BaseURL; baseURL != "" {
		config.BaseURL = baseURL
	}
	return &RemoteAPIChat{
		modelName: chatConfig.ModelName,
		client:    openai.NewClientWithConfig(config),
		modelID:   chatConfig.ModelID,
		baseURL:   chatConfig.BaseURL,
		apiKey:    apiKey,
	}, nil
}

// convertMessages converts messages to OpenAI format
func (c *RemoteAPIChat) convertMessages(messages []Message) []openai.ChatCompletionMessage {
	openaiMessages := make([]openai.ChatCompletionMessage, len(messages))
	for i, msg := range messages {
		openaiMessages[i] = openai.ChatCompletionMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}
	return openaiMessages
}

// isAliyunQwen3Model checks whether the model is qwen3 on Aliyun compatible endpoint
func (c *RemoteAPIChat) isAliyunQwen3Model() bool {
	return strings.HasPrefix(c.modelName, "qwen3-") && c.baseURL == "https://dashscope.aliyuncs.com/compatible-mode/v1"
}

// buildQwenChatCompletionRequest builds a chat request for qwen models
func (c *RemoteAPIChat) buildQwenChatCompletionRequest(messages []Message,
	opts *ChatOptions, isStream bool,
) QwenChatCompletionRequest {
	req := QwenChatCompletionRequest{
		ChatCompletionRequest: c.buildChatCompletionRequest(messages, opts, isStream),
	}

	// For qwen models, force enable_thinking=false on non-stream calls
	if !isStream {
		enableThinking := false
		req.EnableThinking = &enableThinking
	}
	return req
}

// buildChatCompletionRequest builds a standard chat completion request
func (c *RemoteAPIChat) buildChatCompletionRequest(messages []Message,
	opts *ChatOptions, isStream bool,
) openai.ChatCompletionRequest {
	req := openai.ChatCompletionRequest{
		Model:    c.modelName,
		Messages: c.convertMessages(messages),
		Stream:   isStream,
	}

	// Add optional parameters
	if opts != nil {
		if opts.Temperature > 0 {
			req.Temperature = float32(opts.Temperature)
		}
		if opts.TopP > 0 {
			req.TopP = float32(opts.TopP)
		}
		if opts.MaxTokens > 0 {
			req.MaxTokens = opts.MaxTokens
		}
		if opts.MaxCompletionTokens > 0 {
			req.MaxCompletionTokens = opts.MaxCompletionTokens
		}
		if opts.FrequencyPenalty > 0 {
			req.FrequencyPenalty = float32(opts.FrequencyPenalty)
		}
		if opts.PresencePenalty > 0 {
			req.PresencePenalty = float32(opts.PresencePenalty)
		}
	}

	return req
}

// Chat performs a non-streaming chat request
func (c *RemoteAPIChat) Chat(ctx context.Context, messages []Message, opts *ChatOptions) (*types.ChatResponse, error) {
	// Use custom path for qwen models
	if c.isAliyunQwen3Model() {
		return c.chatWithQwen(ctx, messages, opts)
	}

	// Build request
	req := c.buildChatCompletionRequest(messages, opts, false)

	// Send request
	resp, err := c.client.CreateChatCompletion(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("create chat completion: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response from OpenAI")
	}

	// Convert response
	return &types.ChatResponse{
		Content: resp.Choices[0].Message.Content,
		Usage: struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		}{
			PromptTokens:     resp.Usage.PromptTokens,
			CompletionTokens: resp.Usage.CompletionTokens,
			TotalTokens:      resp.Usage.TotalTokens,
		},
	}, nil
}

// chatWithQwen handles qwen models with a custom request
func (c *RemoteAPIChat) chatWithQwen(ctx context.Context, messages []Message, opts *ChatOptions) (*types.ChatResponse, error) {
	// Build qwen request
	req := c.buildQwenChatCompletionRequest(messages, opts, false)

	// Marshal request
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	// Build endpoint URL
	endpoint := c.baseURL + "/chat/completions"

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	// Send request
	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status: %d", resp.StatusCode)
	}

	// Decode response
	var chatResp openai.ChatCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if len(chatResp.Choices) == 0 {
		return nil, fmt.Errorf("no response from API")
	}

	// Convert response
	return &types.ChatResponse{
		Content: chatResp.Choices[0].Message.Content,
		Usage: struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		}{
			PromptTokens:     chatResp.Usage.PromptTokens,
			CompletionTokens: chatResp.Usage.CompletionTokens,
			TotalTokens:      chatResp.Usage.TotalTokens,
		},
	}, nil
}

// ChatStream performs streaming chat
func (c *RemoteAPIChat) ChatStream(ctx context.Context,
	messages []Message, opts *ChatOptions,
) (<-chan types.StreamResponse, error) {
	// Build request
	req := c.buildChatCompletionRequest(messages, opts, true)

	// Create stream response channel
	streamChan := make(chan types.StreamResponse)

	// Start stream request
	stream, err := c.client.CreateChatCompletionStream(ctx, req)
	if err != nil {
		close(streamChan)
		return nil, fmt.Errorf("create chat completion stream: %w", err)
	}

	// Consume stream in background
	go func() {
		defer close(streamChan)
		defer stream.Close()

		for {
			response, err := stream.Recv()
			if err != nil {
				streamChan <- types.StreamResponse{
					ResponseType: types.ResponseTypeAnswer,
					Done:         true,
				}
				return
			}
			if len(response.Choices) > 0 {
				streamChan <- types.StreamResponse{
					ResponseType: types.ResponseTypeAnswer,
					Content:      response.Choices[0].Delta.Content,
					Done:         false,
				}
			}
		}
	}()

	return streamChan, nil
}

// GetModelName returns the model name
func (c *RemoteAPIChat) GetModelName() string {
	return c.modelName
}

// GetModelID returns the model ID
func (c *RemoteAPIChat) GetModelID() string {
	return c.modelID
}
