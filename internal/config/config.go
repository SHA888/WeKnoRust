package config

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/viper"
)

// Config holds the application's configuration
type Config struct {
	Conversation   *ConversationConfig   `yaml:"conversation" json:"conversation"`
	Server         *ServerConfig         `yaml:"server" json:"server"`
	KnowledgeBase  *KnowledgeBaseConfig  `yaml:"knowledge_base" json:"knowledge_base"`
	Tenant         *TenantConfig         `yaml:"tenant" json:"tenant"`
	Models         []ModelConfig         `yaml:"models" json:"models"`
	Asynq          *AsynqConfig          `yaml:"asynq" json:"asynq"`
	VectorDatabase *VectorDatabaseConfig `yaml:"vector_database" json:"vector_database"`
	DocReader      *DocReaderConfig      `yaml:"docreader" json:"docreader"`
	StreamManager  *StreamManagerConfig  `yaml:"stream_manager" json:"stream_manager"`
}

type DocReaderConfig struct {
	Addr string `yaml:"addr" json:"addr"`
}

type VectorDatabaseConfig struct {
	Driver string `yaml:"driver" json:"driver"`
}

// ConversationConfig configures the conversation service
type ConversationConfig struct {
	MaxRounds                  int            `yaml:"max_rounds" json:"max_rounds"`
	KeywordThreshold           float64        `yaml:"keyword_threshold" json:"keyword_threshold"`
	EmbeddingTopK              int            `yaml:"embedding_top_k" json:"embedding_top_k"`
	VectorThreshold            float64        `yaml:"vector_threshold" json:"vector_threshold"`
	RerankTopK                 int            `yaml:"rerank_top_k" json:"rerank_top_k"`
	RerankThreshold            float64        `yaml:"rerank_threshold" json:"rerank_threshold"`
	FallbackStrategy           string         `yaml:"fallback_strategy" json:"fallback_strategy"`
	FallbackResponse           string         `yaml:"fallback_response" json:"fallback_response"`
	FallbackPrompt             string         `yaml:"fallback_prompt" json:"fallback_prompt"`
	EnableRewrite              bool           `yaml:"enable_rewrite" json:"enable_rewrite"`
	EnableRerank               bool           `yaml:"enable_rerank" json:"enable_rerank"`
	Summary                    *SummaryConfig `yaml:"summary" json:"summary"`
	GenerateSessionTitlePrompt string         `yaml:"generate_session_title_prompt" json:"generate_session_title_prompt"`
	GenerateSummaryPrompt      string         `yaml:"generate_summary_prompt" json:"generate_summary_prompt"`
	RewritePromptSystem        string         `yaml:"rewrite_prompt_system" json:"rewrite_prompt_system"`
	RewritePromptUser          string         `yaml:"rewrite_prompt_user" json:"rewrite_prompt_user"`
	SimplifyQueryPrompt        string         `yaml:"simplify_query_prompt" json:"simplify_query_prompt"`
	SimplifyQueryPromptUser    string         `yaml:"simplify_query_prompt_user" json:"simplify_query_prompt_user"`
	ExtractEntitiesPrompt      string         `yaml:"extract_entities_prompt" json:"extract_entities_prompt"`
	ExtractRelationshipsPrompt string         `yaml:"extract_relationships_prompt" json:"extract_relationships_prompt"`
}

// SummaryConfig configures summarization behavior
type SummaryConfig struct {
	MaxTokens           int     `yaml:"max_tokens" json:"max_tokens"`
	RepeatPenalty       float64 `yaml:"repeat_penalty" json:"repeat_penalty"`
	TopK                int     `yaml:"top_k" json:"top_k"`
	TopP                float64 `yaml:"top_p" json:"top_p"`
	FrequencyPenalty    float64 `yaml:"frequency_penalty" json:"frequency_penalty"`
	PresencePenalty     float64 `yaml:"presence_penalty" json:"presence_penalty"`
	Prompt              string  `yaml:"prompt" json:"prompt"`
	ContextTemplate     string  `yaml:"context_template" json:"context_template"`
	Temperature         float64 `yaml:"temperature" json:"temperature"`
	Seed                int     `yaml:"seed" json:"seed"`
	MaxCompletionTokens int     `yaml:"max_completion_tokens" json:"max_completion_tokens"`
	NoMatchPrefix       string  `yaml:"no_match_prefix" json:"no_match_prefix"`
}

// ServerConfig configures the HTTP server
type ServerConfig struct {
	Port            int           `yaml:"port" json:"port"`
	Host            string        `yaml:"host" json:"host"`
	LogPath         string        `yaml:"log_path" json:"log_path"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout" json:"shutdown_timeout" default:"30s"`
}

// KnowledgeBaseConfig configures the knowledge base
type KnowledgeBaseConfig struct {
	ChunkSize       int                    `yaml:"chunk_size" json:"chunk_size"`
	ChunkOverlap    int                    `yaml:"chunk_overlap" json:"chunk_overlap"`
	SplitMarkers    []string               `yaml:"split_markers" json:"split_markers"`
	KeepSeparator   bool                   `yaml:"keep_separator" json:"keep_separator"`
	ImageProcessing *ImageProcessingConfig `yaml:"image_processing" json:"image_processing"`
}

// ImageProcessingConfig configures image processing
type ImageProcessingConfig struct {
	EnableMultimodal bool `yaml:"enable_multimodal" json:"enable_multimodal"`
}

// TenantConfig configures tenant defaults
type TenantConfig struct {
	DefaultSessionName        string `yaml:"default_session_name" json:"default_session_name"`
	DefaultSessionTitle       string `yaml:"default_session_title" json:"default_session_title"`
	DefaultSessionDescription string `yaml:"default_session_description" json:"default_session_description"`
}

// ModelConfig configures models available to the system
type ModelConfig struct {
	Type       string                 `yaml:"type" json:"type"`
	Source     string                 `yaml:"source" json:"source"`
	ModelName  string                 `yaml:"model_name" json:"model_name"`
	Parameters map[string]interface{} `yaml:"parameters" json:"parameters"`
}

type AsynqConfig struct {
	Addr         string        `yaml:"addr" json:"addr"`
	Username     string        `yaml:"username" json:"username"`
	Password     string        `yaml:"password" json:"password"`
	ReadTimeout  time.Duration `yaml:"read_timeout" json:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout" json:"write_timeout"`
	Concurrency  int           `yaml:"concurrency" json:"concurrency"`
}

// StreamManagerConfig configures the stream manager
type StreamManagerConfig struct {
	Type           string        `yaml:"type" json:"type"`                       // Type: "memory" or "redis"
	Redis          RedisConfig   `yaml:"redis" json:"redis"`                     // Redis configuration
	CleanupTimeout time.Duration `yaml:"cleanup_timeout" json:"cleanup_timeout"` // Cleanup timeout
}

// RedisConfig holds Redis connection settings
type RedisConfig struct {
	Address  string        `yaml:"address" json:"address"`   // Redis address
	Password string        `yaml:"password" json:"password"` // Redis password
	DB       int           `yaml:"db" json:"db"`             // Redis database index
	Prefix   string        `yaml:"prefix" json:"prefix"`     // Key prefix
	TTL      time.Duration `yaml:"ttl" json:"ttl"`           // Expiration duration
}

// LoadConfig loads configuration from files and environment
func LoadConfig() (*Config, error) {
	// Set config file name and search paths
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")
	viper.AddConfigPath("$HOME/.appname")
	viper.AddConfigPath("/etc/appname/")

	// Enable environment variables
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Read configuration file
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	// Replace environment variable placeholders
	configFileContent, err := os.ReadFile(viper.ConfigFileUsed())
	if err != nil {
		return nil, fmt.Errorf("error reading config file content: %w", err)
	}

	// Replace ${ENV_VAR} placeholders with env values
	re := regexp.MustCompile(`\${([^}]+)}`)
	result := re.ReplaceAllStringFunc(string(configFileContent), func(match string) string {
		// Extract env var name (strip ${})
		envVar := match[2 : len(match)-1]
		// Replace with env value if present
		if value := os.Getenv(envVar); value != "" {
			return value
		}
		return match
	})

	// Use the processed configuration content
	viper.ReadConfig(strings.NewReader(result))

	// Unmarshal into struct
	var cfg Config
	if err := viper.Unmarshal(&cfg, func(dc *mapstructure.DecoderConfig) {
		dc.TagName = "yaml"
	}); err != nil {
		return nil, fmt.Errorf("unable to decode config into struct: %w", err)
	}
	fmt.Printf("Using configuration file: %s\n", viper.ConfigFileUsed())
	return &cfg, nil
}
