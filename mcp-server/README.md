# WeKnowRust MCP Server

This is a Model Context Protocol (MCP) server that provides access to the WeKnowRust knowledge management API.

## Quick Start

### 1) Install dependencies
```bash
pip install -r requirements.txt
```

### 2) Configure environment variables
```bash
# Linux/macOS
export WEKNOWRUST_BASE_URL="http://localhost:8080/api/v1"
export WEKNOWRUST_API_KEY="your_api_key_here"

# Windows PowerShell
$env:WEKNOWRUST_BASE_URL="http://localhost:8080/api/v1"
$env:WEKNOWRUST_API_KEY="your_api_key_here"

# Windows CMD
set WEKNOWRUST_BASE_URL=http://localhost:8080/api/v1
set WEKNOWRUST_API_KEY=your_api_key_here
```

### 3) Run the server

**Recommended – using the main entry point:**
```bash
python main.py
```

**Other ways to run:**
```bash
# Use the original startup script
python run_server.py

# Use helper script
python run.py

# Run the server module directly
python weknowrust_mcp_server.py

# Run as a Python module
python -m weknowrust_mcp_server
```

### 4) CLI options
```bash
python main.py --help                 # Show help
python main.py --check-only           # Check environment only
python main.py --verbose              # Enable verbose logs
python main.py --version              # Show version
```

## Install as a Python package

### Development install
```bash
pip install -e .
```

After installation you can use the CLI tools:
```bash
weknowrust-mcp-server
# Or
weknowrust-server
```

### Production install
```bash
pip install .
```

### Build distributions
```bash
# Using setuptools
python setup.py sdist bdist_wheel

# Using modern build tools
pip install build
python -m build
```

## Test module

Run the test script to verify the module works:
```bash
python test_module.py
```

## Features

This MCP server provides the following tools:

### Tenant management
- `create_tenant` - Create a new tenant
- `list_tenants` - List all tenants

### Knowledge base management
- `create_knowledge_base` - Create a knowledge base
- `list_knowledge_bases` - List knowledge bases
- `get_knowledge_base` - Get knowledge base details
- `delete_knowledge_base` - Delete a knowledge base
- `hybrid_search` - Hybrid search

### Knowledge management
- `create_knowledge_from_url` - Create knowledge from URL
- `list_knowledge` - List knowledge
- `get_knowledge` - Get knowledge details
- `delete_knowledge` - Delete knowledge

### Model management
- `create_model` - Create model
- `list_models` - List models
- `get_model` - Get model details

### Session management
- `create_session` - Create a chat session
- `get_session` - Get session details
- `list_sessions` - List sessions
- `delete_session` - Delete session

### Chat
- `chat` - Send chat messages

### Chunk management
- `list_chunks` - List knowledge chunks
- `delete_chunk` - Delete a knowledge chunk

## Troubleshooting

If you encounter import errors, ensure that:
1. All required dependencies are installed
2. You are using a compatible Python version (3.8+ recommended)
3. There are no filename conflicts (avoid naming files `mcp.py`)