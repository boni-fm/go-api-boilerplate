@echo off
title Setup Go Environment ITSD3/SD3

:: 1. Auto-Elevate ke Administrator
net session >nul 2>&1
if %errorLevel% NEQ 0 (
    echo [INFO] Requesting Administrator access...
    powershell -Command "Start-Process -FilePath \"%~f0\" -Verb RunAs"
    exit /b
)

:: 2. Header
echo.
echo -----------------------------------------------------
echo   GO ENVIRONMENT SETUP - ITSD3/SD3
echo -----------------------------------------------------
echo.
echo   This script will:
echo     1. Create SDK folders at the configured root
echo     2. Download Go versions from https://go.dev/dl/
echo     3. Set default environment variables (GOROOT, GOPATH, PATH)
echo.
echo   WARNING: Step 3 permanently modifies your user environment
echo   variables. If you only need a session-specific version, use
echo   goswitch.bat instead (session-only, no permanent changes).
echo.
echo   Press Ctrl+C to cancel, or
pause

:: 3. Ekstrak baris #PS# ke temp file lalu jalankan
set "TEMP_PS=%temp%\setup_go_itsd3_%RANDOM%.ps1"

:: Gunakan findstr untuk ekstrak baris #PS# (lebih reliable daripada PowerShell parsing)
:: Lalu pakai PowerShell hanya untuk strip prefix
findstr /B "#PS#" "%~f0" > "%TEMP_PS%.raw"
powershell -NoProfile -Command "(Get-Content -LiteralPath '%TEMP_PS%.raw') -replace '^#PS# ?','' | Set-Content -LiteralPath '%TEMP_PS%' -Encoding UTF8"
del "%TEMP_PS%.raw"

:: Jalankan script yang sudah bersih
powershell -NoProfile -ExecutionPolicy Bypass -File "%TEMP_PS%"

:: Bersihkan
del "%TEMP_PS%"

echo.
echo Process complete. Review the log above for any errors.
pause
exit /b

:: =====================================================================
:: POWERSHELL CODE BLOCK (Do not remove the #PS# prefix)
:: =====================================================================
#PS# # --- Configuration ---
#PS# # Override these variables to install to a different location.
#PS# # By default SDKs go to D:\go-apps\go-sdks and GOPATH to D:\go-apps\gopath.
#PS# # You can also set the GOSWITCH_SDK_ROOT env var before running this script.
#PS# $envRoot = $env:GOSWITCH_SDK_ROOT
#PS# if ($envRoot) {
#PS#     $targetRoot = Split-Path $envRoot -Parent
#PS#     $goRootBase = $envRoot
#PS# } else {
#PS#     $targetRoot = 'D:\go-apps'
#PS#     $goRootBase = "$targetRoot\go-sdks"
#PS# }
#PS# $goVersions   = @('1.24.6', '1.25.8')
#PS# $globalVer    = '1.25.8'
#PS# $success      = $true
#PS#
#PS# Write-Host ''
#PS# Write-Host '--- Step 1: Preparing folders ---' -ForegroundColor Cyan
#PS# Write-Host "    SDK root : $goRootBase"       -ForegroundColor Gray
#PS# Write-Host "    GOPATH   : $targetRoot\gopath" -ForegroundColor Gray
#PS# Write-Host ''
#PS#
#PS# if (!(Test-Path $targetRoot)) {
#PS#     try { New-Item -Path $targetRoot -ItemType Directory -Force | Out-Null }
#PS#     catch {
#PS#         Write-Host "[FAIL] Cannot create folder $targetRoot" -ForegroundColor Red
#PS#         $success = $false
#PS#     }
#PS# }
#PS# if ($success -and !(Test-Path $goRootBase)) {
#PS#     New-Item -Path $goRootBase -ItemType Directory -Force | Out-Null
#PS# }
#PS#
#PS# if ($success) {
#PS#     Write-Host '--- Step 2: Download & Install Go SDKs ---' -ForegroundColor Cyan
#PS#     foreach ($ver in $goVersions) {
#PS#         $installDir = "$goRootBase\go$ver"
#PS#         if (Test-Path "$installDir\bin\go.exe") {
#PS#             Write-Host "[SKIP] Go $ver already installed at $installDir" -ForegroundColor Green
#PS#             continue
#PS#         }
#PS#
#PS#         $zipName = "go${ver}.windows-amd64.zip"
#PS#         $url     = "https://go.dev/dl/$zipName"
#PS#         $zipPath = "$env:TEMP\$zipName"
#PS#
#PS#         Write-Host "[DOWN] Downloading Go $ver from $url ..." -ForegroundColor Yellow
#PS#         try {
#PS#             [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12
#PS#             Invoke-WebRequest -Uri $url -OutFile $zipPath -UseBasicParsing
#PS#         } catch {
#PS#             Write-Host "[FAIL] Could not download Go $ver : $_" -ForegroundColor Red
#PS#             Write-Host "[HINT] Make sure version $ver is available at https://go.dev/dl/" -ForegroundColor Yellow
#PS#             $success = $false
#PS#             continue
#PS#         }
#PS#
#PS#         $fileSize = (Get-Item $zipPath).Length
#PS#         if ($fileSize -lt 1MB) {
#PS#             Write-Host "[FAIL] Downloaded file too small ($fileSize bytes). The version may not exist." -ForegroundColor Red
#PS#             Remove-Item $zipPath -Force
#PS#             $success = $false
#PS#             continue
#PS#         }
#PS#
#PS#         Write-Host "[EXTR] Extracting to $installDir ..." -ForegroundColor Yellow
#PS#         $tempExtract = "$env:TEMP\go_extract_$ver"
#PS#         if (Test-Path $tempExtract) { Remove-Item $tempExtract -Recurse -Force }
#PS#         try {
#PS#             Expand-Archive -Path $zipPath -DestinationPath $tempExtract -Force
#PS#         } catch {
#PS#             Write-Host "[FAIL] Extraction failed: $_" -ForegroundColor Red
#PS#             Remove-Item $zipPath -Force
#PS#             $success = $false
#PS#             continue
#PS#         }
#PS#
#PS#         Move-Item -Path "$tempExtract\go" -Destination $installDir -Force
#PS#         Remove-Item $tempExtract -Recurse -Force
#PS#         Remove-Item $zipPath -Force
#PS#
#PS#         if (Test-Path "$installDir\bin\go.exe") {
#PS#             Write-Host "[PASS] Go $ver installed successfully." -ForegroundColor Green
#PS#         } else {
#PS#             Write-Host "[FAIL] Go $ver installation failed." -ForegroundColor Red
#PS#             $success = $false
#PS#         }
#PS#     }
#PS# }
#PS#
#PS# if ($success) {
#PS#     Write-Host ''
#PS#     Write-Host "--- Step 3: Set Go $globalVer as default (permanent) ---" -ForegroundColor Cyan
#PS#     Write-Host '    NOTE: This modifies your user-level GOROOT, GOPATH, and PATH.' -ForegroundColor Yellow
#PS#     Write-Host '    To switch versions per-session without permanent changes,' -ForegroundColor Yellow
#PS#     Write-Host '    use goswitch.bat instead.'                                    -ForegroundColor Yellow
#PS#     Write-Host ''
#PS#     $activeGoRoot = "$goRootBase\go$globalVer"
#PS#     $activeGoBin  = "$activeGoRoot\bin"
#PS#     $goPath       = "$targetRoot\gopath"
#PS#
#PS#     if (!(Test-Path $goPath)) { New-Item -Path $goPath -ItemType Directory -Force | Out-Null }
#PS#
#PS#     [Environment]::SetEnvironmentVariable('GOROOT', $activeGoRoot, 'User')
#PS#     [Environment]::SetEnvironmentVariable('GOPATH', $goPath, 'User')
#PS#     [Environment]::SetEnvironmentVariable('GOSWITCH_SDK_ROOT', $goRootBase, 'User')
#PS#
#PS#     $userPath = [Environment]::GetEnvironmentVariable('Path', 'User')
#PS#     $cleanParts = ($userPath -split ';') | Where-Object {
#PS#         $_ -ne '' -and $_ -notlike "*$goRootBase*" -and $_ -notlike '*\Go\bin*'
#PS#     }
#PS#     $newParts = @($activeGoBin, "$goPath\bin") + $cleanParts
#PS#     $newPath  = ($newParts | Select-Object -Unique) -join ';'
#PS#     [Environment]::SetEnvironmentVariable('Path', $newPath, 'User')
#PS#
#PS#     Write-Host "    GOROOT           = $activeGoRoot" -ForegroundColor Green
#PS#     Write-Host "    GOPATH           = $goPath" -ForegroundColor Green
#PS#     Write-Host "    GOSWITCH_SDK_ROOT = $goRootBase" -ForegroundColor Green
#PS#     Write-Host "    PATH             = Updated" -ForegroundColor Green
#PS# }
#PS#
#PS# if ($success) {
#PS#     Write-Host ''
#PS#     Write-Host '--- Step 4: Verification ---' -ForegroundColor Cyan
#PS#     foreach ($ver in $goVersions) {
#PS#         $exePath = "$goRootBase\go$ver\bin\go.exe"
#PS#         if (Test-Path $exePath) {
#PS#             $out = & $exePath version
#PS#             Write-Host "[PASS] $out" -ForegroundColor Green
#PS#         } else {
#PS#             Write-Host "[FAIL] go$ver not found!" -ForegroundColor Red
#PS#             $success = $false
#PS#         }
#PS#     }
#PS# }
#PS#
#PS# Write-Host ''
#PS# if ($success) {
#PS#     Write-Host '---------------------------------------------------' -ForegroundColor Green
#PS#     Write-Host 'SETUP COMPLETE! Environment is ready.'               -ForegroundColor White
#PS#     Write-Host "Default: Go $globalVer"                              -ForegroundColor White
#PS#     Write-Host ''
#PS#     Write-Host 'Folder structure:'                                   -ForegroundColor Magenta
#PS#     Write-Host "  $goRootBase\go1.24.6\"                             -ForegroundColor Yellow
#PS#     Write-Host "  $goRootBase\go1.25.8\  (default)"                  -ForegroundColor Yellow
#PS#     Write-Host "  $targetRoot\gopath\"                               -ForegroundColor Yellow
#PS#     Write-Host ''
#PS#     Write-Host 'To switch version (current terminal only):'          -ForegroundColor Magenta
#PS#     Write-Host '  goswitch 1.24.6'                                   -ForegroundColor Yellow
#PS#     Write-Host '  goswitch 1.25.8'                                   -ForegroundColor Yellow
#PS#     Write-Host ''
#PS#     Write-Host 'goswitch is SESSION-ONLY — it will NOT affect'       -ForegroundColor Cyan
#PS#     Write-Host 'other terminals or other projects.'                  -ForegroundColor Cyan
#PS#     Write-Host ''
#PS#     Write-Host 'Then RESTART your terminal or VS Code.'              -ForegroundColor Magenta
#PS#     Write-Host '---------------------------------------------------' -ForegroundColor Green
#PS# } else {
#PS#     Write-Host '---------------------------------------------------' -ForegroundColor Red
#PS#     Write-Host 'ERRORS DETECTED! Review the log above.'              -ForegroundColor Red
#PS#     Write-Host '---------------------------------------------------' -ForegroundColor Red
#PS# }