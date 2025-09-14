// Package service provides the application's core business logic service layer
// This package includes core features such as knowledge base management,
// tenant/user management, and model services
package service

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/Tencent/WeKnowRust/internal/config"
	"github.com/Tencent/WeKnowRust/internal/logger"
	"github.com/Tencent/WeKnowRust/internal/models/chat"
	"github.com/Tencent/WeKnowRust/internal/models/embedding"
	"github.com/Tencent/WeKnowRust/internal/models/rerank"
	"github.com/Tencent/WeKnowRust/internal/models/utils/ollama"
	"github.com/Tencent/WeKnowRust/internal/types"
	"github.com/Tencent/WeKnowRust/internal/types/interfaces"
)

// TestDataService is responsible for initializing test data, including creating
// a test tenant and a test knowledge base, and configuring the required model services
type TestDataService struct {
	config        *config.Config                     // Application configuration
	kbRepo        interfaces.KnowledgeBaseRepository // Knowledge base repository interface
	tenantService interfaces.TenantService           // Tenant service interface
	ollamaService *ollama.OllamaService              // Ollama model service
	modelService  interfaces.ModelService            // Model service interface
	EmbedModel    embedding.Embedder                 // Embedding model instance
	RerankModel   rerank.Reranker                    // Rerank model instance
	LLMModel      chat.Chat                          // Large language model instance
}

// NewTestDataService creates a new TestDataService
// Injects required dependent services and components
func NewTestDataService(
	config *config.Config,
	kbRepo interfaces.KnowledgeBaseRepository,
	tenantService interfaces.TenantService,
	ollamaService *ollama.OllamaService,
	modelService interfaces.ModelService,
) *TestDataService {
	return &TestDataService{
		config:        config,
		kbRepo:        kbRepo,
		tenantService: tenantService,
		ollamaService: ollamaService,
		modelService:  modelService,
	}
}

// initTenant initializes the test tenant
// It reads the tenant ID from environment variables; creates a new tenant if
// it doesn't exist, otherwise updates the existing tenant. It also configures
// the tenant's retriever engine parameters.
func (s *TestDataService) initTenant(ctx context.Context) error {
	logger.Info(ctx, "Start initializing test tenant")

	// Read tenant ID from environment
	tenantID := os.Getenv("INIT_TEST_TENANT_ID")
	logger.Infof(ctx, "Test tenant ID from environment: %s", tenantID)

	// Convert the string ID to uint64
	tenantIDUint, err := strconv.ParseUint(tenantID, 10, 64)
	if err != nil {
		logger.Errorf(ctx, "Failed to parse tenant ID: %v", err)
		return err
	}

	// Build tenant configuration
	tenantConfig := &types.Tenant{
		Name:        "Test Tenant",
		Description: "Test Tenant for Testing",
		RetrieverEngines: types.RetrieverEngines{
			Engines: []types.RetrieverEngineParams{
				{
					RetrieverType:       types.KeywordsRetrieverType,
					RetrieverEngineType: types.PostgresRetrieverEngineType,
				},
				{
					RetrieverType:       types.VectorRetrieverType,
					RetrieverEngineType: types.PostgresRetrieverEngineType,
				},
			},
		},
	}

	// Get or create the test tenant
	logger.Infof(ctx, "Attempting to get tenant with ID: %d", tenantIDUint)
	tenant, err := s.tenantService.GetTenantByID(ctx, uint(tenantIDUint))
	if err != nil {
		// Tenant not found; create a new one
		logger.Info(ctx, "Tenant not found, creating a new test tenant")
		tenant, err = s.tenantService.CreateTenant(ctx, tenantConfig)
		if err != nil {
			logger.Errorf(ctx, "Failed to create tenant: %v", err)
			return err
		}
		logger.Infof(ctx, "Created new test tenant with ID: %d", tenant.ID)
	} else {
		// Tenant exists; update retriever engine configuration
		logger.Info(ctx, "Test tenant found, updating retriever engines")
		tenant.RetrieverEngines = tenantConfig.RetrieverEngines
		tenant, err = s.tenantService.UpdateTenant(ctx, tenant)
		if err != nil {
			logger.Errorf(ctx, "Failed to update tenant: %v", err)
			return err
		}
		logger.Info(ctx, "Test tenant updated successfully")
	}

	logger.Infof(ctx, "Test tenant configured - ID: %d, Name: %s, API Key: %s",
		tenant.ID, tenant.Name, tenant.APIKey)
	return nil
}

// initKnowledgeBase initializes the test knowledge base
// It reads the KB ID from environment variables, creates or updates the KB,
// and configures chunking strategy, embedding model, and summary model
func (s *TestDataService) initKnowledgeBase(ctx context.Context) error {
	logger.Info(ctx, "Start initializing test knowledge base")

	// Check tenant ID in context
	if ctx.Value(types.TenantIDContextKey).(uint) == 0 {
		logger.Warn(ctx, "Tenant ID is 0, skipping knowledge base initialization")
		return nil
	}

	// Read knowledge base ID from environment
	knowledgeBaseID := os.Getenv("INIT_TEST_KNOWLEDGE_BASE_ID")
	logger.Infof(ctx, "Test knowledge base ID from environment: %s", knowledgeBaseID)

	// Build knowledge base configuration
	kbConfig := &types.KnowledgeBase{
		ID:          knowledgeBaseID,
		Name:        "Test Knowledge Base",
		Description: "Knowledge Base for Testing",
		TenantID:    ctx.Value(types.TenantIDContextKey).(uint),
		ChunkingConfig: types.ChunkingConfig{
			ChunkSize:        s.config.KnowledgeBase.ChunkSize,
			ChunkOverlap:     s.config.KnowledgeBase.ChunkOverlap,
			Separators:       s.config.KnowledgeBase.SplitMarkers,
			EnableMultimodal: s.config.KnowledgeBase.ImageProcessing.EnableMultimodal,
		},
		EmbeddingModelID: s.EmbedModel.GetModelID(),
		SummaryModelID:   s.LLMModel.GetModelID(),
		RerankModelID:    s.RerankModel.GetModelID(),
	}

	// Initialize the test knowledge base
	logger.Info(ctx, "Attempting to get existing knowledge base")
	_, err := s.kbRepo.GetKnowledgeBaseByID(ctx, knowledgeBaseID)
	if err != nil {
		// Knowledge base not found; create a new one
		logger.Info(ctx, "Knowledge base not found, creating a new one")
		logger.Infof(ctx, "Creating knowledge base with ID: %s, tenant ID: %d",
			kbConfig.ID, kbConfig.TenantID)

		if err := s.kbRepo.CreateKnowledgeBase(ctx, kbConfig); err != nil {
			logger.Errorf(ctx, "Failed to create knowledge base: %v", err)
			return err
		}
		logger.Info(ctx, "Knowledge base created successfully")
	} else {
		// Knowledge base exists; update its configuration
		logger.Info(ctx, "Knowledge base found, updating configuration")
		logger.Infof(ctx, "Updating knowledge base with ID: %s", kbConfig.ID)

		err = s.kbRepo.UpdateKnowledgeBase(ctx, kbConfig)
		if err != nil {
			logger.Errorf(ctx, "Failed to update knowledge base: %v", err)
			return err
		}
		logger.Info(ctx, "Knowledge base updated successfully")
	}

	logger.Infof(ctx, "Test knowledge base configured - ID: %s, Name: %s", kbConfig.ID, kbConfig.Name)
	return nil
}

// InitializeTestData orchestrates initialization of test data
// It initializes tenant, embedding model, rerank model, LLM model, and knowledge base
func (s *TestDataService) InitializeTestData(ctx context.Context) error {
	logger.Info(ctx, "Start initializing test data")

	// Read tenant ID from environment
	tenantID := os.Getenv("INIT_TEST_TENANT_ID")
	logger.Infof(ctx, "Test tenant ID from environment: %s", tenantID)

	// Parse tenant ID
	tenantIDUint, err := strconv.ParseUint(tenantID, 10, 64)
	if err != nil {
		// If parse fails, use default value 0
		logger.Warn(ctx, "Failed to parse tenant ID, using default value 0")
		tenantIDUint = 0
	} else {
		// Initialize tenant
		logger.Info(ctx, "Initializing tenant")
		err = s.initTenant(ctx)
		if err != nil {
			logger.Errorf(ctx, "Failed to initialize tenant: %v", err)
			return err
		}
		logger.Info(ctx, "Tenant initialized successfully")
	}

	// Create a new context with the tenant ID
	newCtx := context.Background()
	newCtx = context.WithValue(newCtx, types.TenantIDContextKey, uint(tenantIDUint))
	logger.Infof(ctx, "Created new context with tenant ID: %d", tenantIDUint)

	// Initialize models
	modelInitFuncs := []struct {
		name string
		fn   func(context.Context) error
	}{
		{"embedding model", s.initEmbeddingModel},
		{"rerank model", s.initRerankModel},
		{"LLM model", s.initLLMModel},
	}

	for _, initFunc := range modelInitFuncs {
		logger.Infof(ctx, "Initializing %s", initFunc.name)
		if err := initFunc.fn(newCtx); err != nil {
			logger.Errorf(ctx, "Failed to initialize %s: %v", initFunc.name, err)
			return err
		}
		logger.Infof(ctx, "%s initialized successfully", initFunc.name)
	}

	// Initialize knowledge base
	logger.Info(ctx, "Initializing knowledge base")
	if err := s.initKnowledgeBase(newCtx); err != nil {
		logger.Errorf(ctx, "Failed to initialize knowledge base: %v", err)
		return err
	}
	logger.Info(ctx, "Knowledge base initialized successfully")

	logger.Info(ctx, "Test data initialization completed")
	return nil
}

// getEnvOrError reads an environment variable or returns an error if unset
func (s *TestDataService) getEnvOrError(name string) (string, error) {
	value := os.Getenv(name)
	if value == "" {
		return "", fmt.Errorf("%s environment variable is not set", name)
	}
	return value, nil
}

// updateOrCreateModel updates or creates a model record
func (s *TestDataService) updateOrCreateModel(ctx context.Context, modelConfig *types.Model) error {
	model, err := s.modelService.GetModelByID(ctx, modelConfig.ID)
	if err != nil {
		// Model not found; create a new one
		return s.modelService.CreateModel(ctx, modelConfig)
	}

	// Model exists; update attributes
	model.TenantID = modelConfig.TenantID
	model.Name = modelConfig.Name
	model.Source = modelConfig.Source
	model.Type = modelConfig.Type
	model.Parameters = modelConfig.Parameters
	model.Status = modelConfig.Status

	return s.modelService.UpdateModel(ctx, model)
}

// initEmbeddingModel initializes the embedding model
func (s *TestDataService) initEmbeddingModel(ctx context.Context) error {
	// Read model parameters from environment
	modelName, err := s.getEnvOrError("INIT_EMBEDDING_MODEL_NAME")
	if err != nil {
		return err
	}

	dimensionStr := os.Getenv("INIT_EMBEDDING_MODEL_DIMENSION")
	dimension, err := strconv.Atoi(dimensionStr)
	if err != nil || dimension == 0 {
		return fmt.Errorf("invalid embedding model dimension: %s", dimensionStr)
	}

	baseURL := os.Getenv("INIT_EMBEDDING_MODEL_BASE_URL")
	apiKey := os.Getenv("INIT_EMBEDDING_MODEL_API_KEY")

	// Determine model source
	source := types.ModelSourceRemote
	if baseURL == "" {
		source = types.ModelSourceLocal
	}

	// Determine model ID
	modelID := os.Getenv("INIT_EMBEDDING_MODEL_ID")
	if modelID == "" {
		modelID = fmt.Sprintf("builtin:%s:%d", modelName, dimension)
	}

	// Create embedding model instance
	s.EmbedModel, err = embedding.NewEmbedder(embedding.Config{
		Source:     source,
		BaseURL:    baseURL,
		ModelName:  modelName,
		APIKey:     apiKey,
		Dimensions: dimension,
		ModelID:    modelID,
	})
	if err != nil {
		return fmt.Errorf("failed to create embedder: %w", err)
	}

	// If local model, pull it via Ollama
	if source == types.ModelSourceLocal && s.ollamaService != nil {
		if err := s.ollamaService.PullModel(context.Background(), modelName); err != nil {
			return fmt.Errorf("failed to pull embedding model: %w", err)
		}
	}

	// Build model configuration
	modelConfig := &types.Model{
		ID:       modelID,
		TenantID: ctx.Value(types.TenantIDContextKey).(uint),
		Name:     modelName,
		Source:   source,
		Type:     types.ModelTypeEmbedding,
		Parameters: types.ModelParameters{
			BaseURL: baseURL,
			APIKey:  apiKey,
			EmbeddingParameters: types.EmbeddingParameters{
				Dimension: dimension,
			},
		},
		Status: "active",
	}

	// Update or create model
	return s.updateOrCreateModel(ctx, modelConfig)
}

// initRerankModel initializes the rerank model
func (s *TestDataService) initRerankModel(ctx context.Context) error {
	// Read model parameters from environment
	modelName, err := s.getEnvOrError("INIT_RERANK_MODEL_NAME")
	if err != nil {
		logger.Warnf(ctx, "Skip Rerank Model: %v", err)
		return nil
	}

	baseURL, err := s.getEnvOrError("INIT_RERANK_MODEL_BASE_URL")
	if err != nil {
		return err
	}

	apiKey := os.Getenv("INIT_RERANK_MODEL_API_KEY")
	modelID := fmt.Sprintf("builtin:%s:rerank:%s", types.ModelSourceRemote, modelName)

	// Create rerank model instance
	s.RerankModel, err = rerank.NewReranker(&rerank.RerankerConfig{
		Source:    types.ModelSourceRemote,
		BaseURL:   baseURL,
		ModelName: modelName,
		APIKey:    apiKey,
		ModelID:   modelID,
	})
	if err != nil {
		return fmt.Errorf("failed to create reranker: %w", err)
	}

	// Build model configuration
	modelConfig := &types.Model{
		ID:       modelID,
		TenantID: ctx.Value(types.TenantIDContextKey).(uint),
		Name:     modelName,
		Source:   types.ModelSourceRemote,
		Type:     types.ModelTypeRerank,
		Parameters: types.ModelParameters{
			BaseURL: baseURL,
			APIKey:  apiKey,
		},
		Status: "active",
	}

	// Update or create model
	return s.updateOrCreateModel(ctx, modelConfig)
}

// initLLMModel initializes the large language model
func (s *TestDataService) initLLMModel(ctx context.Context) error {
	// Read model parameters from environment
	modelName, err := s.getEnvOrError("INIT_LLM_MODEL_NAME")
	if err != nil {
		return err
	}

	baseURL := os.Getenv("INIT_LLM_MODEL_BASE_URL")
	apiKey := os.Getenv("INIT_LLM_MODEL_API_KEY")

	// Determine model source
	source := types.ModelSourceRemote
	if baseURL == "" {
		source = types.ModelSourceLocal
	}

	// Determine model ID
	modelID := fmt.Sprintf("builtin:%s:llm:%s", source, modelName)

	// Create LLM instance
	s.LLMModel, err = chat.NewChat(&chat.ChatConfig{
		Source:    source,
		BaseURL:   baseURL,
		ModelName: modelName,
		APIKey:    apiKey,
		ModelID:   modelID,
	})
	if err != nil {
		return fmt.Errorf("failed to create llm: %w", err)
	}

	// If local model, pull it via Ollama
	if source == types.ModelSourceLocal && s.ollamaService != nil {
		if err := s.ollamaService.PullModel(context.Background(), modelName); err != nil {
			return fmt.Errorf("failed to pull llm model: %w", err)
		}
	}

	// Build model configuration
	modelConfig := &types.Model{
		ID:       modelID,
		TenantID: ctx.Value(types.TenantIDContextKey).(uint),
		Name:     modelName,
		Source:   source,
		Type:     types.ModelTypeKnowledgeQA,
		Parameters: types.ModelParameters{
			BaseURL: baseURL,
			APIKey:  apiKey,
		},
		Status: "active",
	}

	// Update or create model
	return s.updateOrCreateModel(ctx, modelConfig)
}
