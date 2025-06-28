#!/bin/bash

# WebShell Install Script
set -e

# Configuration
REPO="adaptive-scale/webshell"
VERSION=${1:-"latest"}
BINARY_NAME="webshell"

echo "WebShell Install Script"
echo "Repository: ${REPO}"
echo "Version: ${VERSION}"
echo ""

# Detect platform
OS=""
ARCH=""

case "$(uname -s)" in
    Linux*)     OS="linux";;
    Darwin*)    OS="darwin";;
    CYGWIN*)    OS="windows";;
    MINGW*)     OS="windows";;
    MSYS*)      OS="windows";;
    *)          echo "Unsupported operating system"; exit 1;;
esac

case "$(uname -m)" in
    x86_64)     ARCH="amd64";;
    amd64)      ARCH="amd64";;
    arm64)      ARCH="arm64";;
    aarch64)    ARCH="arm64";;
    *)          echo "Unsupported architecture: $(uname -m)"; exit 1;;
esac

PLATFORM="${OS}_${ARCH}"
echo "Detected: ${OS} ${ARCH}"

# Get version
if [ "$VERSION" = "latest" ]; then
    VERSION=$(curl -s "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
fi
VERSION=${VERSION#v}

# Set filename
FILENAME="webshell_${PLATFORM}"
if [ "$PLATFORM" = "windows_amd64" ]; then
    FILENAME="webshell.exe"
fi

DOWNLOAD_URL="https://github.com/${REPO}/releases/download/v${VERSION}/${FILENAME}"

echo "Downloading WebShell v${VERSION} for ${PLATFORM}"
echo "URL: ${DOWNLOAD_URL}"

# Download
if curl -L -o "${FILENAME}" "${DOWNLOAD_URL}"; then
    echo "Download successful"
else
    echo "Download failed"
    echo "Make sure the release exists at: https://github.com/${REPO}/releases/tag/v${VERSION}"
    exit 1
fi

# Install
echo "Installing WebShell"
chmod +x "${FILENAME}"

if [ -w "/usr/local/bin" ] || sudo -n true 2>/dev/null; then
    sudo mv "${FILENAME}" "/usr/local/bin/${BINARY_NAME}"
    echo "Installed to /usr/local/bin/${BINARY_NAME}"
else
    mkdir -p "$HOME/.local/bin"
    mv "${FILENAME}" "$HOME/.local/bin/${BINARY_NAME}"
    echo "Installed to $HOME/.local/bin/${BINARY_NAME}"
    echo "Add $HOME/.local/bin to your PATH if not already added"
    echo "Add this line to your shell profile:"
    echo "export PATH=\"\$HOME/.local/bin:\$PATH\""
    echo "Alternative: you can run ~/.local/bin/webshell"
fi

# Verify
echo "Verifying installation"
if command -v "$BINARY_NAME" >/dev/null 2>&1; then
    VERSION_INFO=$("$BINARY_NAME" --version 2>/dev/null || echo "unknown version")
    echo "WebShell installed successfully"
    echo "Version: ${VERSION_INFO}"
    echo "Location: $(which $BINARY_NAME)"
else
    echo "Installation verification failed"
    echo "Try adding the installation directory to your PATH"
    exit 1
fi

echo ""
echo "WebShell installation completed successfully"
echo ""
echo "Usage Information:"
echo "Start WebShell: $BINARY_NAME"
echo "Access Web Interface: http://localhost:8080"
echo "Access Web Terminal: http://localhost:8080/terminal"
echo "Documentation: https://github.com/${REPO}"
echo "Custom Port: PORT=3000 $BINARY_NAME"

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
        ;;
esac 