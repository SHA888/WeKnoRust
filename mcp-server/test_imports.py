#!/usr/bin/env python3
"""
Test MCP imports
"""

try:
    import mcp.types as types
    print("✓ mcp.types import OK")
except ImportError as e:
    print(f"✗ mcp.types import FAILED: {e}")

try:
    from mcp.server import Server, NotificationOptions
    print("✓ mcp.server import OK")
except ImportError as e:
    print(f"✗ mcp.server import FAILED: {e}")

try:
    import mcp.server.stdio
    print("✓ mcp.server.stdio import OK")
except ImportError as e:
    print(f"✗ mcp.server.stdio import FAILED: {e}")

try:
    from mcp.server.models import InitializationOptions
    print("✓ InitializationOptions import from mcp.server.models OK")
except ImportError:
    try:
        from mcp import InitializationOptions
        print("✓ InitializationOptions import from mcp OK")
    except ImportError as e:
        print(f"✗ InitializationOptions import FAILED: {e}")

# Check MCP package structure
import mcp
print(f"\nMCP version: {getattr(mcp, '__version__', 'unknown')}")
print(f"MCP path: {mcp.__file__}")
print(f"MCP contents: {dir(mcp)}")