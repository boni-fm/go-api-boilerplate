#!/usr/bin/env bash
# setup.sh — Developer environment initialisation for the Go API Boilerplate.
#
# Run this script once after cloning the repository to install tools,
# verify dependencies, and generate the initial Swagger documentation.
#
# Usage:
#   chmod +x setup.sh && ./setup.sh

set -euo pipefail

BOLD="\033[1m"
GREEN="\033[0;32m"
YELLOW="\033[0;33m"
CYAN="\033[0;36m"
RESET="\033[0m"

echo ""
echo -e "${BOLD}┌──────────────────────────────────────────────────────────┐${RESET}"
echo -e "${BOLD}│      IT SD3 Go API Boilerplate — Developer Setup         │${RESET}"
echo -e "${BOLD}└──────────────────────────────────────────────────────────┘${RESET}"
echo ""

# ---------------------------------------------------------------------------
# 1. Verify Go installation
# ---------------------------------------------------------------------------
echo -e "${CYAN}[1/5] Checking Go installation...${RESET}"
if ! command -v go &>/dev/null; then
    echo -e "${YELLOW}  ✗ Go is not installed. Download it from https://go.dev/dl/${RESET}"
    exit 1
fi

GO_VERSION=$(go version | awk '{print $3}')
echo -e "  ✓ Found ${GO_VERSION}"

# Extract the required version from go.mod and warn if the local version is older.
REQUIRED=$(grep '^go ' go.mod | awk '{print "go"$2}')
if [ "$GO_VERSION" \< "$REQUIRED" ]; then
    echo -e "${YELLOW}  ⚠ Your Go version (${GO_VERSION}) is older than required (${REQUIRED}).${RESET}"
    echo -e "${YELLOW}    Please upgrade: https://go.dev/dl/${RESET}"
fi

# ---------------------------------------------------------------------------
# 2. Download / tidy modules
# ---------------------------------------------------------------------------
echo ""
echo -e "${CYAN}[2/5] Downloading Go modules...${RESET}"
go mod download
go mod tidy
echo -e "  ✓ Modules up to date"

# ---------------------------------------------------------------------------
# 3. Install swag CLI (Swagger code-generator)
# ---------------------------------------------------------------------------
echo ""
echo -e "${CYAN}[3/5] Installing swag CLI...${RESET}"
if command -v swag &>/dev/null; then
    SWAG_VER=$(swag --version 2>&1 | head -1)
    echo -e "  ✓ swag already installed: ${SWAG_VER}"
else
    go install github.com/swaggo/swag/cmd/swag@latest
    echo -e "  ✓ swag installed"
fi

# ---------------------------------------------------------------------------
# 4. Generate Swagger documentation
# ---------------------------------------------------------------------------
echo ""
echo -e "${CYAN}[4/5] Generating Swagger documentation...${RESET}"
if command -v swag &>/dev/null; then
    swag init -g main.go -o docs
    echo -e "  ✓ docs/ updated"
else
    echo -e "${YELLOW}  ⚠ swag not found in PATH — skipping doc generation.${RESET}"
    echo -e "    Ensure \$(go env GOPATH)/bin is in your PATH and re-run setup.sh."
fi

# ---------------------------------------------------------------------------
# 5. Verify the project builds cleanly
# ---------------------------------------------------------------------------
echo ""
echo -e "${CYAN}[5/5] Building project...${RESET}"
go build ./...
echo -e "  ✓ Build successful"

# ---------------------------------------------------------------------------
# Summary
# ---------------------------------------------------------------------------
echo ""
echo -e "${GREEN}${BOLD}✓ Setup complete!${RESET}"
echo ""
echo -e "  ${BOLD}What changed / was verified:${RESET}"
echo -e "   • Go modules downloaded and tidied"
echo -e "   • swag CLI installed (Swagger doc generator)"
echo -e "   • Swagger docs generated in docs/"
echo -e "   • Project compiled successfully"
echo ""
echo -e "  ${BOLD}Next steps:${RESET}"
echo -e "   1. Copy appsettings.ini and fill in your database credentials (Kunci)."
echo -e "   2. Set IsDevelopment = true for local development (enables Swagger UI)."
echo -e "   3. Set Timezone to your region (e.g. Asia/Jakarta)."
echo -e "   4. Run the server:  go run main.go"
echo -e "   5. Open Swagger UI: http://localhost:8080/swagger"
echo ""
echo -e "   To build a release binary:  ${CYAN}./build/build.sh [GOOS] [GOARCH]${RESET}"
echo ""
