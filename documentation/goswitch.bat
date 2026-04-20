@echo off
:: goswitch.bat — Switch Go version for the CURRENT terminal session only.
::
:: This script sets GOROOT and PATH for the current terminal session WITHOUT
:: modifying permanent (user/system) environment variables. This means each
:: terminal can use a different Go version, and switching in one terminal will
:: NOT interfere with other projects or other open terminals.
::
:: Configuration:
::   SDK_ROOT — set via the GOSWITCH_SDK_ROOT environment variable, or edit
::              the default below. The folder must contain sub-folders named
::              go<version> (e.g. go1.24.6, go1.25.8).
::
:: Usage:
::   goswitch              — list available versions
::   goswitch <version>    — switch to <version> in this terminal

setlocal enabledelayedexpansion

:: ---------------------------------------------------------------------------
:: SDK root: configurable via GOSWITCH_SDK_ROOT env var, falls back to default
:: ---------------------------------------------------------------------------
if defined GOSWITCH_SDK_ROOT (
    set "SDK_ROOT=!GOSWITCH_SDK_ROOT!"
) else (
    set "SDK_ROOT=D:\go-apps\go-sdks"
)

:: GOPATH: use the standard Go GOPATH if available, otherwise a sensible default
for /f "tokens=*" %%P in ('go env GOPATH 2^>nul') do set "GOPATH_DIR=%%P"
if not defined GOPATH_DIR set "GOPATH_DIR=%USERPROFILE%\go"

:: ---------------------------------------------------------------------------
:: No arguments → list available versions
:: ---------------------------------------------------------------------------
if "%~1"=="" (
    echo.
    echo   GOSWITCH - Go Version Switcher  [session-only]
    echo   ==============================================
    echo.
    echo   Usage:    goswitch ^<version^>
    echo   Example:  goswitch 1.24.6
    echo.
    echo   SDK root: !SDK_ROOT!
    echo   ^(Override with: set GOSWITCH_SDK_ROOT=^<path^>^)
    echo.
    echo   Installed versions:
    set "_found=0"
    for /D %%d in ("!SDK_ROOT!\go*") do (
        set "folder=%%~nxd"
        set "ver=!folder:go=!"
        set "marker=  "
        if /i "%GOROOT%"=="%%d" set "marker=* "
        echo     !marker! !ver!
        set "_found=1"
    )
    if "!_found!"=="0" (
        echo     ^(none found — run setup-goenv.bat first^)
    )
    echo.
    echo   * = active version in this terminal
    echo.
    echo   NOTE: goswitch only changes the CURRENT terminal session.
    echo         Other terminals and projects are NOT affected.
    echo.
    exit /b
)

:: ---------------------------------------------------------------------------
:: Validate requested version
:: ---------------------------------------------------------------------------
set "TARGET=!SDK_ROOT!\go%~1"

if not exist "!TARGET!\bin\go.exe" (
    echo.
    echo   [FAIL] Go %~1 not found at !TARGET!
    echo.
    echo   Installed versions:
    set "_found=0"
    for /D %%d in ("!SDK_ROOT!\go*") do (
        set "folder=%%~nxd"
        echo     !folder:go=!
        set "_found=1"
    )
    if "!_found!"=="0" echo     ^(none^)
    echo.
    echo   To install a new version, run setup-goenv.bat or download from https://go.dev/dl/
    echo.
    exit /b 1
)

:: ---------------------------------------------------------------------------
:: Switch — current session only (no setx, no permanent changes)
:: ---------------------------------------------------------------------------
endlocal & (
    set "GOROOT=%TARGET%"
    set "GOPATH=%GOPATH_DIR%"
    set "PATH=%TARGET%\bin;%GOPATH_DIR%\bin;%PATH%"
)

echo.
echo   [PASS] Switched to Go %~1  [this terminal only]
echo   ================================================
"%GOROOT%\bin\go.exe" version
echo.
echo   GOROOT = %GOROOT%
echo   GOPATH = %GOPATH%
echo.
echo   NOTE: This change applies to the CURRENT terminal only.
echo         Other terminals and projects are NOT affected.
echo         To make this the default, run: setup-goenv.bat
echo.