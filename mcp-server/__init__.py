#!/usr/bin/env python3
"""
WeKnoRust MCP Server Package

A Model Context Protocol server that provides access to the WeKnoRust knowledge management API.
"""

__version__ = "1.0.0"
__author__ = "WeKnoRust Team"
__description__ = "WeKnoRust MCP Server - Model Context Protocol server for WeKnoRust API"

from .weknorust_mcp_server import WeKnoRustClient, run

__all__ = ["WeKnoRustClient", "run"]