package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"strconv"

	"github.com/Tencent/WeKnowRust/internal/config"
	"github.com/Tencent/WeKnowRust/internal/errors"
	"github.com/Tencent/WeKnowRust/internal/logger"
	"github.com/Tencent/WeKnowRust/internal/models/embedding"
	"github.com/Tencent/WeKnowRust/internal/models/utils/ollama"
	"github.com/Tencent/WeKnowRust/internal/types"
	"github.com/Tencent/WeKnowRust/internal/types/interfaces"
	"github.com/Tencent/WeKnowRust/services/docreader/src/client"
	"github.com/Tencent/WeKnowRust/services/docreader/src/proto"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ollama/ollama/api"
)

// DownloadTask represents a model download task
type DownloadTask struct {
	ID        string     `json:"id"`
	ModelName string     `json:"modelName"`
	Status    string     `json:"status"` // pending, downloading, completed, failed
	Progress  float64    `json:"progress"`
	Message   string     `json:"message"`
	StartTime time.Time  `json:"startTime"`
	EndTime   *time.Time `json:"endTime,omitempty"`
}

// Global download task manager
var (
	downloadTasks = make(map[string]*DownloadTask)
	tasksMutex    sync.RWMutex
)

// InitializationHandler handles system initialization
type InitializationHandler struct {
	config           *config.Config
	tenantService    interfaces.TenantService
	modelService     interfaces.ModelService
	kbService        interfaces.KnowledgeBaseService
	kbRepository     interfaces.KnowledgeBaseRepository
	knowledgeService interfaces.KnowledgeService
	ollamaService    *ollama.OllamaService
	docReaderClient  *client.Client
}

// NewInitializationHandler creates a new InitializationHandler
func NewInitializationHandler(
	config *config.Config,
	tenantService interfaces.TenantService,
	modelService interfaces.ModelService,
	kbService interfaces.KnowledgeBaseService,
	kbRepository interfaces.KnowledgeBaseRepository,
	knowledgeService interfaces.KnowledgeService,
	ollamaService *ollama.OllamaService,
	docReaderClient *client.Client,
) *InitializationHandler {
	return &InitializationHandler{
		config:           config,
		tenantService:    tenantService,
		modelService:     modelService,
		kbService:        kbService,
		kbRepository:     kbRepository,
		knowledgeService: knowledgeService,
		ollamaService:    ollamaService,
		docReaderClient:  docReaderClient,
	}
}

// InitializationRequest represents the initialization request payload
type InitializationRequest struct {
	// Storage type passed from frontend: "cos" or "minio"
	StorageType string `json:"storageType"`
	LLM         struct {
		Source    string `json:"source" binding:"required"`
		ModelName string `json:"modelName" binding:"required"`
		BaseURL   string `json:"baseUrl"`
		APIKey    string `json:"apiKey"`
	} `json:"llm" binding:"required"`

	Embedding struct {
		Source    string `json:"source" binding:"required"`
		ModelName string `json:"modelName" binding:"required"`
		BaseURL   string `json:"baseUrl"`
		APIKey    string `json:"apiKey"`
		Dimension int    `json:"dimension"` // Embedding vector dimension
	} `json:"embedding" binding:"required"`

	Rerank struct {
		Enabled   bool   `json:"enabled"`
		ModelName string `json:"modelName"`
		BaseURL   string `json:"baseUrl"`
		APIKey    string `json:"apiKey"`
	} `json:"rerank"`

	Multimodal struct {
		Enabled bool `json:"enabled"`
		VLM     *struct {
			ModelName     string `json:"modelName"`
			BaseURL       string `json:"baseUrl"`
			APIKey        string `json:"apiKey"`
			InterfaceType string `json:"interfaceType"` // "ollama" or "openai"
		} `json:"vlm,omitempty"`
		COS *struct {
			SecretID   string `json:"secretId"`
			SecretKey  string `json:"secretKey"`
			Region     string `json:"region"`
			BucketName string `json:"bucketName"`
			AppID      string `json:"appId"`
			PathPrefix string `json:"pathPrefix"`
		} `json:"cos,omitempty"`
		Minio *struct {
			BucketName string `json:"bucketName"`
			PathPrefix string `json:"pathPrefix"`
		} `json:"minio,omitempty"`
	} `json:"multimodal"`

	DocumentSplitting struct {
		ChunkSize    int      `json:"chunkSize" binding:"required,min=100,max=10000"`
		ChunkOverlap int      `json:"chunkOverlap" binding:"required,min=0"`
		Separators   []string `json:"separators" binding:"required,min=1"`
	} `json:"documentSplitting" binding:"required"`
}

// CheckStatus checks whether the system has been initialized
func (h *InitializationHandler) CheckStatus(c *gin.Context) {
	ctx := c.Request.Context()
	logger.Info(ctx, "Checking system initialization status")

	// Check if tenant exists
	tenant, err := h.tenantService.GetTenantByID(ctx, types.InitDefaultTenantID)
	if err != nil {
		logger.ErrorWithFields(ctx, err, nil)
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": gin.H{
				"initialized": false,
			},
		})
		return
	}

	// If no tenant exists, the system is not initialized
	if tenant == nil {
		logger.Info(ctx, "No tenants found, system not initialized")
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": gin.H{
				"initialized": false,
			},
		})
		return
	}
	ctx = context.WithValue(ctx, types.TenantIDContextKey, types.InitDefaultTenantID)

	// Check if models exist
	models, err := h.modelService.ListModels(ctx)
	if err != nil || len(models) == 0 {
		logger.Info(ctx, "No models found, system not initialized")
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": gin.H{
				"initialized": false,
			},
		})
		return
	}

	logger.Info(ctx, "System is already initialized")
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"initialized": true,
		},
	})
}

// Initialize performs system initialization
func (h *InitializationHandler) Initialize(c *gin.Context) {
	ctx := c.Request.Context()

	logger.Info(ctx, "Starting system initialization")

	var req InitializationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error(ctx, "Failed to parse initialization request", err)
		c.Error(errors.NewBadRequestError(err.Error()))
		return
	}

	// Validate multimodal configuration
	if req.Multimodal.Enabled {
		storageType := strings.ToLower(req.StorageType)
		if req.Multimodal.VLM == nil {
			logger.Error(ctx, "Multimodal enabled but missing VLM configuration")
			c.Error(errors.NewBadRequestError("VLM configuration is required when multimodal is enabled"))
			return
		}
		if req.Multimodal.VLM.ModelName == "" || req.Multimodal.VLM.BaseURL == "" {
			logger.Error(ctx, "VLM configuration incomplete")
			c.Error(errors.NewBadRequestError("VLM configuration is incomplete"))
			return
		}
		switch storageType {
		case "cos":
			if req.Multimodal.COS == nil || req.Multimodal.COS.SecretID == "" || req.Multimodal.COS.SecretKey == "" ||
				req.Multimodal.COS.Region == "" || req.Multimodal.COS.BucketName == "" ||
				req.Multimodal.COS.AppID == "" {
				logger.Error(ctx, "COS configuration incomplete")
				c.Error(errors.NewBadRequestError("COS configuration is incomplete"))
				return
			}
		case "minio":
			if req.Multimodal.Minio == nil || req.Multimodal.Minio.BucketName == "" ||
				os.Getenv("MINIO_ACCESS_KEY_ID") == "" || os.Getenv("MINIO_SECRET_ACCESS_KEY") == "" {
				logger.Error(ctx, "MinIO configuration incomplete")
				c.Error(errors.NewBadRequestError("MinIO configuration is incomplete"))
				return
			}
		}
	}

	// Validate rerank configuration (when enabled)
	if req.Rerank.Enabled {
		if req.Rerank.ModelName == "" || req.Rerank.BaseURL == "" {
			logger.Error(ctx, "Rerank configuration incomplete")
			c.Error(errors.NewBadRequestError("When Rerank is enabled, both model name and Base URL are required"))
			return
		}
	}

	// Validate document splitting configuration
	if req.DocumentSplitting.ChunkOverlap >= req.DocumentSplitting.ChunkSize {
		logger.Error(ctx, "Chunk overlap must be less than chunk size")
		c.Error(errors.NewBadRequestError("Chunk overlap must be less than chunk size"))
		return
	}
	if len(req.DocumentSplitting.Separators) == 0 {
		logger.Error(ctx, "Document separators cannot be empty")
		c.Error(errors.NewBadRequestError("Document separators cannot be empty"))
		return
	}
	var err error
	// 1. Handle tenant - check existence, create if missing
	tenant, _ := h.tenantService.GetTenantByID(ctx, types.InitDefaultTenantID)
	if tenant == nil {
		logger.Info(ctx, "Tenant not found, creating tenant")
		// Create default tenant
		tenant = &types.Tenant{
			ID:          types.InitDefaultTenantID,
			Name:        "Default Tenant",
			Description: "System Default Tenant",
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
		logger.Info(ctx, "Creating default tenant")
		tenant, err = h.tenantService.CreateTenant(ctx, tenant)
		if err != nil {
			logger.ErrorWithFields(ctx, err, nil)
			c.Error(errors.NewInternalServerError("Failed to create tenant"))
			return
		}
	} else {
		logger.Info(ctx, "Tenant exists, updating if needed")
		// Update tenant info if needed
		updated := false
		if tenant.Name != "Default Tenant" {
			tenant.Name = "Default Tenant"
			updated = true
		}
		if tenant.Description != "System Default Tenant" {
			tenant.Description = "System Default Tenant"
			updated = true
		}

		if updated {
			_, err = h.tenantService.UpdateTenant(ctx, tenant)
			if err != nil {
				logger.ErrorWithFields(ctx, err, nil)
				c.Error(errors.NewInternalServerError("Failed to update tenant: " + err.Error()))
				return
			}
			logger.Info(ctx, "Tenant updated successfully")
		}
	}

	// Create a new context with tenant ID
	newCtx := context.WithValue(ctx, types.TenantIDContextKey, types.InitDefaultTenantID)

	// 2. Handle models - update existing or create new ones
	existingModels, err := h.modelService.ListModels(newCtx)
	if err != nil {
		logger.ErrorWithFields(ctx, err, nil)
		// If listing fails, continue with creation flow
		existingModels = []*types.Model{}
	}

	// Build a map of models by type
	modelMap := make(map[types.ModelType]*types.Model)
	for _, model := range existingModels {
		modelMap[model.Type] = model
	}

	// Model configurations to process
	modelsToProcess := []struct {
		modelType   types.ModelType
		name        string
		source      types.ModelSource
		description string
		baseURL     string
		apiKey      string
		dimension   int
	}{
		{
			modelType:   types.ModelTypeKnowledgeQA,
			name:        req.LLM.ModelName,
			source:      types.ModelSource(req.LLM.Source),
			description: "LLM Model for Knowledge QA",
			baseURL:     req.LLM.BaseURL,
			apiKey:      req.LLM.APIKey,
		},
		{
			modelType:   types.ModelTypeEmbedding,
			name:        req.Embedding.ModelName,
			source:      types.ModelSource(req.Embedding.Source),
			description: "Embedding Model",
			baseURL:     req.Embedding.BaseURL,
			apiKey:      req.Embedding.APIKey,
			dimension:   req.Embedding.Dimension,
		},
	}

	// If rerank is enabled, append a rerank model
	if req.Rerank.Enabled {
		modelsToProcess = append(modelsToProcess, struct {
			modelType   types.ModelType
			name        string
			source      types.ModelSource
			description string
			baseURL     string
			apiKey      string
			dimension   int
		}{
			modelType:   types.ModelTypeRerank,
			name:        req.Rerank.ModelName,
			source:      types.ModelSourceRemote,
			description: "Rerank Model",
			baseURL:     req.Rerank.BaseURL,
			apiKey:      req.Rerank.APIKey,
		})
	}

	// If multimodal is enabled, append a VLM model
	if req.Multimodal.Enabled && req.Multimodal.VLM != nil {
		modelsToProcess = append(modelsToProcess, struct {
			modelType   types.ModelType
			name        string
			source      types.ModelSource
			description string
			baseURL     string
			apiKey      string
			dimension   int
		}{
			modelType:   types.ModelTypeVLLM,
			name:        req.Multimodal.VLM.ModelName,
			source:      types.ModelSourceRemote,
			description: "Vision Language Model",
			baseURL:     req.Multimodal.VLM.BaseURL,
			apiKey:      req.Multimodal.VLM.APIKey,
		})
	}

	// Process each model
	var processedModels []*types.Model
	for _, modelConfig := range modelsToProcess {
		existingModel, exists := modelMap[modelConfig.modelType]

		if exists {
			// Update existing model
			logger.Infof(ctx, "Updating existing model: %s (%s)",
				modelConfig.name, modelConfig.modelType,
			)
			existingModel.Name = modelConfig.name
			existingModel.Source = modelConfig.source
			existingModel.Description = modelConfig.description
			existingModel.Parameters = types.ModelParameters{
				BaseURL: modelConfig.baseURL,
				APIKey:  modelConfig.apiKey,
				EmbeddingParameters: types.EmbeddingParameters{
					Dimension: modelConfig.dimension,
				},
			}
			existingModel.IsDefault = true
			existingModel.Status = types.ModelStatusActive

			err := h.modelService.UpdateModel(newCtx, existingModel)
			if err != nil {
				logger.ErrorWithFields(ctx, err, map[string]interface{}{
					"model_name": modelConfig.name,
					"model_type": modelConfig.modelType,
				})
				c.Error(errors.NewInternalServerError("Failed to update model: " + err.Error()))
				return
			}
			processedModels = append(processedModels, existingModel)
		} else {
			// Create a new model
			logger.Infof(ctx,
				"Creating new model: %s (%s)",
				modelConfig.name, modelConfig.modelType,
			)
			newModel := &types.Model{
				TenantID:    types.InitDefaultTenantID,
				Name:        modelConfig.name,
				Type:        modelConfig.modelType,
				Source:      modelConfig.source,
				Description: modelConfig.description,
				Parameters: types.ModelParameters{
					BaseURL: modelConfig.baseURL,
					APIKey:  modelConfig.apiKey,
					EmbeddingParameters: types.EmbeddingParameters{
						Dimension: modelConfig.dimension,
					},
				},
				IsDefault: true,
				Status:    types.ModelStatusActive,
			}

			err := h.modelService.CreateModel(newCtx, newModel)
			if err != nil {
				logger.ErrorWithFields(ctx, err, map[string]interface{}{
					"model_name": modelConfig.name,
					"model_type": modelConfig.modelType,
				})
				c.Error(errors.NewInternalServerError("Failed to create model: " + err.Error()))
				return
			}
			processedModels = append(processedModels, newModel)
		}
	}

	// Delete VLM model if not needed (when multimodal is disabled)
	if !req.Multimodal.Enabled {
		if existingVLM, exists := modelMap[types.ModelTypeVLLM]; exists {
			logger.Info(ctx, "Deleting VLM model as multimodal is disabled")
			err := h.modelService.DeleteModel(newCtx, existingVLM.ID)
			if err != nil {
				logger.ErrorWithFields(ctx, err, map[string]interface{}{
					"model_id": existingVLM.ID,
				})
				// Log the error but do not block the flow
				logger.Warn(ctx, "Failed to delete VLM model, but continuing")
			}
		}
	}

	// Delete Rerank model if not needed (when rerank is disabled)
	if !req.Rerank.Enabled {
		if existingRerank, exists := modelMap[types.ModelTypeRerank]; exists {
			logger.Info(ctx, "Deleting Rerank model as rerank is disabled")
			err := h.modelService.DeleteModel(newCtx, existingRerank.ID)
			if err != nil {
				logger.ErrorWithFields(ctx, err, map[string]interface{}{
					"model_id": existingRerank.ID,
				})
				// Log the error but do not block the flow
				logger.Warn(ctx, "Failed to delete Rerank model, but continuing")
			}
		}
	}

	// 3. Handle knowledge base - create if missing, otherwise update
	kb, err := h.kbService.GetKnowledgeBaseByID(newCtx, types.InitDefaultKnowledgeBaseID)

	// Find embedding model ID and LLM model ID
	var embeddingModelID, llmModelID, rerankModelID, vlmModelID string
	for _, model := range processedModels {
		if model.Type == types.ModelTypeEmbedding {
			embeddingModelID = model.ID
		}
		if model.Type == types.ModelTypeKnowledgeQA {
			llmModelID = model.ID
		}
		if model.Type == types.ModelTypeRerank && req.Rerank.Enabled {
			rerankModelID = model.ID
		}
		if model.Type == types.ModelTypeVLLM {
			vlmModelID = model.ID
		}
	}

	if kb == nil {
		// Create default knowledge base
		logger.Info(ctx, "Creating default knowledge base")
		kb = &types.KnowledgeBase{
			ID:          types.InitDefaultKnowledgeBaseID,
			Name:        "Default Knowledge Base",
			Description: "System Default Knowledge Base",
			TenantID:    types.InitDefaultTenantID,
			ChunkingConfig: types.ChunkingConfig{
				ChunkSize:        req.DocumentSplitting.ChunkSize,
				ChunkOverlap:     req.DocumentSplitting.ChunkOverlap,
				Separators:       req.DocumentSplitting.Separators,
				EnableMultimodal: req.Multimodal.Enabled,
			},
			EmbeddingModelID: embeddingModelID,
			SummaryModelID:   llmModelID,
			RerankModelID:    rerankModelID,
			VLMModelID:       vlmModelID,
			VLMConfig: types.VLMConfig{
				ModelName:     req.Multimodal.VLM.ModelName,
				BaseURL:       req.Multimodal.VLM.BaseURL,
				APIKey:        req.Multimodal.VLM.APIKey,
				InterfaceType: req.Multimodal.VLM.InterfaceType,
			},
		}
		switch req.StorageType {
		case "cos":
			if req.Multimodal.COS != nil {
				kb.StorageConfig = types.StorageConfig{
					Provider:   req.StorageType,
					BucketName: req.Multimodal.COS.BucketName,
					AppID:      req.Multimodal.COS.AppID,
					PathPrefix: req.Multimodal.COS.PathPrefix,
					SecretID:   req.Multimodal.COS.SecretID,
					SecretKey:  req.Multimodal.COS.SecretKey,
					Region:     req.Multimodal.COS.Region,
				}
			}
		case "minio":
			if req.Multimodal.Minio != nil {
				kb.StorageConfig = types.StorageConfig{
					Provider:   req.StorageType,
					BucketName: req.Multimodal.Minio.BucketName,
					PathPrefix: req.Multimodal.Minio.PathPrefix,
					SecretID:   os.Getenv("MINIO_ACCESS_KEY_ID"),
					SecretKey:  os.Getenv("MINIO_SECRET_ACCESS_KEY"),
				}
			}
		}

		_, err = h.kbService.CreateKnowledgeBase(newCtx, kb)
		if err != nil {
			logger.ErrorWithFields(ctx, err, nil)
			c.Error(errors.NewInternalServerError("Failed to create knowledge base: " + err.Error()))
			return
		}
	} else {
		// Update existing knowledge base
		logger.Info(ctx, "Updating existing knowledge base")

		// Check if there are files, if so, do not allow updating the Embedding model
		// ...

		// Update basic information and configuration
		err = h.kbRepository.UpdateKnowledgeBase(newCtx, kb)
		if err != nil {
			logger.ErrorWithFields(ctx, err, nil)
			c.Error(errors.NewInternalServerError("Failed to update knowledge base configuration: " + err.Error()))
			return
		}

		// If necessary, update model IDs using the repository directly
		if !hasFiles || kb.SummaryModelID != llmModelID {
			// Refresh the knowledge base object to get the latest information
			kb, err = h.kbService.GetKnowledgeBaseByID(newCtx, types.InitDefaultKnowledgeBaseID)
			if err != nil {
				logger.ErrorWithFields(ctx, err, nil)
				c.Error(errors.NewInternalServerError("Failed to fetch updated knowledge base: " + err.Error()))
				return
			}

			// Update model IDs
			kb.SummaryModelID = llmModelID
			if req.Rerank.Enabled {
				kb.RerankModelID = rerankModelID
			} else {
				kb.RerankModelID = "" // Clear Rerank model ID
			}

			// Use the repository to update model IDs directly
			err = h.kbRepository.UpdateKnowledgeBase(newCtx, kb)
			if err != nil {
				logger.ErrorWithFields(ctx, err, nil)
				c.Error(errors.NewInternalServerError("Failed to update knowledge base model IDs: " + err.Error()))
				return
			}

			logger.Info(ctx, "Model IDs updated successfully")
		}
	}

	// ...

	available, err := h.ollamaService.IsModelAvailable(ctx, req.ModelName)
	if err != nil {
		logger.ErrorWithFields(ctx, err, map[string]interface{}{
			"model_name": req.ModelName,
		})
		c.Error(errors.NewInternalServerError("Failed to check model status: " + err.Error()))
		return
	}

	if available {
		logger.Infof(ctx, "Model %s already exists", req.ModelName)
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Model already exists",
			"data": gin.H{
				"modelName": req.ModelName,
				"status":    "completed",
				"progress":  100.0,
			},
		})
		return
	}

	// ...

	tasksMutex.RLock()
	for _, task := range downloadTasks {
		if task.ModelName == req.ModelName && (task.Status == "pending" || task.Status == "downloading") {
			tasksMutex.RUnlock()
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": "A download task for this model already exists",
				"data": gin.H{
					"taskId":    task.ID,
					"modelName": task.ModelName,
					"status":    task.Status,
					"progress":  task.Progress,
				},
			})
			return
		}
	}
	tasksMutex.RUnlock()

	// Create a download task
	taskID := uuid.New().String()
	task := &DownloadTask{
		ID:        taskID,
		ModelName: req.ModelName,
		Status:    "pending",
		Progress:  0.0,
		Message:   "Preparing download",
		StartTime: time.Now(),
	}

	tasksMutex.Lock()
	downloadTasks[taskID] = task
	tasksMutex.Unlock()

	// Start asynchronous download
	newCtx, cancel := context.WithTimeout(context.Background(), 12*time.Hour)
	go func() {
		defer cancel()
		h.downloadModelAsync(newCtx, taskID, req.ModelName)
	}()

	logger.Infof(ctx, "Created download task for model: %s, task ID: %s", req.ModelName, taskID)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Model download task created",
		"data": gin.H{
			"taskId":    taskID,
			"modelName": req.ModelName,
			"status":    "pending",
			"progress":  0.0,
		},
	})
}

// GetDownloadProgress returns the download progress
func (h *InitializationHandler) GetDownloadProgress(c *gin.Context) {
	taskID := c.Param("taskId")

	if taskID == "" {
		c.Error(errors.NewBadRequestError("taskId cannot be empty"))
		return
	}

	tasksMutex.RLock()
	task, exists := downloadTasks[taskID]
	tasksMutex.RUnlock()

	if !exists {
		c.Error(errors.NewNotFoundError("Download task not found"))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    task,
	})
}

// ListDownloadTasks lists all download tasks
func (h *InitializationHandler) ListDownloadTasks(c *gin.Context) {
	tasksMutex.RLock()
	tasks := make([]*DownloadTask, 0, len(downloadTasks))
	for _, task := range downloadTasks {
		tasks = append(tasks, task)
	}
	tasksMutex.RUnlock()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    tasks,
	})
}

// ListOllamaModels lists installed Ollama models
func (h *InitializationHandler) ListOllamaModels(c *gin.Context) {
	ctx := c.Request.Context()

	logger.Info(ctx, "Listing installed Ollama models")

	// Ensure service is available
	if !h.ollamaService.IsAvailable() {
		if err := h.ollamaService.StartService(ctx); err != nil {
			logger.ErrorWithFields(ctx, err, nil)
			c.Error(errors.NewInternalServerError("Ollama service is unavailable: " + err.Error()))
			return
		}
	}

	models, err := h.ollamaService.ListModels(ctx)
	if err != nil {
		logger.ErrorWithFields(ctx, err, nil)
		c.Error(errors.NewInternalServerError("Failed to list models: " + err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"models": models,
		},
	})
}

// downloadModelAsync downloads a model asynchronously
func (h *InitializationHandler) downloadModelAsync(ctx context.Context,
	taskID, modelName string,
) {
	logger.Infof(ctx, "Starting async download for model: %s, task: %s", modelName, taskID)

	// Update task status to downloading
	h.updateTaskStatus(taskID, "downloading", 0.0, "Starting model download")

	// Perform download with progress callback
	err := h.pullModelWithProgress(ctx, modelName, func(progress float64, message string) {
		h.updateTaskStatus(taskID, "downloading", progress, message)
	})

	if err != nil {
		logger.ErrorWithFields(ctx, err, map[string]interface{}{
			"model_name": modelName,
			"task_id":    taskID,
		})
		h.updateTaskStatus(taskID, "failed", 0.0, fmt.Sprintf("Download failed: %v", err))
		return
	}

	// Download succeeded
	logger.Infof(ctx, "Model %s downloaded successfully, task: %s", modelName, taskID)
	h.updateTaskStatus(taskID, "completed", 100.0, "Download completed")
}

// pullModelWithProgress pulls a model and reports progress
func (h *InitializationHandler) pullModelWithProgress(ctx context.Context,
	modelName string,
	progressCallback func(float64, string),
) error {
	// Ensure service is available
	if err := h.ollamaService.StartService(ctx); err != nil {
		logger.ErrorWithFields(ctx, err, nil)
		return err
	}

	// Check if the model already exists
	available, err := h.ollamaService.IsModelAvailable(ctx, modelName)
	if err != nil {
		logger.ErrorWithFields(ctx, err, map[string]interface{}{
			"model_name": modelName,
		})
		return err
	}
	if available {
		progressCallback(100.0, "Model already exists")
		return nil
	}

	logger.GetLogger(ctx).Infof("Pulling model %s...", modelName)

	// Create a download request
	pullReq := &api.PullRequest{
		Name: modelName,
	}

	// Use Ollama client Pull with progress callback
	err = h.ollamaService.GetClient().Pull(ctx, pullReq, func(progress api.ProgressResponse) error {
		var progressPercent float64 = 0.0
		var message string = "Downloading"

		if progress.Total > 0 && progress.Completed > 0 {
			progressPercent = float64(progress.Completed) / float64(progress.Total) * 100
			message = fmt.Sprintf("Downloading: %.1f%% (%s)", progressPercent, progress.Status)
		} else if progress.Status != "" {
			message = progress.Status
		}

		// Invoke progress callback
		progressCallback(progressPercent, message)

		logger.Infof(ctx,
			"Download progress for %s: %.2f%% - %s",
			modelName, progressPercent, message,
		)
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to pull model: %w", err)
	}

	return nil
}

// updateTaskStatus updates a task's status
func (h *InitializationHandler) updateTaskStatus(
	taskID, status string, progress float64, message string,
) {
	tasksMutex.Lock()
	defer tasksMutex.Unlock()

	if task, exists := downloadTasks[taskID]; exists {
		task.Status = status
		task.Progress = progress
		task.Message = message

		if status == "completed" || status == "failed" {
			now := time.Now()
			task.EndTime = &now
		}
	}
}

// GetCurrentConfig retrieves current system configuration
func (h *InitializationHandler) GetCurrentConfig(c *gin.Context) {
	ctx := c.Request.Context()

	logger.Info(ctx, "Getting current system configuration")

	// Set tenant context
	newCtx := context.WithValue(ctx, types.TenantIDContextKey, types.InitDefaultTenantID)

	// Get model information
	models, err := h.modelService.ListModels(newCtx)
	if err != nil {
		logger.ErrorWithFields(ctx, err, nil)
		c.Error(errors.NewInternalServerError("Failed to get model list: " + err.Error()))
		return
	}

	// Get knowledge base information
	kb, err := h.kbService.GetKnowledgeBaseByID(newCtx, types.InitDefaultKnowledgeBaseID)
	if err != nil {
		logger.ErrorWithFields(ctx, err, nil)
		c.Error(errors.NewInternalServerError("Failed to get knowledge base info: " + err.Error()))
		return
	}

	// Check if the knowledge base has files
	knowledgeList, err := h.knowledgeService.ListPagedKnowledgeByKnowledgeBaseID(newCtx,
		types.InitDefaultKnowledgeBaseID, &types.Pagination{
			Page:     1,
			PageSize: 1,
		})
	hasFiles := false
	if err == nil && knowledgeList != nil && knowledgeList.Total > 0 {
		hasFiles = true
	}

	// Build configuration response
	config := buildConfigResponse(models, kb, hasFiles)

	logger.Info(ctx, "Current system configuration retrieved successfully")
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    config,
	})
}

// buildConfigResponse builds the configuration response payload
func buildConfigResponse(models []*types.Model,
	kb *types.KnowledgeBase, hasFiles bool,
) map[string]interface{} {
	config := map[string]interface{}{
		"hasFiles": hasFiles,
	}

	// Group models by type
	for _, model := range models {
		switch model.Type {
		case types.ModelTypeKnowledgeQA:
			config["llm"] = map[string]interface{}{
				"source":    string(model.Source),
				"modelName": model.Name,
				"baseUrl":   model.Parameters.BaseURL,
				"apiKey":    model.Parameters.APIKey,
			}
		case types.ModelTypeEmbedding:
			config["embedding"] = map[string]interface{}{
				"source":    string(model.Source),
				"modelName": model.Name,
				"baseUrl":   model.Parameters.BaseURL,
				"apiKey":    model.Parameters.APIKey,
				"dimension": model.Parameters.EmbeddingParameters.Dimension,
			}
		case types.ModelTypeRerank:
			config["rerank"] = map[string]interface{}{
				"enabled":   true,
				"modelName": model.Name,
				"baseUrl":   model.Parameters.BaseURL,
				"apiKey":    model.Parameters.APIKey,
			}
		case types.ModelTypeVLLM:
			if config["multimodal"] == nil {
				config["multimodal"] = map[string]interface{}{
					"enabled": true,
				}
			}
			multimodal := config["multimodal"].(map[string]interface{})
			multimodal["vlm"] = map[string]interface{}{
				"modelName":     model.Name,
				"baseUrl":       model.Parameters.BaseURL,
				"apiKey":        model.Parameters.APIKey,
				"interfaceType": kb.VLMConfig.InterfaceType,
			}
		}
	}

	// If no VLM model, set multimodal to disabled
	if config["multimodal"] == nil {
		config["multimodal"] = map[string]interface{}{
			"enabled": false,
		}
	}

	// If no rerank model, set rerank to disabled
	if config["rerank"] == nil {
		config["rerank"] = map[string]interface{}{
			"enabled":   false,
			"modelName": "",
			"baseUrl":   "",
			"apiKey":    "",
		}
	}

	// Add knowledge base document splitting configuration
	if kb != nil {
		config["documentSplitting"] = map[string]interface{}{
			"chunkSize":    kb.ChunkingConfig.ChunkSize,
			"chunkOverlap": kb.ChunkingConfig.ChunkOverlap,
			"separators":   kb.ChunkingConfig.Separators,
		}

		// Add multimodal storage configuration (COS/MinIO)
		if kb.StorageConfig.SecretID != "" {
			if config["multimodal"] == nil {
				config["multimodal"] = map[string]interface{}{
					"enabled": true,
				}
			}
			multimodal := config["multimodal"].(map[string]interface{})
			multimodal["storageType"] = kb.StorageConfig.Provider
			switch kb.StorageConfig.Provider {
			case "cos":
				multimodal["cos"] = map[string]interface{}{
					"secretId":   kb.StorageConfig.SecretID,
					"secretKey":  kb.StorageConfig.SecretKey,
					"region":     kb.StorageConfig.Region,
					"bucketName": kb.StorageConfig.BucketName,
					"appId":      kb.StorageConfig.AppID,
					"pathPrefix": kb.StorageConfig.PathPrefix,
				}
			case "minio":
				multimodal["minio"] = map[string]interface{}{
					"bucketName": kb.StorageConfig.BucketName,
					"pathPrefix": kb.StorageConfig.PathPrefix,
				}
			}
		}
	}

	return config
}

// RemoteModelCheckRequest represents the remote model check request payload
type RemoteModelCheckRequest struct {
	ModelName string `json:"modelName" binding:"required"`
	BaseURL   string `json:"baseUrl" binding:"required"`
	APIKey    string `json:"apiKey"`
}

// CheckRemoteModel checks connectivity to a remote API model
func (h *InitializationHandler) CheckRemoteModel(c *gin.Context) {
	ctx := c.Request.Context()

	logger.Info(ctx, "Checking remote model connection")

	var req RemoteModelCheckRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error(ctx, "Failed to parse remote model check request", err)
		c.Error(errors.NewBadRequestError(err.Error()))
		return
	}

	// Validate request parameters
	if req.ModelName == "" || req.BaseURL == "" {
		logger.Error(ctx, "Model name and base URL are required")
		c.Error(errors.NewBadRequestError("Model name and Base URL cannot be empty"))
		return
	}

	// Create a temporary model configuration for testing
	modelConfig := &types.Model{
		Name:   req.ModelName,
		Source: "remote",
		Parameters: types.ModelParameters{
			BaseURL: req.BaseURL,
			APIKey:  req.APIKey,
		},
		Type: "llm", // Default type; the check does not depend on the exact type
	}

	// Check remote model connection
	available, message := h.checkRemoteModelConnection(ctx, modelConfig)

	logger.Info(ctx,
		fmt.Sprintf(
			"Remote model check completed: modelName=%s, baseUrl=%s, available=%v, message=%s",
			req.ModelName, req.BaseURL, available, message,
		),
	)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"available": available,
			"message":   message,
		},
	})
}

// TestEmbeddingModel tests whether the Embedding interface (local or remote) is available
func (h *InitializationHandler) TestEmbeddingModel(c *gin.Context) {
	ctx := c.Request.Context()

	logger.Info(ctx, "Testing embedding model connectivity and functionality")

	var req struct {
		Source    string `json:"source" binding:"required"`
		ModelName string `json:"modelName" binding:"required"`
		BaseURL   string `json:"baseUrl"`
		APIKey    string `json:"apiKey"`
		Dimension int    `json:"dimension"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error(ctx, "Failed to parse embedding test request", err)
		c.Error(errors.NewBadRequestError(err.Error()))
		return
	}

	// Build embedder configuration
	cfg := embedding.Config{
		Source:               types.ModelSource(strings.ToLower(req.Source)),
		BaseURL:              req.BaseURL,
		ModelName:            req.ModelName,
		APIKey:               req.APIKey,
		TruncatePromptTokens: 256,
		Dimensions:           req.Dimension,
		ModelID:              "",
	}

	emb, err := embedding.NewEmbedder(cfg)
	if err != nil {
		logger.ErrorWithFields(ctx, err, map[string]interface{}{"model": req.ModelName})
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": gin.H{
				"available": false,
				"message":   fmt.Sprintf("Failed to create embedder: %v", err),
				"dimension": 0,
			},
		})
		return
	}

	// Execute a minimal embedding call
	sample := "hello"
	vec, err := emb.Embed(ctx, sample)
	if err != nil {
		logger.ErrorWithFields(ctx, err, map[string]interface{}{"model": req.ModelName})
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": gin.H{
				"available": false,
				"message":   fmt.Sprintf("Embedding call failed: %v", err),
				"dimension": 0,
			},
		})
		return
	}

	logger.Infof(ctx, "Embedding test succeeded, dim=%d", len(vec))
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"available": true,
			"message":   fmt.Sprintf("Test succeeded, vector dimension=%d", len(vec)),
			"dimension": len(vec),
		},
	})
}

// checkRemoteModelConnection is an internal helper that checks remote model connectivity
func (h *InitializationHandler) checkRemoteModelConnection(ctx context.Context,
	model *types.Model,
) (bool, string) {
	// Use the /chat/completions endpoint for connectivity check
	// Send a minimal test request to validate connectivity and authentication

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Build test request
	testEndpoint := ""
	if model.Parameters.BaseURL != "" {
		testEndpoint = model.Parameters.BaseURL + "/chat/completions"
	}

	// Build test request body
	testRequest := map[string]interface{}{
		"model": model.Name,
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": "test",
			},
		},
		"max_tokens":      1,
		"enable_thinking": false, // for dashscope.aliyuncs qwen3-32b
	}

	jsonData, err := json.Marshal(testRequest)
	if err != nil {
		return false, fmt.Sprintf("Failed to build request body: %v", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", testEndpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return false, fmt.Sprintf("Failed to create request: %v", err)
	}

	// Add auth header if provided
	if model.Parameters.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+model.Parameters.APIKey)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Sprintf("Connection failed: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err == nil {
		logger.Infof(ctx, "Response body: %s", string(body))
	}

	// Check response status
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		// Connection OK, model available
		return true, "Connection OK, model available"
	} else if resp.StatusCode == 401 {
		return false, "Authentication failed, please check API key"
	} else if resp.StatusCode == 403 {
		return false, "Insufficient permissions, please check API key permissions"
	} else if resp.StatusCode == 404 {
		return false, "API endpoint not found, please check Base URL"
	} else {
		return false, fmt.Sprintf("API returned error status: %d", resp.StatusCode)
	}
}

// checkRerankModelConnection checks rerank model connectivity and basic functionality
func (h *InitializationHandler) checkRerankModelConnection(ctx context.Context,
	modelName, baseURL, apiKey string) (bool, string) {
	client := &http.Client{
		Timeout: 15 * time.Second,
	}

	// Build rerank API endpoint
	rerankEndpoint := baseURL + "/rerank"

	// Mock test data
	testQuery := "What is artificial intelligence?"
	testPassages := []string{
		"Machine learning is a subfield of AI focusing on algorithms and statistical models that improve with experience.",
		"Deep learning is a subset of machine learning using artificial neural networks to simulate the human brain.",
	}

	// Build rerank request
	rerankRequest := map[string]interface{}{
		"model":                  modelName,
		"query":                  testQuery,
		"documents":              testPassages,
		"truncate_prompt_tokens": 512,
	}

	jsonData, err := json.Marshal(rerankRequest)
	if err != nil {
		return false, fmt.Sprintf("Failed to build request: %v", err)
	}

	logger.Infof(ctx, "Rerank request: %s, modelName=%s, baseURL=%s, apiKey=%s",
		string(jsonData), modelName, baseURL, apiKey)

	req, err := http.NewRequestWithContext(
		ctx, "POST", rerankEndpoint, strings.NewReader(string(jsonData)),
	)
	if err != nil {
		return false, fmt.Sprintf("Failed to create request: %v", err)
	}

	// Add auth header if provided
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Sprintf("Connection failed: %v", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Sprintf("Failed to read response: %v", err)
	}

	// Check response status
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		// Try to parse rerank response
		var rerankResp struct {
			Results []struct {
				Index          int     `json:"index"`
				Document       string  `json:"document"`
				RelevanceScore float64 `json:"relevance_score"`
			} `json:"results"`
		}

		if err := json.Unmarshal(body, &rerankResp); err != nil {
			// If response is not in standard rerank format, treat connectivity as OK
			return true, "Connection OK, but response format is non-standard"
		}

		// Check if rerank results are returned
		if len(rerankResp.Results) > 0 {
			return true, fmt.Sprintf("Rerank works, returned %d results", len(rerankResp.Results))
		} else {
			return false, "Rerank API connected, but no results returned"
		}
	} else if resp.StatusCode == 401 {
		return false, "Authentication failed, please check API key"
	} else if resp.StatusCode == 403 {
		return false, "Insufficient permissions, please check API key permissions"
	} else if resp.StatusCode == 404 {
		return false, "Rerank API endpoint not found, please check Base URL"
	} else if resp.StatusCode == 422 {
		return false, fmt.Sprintf("Request parameter error: %s", string(body))
	} else {
		return false, fmt.Sprintf("API returned error status: %d, response: %s", resp.StatusCode, string(body))
	}
}

// CheckRerankModel checks the rerank model connectivity and functionality
func (h *InitializationHandler) CheckRerankModel(c *gin.Context) {
	ctx := c.Request.Context()

	logger.Info(ctx, "Checking rerank model connection and functionality")

	var req struct {
		ModelName string `json:"modelName" binding:"required"`
		BaseURL   string `json:"baseUrl" binding:"required"`
		APIKey    string `json:"apiKey"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error(ctx, "Failed to parse rerank model check request", err)
		c.Error(errors.NewBadRequestError(err.Error()))
		return
	}

	// Validate request parameters
	if req.ModelName == "" || req.BaseURL == "" {
		logger.Error(ctx, "Model name and base URL are required")
		c.Error(errors.NewBadRequestError("Model name and Base URL cannot be empty"))
		return
	}

	// Check rerank model connection and functionality
	available, message := h.checkRerankModelConnection(
		ctx, req.ModelName, req.BaseURL, req.APIKey,
	)

	logger.Info(ctx,
		fmt.Sprintf("Rerank model check completed: modelName=%s, baseUrl=%s, available=%v, message=%s",
			req.ModelName, req.BaseURL, available, message,
		),
	)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"available": available,
			"message":   message,
		},
	})
}

// testMultimodalForm is used to parse multipart form data
type testMultimodalForm struct {
	VLMModel         string `form:"vlm_model"`
	VLMBaseURL       string `form:"vlm_base_url"`
	VLMAPIKey        string `form:"vlm_api_key"`
	VLMInterfaceType string `form:"vlm_interface_type"`

	StorageType string `form:"storage_type"`

	// COS configuration
	COSSecretID   string `form:"cos_secret_id"`
	COSSecretKey  string `form:"cos_secret_key"`
	COSRegion     string `form:"cos_region"`
	COSBucketName string `form:"cos_bucket_name"`
	COSAppID      string `form:"cos_app_id"`
	COSPathPrefix string `form:"cos_path_prefix"`

	// MinIO configuration (when storage is minio)
	MinioBucketName string `form:"minio_bucket_name"`
	MinioPathPrefix string `form:"minio_path_prefix"`

	// Document splitting configuration (strings parsed later to avoid binding issues)
	ChunkSize     string `form:"chunk_size"`
	ChunkOverlap  string `form:"chunk_overlap"`
	SeparatorsRaw string `form:"separators"`
}

// TestMultimodalFunction tests the multimodal processing flow
func (h *InitializationHandler) TestMultimodalFunction(c *gin.Context) {
	ctx := c.Request.Context()

	logger.Info(ctx, "Testing multimodal functionality")

	var req testMultimodalForm
	if err := c.ShouldBind(&req); err != nil {
		logger.Error(ctx, "Failed to parse form data", err)
		c.Error(errors.NewBadRequestError("Failed to parse form parameters"))
		return
	}

	// For ollama interface, auto-append /v1 to base URL from env
	if req.VLMInterfaceType == "ollama" {
		req.VLMBaseURL = os.Getenv("OLLAMA_BASE_URL") + "/v1"
	}

	req.StorageType = strings.ToLower(req.StorageType)

	if req.VLMModel == "" || req.VLMBaseURL == "" {
		logger.Error(ctx, "VLM model name and base URL are required")
		c.Error(errors.NewBadRequestError("VLM model name and Base URL cannot be empty"))
		return
	}

	switch req.StorageType {
	case "cos":
		logger.Infof(ctx, "COS config: Region=%s, Bucket=%s, App=%s, Prefix=%s",
			req.COSRegion, req.COSBucketName, req.COSAppID, req.COSPathPrefix)
		// Required: SecretID/SecretKey/Region/BucketName/AppID; PathPrefix optional
		if req.COSSecretID == "" || req.COSSecretKey == "" ||
			req.COSRegion == "" || req.COSBucketName == "" ||
			req.COSAppID == "" {
			logger.Error(ctx, "COS configuration is required")
			c.Error(errors.NewBadRequestError("COS configuration cannot be empty"))
			return
		}
	case "minio":
		logger.Infof(ctx, "MinIO config: Bucket=%s, PathPrefix=%s", req.MinioBucketName, req.MinioPathPrefix)
		if req.MinioBucketName == "" {
			logger.Error(ctx, "MinIO configuration is required")
			c.Error(errors.NewBadRequestError("MinIO configuration cannot be empty"))
			return
		}
	default:
		logger.Error(ctx, "Invalid storage type")
		c.Error(errors.NewBadRequestError("Invalid storage type"))
		return
	}

	logger.Infof(ctx, "VLM config: Model=%s, URL=%s, HasKey=%v, Type=%s",
		req.VLMModel, req.VLMBaseURL, req.VLMAPIKey != "", req.VLMInterfaceType)

	// Get uploaded image file
	file, header, err := c.Request.FormFile("image")
	if err != nil {
		logger.Error(ctx, "Failed to get uploaded image", err)
		c.Error(errors.NewBadRequestError("Failed to get uploaded image"))
		return
	}
	defer file.Close()

	// Validate file type
	if !strings.HasPrefix(header.Header.Get("Content-Type"), "image/") {
		logger.Error(ctx, "Invalid file type, only images are allowed")
		c.Error(errors.NewBadRequestError("Only image files are allowed"))
		return
	}

	// Validate file size (10MB)
	if header.Size > 10*1024*1024 {
		logger.Error(ctx, "File size too large")
		c.Error(errors.NewBadRequestError("Image file size must not exceed 10MB"))
		return
	}

	logger.Infof(ctx, "Processing image: %s, size: %d bytes", header.Filename, header.Size)

	// Parse document splitting configuration
	chunkSize, err := strconv.Atoi(req.ChunkSize)
	if err != nil || chunkSize < 100 || chunkSize > 10000 {
		chunkSize = 1000
	}

	chunkOverlap, err := strconv.Atoi(req.ChunkOverlap)
	if err != nil || chunkOverlap < 0 || chunkOverlap >= chunkSize {
		chunkOverlap = 200
	}

	var separators []string
	if req.SeparatorsRaw != "" {
		if err := json.Unmarshal([]byte(req.SeparatorsRaw), &separators); err != nil {
			separators = []string{"\n\n", "\n", ".", "!", "?", ";", ";"}
		}
	} else {
		separators = []string{"\n\n", "\n", ".", "!", "?", ";", ";"}
	}

	// Read image file content
	imageContent, err := io.ReadAll(file)
	if err != nil {
		logger.Error(ctx, "Failed to read image file", err)
		c.Error(errors.NewBadRequestError("Failed to read image file"))
		return
	}

	// Call multimodal test
	startTime := time.Now()
	result, err := h.testMultimodalWithDocReader(
		ctx,
		imageContent, header.Filename,
		chunkSize, chunkOverlap, separators, &req,
	)
	processingTime := time.Since(startTime).Milliseconds()

	if err != nil {
		logger.ErrorWithFields(ctx, err, map[string]interface{}{
			"vlm_model":    req.VLMModel,
			"vlm_base_url": req.VLMBaseURL,
			"filename":     header.Filename,
		})
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": gin.H{
				"success":         false,
				"message":         err.Error(),
				"processing_time": processingTime,
			},
		})
		return
	}

	logger.Info(ctx, fmt.Sprintf("Multimodal test completed successfully in %dms", processingTime))

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"success":         true,
			"caption":         result["caption"],
			"ocr":             result["ocr"],
			"processing_time": processingTime,
		},
	})
}

// testMultimodalWithDocReader calls the DocReader service for multimodal processing
func (h *InitializationHandler) testMultimodalWithDocReader(
	ctx context.Context,
	imageContent []byte, filename string,
	chunkSize, chunkOverlap int, separators []string,
	req *testMultimodalForm,
) (map[string]string, error) {
	// Get file extension
	fileExt := ""
	if idx := strings.LastIndex(filename, "."); idx != -1 {
		fileExt = strings.ToLower(filename[idx+1:])
	}

	// Ensure docreader client is configured
	if h.docReaderClient == nil {
		return nil, fmt.Errorf("DocReader service not configured")
	}

	// Build request
	request := &proto.ReadFromFileRequest{
		FileContent: imageContent,
		FileName:    filename,
		FileType:    fileExt,
		ReadConfig: &proto.ReadConfig{
			ChunkSize:        int32(chunkSize),
			ChunkOverlap:     int32(chunkOverlap),
			Separators:       separators,
			EnableMultimodal: true, // enable multimodal processing
			VlmConfig: &proto.VLMConfig{
				ModelName:     req.VLMModel,
				BaseUrl:       req.VLMBaseURL,
				ApiKey:        req.VLMAPIKey,
				InterfaceType: req.VLMInterfaceType,
			},
		},
		RequestId: ctx.Value(types.RequestIDContextKey).(string),
	}

	// Set object storage configuration
	switch strings.ToLower(req.StorageType) {
	case "cos":
		request.ReadConfig.StorageConfig = &proto.StorageConfig{
			Provider:        proto.StorageProvider_COS,
			Region:          req.COSRegion,
			BucketName:      req.COSBucketName,
			AccessKeyId:     req.COSSecretID,
			SecretAccessKey: req.COSSecretKey,
			AppId:           req.COSAppID,
			PathPrefix:      req.COSPathPrefix,
		}
	case "minio":
		request.ReadConfig.StorageConfig = &proto.StorageConfig{
			Provider:        proto.StorageProvider_MINIO,
			BucketName:      req.MinioBucketName,
			PathPrefix:      req.MinioPathPrefix,
			AccessKeyId:     os.Getenv("MINIO_ACCESS_KEY_ID"),
			SecretAccessKey: os.Getenv("MINIO_SECRET_ACCESS_KEY"),
		}
	}

	// Call DocReader service
	response, err := h.docReaderClient.ReadFromFile(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("DocReader service call failed: %v", err)
	}

	if response.Error != "" {
		return nil, fmt.Errorf("DocReader service returned error: %s", response.Error)
	}

	// Process response to extract caption and OCR information
	result := make(map[string]string)
	var allCaptions, allOCRTexts []string

	for _, chunk := range response.Chunks {
		if len(chunk.Images) > 0 {
			for _, image := range chunk.Images {
				if image.Caption != "" {
					allCaptions = append(allCaptions, image.Caption)
				}
				if image.OcrText != "" {
					allOCRTexts = append(allOCRTexts, image.OcrText)
				}
			}
		}
	}

	// Merge all captions and OCR results
	result["caption"] = strings.Join(allCaptions, "; ")
	result["ocr"] = strings.Join(allOCRTexts, "; ")

	return result, nil
}
