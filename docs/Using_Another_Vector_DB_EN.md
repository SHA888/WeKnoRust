### How to Integrate a New Vector Database

This guide provides a complete walkthrough for adding support for a new vector database to the WeKnoRust project. By implementing standardized interfaces and following a structured process, you can efficiently integrate a custom vector database.

### Integration Workflow

#### 1. Implement the base retrieve engine interface
Implement the `RetrieveEngine` interface from the `interfaces` package to define the core capabilities of the retrieval engine:

```go
type RetrieveEngine interface {
    // Return the engine type identifier
    EngineType() types.RetrieverEngineType

    // Execute the retrieval operation and return matches
    Retrieve(ctx context.Context, params types.RetrieveParams) ([]*types.RetrieveResult, error)

    // Return the list of retrieval types supported by this engine
    Support() []types.RetrieverType
}
```

#### 2. Implement the repository (storage) interface
Implement the `RetrieveEngineRepository` interface to extend the base engine with index management capabilities:

```go
type RetrieveEngineRepository interface {
    // Save a single index entry
    Save(ctx context.Context, indexInfo *types.IndexInfo, params map[string]any) error
    
    // Save multiple index entries in batch
    BatchSave(ctx context.Context, indexInfoList []*types.IndexInfo, params map[string]any) error
    
    // Estimate storage size required by the indices
    EstimateStorageSize(ctx context.Context, indexInfoList []*types.IndexInfo, params map[string]any) int64
    
    // Delete indices by a list of chunk IDs
    DeleteByChunkIDList(ctx context.Context, indexIDList []string, dimension int) error
    
    // Copy indices to avoid recomputing embeddings
    CopyIndices(
        ctx context.Context,
        sourceKnowledgeBaseID string,
        sourceToTargetKBIDMap map[string]string,
        sourceToTargetChunkIDMap map[string]string,
        targetKnowledgeBaseID string,
        dimension int,
    ) error
    
    // Delete indices by a list of knowledge IDs
    DeleteByKnowledgeIDList(ctx context.Context, knowledgeIDList []string, dimension int) error
    
    // Inherit from RetrieveEngine
    RetrieveEngine
}
```

#### 3. Implement the service-layer interface
Create a service that implements `RetrieveEngineService` and handles business logic for index creation and management:

```go
type RetrieveEngineService interface {
    // Create a single index
    Index(ctx context.Context,
        embedder embedding.Embedder,
        indexInfo *types.IndexInfo,
        retrieverTypes []types.RetrieverType,
    ) error

    // Create indices in batch
    BatchIndex(ctx context.Context,
        embedder embedding.Embedder,
        indexInfoList []*types.IndexInfo,
        retrieverTypes []types.RetrieverType,
    ) error

    // Estimate storage size for the indices
    EstimateStorageSize(ctx context.Context,
        embedder embedding.Embedder,
        indexInfoList []*types.IndexInfo,
        retrieverTypes []types.RetrieverType,
    ) int64
    
    // Copy indices
    CopyIndices(
        ctx context.Context,
        sourceKnowledgeBaseID string,
        sourceToTargetKBIDMap map[string]string,
        sourceToTargetChunkIDMap map[string]string,
        targetKnowledgeBaseID string,
        dimension int,
    ) error

    // Delete indices
    DeleteByChunkIDList(ctx context.Context, indexIDList []string, dimension int) error
    DeleteByKnowledgeIDList(ctx context.Context, knowledgeIDList []string, dimension int) error

    // Inherit from RetrieveEngine
    RetrieveEngine
}
```

#### 4. Add environment variables
Add the necessary connection parameters for the new database to your environment configuration:

```
# Add the new driver name to RETRIEVE_DRIVER (comma-separated for multiple drivers)
RETRIEVE_DRIVER=postgres,elasticsearch_v8,your_database

# Connection parameters for the new database
YOUR_DATABASE_ADDR=your_database_host:port
YOUR_DATABASE_USERNAME=username
YOUR_DATABASE_PASSWORD=password
# Other required parameters...
```

#### 5. Register the retrieve engine
Add initialization and registration logic for the new database in the `initRetrieveEngineRegistry` function in `internal/container/container.go`:

```go
func initRetrieveEngineRegistry(db *gorm.DB, cfg *config.Config) (interfaces.RetrieveEngineRegistry, error) {
    registry := retriever.NewRetrieveEngineRegistry()
    retrieveDriver := strings.Split(os.Getenv("RETRIEVE_DRIVER"), ",")
    log := logger.GetLogger(context.Background())

    // Existing PostgreSQL and Elasticsearch initialization code...
    
    // Add initialization for the new vector database
    if slices.Contains(retrieveDriver, "your_database") {
        // Initialize the database client
        client, err := your_database.NewClient(your_database.Config{
            Addresses: []string{os.Getenv("YOUR_DATABASE_ADDR")},
            Username:  os.Getenv("YOUR_DATABASE_USERNAME"),
            Password:  os.Getenv("YOUR_DATABASE_PASSWORD"),
            // Other connection params...
        })
        
        if err != nil {
            log.Errorf("Create your_database client failed: %v", err)
        } else {
            // Create repository
            yourDatabaseRepo := your_database.NewYourDatabaseRepository(client, cfg)
            
            // Register retrieve engine
            if err := registry.Register(
                retriever.NewKVHybridRetrieveEngine(
                    yourDatabaseRepo, types.YourDatabaseRetrieverEngineType,
                ),
            ); err != nil {
                log.Errorf("Register your_database retrieve engine failed: %v", err)
            } else {
                log.Infof("Register your_database retrieve engine success")
            }
        }
    }

    return registry, nil
}
```

#### 6. Define a new retriever engine type constant
Add a new retriever engine type constant in `internal/types/retriever.go`:

```go
// RetrieverEngineType defines the retriever engine type
const (
    ElasticsearchRetrieverEngineType RetrieverEngineType = "elasticsearch"
    PostgresRetrieverEngineType      RetrieverEngineType = "postgres"
    YourDatabaseRetrieverEngineType  RetrieverEngineType = "your_database" // New database type
)
```

## Reference Implementations
We recommend using existing PostgreSQL and Elasticsearch implementations as templates. See:

- PostgreSQL: `internal/application/repository/retriever/postgres/`
- Elasticsearch V7: `internal/application/repository/retriever/elasticsearch/v7/`
- Elasticsearch V8: `internal/application/repository/retriever/elasticsearch/v8/`

By following the steps above and referencing the existing implementations, you can successfully integrate a new vector database into the WeKnoRust system and extend its vector retrieval capabilities.
