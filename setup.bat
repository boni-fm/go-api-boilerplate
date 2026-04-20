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
    echo   X Go is not installed. Download it from https://go.dev/dl/
    exit /b 1
)

:: Read required version from go.mod (e.g. 1.25.8)
for /f "tokens=2" %%V in ('findstr /r "^go " go.mod') do set "REQUIRED_VERSION=%%V"
:: Read current version
for /f "tokens=3" %%V in ('go version') do set "RAW_VER=%%V"
set "CURRENT_VERSION=!RAW_VER:go=!"

echo   System Go : go!CURRENT_VERSION!
echo   Required  : go!REQUIRED_VERSION!

:: Default go command to use for the rest of the script
set "GO_CMD=go"

if "!CURRENT_VERSION!"=="!REQUIRED_VERSION!" (
    echo   OK Version matches -- no switch needed.
) else (
    echo   WARNING Version mismatch. Installing go!REQUIRED_VERSION! via golang.org/dl...
    echo          ^(Your system go!CURRENT_VERSION! will NOT be modified.^)

    go install "golang.org/dl/go!REQUIRED_VERSION!@latest" 2>nul
    if errorlevel 1 (
        echo   WARNING Could not install wrapper. Continuing with system go.
        echo   Manual install: https://go.dev/dl/#go!REQUIRED_VERSION!
        goto :version_done
    )

    :: Determine GOPATH\bin path for the versioned wrapper
    for /f %%P in ('go env GOPATH') do set "GOPATH_DIR=%%P"
    set "GOSWITCH_BIN=!GOPATH_DIR!\bin\go!REQUIRED_VERSION!.exe"

    if exist "!GOSWITCH_BIN!" (
        "!GOSWITCH_BIN!" download 2>nul
        set "GO_CMD=!GOSWITCH_BIN!"
        echo   OK Using go!REQUIRED_VERSION! for this setup run.
        echo      Tip: add %%GOPATH%%\bin to your PATH and run:
        echo           go!REQUIRED_VERSION! run main.go
    ) else (
        echo   WARNING Wrapper not found. Continuing with system go.
    )
)
:version_done

:: ---------------------------------------------------------------------------
:: 2. Download / tidy modules
:: ---------------------------------------------------------------------------
echo.
echo [2/6] Downloading Go modules...
"!GO_CMD!" mod download
if errorlevel 1 ( echo   FAILED & exit /b 1 )
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
    "!GO_CMD!" install github.com/swaggo/swag/cmd/swag@latest
    if errorlevel 1 ( echo   FAILED to install swag & exit /b 1 )
    echo   OK swag installed
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
where swag >nul 2>&1
if errorlevel 1 (
    echo   WARNING swag not in PATH -- skipping doc generation.
    echo   Ensure %%GOPATH%%\bin is in your PATH and re-run setup.bat.
) else (
    swag init -g main.go -o docs
    if errorlevel 1 ( echo   FAILED & exit /b 1 )
    echo   OK docs\ updated
)

:: ---------------------------------------------------------------------------
:: 5. Verify the project builds cleanly
:: ---------------------------------------------------------------------------
echo.
echo [5/6] Building project...
"!GO_CMD!" build ./...
if errorlevel 1 ( echo   Build FAILED & exit /b 1 )
echo   OK Build successful

:: ---------------------------------------------------------------------------
:: 6. Run tests
:: ---------------------------------------------------------------------------
echo.
echo [6/6] Running tests...
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
echo Setup complete!
echo.
echo   What changed / was verified:
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
echo    5. Run the server:  !GO_CMD! run main.go
echo    6. Open Swagger UI: http://localhost:8080/swagger
echo.
echo   To build a release binary:  build\build.bat [GOOS] [GOARCH]
echo.
endlocal
