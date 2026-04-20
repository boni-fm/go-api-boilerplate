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
RED="\033[0;31m"
RESET="\033[0m"

echo ""
echo -e "${BOLD}┌──────────────────────────────────────────────────────────┐${RESET}"
echo -e "${BOLD}│      IT SD3 Go API Boilerplate — Developer Setup         │${RESET}"
echo -e "${BOLD}└──────────────────────────────────────────────────────────┘${RESET}"
echo ""
echo -e "  This script will:"
echo -e "    1. Check your Go version and install the required one if needed"
echo -e "    2. Download Go modules"
echo -e "    3. Install the swag CLI (Swagger doc generator)"
echo -e "    4. Generate Swagger documentation"
echo -e "    5. Build the project"
echo -e "    6. Run the tests"
echo ""
echo -e "  ${CYAN}Your system Go installation will NOT be modified.${RESET}"
echo -e "  If a version switch is needed, a project-local wrapper is"
echo -e "  installed via golang.org/dl (stored in \$(go env GOPATH)/bin/)."
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
    echo -e "${RED}  ✗ Go is not installed.${RESET}"
    echo -e "    Please install Go from https://go.dev/dl/ and re-run this script."
    exit 1
fi

# Read the exact version from go.mod (e.g. "1.25.8").
REQUIRED_VERSION=$(grep '^go ' go.mod | awk '{print $2}')
CURRENT_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
GOPATH_BIN="$(go env GOPATH)/bin"

echo -e "  System Go : go${CURRENT_VERSION}"
echo -e "  Required  : go${REQUIRED_VERSION}"
echo -e "  GOPATH/bin: ${GOPATH_BIN}"

# We use GO_CMD to point to the correct binary for the rest of the script.
# Default to the system go; override below if a switch is needed.
GO_CMD="go"

if [ "$CURRENT_VERSION" = "$REQUIRED_VERSION" ]; then
    echo -e "  ${GREEN}✓ Version matches — no switch needed.${RESET}"
else
    echo ""
    echo -e "  ${YELLOW}⚠ Version mismatch detected.${RESET}"
    echo -e "  ${YELLOW}  Installing go${REQUIRED_VERSION} via golang.org/dl...${RESET}"
    echo -e "  ${YELLOW}  (Your system go${CURRENT_VERSION} will NOT be modified.)${RESET}"
    echo ""

    # Install the version wrapper using the current (system) go.
    # golang.org/dl installs a binary named go<version> into GOPATH/bin.
    echo -e "  Installing wrapper binary..."
    go install "golang.org/dl/go${REQUIRED_VERSION}@latest" 2>/dev/null || {
        echo -e "  ${YELLOW}  ✗ Could not install wrapper via golang.org/dl.${RESET}"
        echo -e "    This usually means go${REQUIRED_VERSION} is not yet published."
        echo -e "    Continuing with system go${CURRENT_VERSION}."
        echo -e "    To install manually: https://go.dev/dl/#go${REQUIRED_VERSION}"
    }

    GOSWITCH_BIN="${GOPATH_BIN}/go${REQUIRED_VERSION}"
    if [ -x "$GOSWITCH_BIN" ]; then
        # Download the toolchain if not already cached.
        echo -e "  Downloading toolchain (this may take a minute on first run)..."
        "$GOSWITCH_BIN" download 2>/dev/null || true
        GO_CMD="$GOSWITCH_BIN"
        echo ""
        echo -e "  ${GREEN}✓ Using go${REQUIRED_VERSION} for this setup run.${RESET}"
        echo ""
        echo -e "  ${BOLD}How to use go${REQUIRED_VERSION} in future terminal sessions:${RESET}"
        echo -e "  ┌─────────────────────────────────────────────────────────────┐"
        echo -e "  │ Option A — Add GOPATH/bin to PATH (recommended):            │"
        echo -e "  │   export PATH=\"\$(go env GOPATH)/bin:\$PATH\"                  │"
        echo -e "  │   go${REQUIRED_VERSION} run main.go                                    │"
        echo -e "  │                                                             │"
        echo -e "  │ Option B — Use the full path:                               │"
        echo -e "  │   ${GOSWITCH_BIN} run main.go              │"
        echo -e "  │                                                             │"
        echo -e "  │ NOTE: This does NOT change your system Go. Other projects   │"
        echo -e "  │ using a different Go version are not affected.              │"
        echo -e "  └─────────────────────────────────────────────────────────────┘"
        echo ""
    else
        echo -e "  ${YELLOW}  ⚠ Wrapper not found at ${GOSWITCH_BIN}.${RESET}"
        echo -e "    Continuing with system go${CURRENT_VERSION}."
    fi
fi

echo -e "  ${CYAN}Using: $($GO_CMD version)${RESET}"

# ---------------------------------------------------------------------------
# 2. Download / tidy modules (using the resolved GO_CMD)
# ---------------------------------------------------------------------------
echo ""
echo -e "${CYAN}[2/6] Downloading Go modules...${RESET}"
echo -e "  Running: $GO_CMD mod download"
"$GO_CMD" mod download
echo -e "  Running: $GO_CMD mod tidy"
"$GO_CMD" mod tidy
echo -e "  ${GREEN}✓ Modules up to date${RESET}"

# ---------------------------------------------------------------------------
# 3. Install swag CLI (Swagger code-generator)
# ---------------------------------------------------------------------------
echo ""
echo -e "${CYAN}[3/6] Installing swag CLI...${RESET}"
if command -v swag &>/dev/null; then
    SWAG_VER=$(swag --version 2>&1 | head -1)
    echo -e "  ${GREEN}✓ swag already installed: ${SWAG_VER}${RESET}"
else
    echo -e "  swag not found in PATH — installing..."
    "$GO_CMD" install github.com/swaggo/swag/cmd/swag@latest
    # Verify it's accessible
    if command -v swag &>/dev/null; then
        echo -e "  ${GREEN}✓ swag installed and available in PATH${RESET}"
    elif [ -x "${GOPATH_BIN}/swag" ]; then
        echo -e "  ${GREEN}✓ swag installed at ${GOPATH_BIN}/swag${RESET}"
        echo -e "  ${YELLOW}  ⚠ ${GOPATH_BIN} is not in your PATH.${RESET}"
        echo -e "    Add it with: export PATH=\"\$(go env GOPATH)/bin:\$PATH\""
    else
        echo -e "  ${YELLOW}  ⚠ swag installation may have failed — check output above.${RESET}"
    fi
fi

# ---------------------------------------------------------------------------
# 4. Generate Swagger documentation
# ---------------------------------------------------------------------------
echo ""
echo -e "${CYAN}[4/6] Generating Swagger documentation...${RESET}"
SWAG_BIN=""
if command -v swag &>/dev/null; then
    SWAG_BIN="swag"
elif [ -x "${GOPATH_BIN}/swag" ]; then
    SWAG_BIN="${GOPATH_BIN}/swag"
fi

if [ -n "$SWAG_BIN" ]; then
    echo -e "  Running: $SWAG_BIN init -g main.go -o docs"
    "$SWAG_BIN" init -g main.go -o docs
    echo -e "  ${GREEN}✓ docs/ updated${RESET}"
else
    echo -e "${YELLOW}  ⚠ swag not found in PATH or GOPATH/bin — skipping doc generation.${RESET}"
    echo -e "    Ensure \$(go env GOPATH)/bin is in your PATH and re-run setup.sh."
fi

# ---------------------------------------------------------------------------
# 5. Verify the project builds cleanly
# ---------------------------------------------------------------------------
echo ""
echo -e "${CYAN}[5/6] Building project...${RESET}"
echo -e "  Running: $GO_CMD build ./..."
"$GO_CMD" build ./...
echo -e "  ${GREEN}✓ Build successful${RESET}"

# ---------------------------------------------------------------------------
# 6. Run tests
# ---------------------------------------------------------------------------
echo ""
echo -e "${CYAN}[6/6] Running tests...${RESET}"
echo -e "  Running: $GO_CMD test ./..."
"$GO_CMD" test ./... && echo -e "  ${GREEN}✓ All tests passed${RESET}" || echo -e "${YELLOW}  ⚠ Some tests failed — check output above.${RESET}"

# ---------------------------------------------------------------------------
# Summary
# ---------------------------------------------------------------------------
echo ""
echo -e "${GREEN}${BOLD}┌──────────────────────────────────────────────────────────┐${RESET}"
echo -e "${GREEN}${BOLD}│                    ✓ Setup complete!                      │${RESET}"
echo -e "${GREEN}${BOLD}└──────────────────────────────────────────────────────────┘${RESET}"
echo ""
echo -e "  ${BOLD}What was done:${RESET}"
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
echo -e "   5. Run the server:  ${CYAN}\"${GO_CMD}\" run main.go${RESET}"
echo -e "   6. Open Swagger UI: http://localhost:8080/swagger"
echo ""
echo -e "  ${BOLD}Go version switching (does NOT affect other projects):${RESET}"
echo -e "   • This setup uses golang.org/dl to install a project-local Go wrapper."
echo -e "   • Your system Go is never modified."
echo -e "   • To use go${REQUIRED_VERSION} in any terminal:  ${CYAN}go${REQUIRED_VERSION} run main.go${RESET}"
echo -e "   • Or add GOPATH/bin to PATH:  ${CYAN}export PATH=\"\$(go env GOPATH)/bin:\$PATH\"${RESET}"
echo -e "   • See documentation/setting-up-goenv.md for more details."
echo ""
echo -e "   To build a release binary:  ${CYAN}./build/build.sh [GOOS] [GOARCH]${RESET}"
echo ""
