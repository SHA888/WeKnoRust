#!/usr/bin/env python3
"""
WeKnoRust MCP Server setup script
"""

from setuptools import setup
import os

# Read README file
def read_readme():
    try:
        with open("README.md", "r", encoding="utf-8") as f:
            return f.read()
    except FileNotFoundError:
        return "WeKnowRust MCP Server - Model Context Protocol server for WeKnowRust API"

# Read requirements
def read_requirements():
    try:
        with open("requirements.txt", "r", encoding="utf-8") as f:
            return [line.strip() for line in f if line.strip() and not line.startswith("#")]
    except FileNotFoundError:
        return ["mcp>=1.0.0", "requests>=2.31.0"]

setup(
    name="weknorust-mcp-server",
    version="1.0.0",
    author="WeKnoRust Team",
    author_email="support@weknorust.com",
    description="WeKnoRust MCP Server - Model Context Protocol server for WeKnoRust API",
    long_description=read_readme(),
    long_description_content_type="text/markdown",
    url="https://github.com/SHA888/WeKnoRustMCP",
    py_modules=["weknorust_mcp_server", "main", "run_server", "run", "test_module"],
    classifiers=[
        "Development Status :: 4 - Beta",
        "Intended Audience :: Developers",
        "License :: OSI Approved :: MIT License",
        "Operating System :: OS Independent",
        "Programming Language :: Python :: 3",
        "Programming Language :: Python :: 3.8",
        "Programming Language :: Python :: 3.9",
        "Programming Language :: Python :: 3.10",
        "Programming Language :: Python :: 3.11",
        "Programming Language :: Python :: 3.12",
        "Topic :: Software Development :: Libraries :: Python Modules",
        "Topic :: Internet :: WWW/HTTP :: HTTP Servers",
        "Topic :: Scientific/Engineering :: Artificial Intelligence",
    ],
    python_requires=">=3.8",
    install_requires=read_requirements(),
    entry_points={
        "console_scripts": [
            "weknorust-mcp-server=main:sync_main",
            "weknorust-server=run_server:main",
            # no backward compatibility aliases
        ],
    },
    include_package_data=True,
    data_files=[
        ("", ["README.md", "requirements.txt", "LICENSE"]),
    ],
    keywords="mcp model-context-protocol weknowrust knowledge-management api-server",
)