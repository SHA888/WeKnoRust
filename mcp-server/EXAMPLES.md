# WeKnowRust MCP Server Examples

This document provides detailed usage examples for the WeKnowRust MCP Server.

## Basic Usage

### 1) Start the server

```bash
# Recommended – main entry point
python main.py

# Check environment configuration
python main.py --check-only

# Enable verbose logging
python main.py --verbose
```

### 2) Environment configuration example

```bash
# Set environment variables
export WEKNOWRUST_BASE_URL="http://localhost:8080/api/v1"
export WEKNOWRUST_API_KEY="your_api_key_here"

# Or set them in a .env file
echo "WEKNOWRUST_BASE_URL=http://localhost:8080/api/v1" > .env
echo "WEKNOWRUST_API_KEY=your_api_key_here" >> .env
```

## MCP Tool Examples

Below are examples for various MCP tools:

### Tenant management

#### Create tenant
```json
{
  "tool": "create_tenant",
  "arguments": {
    "name": "我的公司",
    "description": "公司知识管理系统",
    "business": "technology",
    "retriever_engines": {
      "engines": [
        {"retriever_type": "keywords", "retriever_engine_type": "postgres"},
        {"retriever_type": "vector", "retriever_engine_type": "postgres"}
      ]
    }
  }
}
```

#### List tenants
```json
{
  "tool": "list_tenants",
  "arguments": {}
}
```

### Knowledge base management

#### Create knowledge base
```json
{
  "tool": "create_knowledge_base",
  "arguments": {
    "name": "产品文档库",
    "description": "产品相关文档和资料",
    "embedding_model_id": "text-embedding-ada-002",
    "summary_model_id": "gpt-3.5-turbo"
  }
}
```

#### List knowledge bases
```json
{
  "tool": "list_knowledge_bases",
  "arguments": {}
}
```

#### Get knowledge base details
```json
{
  "tool": "get_knowledge_base",
  "arguments": {
    "kb_id": "kb_123456"
  }
}
```

#### Hybrid search
```json
{
  "tool": "hybrid_search",
  "arguments": {
    "kb_id": "kb_123456",
    "query": "如何使用API",
    "vector_threshold": 0.7,
    "keyword_threshold": 0.5,
    "match_count": 10
  }
}
```

### Knowledge management

#### Create knowledge from URL
```json
{
  "tool": "create_knowledge_from_url",
  "arguments": {
    "kb_id": "kb_123456",
    "url": "https://docs.example.com/api-guide",
    "enable_multimodel": true
  }
}
```

#### List knowledge
```json
{
  "tool": "list_knowledge",
  "arguments": {
    "kb_id": "kb_123456",
    "page": 1,
    "page_size": 20
  }
}
```

#### Get knowledge details
```json
{
  "tool": "get_knowledge",
  "arguments": {
    "knowledge_id": "know_789012"
  }
}
```

### Model management

#### Create model
```json
{
  "tool": "create_model",
  "arguments": {
    "name": "GPT-4 Chat Model",
    "type": "KnowledgeQA",
    "source": "openai",
    "description": "OpenAI GPT-4 模型用于知识问答",
    "base_url": "https://api.openai.com/v1",
    "api_key": "sk-...",
    "is_default": true
  }
}
```

#### List models
```json
{
  "tool": "list_models",
  "arguments": {}
}
```

### Session management

#### Create chat session
```json
{
  "tool": "create_session",
  "arguments": {
    "kb_id": "kb_123456",
    "max_rounds": 10,
    "enable_rewrite": true,
    "fallback_response": "Sorry, I cannot answer this question.",
    "summary_model_id": "gpt-3.5-turbo"
  }
}
```

#### Get session details
```json
{
  "tool": "get_session",
  "arguments": {
    "session_id": "sess_345678"
  }
}
```

#### List sessions
```json
{
  "tool": "list_sessions",
  "arguments": {
    "page": 1,
    "page_size": 10
  }
}
```

### Chat

#### Send chat message
```json
{
  "tool": "chat",
  "arguments": {
    "session_id": "sess_345678",
    "query": "请介绍一下产品的主要功能"
  }
}
```

### Chunk management

#### List knowledge chunks
```json
{
  "tool": "list_chunks",
  "arguments": {
    "knowledge_id": "know_789012",
    "page": 1,
    "page_size": 50
  }
}
```

#### Delete knowledge chunk
```json
{
  "tool": "delete_chunk",
  "arguments": {
    "knowledge_id": "know_789012",
    "chunk_id": "chunk_456789"
  }
}
```

## Full Workflow Example

### Scenario: Build a complete knowledge Q&A system

```bash
# 1. 启动服务器
python main.py --verbose

# 2. In your MCP client, perform the following steps:
```

#### Step 1: Create a tenant
```json
{
  "tool": "create_tenant",
  "arguments": {
    "name": "Tech Docs Center",
    "description": "Company technical documentation knowledge management",
    "business": "technology"
  }
}
```

#### Step 2: Create a knowledge base
```json
{
  "tool": "create_knowledge_base",
  "arguments": {
    "name": "API Docs",
    "description": "All API-related documentation"
  }
}
```

#### Step 3: Add knowledge content
```json
{
  "tool": "create_knowledge_from_url",
  "arguments": {
    "kb_id": "returned_knowledge_base_id",
    "url": "https://docs.company.com/api",
    "enable_multimodel": true
  }
}
```

#### Step 4: Create a chat session
```json
{
  "tool": "create_session",
  "arguments": {
    "kb_id": "knowledge_base_id",
    "max_rounds": 5,
    "enable_rewrite": true
  }
}
```

#### Step 5: Start a conversation
```json
{
  "tool": "chat",
  "arguments": {
    "session_id": "session_id",
    "query": "How to use the user authentication API?"
  }
}
```

## Error Handling Examples

### Common errors and solutions

#### 1) Connection error
```json
{
  "error": "Connection refused",
  "solution": "Verify WEKNOWRUST_BASE_URL and ensure the service is running"
}
```

#### 2) Authentication error
```json
{
  "error": "Unauthorized",
  "solution": "Check WEKNOWRUST_API_KEY is set correctly"
}
```

#### 3) Resource not found
```json
{
  "error": "Knowledge base not found",
  "solution": "Confirm the knowledge base ID is correct, or create the knowledge base first"
}
```

## Advanced Configuration Examples

### Custom retriever configuration
```json
{
  "tool": "hybrid_search",
  "arguments": {
    "kb_id": "kb_123456",
    "query": "search query",
    "vector_threshold": 0.8,
    "keyword_threshold": 0.6,
    "match_count": 15
  }
}
```

### Custom session policy
```json
{
  "tool": "create_session",
  "arguments": {
    "kb_id": "kb_123456",
    "max_rounds": 20,
    "enable_rewrite": true,
    "fallback_response": "Based on the current knowledge, I can’t accurately answer your question. Please rephrase or contact support."
  }
}
```

## Performance Optimization Tips

1. **Batch operations**: Batch knowledge creation and updates where possible
2. **Tuning thresholds**: Adjust search thresholds to balance accuracy and performance
3. **Session management**: Clean up unneeded sessions to save resources
4. **Monitor logs**: Use `--verbose` to monitor performance indicators

## Integration Examples

### Integration with Claude Desktop
Add the following to your Claude Desktop configuration:
```json
{
  "mcpServers": {
    "weknowrust": {
      "command": "python",
      "args": ["path/to/main.py"],
      "env": {
        "WEKNOWRUST_BASE_URL": "http://localhost:8080/api/v1",
        "WEKNOWRUST_API_KEY": "your_api_key"
      }
    }
  }
}
```

Project repository: https://github.com/SHA888/WeKnoRustMCP

### Integration with other MCP clients
Refer to each client’s documentation to configure the server command and environment variables.

## Troubleshooting

If you encounter issues:
1. Run `python main.py --check-only` to check the environment
2. Use `python main.py --verbose` for detailed logs
3. Ensure the WeKnowRust service is running
4. Verify network connectivity and firewall rules