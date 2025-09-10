# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/).

## [1.0.0] - 2024-01-XX

### Added
- Initial release
- WeKnowRust MCP Server core features
- Full WeKnowRust API integration
- Tenant management tools
- Knowledge base management tools
- Knowledge management tools
- Model management tools
- Session management tools
- Chat tool
- Chunk management tools
- Multiple startup methods supported
- CLI arguments support
- Environment variable configuration
- Full package installation support
- Development and production modes
- Detailed documentation and installation guide

### Tool List
- `create_tenant` - Create a new tenant
- `list_tenants` - List all tenants
- `create_knowledge_base` - Create a knowledge base
- `list_knowledge_bases` - List knowledge bases
- `get_knowledge_base` - Get knowledge base details
- `delete_knowledge_base` - Delete a knowledge base
- `hybrid_search` - Hybrid search
- `create_knowledge_from_url` - Create knowledge from URL
- `list_knowledge` - List knowledge
- `get_knowledge` - Get knowledge details
- `delete_knowledge` - Delete knowledge
- `create_model` - Create model
- `list_models` - List models
- `get_model` - Get model details
- `create_session` - Create chat session
- `get_session` - Get session details
- `list_sessions` - List sessions
- `delete_session` - Delete session
- `chat` - Send chat message
- `list_chunks` - List knowledge chunks
- `delete_chunk` - Delete knowledge chunk

### File Structure
```
WeKnowRustMCP/
├── __init__.py              # Package init
├── main.py                  # Main entry point (recommended)
├── run.py                   # Helper startup script
├── run_server.py            # Original startup script
├── weknowrust_mcp_server.py # MCP server implementation
├── test_module.py           # Module test
├── requirements.txt         # Dependencies
├── setup.py                 # Setup script (legacy)
├── pyproject.toml           # Modern project config
├── MANIFEST.in              # Package manifest
├── LICENSE                  # MIT license
├── README.md                # Project README
├── INSTALL.md               # Installation guide
└── CHANGELOG.md             # Changelog
```

### Startup Methods
1. `python main.py` - Main entry point (recommended)
2. `python run_server.py` - Original startup script
3. `python run.py` - Helper startup script
4. `python weknowrust_mcp_server.py` - Run directly
5. `python -m weknowrust_mcp_server` - Module execution
6. `weknowrust-mcp-server` - CLI after install
7. `weknowrust-server` - CLI alias after install

### Technical Features
- Based on Model Context Protocol (MCP) 1.0.0+
- Async I/O support
- Comprehensive error handling
- Detailed logging
- Environment variable configuration
- CLI arguments support
- Multiple installation methods
- Dev and production modes
- Comprehensive test coverage

### Dependencies
- Python 3.8+
- mcp >= 1.0.0
- requests >= 2.31.0

### Compatibility
- Windows, macOS, Linux
- Python 3.8–3.12
- Compatible with modern Python packaging tools