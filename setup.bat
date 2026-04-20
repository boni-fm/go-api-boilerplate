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
:: 1. Verify Go installation
:: ---------------------------------------------------------------------------
echo [1/5] Checking Go installation...
where go >nul 2>&1
if errorlevel 1 (
    echo   X Go is not installed. Download it from https://go.dev/dl/
    exit /b 1
)
for /f "tokens=3" %%V in ('go version') do set "GO_VERSION=%%V"
echo   OK Found !GO_VERSION!

:: ---------------------------------------------------------------------------
:: 2. Download / tidy modules
:: ---------------------------------------------------------------------------
echo.
echo [2/5] Downloading Go modules...
go mod download
if errorlevel 1 ( echo   FAILED & exit /b 1 )
go mod tidy
if errorlevel 1 ( echo   FAILED & exit /b 1 )
echo   OK Modules up to date

:: ---------------------------------------------------------------------------
:: 3. Install swag CLI
:: ---------------------------------------------------------------------------
echo.
echo [3/5] Installing swag CLI...
where swag >nul 2>&1
if errorlevel 1 (
    go install github.com/swaggo/swag/cmd/swag@latest
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
echo [4/5] Generating Swagger documentation...
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
echo [5/5] Building project...
go build ./...
if errorlevel 1 ( echo   Build FAILED & exit /b 1 )
echo   OK Build successful

:: ---------------------------------------------------------------------------
:: Summary
:: ---------------------------------------------------------------------------
echo.
echo Setup complete!
echo.
echo   What changed / was verified:
echo    - Go modules downloaded and tidied
echo    - swag CLI installed (Swagger doc generator)
echo    - Swagger docs generated in docs\
echo    - Project compiled successfully
echo.
echo   Next steps:
echo    1. Edit appsettings.ini with your database credentials (Kunci).
echo    2. Set IsDevelopment = true for local development (enables Swagger UI).
echo    3. Set Timezone to your region (e.g. Asia/Jakarta).
echo    4. Run the server:  go run main.go
echo    5. Open Swagger UI: http://localhost:8080/swagger
echo.
echo   To build a release binary:  build\build.bat [GOOS] [GOARCH]
echo.
endlocal
