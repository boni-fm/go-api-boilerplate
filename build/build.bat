@echo off
:: build\build.bat — Cross-platform build script for the Go API Boilerplate.
::
:: Usage:
::   build\build.bat [GOOS] [GOARCH]
::
:: Examples:
::   build\build.bat                     :: build for current OS/arch
::   build\build.bat linux   amd64       :: Linux 64-bit binary
::   build\build.bat windows amd64       :: Windows 64-bit binary
::   build\build.bat darwin  arm64       :: macOS Apple Silicon binary
::
:: The output binary is written to .\bin\<AppName>[.exe].
:: AppName is read from appsettings.ini; if the file is missing the
:: directory name is used as a fallback.

setlocal enabledelayedexpansion

:: ---------------------------------------------------------------------------
:: Read AppName from appsettings.ini
:: ---------------------------------------------------------------------------
set "INI_FILE=appsettings.ini"
set "APP_NAME="

if exist "%INI_FILE%" (
    for /f "tokens=2 delims==" %%A in ('findstr /i "AppName" "%INI_FILE%"') do (
        set "APP_NAME=%%A"
        :: trim leading/trailing spaces
        set "APP_NAME=!APP_NAME: =!"
    )
)

if "!APP_NAME!"=="" (
    for %%I in (.) do set "APP_NAME=%%~nxI"
    echo [build] AppName not found in %INI_FILE% -- using directory name: !APP_NAME!
)

:: ---------------------------------------------------------------------------
:: Resolve target OS and architecture
:: ---------------------------------------------------------------------------
if "%~1"=="" (
    for /f %%A in ('go env GOOS') do set "TARGET_OS=%%A"
) else (
    set "TARGET_OS=%~1"
)

if "%~2"=="" (
    for /f %%A in ('go env GOARCH') do set "TARGET_ARCH=%%A"
) else (
    set "TARGET_ARCH=%~2"
)

:: Append .exe for Windows binaries
set "OUTPUT_EXT="
if /i "!TARGET_OS!"=="windows" set "OUTPUT_EXT=.exe"

:: ---------------------------------------------------------------------------
:: Build
:: ---------------------------------------------------------------------------
set "OUTPUT_DIR=bin"
set "OUTPUT_BIN=%OUTPUT_DIR%\!APP_NAME!!OUTPUT_EXT!"

if not exist "%OUTPUT_DIR%" mkdir "%OUTPUT_DIR%"

echo ------------------------------------------------------------
echo   App     : !APP_NAME!
echo   Target  : !TARGET_OS!/!TARGET_ARCH!
echo   Output  : !OUTPUT_BIN!
echo ------------------------------------------------------------

set GOOS=!TARGET_OS!
set GOARCH=!TARGET_ARCH!
go build -ldflags="-s -w" -o "!OUTPUT_BIN!" .

if errorlevel 1 (
    echo [build] FAILED
    exit /b 1
)

echo [build] Done -^> !OUTPUT_BIN!
endlocal
