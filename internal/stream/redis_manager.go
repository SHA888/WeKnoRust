package stream

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Tencent/WeKnowRust/internal/types"
	"github.com/Tencent/WeKnowRust/internal/types/interfaces"
	"github.com/redis/go-redis/v9"
)

// redisStreamInfo represents stream information stored in Redis
type redisStreamInfo struct {
	SessionID           string           `json:"session_id"`
	RequestID           string           `json:"request_id"`
	Query               string           `json:"query"`
	Content             string           `json:"content"`
	KnowledgeReferences types.References `json:"knowledge_references"`
	LastUpdated         time.Time        `json:"last_updated"`
	IsCompleted         bool             `json:"is_completed"`
}

// RedisStreamManager is a Redis-based stream manager implementation
type RedisStreamManager struct {
	client *redis.Client
	ttl    time.Duration // TTL for stream data in Redis
	prefix string        // Redis key prefix
}

// NewRedisStreamManager creates a new Redis stream manager
func NewRedisStreamManager(redisAddr, redisPassword string,
	redisDB int, prefix string, ttl time.Duration,
) (*RedisStreamManager, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       redisDB,
	})

	// Validate connection
	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	if ttl == 0 {
		ttl = 24 * time.Hour // default TTL is 24 hours
	}

	if prefix == "" {
		prefix = "stream:" // default prefix
	}

	return &RedisStreamManager{
		client: client,
		ttl:    ttl,
		prefix: prefix,
	}, nil
}

// buildKey builds the Redis key
func (r *RedisStreamManager) buildKey(sessionID, requestID string) string {
	return fmt.Sprintf("%s:%s:%s", r.prefix, sessionID, requestID)
}

// RegisterStream registers a new stream
func (r *RedisStreamManager) RegisterStream(ctx context.Context, sessionID, requestID, query string) error {
	info := &redisStreamInfo{
		SessionID:   sessionID,
		RequestID:   requestID,
		Query:       query,
		LastUpdated: time.Now(),
	}

	data, err := json.Marshal(info)
	if err != nil {
		return fmt.Errorf("failed to marshal stream info: %w", err)
	}

	key := r.buildKey(sessionID, requestID)
	return r.client.Set(ctx, key, data, r.ttl).Err()
}

// UpdateStream updates the stream content
func (r *RedisStreamManager) UpdateStream(ctx context.Context, sessionID, requestID string, content string, references types.References) error {
	key := r.buildKey(sessionID, requestID)

	// Get current data
	data, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil // key does not exist, might have expired
		}
		return fmt.Errorf("failed to get stream data: %w", err)
	}

	var info redisStreamInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return fmt.Errorf("failed to unmarshal stream data: %w", err)
	}

	// Update data
	info.Content += content
	if len(references) > 0 {
		info.KnowledgeReferences = references
	}
	info.LastUpdated = time.Now()

	// Save back to Redis
	updatedData, err := json.Marshal(info)
	if err != nil {
		return fmt.Errorf("failed to marshal updated stream info: %w", err)
	}

	return r.client.Set(ctx, key, updatedData, r.ttl).Err()
}

// CompleteStream marks a stream as completed
func (r *RedisStreamManager) CompleteStream(ctx context.Context, sessionID, requestID string) error {
	key := r.buildKey(sessionID, requestID)

	// Get current data
	data, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil // key does not exist, might have expired
		}
		return fmt.Errorf("failed to get stream data: %w", err)
	}

	var info redisStreamInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return fmt.Errorf("failed to unmarshal stream data: %w", err)
	}

	// Mark as completed
	info.IsCompleted = true
	info.LastUpdated = time.Now()

	// Save back to Redis
	updatedData, err := json.Marshal(info)
	if err != nil {
		return fmt.Errorf("failed to marshal updated stream info: %w", err)
	}

	// Delete the stream after 30 seconds
	go func() {
		time.Sleep(30 * time.Second)
		r.client.Del(ctx, key)
	}()
	return r.client.Set(ctx, key, updatedData, r.ttl).Err()
}

// GetStream retrieves a specific stream
func (r *RedisStreamManager) GetStream(ctx context.Context, sessionID, requestID string) (*interfaces.StreamInfo, error) {
	key := r.buildKey(sessionID, requestID)

	// Get data
	data, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // key does not exist
		}
		return nil, fmt.Errorf("failed to get stream data: %w", err)
	}

	var info redisStreamInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return nil, fmt.Errorf("failed to unmarshal stream data: %w", err)
	}

	// Convert to interface structure
	return &interfaces.StreamInfo{
		SessionID:           info.SessionID,
		RequestID:           info.RequestID,
		Query:               info.Query,
		Content:             info.Content,
		KnowledgeReferences: info.KnowledgeReferences,
		LastUpdated:         info.LastUpdated,
		IsCompleted:         info.IsCompleted,
	}, nil
}

// Close closes the Redis connection
func (r *RedisStreamManager) Close() error {
	return r.client.Close()
}

// Ensure interface implementation
var _ interfaces.StreamManager = (*RedisStreamManager)(nil)
