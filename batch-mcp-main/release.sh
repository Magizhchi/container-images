#!/bin/bash

# Build script for batch-mcp - builds for multiple platforms

set -e

APP_NAME="batch-mcp"
VERSION=${VERSION:-"1.0.0"}
BUILD_DIR="dist"
SRC_DIR="./cmd/batch-mcp"

# Create build directory
mkdir -p ${BUILD_DIR}

# Define platforms to build for
PLATFORMS=(
    "linux/amd64"
    "linux/arm64"
    "darwin/amd64"
    "darwin/arm64" 
    "windows/amd64"
    "windows/arm64"
)

# Function to create Claude MCP configuration file
create_claude_config() {
    local build_dir="$1"
    local os="$2"
    local binary_name="$3"
    
    case "$os" in
        "linux")
            cat > "${build_dir}/claude-mcp-config.json" << EOF
{
  "mcpServers": {
    "batch-mcp": {
      "command": "/usr/local/bin/${binary_name}",
      "args": [],
      "env": {
        "MATLAB_PATH": "/usr/local/MATLAB/R2024a/bin/matlab"
      }
    }
  }
}
EOF
            ;;
        "darwin")
            cat > "${build_dir}/claude-mcp-config.json" << EOF
{
  "mcpServers": {
    "batch-mcp": {
      "command": "/usr/local/bin/${binary_name}",
      "args": [],
      "env": {
        "MATLAB_PATH": "/Applications/MATLAB_R2024a.app/bin/matlab"
      }
    }
  }
}
EOF
            ;;
        "windows")
            cat > "${build_dir}/claude-mcp-config.json" << EOF
{
  "mcpServers": {
    "batch-mcp": {
      "command": "C:\\\\Program Files\\\\batch-mcp\\\\${binary_name}",
      "args": [],
      "env": {
        "MATLAB_PATH": "C:\\\\Program Files\\\\MATLAB\\\\R2024a\\\\bin\\\\matlab.exe"
      }
    }
  }
}
EOF
            ;;
    esac
    
    echo "✓ Created Claude MCP config: ${build_dir}/claude-mcp-config.json"
}

echo "Building ${APP_NAME} v${VERSION} for multiple platforms..."

for platform in "${PLATFORMS[@]}"; do
    IFS='/' read -r GOOS GOARCH <<< "$platform"
    
    case "$GOOS" in
        "linux")
            case "$GOARCH" in
                "amd64") 
                    output_name="${APP_NAME}-linux-x64"
                    platform_dir="linux-x64"
                    ;;
                "arm64") 
                    output_name="${APP_NAME}-linux-arm64"
                    platform_dir="linux-arm64"
                    ;;
            esac
            ;;
        "darwin")
            case "$GOARCH" in
                "amd64") 
                    output_name="${APP_NAME}-macos-x64"
                    platform_dir="macos-x64"
                    ;;
                "arm64") 
                    output_name="${APP_NAME}-macos-arm64"
                    platform_dir="macos-arm64"
                    ;;
            esac
            ;;
        "windows")
            case "$GOARCH" in
                "amd64") 
                    output_name="${APP_NAME}-windows-x64.exe"
                    platform_dir="windows-x64"
                    ;;
                "arm64") 
                    output_name="${APP_NAME}-windows-arm64.exe"
                    platform_dir="windows-arm64"
                    ;;
            esac
            ;;
    esac
    
    # Create platform-specific directory
    mkdir -p "${BUILD_DIR}/${platform_dir}"
    
    echo "Building for ${GOOS}/${GOARCH}..."
    
    GOOS=$GOOS GOARCH=$GOARCH go build -o "${BUILD_DIR}/${platform_dir}/${output_name}" -ldflags="-s -w" ${SRC_DIR}
    
    if [ $? -eq 0 ]; then
        echo "✓ Successfully built ${BUILD_DIR}/${platform_dir}/${output_name}"
        
        # Create platform-specific Claude MCP config file
        create_claude_config "${BUILD_DIR}/${platform_dir}" "$GOOS" "$output_name"
        
    else
        echo "✗ Failed to build for ${GOOS}/${GOARCH}"
        exit 1
    fi
done

echo ""
echo "Build complete! Platform-specific binaries and configs available in ${BUILD_DIR}/"
ls -la ${BUILD_DIR}/