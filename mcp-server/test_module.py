#!/usr/bin/env python3
"""
WeKnowRust MCP Server module test script

Test various startup methods and features of the module.
"""

import os
import sys
import subprocess
import importlib.util
from pathlib import Path

def test_imports():
    """Test module imports"""
    print("=== Test Module Imports ===")
    
    try:
        # Test base dependencies
        import mcp
        print("✓ mcp module import OK")
        
        import requests
        print("✓ requests module import OK")
        
        # Test main module
        import weknora_mcp_server
        print("✓ weknora_mcp_server module import OK")
        
        # Test package imports
        from weknora_mcp_server import WeKnoraClient, run
        print("✓ WeKnoraClient and run import OK")
        
        # Test main entry point
        import main
        print("✓ main module import OK")
        
        return True
        
    except ImportError as e:
        print(f"✗ Import FAILED: {e}")
        return False

def test_environment():
    """Test environment configuration"""
    print("\n=== Test Environment Configuration ===")
    
    base_url = os.getenv("WEKNOWRUST_BASE_URL")
    api_key = os.getenv("WEKNOWRUST_API_KEY")
    
    print(f"WEKNOWRUST_BASE_URL: {base_url or 'NOT SET (will use default)'}")
    print(f"WEKNOWRUST_API_KEY: {'SET' if api_key else 'NOT SET'}")
    
    if not base_url:
        print("Tip: You can set environment variable WEKNOWRUST_BASE_URL")
    
    if not api_key:
        print("Tip: Consider setting environment variable WEKNOWRUST_API_KEY")
    
    return True

def test_client_creation():
    """Test client creation"""
    print("\n=== Test Client Creation ===")
    
    try:
        from weknora_mcp_server import WeKnoraClient
        
        base_url = os.getenv("WEKNOWRUST_BASE_URL", "http://localhost:8080/api/v1")
        api_key = os.getenv("WEKNOWRUST_API_KEY", "test_key")
        
        client = WeKnoraClient(base_url, api_key)
        print("✓ WeKnoraClient created successfully")
        
        # Verify client properties
        assert client.base_url == base_url
        assert client.api_key == api_key
        print("✓ Client configuration OK")
        
        return True
        
    except Exception as e:
        print(f"✗ Client creation FAILED: {e}")
        return False

def test_file_structure():
    """Test file structure"""
    print("\n=== Test File Structure ===")
    
    required_files = [
        "__init__.py",
        "main.py", 
        "run_server.py",
        "weknora_mcp_server.py",
        "requirements.txt",
        "setup.py",
        "pyproject.toml",
        "README.md",
        "INSTALL.md",
        "LICENSE",
        "MANIFEST.in"
    ]
    
    missing_files = []
    for file in required_files:
        if Path(file).exists():
            print(f"✓ {file}")
        else:
            print(f"✗ {file} (MISSING)")
            missing_files.append(file)
    
    if missing_files:
        print(f"Missing files: {missing_files}")
        return False
    
    print("✓ All required files exist")
    return True

def test_entry_points():
    """Test entry points"""
    print("\n=== Test Entry Points ===")
    
    # Test main.py --help option
    try:
        result = subprocess.run(
            [sys.executable, "main.py", "--help"],
            capture_output=True,
            text=True,
            timeout=10
        )
        if result.returncode == 0:
            print("✓ main.py --help OK")
        else:
            print(f"✗ main.py --help FAILED: {result.stderr}")
            return False
    except subprocess.TimeoutExpired:
        print("✗ main.py --help TIMEOUT")
        return False
    except Exception as e:
        print(f"✗ main.py --help ERROR: {e}")
        return False
    
    # Test environment check
    try:
        result = subprocess.run(
            [sys.executable, "main.py", "--check-only"],
            capture_output=True,
            text=True,
            timeout=10
        )
        if result.returncode == 0:
            print("✓ main.py --check-only OK")
        else:
            print(f"✗ main.py --check-only FAILED: {result.stderr}")
            return False
    except subprocess.TimeoutExpired:
        print("✗ main.py --check-only TIMEOUT")
        return False
    except Exception as e:
        print(f"✗ main.py --check-only ERROR: {e}")
        return False
    
    return True

def test_package_installation():
    """Test package installation (dev mode)"""
    print("\n=== Test Package Installation ===")
    
    try:
        # Check setup.py basic invocation
        result = subprocess.run(
            [sys.executable, "setup.py", "check"],
            capture_output=True,
            text=True,
            timeout=30
        )
        
        if result.returncode == 0:
            print("✓ setup.py check OK")
        else:
            print(f"✗ setup.py check FAILED: {result.stderr}")
            return False
            
    except subprocess.TimeoutExpired:
        print("✗ setup.py check TIMEOUT")
        return False
    except Exception as e:
        print(f"✗ setup.py check ERROR: {e}")
        return False
    
    return True

def main():
    """Run all tests"""
    print("WeKnowRust MCP Server module tests")
    print("=" * 50)
    
    tests = [
        ("Module imports", test_imports),
        ("Environment", test_environment),
        ("Client creation", test_client_creation),
        ("File structure", test_file_structure),
        ("Entry points", test_entry_points),
        ("Package installation", test_package_installation),
    ]
    
    passed = 0
    total = len(tests)
    
    for test_name, test_func in tests:
        try:
            if test_func():
                passed += 1
            else:
                print(f"Test FAILED: {test_name}")
        except Exception as e:
            print(f"Test ERROR: {test_name} - {e}")
    
    print("\n" + "=" * 50)
    print(f"Test results: {passed}/{total} passed")
    
    if passed == total:
        print("✓ All tests passed! Module is usable.")
        return True
    else:
        print("✗ Some tests FAILED. See errors above.")
        return False

if __name__ == "__main__":
    success = main()
    sys.exit(0 if success else 1)