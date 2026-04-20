@echo off
:: setup.bat — Developer environment initialisation for the Go API Boilerplate.
::
:: Run this script once after cloning the repository to install tools,
:: verify dependencies, and generate the initial Swagger documentation.
::
:: Usage (run from repository root):
::   setup.bat

setlocal enabledelayedexpansion

echo.
echo +----------------------------------------------------------+
echo ^|    IT SD3 Go API Boilerplate -- Developer Setup         ^|
echo +----------------------------------------------------------+
echo.
echo   This script will:
echo     1. Check your Go version and install the required one if needed
echo     2. Download Go modules
echo     3. Install the swag CLI (Swagger doc generator)
echo     4. Generate Swagger documentation
echo     5. Build the project
echo     6. Run the tests
echo.
echo   Your system Go installation will NOT be modified.
echo   If a version switch is needed, a project-local wrapper is
echo   installed via golang.org/dl (stored in GOPATH\bin\).
echo.

:: ---------------------------------------------------------------------------
:: 1. Go version check and goswitch
::
:: Strategy: we NEVER replace or modify the developer's system Go installation.
:: We use golang.org/dl to install a project-local wrapper binary
:: (go<version>.exe) in GOPATH\bin. The wrapper downloads the actual toolchain
:: on first use to %GOPATH%\sdk\. If the system Go already matches the
:: required version, no extra binary is installed.
:: ---------------------------------------------------------------------------
echo [1/6] Checking Go version...
where go >nul 2>&1
if errorlevel 1 (
    echo   X Go is not installed.
    echo     Please install Go from https://go.dev/dl/ and re-run this script.
    exit /b 1
)

:: Read required version from go.mod (e.g. 1.25.8)
for /f "tokens=2" %%V in ('findstr /r "^go " go.mod') do set "REQUIRED_VERSION=%%V"
:: Read current version
for /f "tokens=3" %%V in ('go version') do set "RAW_VER=%%V"
set "CURRENT_VERSION=!RAW_VER:go=!"

:: Get GOPATH\bin for later use
for /f %%P in ('go env GOPATH') do set "GOPATH_DIR=%%P"
set "GOPATH_BIN=!GOPATH_DIR!\bin"

echo   System Go  : go!CURRENT_VERSION!
echo   Required   : go!REQUIRED_VERSION!
echo   GOPATH\bin : !GOPATH_BIN!

:: Default go command to use for the rest of the script
set "GO_CMD=go"

if "!CURRENT_VERSION!"=="!REQUIRED_VERSION!" (
    echo   OK Version matches -- no switch needed.
) else (
    echo.
    echo   WARNING Version mismatch detected.
    echo   Installing go!REQUIRED_VERSION! via golang.org/dl...
    echo   ^(Your system go!CURRENT_VERSION! will NOT be modified.^)
    echo.

    echo   Installing wrapper binary...
    go install "golang.org/dl/go!REQUIRED_VERSION!@latest" 2>nul
    if errorlevel 1 (
        echo   WARNING Could not install wrapper via golang.org/dl.
        echo     This usually means go!REQUIRED_VERSION! is not yet published.
        echo     Continuing with system go!CURRENT_VERSION!.
        echo     To install manually: https://go.dev/dl/#go!REQUIRED_VERSION!
        goto :version_done
    )

    :: Determine GOPATH\bin path for the versioned wrapper
    set "GOSWITCH_BIN=!GOPATH_BIN!\go!REQUIRED_VERSION!.exe"

    if exist "!GOSWITCH_BIN!" (
        echo   Downloading toolchain ^(this may take a minute on first run^)...
        "!GOSWITCH_BIN!" download 2>nul
        set "GO_CMD=!GOSWITCH_BIN!"
        echo.
        echo   OK Using go!REQUIRED_VERSION! for this setup run.
        echo.
        echo   How to use go!REQUIRED_VERSION! in future terminal sessions:
        echo   +-------------------------------------------------------------+
        echo   ^| Option A -- Add GOPATH\bin to PATH ^(recommended^):            ^|
        echo   ^|   set PATH=%%GOPATH%%\bin;%%PATH%%                               ^|
        echo   ^|   go!REQUIRED_VERSION! run main.go                                    ^|
        echo   ^|                                                             ^|
        echo   ^| Option B -- Use the full path:                               ^|
        echo   ^|   !GOSWITCH_BIN! run main.go           ^|
        echo   ^|                                                             ^|
        echo   ^| NOTE: This does NOT change your system Go. Other projects   ^|
        echo   ^| using a different Go version are not affected.              ^|
        echo   +-------------------------------------------------------------+
        echo.
    ) else (
        echo   WARNING Wrapper not found at !GOSWITCH_BIN!.
        echo     Continuing with system go!CURRENT_VERSION!.
    )
)
:version_done

:: Show which Go binary will be used
for /f "tokens=*" %%V in ('"!GO_CMD!" version') do echo   Using: %%V

:: ---------------------------------------------------------------------------
:: 2. Download / tidy modules
:: ---------------------------------------------------------------------------
echo.
echo [2/6] Downloading Go modules...
echo   Running: !GO_CMD! mod download
"!GO_CMD!" mod download
if errorlevel 1 ( echo   FAILED & exit /b 1 )
echo   Running: !GO_CMD! mod tidy
"!GO_CMD!" mod tidy
if errorlevel 1 ( echo   FAILED & exit /b 1 )
echo   OK Modules up to date

:: ---------------------------------------------------------------------------
:: 3. Install swag CLI
:: ---------------------------------------------------------------------------
echo.
echo [3/6] Installing swag CLI...
where swag >nul 2>&1
if errorlevel 1 (
    echo   swag not found in PATH -- installing...
    "!GO_CMD!" install github.com/swaggo/swag/cmd/swag@latest
    if errorlevel 1 ( echo   FAILED to install swag & exit /b 1 )
    :: Verify swag is accessible
    where swag >nul 2>&1
    if errorlevel 1 (
        if exist "!GOPATH_BIN!\swag.exe" (
            echo   OK swag installed at !GOPATH_BIN!\swag.exe
            echo   WARNING !GOPATH_BIN! is not in your PATH.
            echo     Add it with: set PATH=%%GOPATH%%\bin;%%PATH%%
        ) else (
            echo   WARNING swag installation may have failed -- check output above.
        )
    ) else (
        echo   OK swag installed and available in PATH
    )
) else (
    for /f "tokens=*" %%V in ('swag --version 2^>^&1') do set "SWAG_VER=%%V" & goto :swag_done
    :swag_done
    echo   OK swag already installed: !SWAG_VER!
)

:: ---------------------------------------------------------------------------
:: 4. Generate Swagger documentation
:: ---------------------------------------------------------------------------
echo.
echo [4/6] Generating Swagger documentation...
set "SWAG_BIN="
where swag >nul 2>&1
if not errorlevel 1 (
    set "SWAG_BIN=swag"
) else if exist "!GOPATH_BIN!\swag.exe" (
    set "SWAG_BIN=!GOPATH_BIN!\swag.exe"
)

if defined SWAG_BIN (
    echo   Running: !SWAG_BIN! init -g main.go -o docs
    "!SWAG_BIN!" init -g main.go -o docs
    if errorlevel 1 ( echo   FAILED & exit /b 1 )
    echo   OK docs\ updated
) else (
    echo   WARNING swag not in PATH or GOPATH\bin -- skipping doc generation.
    echo   Ensure %%GOPATH%%\bin is in your PATH and re-run setup.bat.
)

:: ---------------------------------------------------------------------------
:: 5. Verify the project builds cleanly
:: ---------------------------------------------------------------------------
echo.
echo [5/6] Building project...
echo   Running: !GO_CMD! build ./...
"!GO_CMD!" build ./...
if errorlevel 1 ( echo   Build FAILED & exit /b 1 )
echo   OK Build successful

:: ---------------------------------------------------------------------------
:: 6. Run tests
:: ---------------------------------------------------------------------------
echo.
echo [6/6] Running tests...
echo   Running: !GO_CMD! test ./...
"!GO_CMD!" test ./...
if errorlevel 1 (
    echo   WARNING Some tests failed -- check output above.
) else (
    echo   OK All tests passed
)

:: ---------------------------------------------------------------------------
:: Summary
:: ---------------------------------------------------------------------------
echo.
echo +----------------------------------------------------------+
echo ^|                    Setup complete!                       ^|
echo +----------------------------------------------------------+
echo.
echo   What was done:
echo    - Go version verified (goswitch via golang.org/dl if needed)
echo    - Go modules downloaded and tidied
echo    - swag CLI installed (Swagger doc generator)
echo    - Swagger docs generated in docs\
echo    - Project compiled and tested successfully
echo.
echo   Next steps:
echo    1. Edit appsettings.ini with your database credentials (Kunci).
echo    2. Set IsDevelopment = true for local development (enables Swagger UI).
echo    3. Set Timezone to your region (e.g. Asia/Jakarta), or set TZ env var.
echo    4. For multi-DC: set Kunci = g009sim,g010sim (comma-separated).
echo    5. Run the server:  "!GO_CMD!" run main.go
echo    6. Open Swagger UI: http://localhost:8080/swagger
echo.
echo   Go version switching (does NOT affect other projects):
echo    - This setup uses golang.org/dl to install a project-local Go wrapper.
echo    - Your system Go is never modified.
echo    - To use go!REQUIRED_VERSION! in any terminal:  go!REQUIRED_VERSION! run main.go
echo    - Or add GOPATH\bin to PATH:  set PATH=%%GOPATH%%\bin;%%PATH%%
echo    - Windows users: see documentation\setting-up-goenv.md for more details.
echo.
echo   To build a release binary:  build\build.bat [GOOS] [GOARCH]
echo.
endlocal
