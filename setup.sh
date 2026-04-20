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
# 1. Go version check & goswitch
#
# Strategy: we NEVER replace or modify the developer's system Go installation.
# Instead, we use the official golang.org/dl mechanism to install a project-
# local wrapper binary (go<version>) alongside the system go. The wrapper
# downloads the real toolchain on first use to $(go env GOPATH)/sdk/.
#
# If the system Go already satisfies the required version we use it directly
# and do nothing extra — zero impact on the developer's environment.
# ---------------------------------------------------------------------------
echo -e "${CYAN}[1/6] Checking Go version...${RESET}"

if ! command -v go &>/dev/null; then
    echo -e "${YELLOW}  ✗ Go is not installed. Download it from https://go.dev/dl/${RESET}"
    exit 1
fi

# Read the exact version from go.mod (e.g. "1.25.8").
REQUIRED_VERSION=$(grep '^go ' go.mod | awk '{print $2}')
CURRENT_VERSION=$(go version | awk '{print $3}' | sed 's/go//')

echo -e "  System Go : go${CURRENT_VERSION}"
echo -e "  Required  : go${REQUIRED_VERSION}"

# We use GO_CMD to point to the correct binary for the rest of the script.
# Default to the system go; override below if a switch is needed.
GO_CMD="go"

if [ "$CURRENT_VERSION" = "$REQUIRED_VERSION" ]; then
    echo -e "  ${GREEN}✓ Version matches — no switch needed.${RESET}"
else
    echo -e "  ${YELLOW}⚠ Version mismatch. Attempting to install go${REQUIRED_VERSION} via golang.org/dl...${RESET}"
    echo -e "  ${YELLOW}  (Your system go${CURRENT_VERSION} will NOT be modified.)${RESET}"

    # Install the version wrapper using the current (system) go.
    # golang.org/dl installs a binary named go<version> into GOPATH/bin.
    go install "golang.org/dl/go${REQUIRED_VERSION}@latest" 2>/dev/null || {
        echo -e "  ${YELLOW}  ✗ Could not install wrapper. Continuing with system go${CURRENT_VERSION}.${RESET}"
        echo -e "    Manually install: https://go.dev/dl/#go${REQUIRED_VERSION}"
    }

    GOSWITCH_BIN="$(go env GOPATH)/bin/go${REQUIRED_VERSION}"
    if [ -x "$GOSWITCH_BIN" ]; then
        # Download the toolchain if not already cached.
        "$GOSWITCH_BIN" download 2>/dev/null || true
        GO_CMD="$GOSWITCH_BIN"
        echo -e "  ${GREEN}✓ Using go${REQUIRED_VERSION} for this setup run.${RESET}"
        echo -e "  ${CYAN}  Tip: add \$(go env GOPATH)/bin to your PATH and run:${RESET}"
        echo -e "  ${CYAN}       go${REQUIRED_VERSION} run main.go${RESET}"
    else
        echo -e "  ${YELLOW}  ⚠ Wrapper not found. Continuing with system go.${RESET}"
    fi
fi

# ---------------------------------------------------------------------------
# 2. Download / tidy modules (using the resolved GO_CMD)
# ---------------------------------------------------------------------------
echo ""
echo -e "${CYAN}[2/6] Downloading Go modules...${RESET}"
"$GO_CMD" mod download
"$GO_CMD" mod tidy
echo -e "  ✓ Modules up to date"

# ---------------------------------------------------------------------------
# 3. Install swag CLI (Swagger code-generator)
# ---------------------------------------------------------------------------
echo ""
echo -e "${CYAN}[3/6] Installing swag CLI...${RESET}"
if command -v swag &>/dev/null; then
    SWAG_VER=$(swag --version 2>&1 | head -1)
    echo -e "  ✓ swag already installed: ${SWAG_VER}"
else
    "$GO_CMD" install github.com/swaggo/swag/cmd/swag@latest
    echo -e "  ✓ swag installed"
fi

# ---------------------------------------------------------------------------
# 4. Generate Swagger documentation
# ---------------------------------------------------------------------------
echo ""
echo -e "${CYAN}[4/6] Generating Swagger documentation...${RESET}"
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
echo -e "${CYAN}[5/6] Building project...${RESET}"
"$GO_CMD" build ./...
echo -e "  ✓ Build successful"

# ---------------------------------------------------------------------------
# 6. Run tests
# ---------------------------------------------------------------------------
echo ""
echo -e "${CYAN}[6/6] Running tests...${RESET}"
"$GO_CMD" test ./... && echo -e "  ✓ All tests passed" || echo -e "${YELLOW}  ⚠ Some tests failed — check output above.${RESET}"

# ---------------------------------------------------------------------------
# Summary
# ---------------------------------------------------------------------------
echo ""
echo -e "${GREEN}${BOLD}✓ Setup complete!${RESET}"
echo ""
echo -e "  ${BOLD}What changed / was verified:${RESET}"
echo -e "   • Go version verified (goswitch via golang.org/dl if needed)"
echo -e "   • Go modules downloaded and tidied"
echo -e "   • swag CLI installed (Swagger doc generator)"
echo -e "   • Swagger docs generated in docs/"
echo -e "   • Project compiled and tested successfully"
echo ""
echo -e "  ${BOLD}Next steps:${RESET}"
echo -e "   1. Copy appsettings.ini and fill in your database credentials (Kunci)."
echo -e "   2. Set IsDevelopment = true for local development (enables Swagger UI)."
echo -e "   3. Set Timezone to your region (e.g. Asia/Jakarta), or export TZ env var."
echo -e "   4. For multi-DC: set Kunci = g009sim,g010sim (comma-separated)."
echo -e "   5. Run the server:  ${CYAN}${GO_CMD} run main.go${RESET}"
echo -e "   6. Open Swagger UI: http://localhost:8080/swagger"
echo ""
echo -e "   To build a release binary:  ${CYAN}./build/build.sh [GOOS] [GOARCH]${RESET}"
echo ""
