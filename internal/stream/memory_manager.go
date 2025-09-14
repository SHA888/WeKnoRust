package stream

import (
	"context"
	"sync"
	"time"

	"github.com/Tencent/WeKnowRust/internal/types"
	"github.com/Tencent/WeKnowRust/internal/types/interfaces"
)

// memoryStreamInfo holds in-memory stream information
type memoryStreamInfo struct {
	sessionID           string
	requestID           string
	query               string
	content             string
	knowledgeReferences types.References
	lastUpdated         time.Time
	isCompleted         bool
}

// MemoryStreamManager is an in-memory stream manager implementation
type MemoryStreamManager struct {
	// sessionID -> requestID -> stream data
	activeStreams map[string]map[string]*memoryStreamInfo
	mu            sync.RWMutex
}

// NewMemoryStreamManager creates a new in-memory stream manager
func NewMemoryStreamManager() *MemoryStreamManager {
	return &MemoryStreamManager{
		activeStreams: make(map[string]map[string]*memoryStreamInfo),
	}
}

// RegisterStream registers a new stream
func (m *MemoryStreamManager) RegisterStream(ctx context.Context, sessionID, requestID, query string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	info := &memoryStreamInfo{
		sessionID:   sessionID,
		requestID:   requestID,
		query:       query,
		lastUpdated: time.Now(),
	}

	if _, exists := m.activeStreams[sessionID]; !exists {
		m.activeStreams[sessionID] = make(map[string]*memoryStreamInfo)
	}

	m.activeStreams[sessionID][requestID] = info
	return nil
}

// UpdateStream updates stream content
func (m *MemoryStreamManager) UpdateStream(ctx context.Context,
	sessionID, requestID string, content string, references types.References,
) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if sessionMap, exists := m.activeStreams[sessionID]; exists {
		if stream, found := sessionMap[requestID]; found {
			stream.content += content
			if len(references) > 0 {
				stream.knowledgeReferences = references
			}
			stream.lastUpdated = time.Now()
		}
	}
	return nil
}

// CompleteStream marks a stream as completed
func (m *MemoryStreamManager) CompleteStream(ctx context.Context, sessionID, requestID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if sessionMap, exists := m.activeStreams[sessionID]; exists {
		if stream, found := sessionMap[requestID]; found {
			stream.isCompleted = true
			// Delete the stream after 30 seconds
			go func() {
				time.Sleep(30 * time.Second)
				m.mu.Lock()
				defer m.mu.Unlock()
				delete(sessionMap, requestID)
				if len(sessionMap) == 0 {
					delete(m.activeStreams, sessionID)
				}
			}()
		}
	}
	return nil
}

// GetStream retrieves a specific stream
func (m *MemoryStreamManager) GetStream(ctx context.Context,
	sessionID, requestID string,
) (*interfaces.StreamInfo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if sessionMap, exists := m.activeStreams[sessionID]; exists {
		if stream, found := sessionMap[requestID]; found {
			return &interfaces.StreamInfo{
				SessionID:           stream.sessionID,
				RequestID:           stream.requestID,
				Query:               stream.query,
				Content:             stream.content,
				KnowledgeReferences: stream.knowledgeReferences,
				LastUpdated:         stream.lastUpdated,
				IsCompleted:         stream.isCompleted,
			}, nil
		}
	}
	return nil, nil
}

// Ensure interface implementation
var _ interfaces.StreamManager = (*MemoryStreamManager)(nil)
