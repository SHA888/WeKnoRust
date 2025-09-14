package router

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.uber.org/dig"

	"github.com/Tencent/WeKnowRust/internal/config"
	"github.com/Tencent/WeKnowRust/internal/handler"
	"github.com/Tencent/WeKnowRust/internal/middleware"
	"github.com/Tencent/WeKnowRust/internal/types/interfaces"
)

// RouterParams defines DI parameters for building routes
type RouterParams struct {
	dig.In

	Config                *config.Config
	KBHandler             *handler.KnowledgeBaseHandler
	KnowledgeHandler      *handler.KnowledgeHandler
	TenantHandler         *handler.TenantHandler
	TenantService         interfaces.TenantService
	ChunkHandler          *handler.ChunkHandler
	SessionHandler        *handler.SessionHandler
	MessageHandler        *handler.MessageHandler
	TestDataHandler       *handler.TestDataHandler
	ModelHandler          *handler.ModelHandler
	EvaluationHandler     *handler.EvaluationHandler
	InitializationHandler *handler.InitializationHandler
}

// NewRouter creates and configures a new Gin engine with routes and middleware
func NewRouter(params RouterParams) *gin.Engine {
	r := gin.New()

	// CORS middleware should be added first
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-API-Key", "X-Request-ID"},
		ExposeHeaders:    []string{"Content-Length", "Access-Control-Allow-Origin"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Other middleware
	r.Use(middleware.RequestID())
	r.Use(middleware.Logger())
	r.Use(middleware.Recovery())
	r.Use(middleware.ErrorHandler())
	r.Use(middleware.Auth(params.TenantService, params.Config))

	// Add OpenTelemetry tracing middleware
	r.Use(middleware.TracingMiddleware())

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Test data APIs (no authentication)
	r.GET("/api/v1/test-data", params.TestDataHandler.GetTestData)

	// Initialization APIs (no authentication)
	r.GET("/api/v1/initialization/status", params.InitializationHandler.CheckStatus)
	r.GET("/api/v1/initialization/config", params.InitializationHandler.GetCurrentConfig)
	r.POST("/api/v1/initialization/initialize", params.InitializationHandler.Initialize)

	// Ollama-related APIs (no authentication)
	r.GET("/api/v1/initialization/ollama/status", params.InitializationHandler.CheckOllamaStatus)
	r.GET("/api/v1/initialization/ollama/models", params.InitializationHandler.ListOllamaModels)
	r.POST("/api/v1/initialization/ollama/models/check", params.InitializationHandler.CheckOllamaModels)
	r.POST("/api/v1/initialization/ollama/models/download", params.InitializationHandler.DownloadOllamaModel)
	r.GET("/api/v1/initialization/ollama/download/progress/:taskId", params.InitializationHandler.GetDownloadProgress)
	r.GET("/api/v1/initialization/ollama/download/tasks", params.InitializationHandler.ListDownloadTasks)

	// Remote API related endpoints (no authentication)
	r.POST("/api/v1/initialization/remote/check", params.InitializationHandler.CheckRemoteModel)
	r.POST("/api/v1/initialization/embedding/test", params.InitializationHandler.TestEmbeddingModel)
	r.POST("/api/v1/initialization/rerank/check", params.InitializationHandler.CheckRerankModel)
	r.POST("/api/v1/initialization/multimodal/test", params.InitializationHandler.TestMultimodalFunction)

	// Authenticated API routes
	v1 := r.Group("/api/v1")
	{
		RegisterTenantRoutes(v1, params.TenantHandler)
		RegisterKnowledgeBaseRoutes(v1, params.KBHandler)
		RegisterKnowledgeRoutes(v1, params.KnowledgeHandler)
		RegisterChunkRoutes(v1, params.ChunkHandler)
		RegisterSessionRoutes(v1, params.SessionHandler)
		RegisterChatRoutes(v1, params.SessionHandler)
		RegisterMessageRoutes(v1, params.MessageHandler)
		RegisterModelRoutes(v1, params.ModelHandler)
		RegisterEvaluationRoutes(v1, params.EvaluationHandler)
	}

	return r
}

// RegisterChunkRoutes registers chunk-related routes
func RegisterChunkRoutes(r *gin.RouterGroup, handler *handler.ChunkHandler) {
	// Chunks route group
	chunks := r.Group("/chunks")
	{
		// List chunks
		chunks.GET("/:knowledge_id", handler.ListKnowledgeChunks)
		// Delete a chunk
		chunks.DELETE("/:knowledge_id/:id", handler.DeleteChunk)
		// Delete all chunks under a knowledge item
		chunks.DELETE("/:knowledge_id", handler.DeleteChunksByKnowledgeID)
		// Update chunk info
		chunks.PUT("/:knowledge_id/:id", handler.UpdateChunk)
	}
}

// RegisterKnowledgeRoutes registers knowledge-related routes
func RegisterKnowledgeRoutes(r *gin.RouterGroup, handler *handler.KnowledgeHandler) {
	// Knowledge under knowledge base route group
	kb := r.Group("/knowledge-bases/:id/knowledge")
	{
		// Create knowledge from file
		kb.POST("/file", handler.CreateKnowledgeFromFile)
		// Create knowledge from URL
		kb.POST("/url", handler.CreateKnowledgeFromURL)
		// List knowledge under the knowledge base
		kb.GET("", handler.ListKnowledge)
	}

	// Knowledge route group
	k := r.Group("/knowledge")
	{
		// Batch get knowledge
		k.GET("/batch", handler.GetKnowledgeBatch)
		// Get knowledge detail
		k.GET("/:id", handler.GetKnowledge)
		// Delete knowledge
		k.DELETE("/:id", handler.DeleteKnowledge)
		// Update knowledge
		k.PUT("/:id", handler.UpdateKnowledge)
		// Download knowledge file
		k.GET("/:id/download", handler.DownloadKnowledgeFile)
		// Update image chunk info
		k.PUT("/image/:id/:chunk_id", handler.UpdateImageInfo)
	}
}

// RegisterKnowledgeBaseRoutes registers knowledge base routes
func RegisterKnowledgeBaseRoutes(r *gin.RouterGroup, handler *handler.KnowledgeBaseHandler) {
	// Knowledge base route group
	kb := r.Group("/knowledge-bases")
	{
		// Create knowledge base
		kb.POST("", handler.CreateKnowledgeBase)
		// List knowledge bases
		kb.GET("", handler.ListKnowledgeBases)
		// Get knowledge base detail
		kb.GET("/:id", handler.GetKnowledgeBase)
		// Update knowledge base
		kb.PUT("/:id", handler.UpdateKnowledgeBase)
		// Delete knowledge base
		kb.DELETE("/:id", handler.DeleteKnowledgeBase)
		// Hybrid search
		kb.GET("/:id/hybrid-search", handler.HybridSearch)
		// Copy knowledge base
		kb.POST("/copy", handler.CopyKnowledgeBase)
	}
}

// RegisterMessageRoutes registers message-related routes
func RegisterMessageRoutes(r *gin.RouterGroup, handler *handler.MessageHandler) {
	// Messages route group
	messages := r.Group("/messages")
	{
		// Load older messages for upward infinite scroll
		messages.GET("/:session_id/load", handler.LoadMessages)
		// Delete a message
		messages.DELETE("/:session_id/:id", handler.DeleteMessage)
	}
}

// RegisterSessionRoutes registers session routes
func RegisterSessionRoutes(r *gin.RouterGroup, handler *handler.SessionHandler) {
	sessions := r.Group("/sessions")
	{
		sessions.POST("", handler.CreateSession)
		sessions.GET("/:id", handler.GetSession)
		sessions.GET("", handler.GetSessionsByTenant)
		sessions.PUT("/:id", handler.UpdateSession)
		sessions.DELETE("/:id", handler.DeleteSession)
		sessions.POST("/:session_id/generate_title", handler.GenerateTitle)
		// Continue receiving active stream
		sessions.GET("/continue-stream/:session_id", handler.ContinueStream)
	}
}

// RegisterChatRoutes registers chat-related routes
func RegisterChatRoutes(r *gin.RouterGroup, handler *handler.SessionHandler) {
	knowledgeChat := r.Group("/knowledge-chat")
	{
		knowledgeChat.POST("/:session_id", handler.KnowledgeQA)
	}

	// Knowledge search endpoint without session_id
	knowledgeSearch := r.Group("/knowledge-search")
	{
		knowledgeSearch.POST("", handler.SearchKnowledge)
	}
}

// RegisterTenantRoutes registers tenant-related routes
func RegisterTenantRoutes(r *gin.RouterGroup, handler *handler.TenantHandler) {
	// Tenant route group
	tenantRoutes := r.Group("/tenants")
	{
		tenantRoutes.POST("", handler.CreateTenant)
		tenantRoutes.GET("/:id", handler.GetTenant)
		tenantRoutes.PUT("/:id", handler.UpdateTenant)
		tenantRoutes.DELETE("/:id", handler.DeleteTenant)
		tenantRoutes.GET("", handler.ListTenants)
	}
}

// RegisterModelRoutes registers model-related routes
func RegisterModelRoutes(r *gin.RouterGroup, handler *handler.ModelHandler) {
	// Models route group
	models := r.Group("/models")
	{
		// Create model
		models.POST("", handler.CreateModel)
		// List models
		models.GET("", handler.ListModels)
		// Get model by ID
		models.GET("/:id", handler.GetModel)
		// Update model
		models.PUT("/:id", handler.UpdateModel)
		// Delete model
		models.DELETE("/:id", handler.DeleteModel)
	}
}

func RegisterEvaluationRoutes(r *gin.RouterGroup, handler *handler.EvaluationHandler) {
	evaluationRoutes := r.Group("/evaluation")
	{
		evaluationRoutes.POST("/", handler.Evaluation)
		evaluationRoutes.GET("/", handler.GetEvaluationResult)
	}
}
