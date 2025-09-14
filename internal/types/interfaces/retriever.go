package interfaces

import (
	"context"

	"github.com/Tencent/WeKnowRust/internal/models/embedding"
	"github.com/Tencent/WeKnowRust/internal/types"
)

// RetrieveEngine defines the retrieve engine interface
type RetrieveEngine interface {
	// EngineType gets the retrieve engine type
	EngineType() types.RetrieverEngineType

	// Retrieve executes the retrieve
	Retrieve(ctx context.Context, params types.RetrieveParams) ([]*types.RetrieveResult, error)

	// Support gets the supported retrieve types
	Support() []types.RetrieverType
}

// RetrieveEngineRepository defines the retrieve engine repository interface
type RetrieveEngineRepository interface {
	// Save saves the index info
	Save(ctx context.Context, indexInfo *types.IndexInfo, params map[string]any) error

	// BatchSave saves the index info list
	BatchSave(ctx context.Context, indexInfoList []*types.IndexInfo, params map[string]any) error

	// EstimateStorageSize estimates the storage size
	EstimateStorageSize(ctx context.Context, indexInfoList []*types.IndexInfo, params map[string]any) int64

	// DeleteByChunkIDList deletes the index info by chunk id list
	DeleteByChunkIDList(ctx context.Context, indexIDList []string, dimension int) error
    // Copy indices
    // sourceKnowledgeBaseID: source knowledge base ID
    // sourceToTargetKBIDMap: mapping from source knowledge base ID to target knowledge base ID
    // sourceToTargetChunkIDMap: mapping from source chunk ID to target chunk ID
    // targetKnowledgeBaseID: target knowledge base ID
    // params: additional parameters, e.g., embedding vectors
    CopyIndices(
        ctx context.Context,
        sourceKnowledgeBaseID string,
		sourceToTargetKBIDMap map[string]string,
		sourceToTargetChunkIDMap map[string]string,
		targetKnowledgeBaseID string,
		dimension int,
	) error

	// DeleteByKnowledgeIDList deletes the index info by knowledge id list
	DeleteByKnowledgeIDList(ctx context.Context, knowledgeIDList []string, dimension int) error

	// RetrieveEngine retrieves the engine
	RetrieveEngine
}

// RetrieveEngineRegistry defines the retrieve engine registry interface
type RetrieveEngineRegistry interface {
	// Register registers the retrieve engine service
	Register(indexService RetrieveEngineService) error
	// GetRetrieveEngineService gets the retrieve engine service
	GetRetrieveEngineService(engineType types.RetrieverEngineType) (RetrieveEngineService, error)
	// GetAllRetrieveEngineServices gets all retrieve engine services
	GetAllRetrieveEngineServices() []RetrieveEngineService
}

// RetrieveEngineService defines the retrieve engine service interface
type RetrieveEngineService interface {
	// Index indexes the index info
	Index(ctx context.Context,
		embedder embedding.Embedder,
		indexInfo *types.IndexInfo,
		retrieverTypes []types.RetrieverType,
	) error

	// BatchIndex indexes the index info list
	BatchIndex(ctx context.Context,
		embedder embedding.Embedder,
		indexInfoList []*types.IndexInfo,
		retrieverTypes []types.RetrieverType,
	) error

	// EstimateStorageSize estimates the storage size
	EstimateStorageSize(ctx context.Context,
		embedder embedding.Embedder,
		indexInfoList []*types.IndexInfo,
		retrieverTypes []types.RetrieverType,
	) int64
    // CopyIndices copies indices from the source knowledge base to the target knowledge base,
    // avoiding recomputation of embeddings
    // sourceKnowledgeBaseID: source knowledge base ID
    // sourceToTargetKBIDMap: mapping from source knowledge base ID to target knowledge base ID
    // sourceToTargetChunkIDMap: mapping from source chunk ID to target chunk ID (key: source ID, value: target ID)
    // targetKnowledgeBaseID: target knowledge base ID
    CopyIndices(
        ctx context.Context,
        sourceKnowledgeBaseID string,
        sourceToTargetKBIDMap map[string]string,
        sourceToTargetChunkIDMap map[string]string,
        targetKnowledgeBaseID string,
        dimension int,
    ) error

	// DeleteByChunkIDList deletes the index info by chunk id list
	DeleteByChunkIDList(ctx context.Context, indexIDList []string, dimension int) error

	// DeleteByKnowledgeIDList deletes the index info by knowledge id list
	DeleteByKnowledgeIDList(ctx context.Context, knowledgeIDList []string, dimension int) error

	// RetrieveEngine retrieves the engine
	RetrieveEngine
}
