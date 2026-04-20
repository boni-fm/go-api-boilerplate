@echo off

set "SDK_ROOT=D:\go-apps\go-sdks"
set "GOPATH_DIR=D:\go-apps\gopath"

:: Tanpa argumen = tampilkan versi tersedia
if "%~1"=="" (
    echo.
    echo   GOSWITCH - Go Version Switcher
    echo   ==============================
    echo.
    echo   Penggunaan: goswitch ^<versi^>
    echo   Contoh:     goswitch 1.24.6
    echo.
    echo   Versi tersedia:
    setlocal enabledelayedexpansion
    for /D %%d in ("%SDK_ROOT%\go*") do (
        set "folder=%%~nxd"
        set "ver=!folder:go=!"
        set "marker=  "
        if /i "%GOROOT%"=="%%d" set "marker=* "
        echo     !marker! !ver!
    )
    endlocal
    echo.
    echo   * = versi aktif saat ini
    echo.
    exit /b
)

:: Validasi versi
set "TARGET=%SDK_ROOT%\go%~1"

if not exist "%TARGET%\bin\go.exe" (
    echo.
    echo   [FAIL] Go %~1 tidak ditemukan di %TARGET%
    echo.
    echo   Versi tersedia:
    setlocal enabledelayedexpansion
    for /D %%d in ("%SDK_ROOT%\go*") do (
        set "folder=%%~nxd"
        echo     !folder:go=!
    )
    endlocal
    echo.
    exit /b 1
)

:: Set permanen (berlaku di terminal baru)
setx GOROOT "%TARGET%" >nul 2>&1
setx GOPATH "%GOPATH_DIR%" >nul 2>&1

:: Set untuk terminal saat ini (langsung aktif)
set "GOROOT=%TARGET%"
set "GOPATH=%GOPATH_DIR%"
set "PATH=%TARGET%\bin;%GOPATH_DIR%\bin;%PATH%"

echo.
echo   [PASS] Switched ke Go %~1
echo   ========================
"%TARGET%\bin\go.exe" version
echo.
echo   GOROOT = %GOROOT%
echo   GOPATH = %GOPATH%
echo.