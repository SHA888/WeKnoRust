// Package types defines data structures and types used throughout the system
// These types are shared across different service modules to ensure data consistency
package types

import (
	"time"

	"gorm.io/gorm"
)

// ChunkType defines different types of chunks
type ChunkType string

const (
	// ChunkTypeText represents a standard text chunk
	ChunkTypeText ChunkType = "text"
	// ChunkTypeImageOCR represents a chunk containing OCR text extracted from an image
	ChunkTypeImageOCR ChunkType = "image_ocr"
	// ChunkTypeImageCaption represents a chunk containing an image caption/description
	ChunkTypeImageCaption ChunkType = "image_caption"
	// ChunkTypeSummary represents a summary chunk
	ChunkTypeSummary = "summary"
	// ChunkTypeEntity represents an entity chunk
	ChunkTypeEntity ChunkType = "entity"
	// ChunkTypeRelationship represents a relationship chunk
	ChunkTypeRelationship ChunkType = "relationship"
)

// ImageInfo represents image information associated with a chunk
type ImageInfo struct {
	// Image URL (e.g., COS)
	URL string `json:"url" gorm:"type:text"`
	// Original image URL
	OriginalURL string `json:"original_url" gorm:"type:text"`
	// Start position of the image within the text
	StartPos int `json:"start_pos"`
	// End position of the image within the text
	EndPos int `json:"end_pos"`
	// Image caption/description
	Caption string `json:"caption"`
	// Image OCR text
	OCRText string `json:"ocr_text"`
}

// Chunk represents a document chunk
// Chunks are meaningful text segments extracted from original documents
// and are the basic units of knowledge base retrieval
// Each chunk contains a portion of the original content
// and maintains its positional relationship with the original text
// Chunks can be independently embedded as vectors and retrieved, supporting precise content localization
type Chunk struct {
	// Unique identifier of the chunk, using UUID format
	ID string `json:"id" gorm:"type:varchar(36);primaryKey"`
	// Tenant ID, used for multi-tenant isolation
	TenantID uint `json:"tenant_id"`
	// ID of the parent knowledge, associated with the Knowledge model
	KnowledgeID string `json:"knowledge_id"`
	// ID of the knowledge base, for quick location
	KnowledgeBaseID string `json:"knowledge_base_id"`
	// Actual text content of the chunk
	Content string `json:"content"`
	// Index position of the chunk in the original document
	ChunkIndex int `json:"chunk_index"`
	// Whether the chunk is enabled, can be used to temporarily disable certain chunks
	IsEnabled bool `json:"is_enabled" gorm:"default:true"`
	// Starting character position in the original text
	StartAt int `json:"start_at" `
	// Ending character position in the original text
	EndAt int `json:"end_at"`
	// Previous chunk ID
	PreChunkID string `json:"pre_chunk_id"`
	// Next chunk ID
	NextChunkID string `json:"next_chunk_id"`
	// Chunk type, used to distinguish different chunk categories
	ChunkType ChunkType `json:"chunk_type" gorm:"type:varchar(20);default:'text'"`
	// Parent chunk ID, used to associate image chunks with the original text chunk
	ParentChunkID string `json:"parent_chunk_id" gorm:"type:varchar(36);index"`
	// Relationship chunk IDs, used to associate relation chunks with the original text chunk
	RelationChunks JSON `json:"relation_chunks" gorm:"type:json"`
	// Indirect relationship chunk IDs, used to associate indirect relation chunks with the original text chunk
	IndirectRelationChunks JSON `json:"indirect_relation_chunks" gorm:"type:json"`
	// Image information, stored as JSON
	ImageInfo string `json:"image_info" gorm:"type:text"`
	// Chunk creation time
	CreatedAt time.Time `json:"created_at"`
	// Chunk last update time
	UpdatedAt time.Time `json:"updated_at"`
	// Soft delete marker, supports data recovery
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}
