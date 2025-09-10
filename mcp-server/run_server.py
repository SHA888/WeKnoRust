#!/usr/bin/env python3
"""
WeKnowRust MCP Server startup script
"""

import os
import sys
import asyncio

def check_environment():
    """Check environment configuration"""
    base_url = os.getenv("WEKNOWRUST_BASE_URL")
    api_key = os.getenv("WEKNOWRUST_API_KEY")
    
    if not base_url:
        print("Warning: WEKNOWRUST_BASE_URL not set, using default: http://localhost:8080/api/v1")
    
    if not api_key:
        print("Warning: WEKNOWRUST_API_KEY not set")
    
    print(f"WeKnowRust Base URL: {base_url or 'http://localhost:8080/api/v1'}")
    print(f"API Key: {'SET' if api_key else 'NOT SET'}")

def main():
    """Main function"""
    print("Starting WeKnowRust MCP Server...")
    check_environment()
    
    try:
        from weknora_mcp_server import run
        asyncio.run(run())
    except ImportError as e:
        print(f"Import error: {e}")
        print("Please ensure dependencies are installed: pip install -r requirements.txt")
        sys.exit(1)
    except KeyboardInterrupt:
        print("\nServer stopped")
    except Exception as e:
        print(f"Server runtime error: {e}")
        sys.exit(1)

if __name__ == "__main__":
    main()