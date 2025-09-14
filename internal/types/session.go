package types

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// FallbackStrategy represents the fallback strategy type
type FallbackStrategy string

const (
	FallbackStrategyFixed FallbackStrategy = "fixed" // Fixed response
	FallbackStrategyModel FallbackStrategy = "model" // Model fallback response
)

type SummaryConfig struct {
	// Max tokens
	MaxTokens int `json:"max_tokens"`
	// Repeat penalty
	RepeatPenalty float64 `json:"repeat_penalty"`
	// TopK
	TopK int `json:"top_k"`
	// TopP
	TopP float64 `json:"top_p"`
	// Frequency penalty
	FrequencyPenalty float64 `json:"frequency_penalty"`
	// Presence penalty
	PresencePenalty float64 `json:"presence_penalty"`
	// Prompt
	Prompt string `json:"prompt"`
	// Context template
	ContextTemplate string `json:"context_template"`
	// No match prefix
	NoMatchPrefix string `json:"no_match_prefix"`
	// Temperature
	Temperature float64 `json:"temperature"`
	// Seed
	Seed int `json:"seed"`
	// Max completion tokens
	MaxCompletionTokens int `json:"max_completion_tokens"`
}

// Session represents the session
type Session struct {
	// ID
	ID string `json:"id" gorm:"type:varchar(36);primaryKey"`
	// Title
	Title string `json:"title"`
	// Description
	Description string `json:"description"`
	// Tenant ID
	TenantID uint `json:"tenant_id" gorm:"index"`

	// Strategy configuration
	KnowledgeBaseID   string           `json:"knowledge_base_id"`                   // Associated knowledge base ID
	MaxRounds         int              `json:"max_rounds"`                          // Number of rounds to keep in multi-turn
	EnableRewrite     bool             `json:"enable_rewrite"`                      // Enable multi-turn query rewrite
	FallbackStrategy  FallbackStrategy `json:"fallback_strategy"`                   // Fallback strategy
	FallbackResponse  string           `json:"fallback_response"`                   // Fixed fallback response content
	EmbeddingTopK     int              `json:"embedding_top_k"`                     // Vector recall TopK
	KeywordThreshold  float64          `json:"keyword_threshold"`                   // Keyword recall threshold
	VectorThreshold   float64          `json:"vector_threshold"`                    // Vector recall threshold
	RerankModelID     string           `json:"rerank_model_id"`                     // Rerank model ID
	RerankTopK        int              `json:"rerank_top_k"`                        // Rerank TopK
	RerankThreshold   float64          `json:"rerank_threshold"`                    // Rerank threshold
	SummaryModelID    string           `json:"summary_model_id"`                    // Summary model ID
	SummaryParameters *SummaryConfig   `json:"summary_parameters" gorm:"type:json"` // Summary model parameters

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`

	// Association relationship, not stored in the database
	Messages []Message `json:"-" gorm:"foreignKey:SessionID"`
}

func (s *Session) BeforeCreate(tx *gorm.DB) (err error) {
	s.ID = uuid.New().String()
	return nil
}

type StringArray []string

// Value implements the driver.Valuer interface, used to convert StringArray to database value
func (c StringArray) Value() (driver.Value, error) {
	return json.Marshal(c)
}

// Scan implements the sql.Scanner interface, used to convert database value to StringArray
func (c *StringArray) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	b, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(b, c)
}

// Value implements the driver.Valuer interface, used to convert SummaryConfig to database value
func (c *SummaryConfig) Value() (driver.Value, error) {
	return json.Marshal(c)
}

// Scan implements the sql.Scanner interface, used to convert database value to SummaryConfig
func (c *SummaryConfig) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	b, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(b, c)
}
