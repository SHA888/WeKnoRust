package chat

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/Tencent/WeKnowRust/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRemoteAPIChat performs an end-to-end test of Remote API Chat functionality
func TestRemoteAPIChat(t *testing.T) {
    // Read environment variables
    deepseekAPIKey := os.Getenv("DEEPSEEK_API_KEY")
    aliyunAPIKey := os.Getenv("ALIYUN_API_KEY")

    // Define test configurations
    testConfigs := []struct {
		name    string
		apiKey  string
		config  *ChatConfig
		skipMsg string
	}{
		{
			name:   "DeepSeek API",
			apiKey: deepseekAPIKey,
			config: &ChatConfig{
				Source:    types.ModelSourceRemote,
				BaseURL:   "https://api.deepseek.com/v1",
				ModelName: "deepseek-chat",
				APIKey:    deepseekAPIKey,
				ModelID:   "deepseek-chat",
			},
			skipMsg: "DEEPSEEK_API_KEY environment variable not set",
		},
		{
			name:   "Aliyun DeepSeek",
			apiKey: aliyunAPIKey,
			config: &ChatConfig{
				Source:    types.ModelSourceRemote,
				BaseURL:   "https://dashscope.aliyuncs.com/compatible-mode/v1",
				ModelName: "deepseek-v3.1",
				APIKey:    aliyunAPIKey,
				ModelID:   "deepseek-v3.1",
			},
			skipMsg: "ALIYUN_API_KEY environment variable not set",
		},
		{
			name:   "Aliyun Qwen3-32b",
			apiKey: aliyunAPIKey,
			config: &ChatConfig{
				Source:    types.ModelSourceRemote,
				BaseURL:   "https://dashscope.aliyuncs.com/compatible-mode/v1",
				ModelName: "qwen3-32b",
				APIKey:    aliyunAPIKey,
				ModelID:   "qwen3-32b",
			},
			skipMsg: "ALIYUN_API_KEY environment variable not set",
		},
		{
			name:   "Aliyun Qwen-max",
			apiKey: aliyunAPIKey,
			config: &ChatConfig{
				Source:    types.ModelSourceRemote,
				BaseURL:   "https://dashscope.aliyuncs.com/compatible-mode/v1",
				ModelName: "qwen-max",
				APIKey:    aliyunAPIKey,
				ModelID:   "qwen-max",
			},
			skipMsg: "ALIYUN_API_KEY environment variable not set",
		},
	}

    // Test messages
    testMessages := []Message{
		{
			Role:    "user",
			Content: "test",
		},
	}

    // Test options
    testOptions := &ChatOptions{
        Temperature: 0.7,
        MaxTokens:   100,
    }

    // Create context
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    // Iterate all configurations and run tests
    for _, tc := range testConfigs {
        t.Run(tc.name, func(t *testing.T) {
            // Check API key
            if tc.apiKey == "" {
                t.Skip(tc.skipMsg)
            }

            // Create chat instance
            chat, err := NewRemoteAPIChat(tc.config)
            require.NoError(t, err)
            assert.Equal(t, tc.config.ModelName, chat.GetModelName())
            assert.Equal(t, tc.config.ModelID, chat.GetModelID())

            // Basic chat test
            t.Run("Basic Chat", func(t *testing.T) {
                response, err := chat.Chat(ctx, testMessages, testOptions)
                require.NoError(t, err)
                assert.NotEmpty(t, response.Content)
                assert.Greater(t, response.Usage.TotalTokens, 0)
                assert.Greater(t, response.Usage.PromptTokens, 0)
                assert.Greater(t, response.Usage.CompletionTokens, 0)

				t.Logf("%s Response: %s", tc.name, response.Content)
				t.Logf("Usage: Prompt=%d, Completion=%d, Total=%d",
					response.Usage.PromptTokens,
					response.Usage.CompletionTokens,
					response.Usage.TotalTokens)
			})

		})
	}
}
