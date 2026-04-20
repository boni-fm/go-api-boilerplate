#!/usr/bin/env bash
# build/build.sh — Cross-platform build script for the Go API Boilerplate.
#
# Usage:
#   ./build/build.sh [GOOS] [GOARCH]
#
# Examples:
#   ./build/build.sh                     # build for current OS/arch
#   ./build/build.sh linux amd64         # Linux 64-bit binary
#   ./build/build.sh windows amd64       # Windows 64-bit binary
#   ./build/build.sh darwin arm64        # macOS Apple Silicon binary
#
# The output binary is written to ./bin/<AppName>[.exe].
# AppName is read from appsettings.ini; if the file is missing or the key is
# absent, the directory name is used as a fallback.

set -euo pipefail

# ---------------------------------------------------------------------------
# Read AppName from appsettings.ini
# ---------------------------------------------------------------------------
INI_FILE="appsettings.ini"
APP_NAME=""

if [ -f "$INI_FILE" ]; then
    APP_NAME=$(grep -E "^\s*AppName\s*=" "$INI_FILE" | head -1 | sed 's/.*=\s*//' | tr -d '[:space:]')
fi

if [ -z "$APP_NAME" ]; then
    APP_NAME=$(basename "$(pwd)")
    echo "[build] AppName not found in $INI_FILE — using directory name: $APP_NAME"
fi

# ---------------------------------------------------------------------------
# Resolve target OS and architecture
# ---------------------------------------------------------------------------
TARGET_OS="${1:-$(go env GOOS)}"
TARGET_ARCH="${2:-$(go env GOARCH)}"

# Append .exe for Windows binaries
OUTPUT_EXT=""
if [ "$TARGET_OS" = "windows" ]; then
    OUTPUT_EXT=".exe"
fi

# ---------------------------------------------------------------------------
# Build
# ---------------------------------------------------------------------------
OUTPUT_DIR="bin"
OUTPUT_BIN="${OUTPUT_DIR}/${APP_NAME}${OUTPUT_EXT}"

mkdir -p "$OUTPUT_DIR"

echo "------------------------------------------------------------"
echo "  App     : $APP_NAME"
echo "  Target  : ${TARGET_OS}/${TARGET_ARCH}"
echo "  Output  : $OUTPUT_BIN"
echo "------------------------------------------------------------"

GOOS="$TARGET_OS" GOARCH="$TARGET_ARCH" go build \
    -ldflags="-s -w" \
    -o "$OUTPUT_BIN" \
    .

echo "[build] Done → $OUTPUT_BIN"
