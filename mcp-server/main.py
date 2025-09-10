#!/usr/bin/env python3
"""
WeKnoRust MCP Server main entry point

This file provides a unified entry to start the WeKnoRust MCP server.
You can run it in multiple ways:
1. python main.py
2. python -m weknorust_mcp_server
3. weknorust-mcp-server (after installation)
"""

import os
import sys
import asyncio
import argparse
from pathlib import Path

def setup_environment():
    """Set up environment and paths"""
    # Ensure current dir is on Python path
    current_dir = Path(__file__).parent.absolute()
    if str(current_dir) not in sys.path:
        sys.path.insert(0, str(current_dir))

def check_dependencies():
    """Check required dependencies are installed"""
    try:
        import mcp
        import requests
        return True
    except ImportError as e:
        print(f"Missing dependency: {e}")
        print("Please run: pip install -r requirements.txt")
        return False

def check_environment_variables():
    """Check environment variable configuration"""
    base_url = os.getenv("WEKNORUST_BASE_URL")
    api_key = os.getenv("WEKNORUST_API_KEY")
    
    print("=== WeKnoRust MCP Server Environment Check ===")
    print(f"Base URL: {base_url or 'http://localhost:8080/api/v1 (default)'}")
    print(f"API Key: {'SET' if api_key else 'NOT SET (warning)'}")
    
    if not base_url:
        print("Tip: You can set WEKNORUST_BASE_URL environment variable")
    
    if not api_key:
        print("Warning: It is recommended to set WEKNORUST_API_KEY environment variable")
    
    print("=" * 40)
    return True

def parse_arguments():
    """Parse command line arguments"""
    parser = argparse.ArgumentParser(
        description="WeKnoRust MCP Server - Model Context Protocol server for WeKnoRust API",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
Examples:
  python main.py                    # Start with default configuration
  python main.py --check-only       # Check environment, do not start server
  python main.py --verbose          # Enable verbose logging
  
Environment variables:
  WEKNORUST_BASE_URL    WeKnoRust API base URL (default: http://localhost:8080/api/v1)
  WEKNORUST_API_KEY     WeKnoRust API key
        """
    )
    
    parser.add_argument(
        "--check-only",
        action="store_true",
        help="Check environment only, do not start server"
    )
    
    parser.add_argument(
        "--verbose", "-v",
        action="store_true",
        help="Enable verbose logging"
    )
    
    parser.add_argument(
        "--version",
        action="version",
        version="WeKnowRust MCP Server 1.0.0"
    )
    
    return parser.parse_args()

async def main():
    """Main function"""
    args = parse_arguments()
    
    # Set up environment
    setup_environment()
    
    # Check dependencies
    if not check_dependencies():
        sys.exit(1)
    
    # Check environment variables
    check_environment_variables()
    
    # Exit if only checking environment
    if args.check_only:
        print("Environment check complete.")
        return
    
    # Configure logging level
    if args.verbose:
        import logging
        logging.basicConfig(level=logging.DEBUG)
        print("Verbose logging enabled")
    
    try:
        print("Starting WeKnoRust MCP Server...")
        
        # Import and run server
        from weknorust_mcp_server import run
        await run()
        
    except ImportError as e:
        print(f"Import error: {e}")
        print("Please ensure all files are in the correct locations")
        sys.exit(1)
    except KeyboardInterrupt:
        print("\nServer stopped")
    except Exception as e:
        print(f"Server runtime error: {e}")
        if args.verbose:
            import traceback
            traceback.print_exc()
        sys.exit(1)

def sync_main():
    """Synchronous entry point for entry_points"""
    asyncio.run(main())

if __name__ == "__main__":
    asyncio.run(main())