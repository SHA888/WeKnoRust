#!/usr/bin/env python3
"""
WeKnowRust MCP Server quick start script

This is a simplified startup script that provides the basics.
For more options, please use main.py
"""

import sys
import os
from pathlib import Path

def main():
    """Simple startup function"""
    # Add current directory to Python path
    current_dir = Path(__file__).parent.absolute()
    if str(current_dir) not in sys.path:
        sys.path.insert(0, str(current_dir))
    
    # Check environment variables
    base_url = os.getenv("WEKNOWRUST_BASE_URL", "http://localhost:8080/api/v1")
    api_key = os.getenv("WEKNOWRUST_API_KEY", "")
    
    print("WeKnowRust MCP Server")
    print(f"Base URL: {base_url}")
    print(f"API Key: {'SET' if api_key else 'NOT SET'}")
    print("-" * 40)
    
    try:
        # Import and run
        from main import sync_main
        sync_main()
    except ImportError:
        print("Error: Could not import required modules")
        print("Please run: pip install -r requirements.txt")
        sys.exit(1)
    except KeyboardInterrupt:
        print("\nServer stopped")
    except Exception as e:
        print(f"Error: {e}")
        sys.exit(1)

if __name__ == "__main__":
    main()