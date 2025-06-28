#!/bin/bash

# WebShell Install Script
# Automatically downloads and installs the appropriate WebShell binary for your platform

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Configuration
REPO="adaptive-scale/webshell"
VERSION=${1:-"latest"}
INSTALL_DIR="/usr/local/bin"
BINARY_NAME="webshell"

echo -e "${BLUE}🚀 WebShell Install Script${NC}"
echo -e "${BLUE}Repository: ${REPO}${NC}"
echo -e "${BLUE}Version: ${VERSION}${NC}"
echo ""

# Detect platform and architecture
detect_platform() {
    local OS=""
    local ARCH=""
    
    case "$(uname -s)" in
        Linux*)     OS="linux";;
        Darwin*)    OS="darwin";;
        CYGWIN*)    OS="windows";;
        MINGW*)     OS="windows";;
        MSYS*)      OS="windows";;
        *)          echo -e "${RED}❌ Unsupported operating system${NC}" >&2; exit 1;;
    esac
    
    case "$(uname -m)" in
        x86_64)     ARCH="amd64";;
        amd64)      ARCH="amd64";;
        arm64)      ARCH="arm64";;
        aarch64)    ARCH="arm64";;
        *)          echo -e "${RED}❌ Unsupported architecture: $(uname -m)${NC}" >&2; exit 1;;
    esac
    
    echo -e "${GREEN}✅ Detected: ${OS} ${ARCH}${NC}" >&2
    echo "${OS}_${ARCH}"
}

# Download binary
download_binary() {
    local platform=$1
    local download_url=""
    
    if [ "$VERSION" = "latest" ]; then
        # Get latest release
        local latest_tag=$(curl -s "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
        VERSION=$latest_tag
    fi
    
    # Remove 'v' prefix if present
    VERSION=${VERSION#v}
    
    local filename="webshell_${platform}"
    if [ "$platform" = "windows_amd64" ]; then
        filename="webshell.exe"
    fi
    
    download_url="https://github.com/${REPO}/releases/download/v${VERSION}/${filename}"
    
    echo -e "${YELLOW}📥 Downloading WebShell v${VERSION} for ${platform}...${NC}"
    echo -e "${YELLOW}URL: ${download_url}${NC}"
    
    # Download the binary
    if curl -L -o "${filename}" "${download_url}"; then
        echo -e "${GREEN}✅ Download successful${NC}"
    else
        echo -e "${RED}❌ Download failed${NC}"
        echo -e "${YELLOW}💡 Make sure the release exists at: https://github.com/${REPO}/releases/tag/v${VERSION}${NC}"
        exit 1
    fi
    
    echo "${filename}"
}

# Install binary
install_binary() {
    local binary_file=$1
    
    echo -e "${YELLOW}🔧 Installing WebShell...${NC}"
    
    # Make binary executable
    chmod +x "${binary_file}"
    
    # Check if we can write to /usr/local/bin
    if [ -w "$INSTALL_DIR" ] || sudo -n true 2>/dev/null; then
        # Install to system directory
        if sudo mv "${binary_file}" "${INSTALL_DIR}/${BINARY_NAME}"; then
            echo -e "${GREEN}✅ Installed to ${INSTALL_DIR}/${BINARY_NAME}${NC}"
        else
            echo -e "${RED}❌ Failed to install to system directory${NC}"
            exit 1
        fi
    else
        # Install to user directory
        local user_bin="$HOME/.local/bin"
        mkdir -p "$user_bin"
        mv "${binary_file}" "$user_bin/${BINARY_NAME}"
        echo -e "${GREEN}✅ Installed to ${user_bin}/${BINARY_NAME}${NC}"
        echo -e "${YELLOW}💡 Add ${user_bin} to your PATH if not already added${NC}"
        echo -e "${YELLOW}   Add this line to your shell profile (.bashrc, .zshrc, etc.):${NC}"
        echo -e "${BLUE}   export PATH=\"\$HOME/.local/bin:\$PATH\"${NC}"
    fi
}

# Verify installation
verify_installation() {
    echo -e "${YELLOW}🔍 Verifying installation...${NC}"
    
    if command -v "$BINARY_NAME" >/dev/null 2>&1; then
        local version=$("$BINARY_NAME" --version 2>/dev/null || echo "unknown version")
        echo -e "${GREEN}✅ WebShell installed successfully!${NC}"
        echo -e "${GREEN}   Version: ${version}${NC}"
        echo -e "${GREEN}   Location: $(which $BINARY_NAME)${NC}"
    else
        echo -e "${RED}❌ Installation verification failed${NC}"
        echo -e "${YELLOW}💡 Try adding the installation directory to your PATH${NC}"
        exit 1
    fi
}

# Show usage information
show_usage() {
    echo -e "${BLUE}📖 Usage Information${NC}"
    echo ""
    echo -e "${GREEN}🚀 Start WebShell:${NC}"
    echo "   $BINARY_NAME"
    echo ""
    echo -e "${GREEN}🌐 Access Web Interface:${NC}"
    echo "   http://localhost:8080"
    echo ""
    echo -e "${GREEN}🖥️ Access Web Terminal:${NC}"
    echo "   http://localhost:8080/terminal"
    echo ""
    echo -e "${GREEN}📚 Documentation:${NC}"
    echo "   https://github.com/${REPO}"
    echo ""
    echo -e "${GREEN}🔧 Custom Port:${NC}"
    echo "   PORT=3000 $BINARY_NAME"
}

# Main installation process
main() {
    echo -e "${BLUE}🔍 Detecting your platform...${NC}"
    local platform=$(detect_platform)
    
    echo -e "${BLUE}📦 Downloading WebShell...${NC}"
    local binary_file=$(download_binary "$platform")
    
    echo -e "${BLUE}🔧 Installing WebShell...${NC}"
    install_binary "$binary_file"
    
    echo -e "${BLUE}✅ Verifying installation...${NC}"
    verify_installation
    
    echo ""
    echo -e "${GREEN}🎉 WebShell installation completed successfully!${NC}"
    echo ""
    show_usage
}

# Handle command line arguments
case "${1:-}" in
    -h|--help)
        echo "WebShell Install Script"
        echo ""
        echo "Usage: $0 [version]"
        echo ""
        echo "Arguments:"
        echo "  version    Specific version to install (default: latest)"
        echo ""
        echo "Examples:"
        echo "  $0              # Install latest version"
        echo "  $0 v0.1.6       # Install specific version"
        echo "  $0 0.1.6        # Install specific version (without v prefix)"
        exit 0
        ;;
    *)
        main
        ;;
esac 