# WeKnowRust MCP Server Installation and Usage Guide

## Quick Start

### 1) Install dependencies
```bash
pip install -r requirements.txt
```

### 2) Set environment variables
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

There are multiple ways to run the server:

#### Option 1: Use the main entry point (recommended)
```bash
python main.py
```

#### Option 2: Use the original startup script
```bash
python run_server.py
```

#### Option 3: Run the server module directly
```bash
python weknowrust_mcp_server.py
```

#### Option 4: Run as a Python module
```bash
python -m weknowrust_mcp_server
```

## Install as a Python package

### Development install
```bash
pip install -e .
```

After installation, you can use the CLI tools:
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
# Build source distribution and wheel
python setup.py sdist bdist_wheel

# Or use the build tool
pip install build
python -m build
```

## Command-line options

The main entry point `main.py` supports the following options:

```bash
python main.py --help                 # Show help
python main.py --check-only           # Check environment only
python main.py --verbose              # Enable verbose logs
python main.py --version              # Show version
```

## Environment check

Run the following command to check environment configuration:
```bash
python main.py --check-only
```

This will display:
- WeKnowRust API base URL configuration
- API key setup status
- Dependency installation status

## Troubleshooting

### 1) Import errors
If you encounter `ImportError`, ensure:
- All dependencies are installed: `pip install -r requirements.txt`
- Python version is compatible (3.8+ recommended)
- No filename conflicts

### 2) Connection errors
If you cannot connect to the WeKnowRust API:
- Check `WEKNOWRUST_BASE_URL` is correct
- Ensure the WeKnowRust service is running
- Verify network connectivity

### 3) Authentication errors
If you encounter authentication issues:
- Check `WEKNOWRUST_API_KEY` is set
- Confirm the API key is valid
- Verify permissions

## Development

### Project structure
```
WeKnowRustMCP/
├── __init__.py              # Package init
├── main.py                  # Main entry point
├── run_server.py            # Original startup script
├── weknowrust_mcp_server.py # MCP server implementation
├── requirements.txt         # Dependencies
├── setup.py                 # Setup script
├── MANIFEST.in              # Package manifest
├── LICENSE                  # License
├── README.md                # Project README
└── INSTALL.md               # Installation guide
```

### Adding new features
1. Add new API methods in `WeKnowRustClient`
2. Register new tools in `handle_list_tools()`
3. Implement tool logic in `handle_call_tool()`
4. Update docs and tests

### Testing
```bash
# Run basic tests
python test_imports.py

# Test environment configuration
python main.py --check-only

# Test server startup
python main.py --verbose
```

## Deployment

### Docker deployment
Create a `Dockerfile`:
```dockerfile
FROM python:3.11-slim

WORKDIR /app
COPY requirements.txt .
RUN pip install -r requirements.txt

COPY . .
RUN pip install -e .

ENV WEKNOWRUST_BASE_URL=http://localhost:8080/api/v1
EXPOSE 8000

CMD ["weknowrust-mcp-server"]
```

### System service
Create a systemd service file at `/etc/systemd/system/weknowrust-mcp.service`:
```ini
[Unit]
Description=WeKnowRust MCP Server
After=network.target

[Service]
Type=simple
User=weknowrust
WorkingDirectory=/opt/weknowrust-mcp
Environment=WEKNOWRUST_BASE_URL=http://localhost:8080/api/v1
Environment=WEKNOWRUST_API_KEY=your_api_key
ExecStart=/usr/local/bin/weknowrust-mcp-server
Restart=always

[Install]
WantedBy=multi-user.target
```

Enable the service:
```bash
sudo systemctl enable weknowrust-mcp
sudo systemctl start weknowrust-mcp
```

## Support

If you encounter problems, please:
1. Check the logs
2. Verify environment configuration
3. Refer to the troubleshooting section
4. Open an issue in the repository: https://github.com/SHA888/WeKnoRustMCP/issues