@echo off
title Setup Go Environment ITSD3/SD3

:: 1. Auto-Elevate ke Administrator
net session >nul 2>&1
if %errorLevel% NEQ 0 (
    echo [INFO] Meminta akses Administrator...
    powershell -Command "Start-Process -FilePath \"%~f0\" -Verb RunAs"
    exit /b
)

:: 2. Header
echo -----------------------------------------------------
echo STANDARISASI GO ENVIRONMENT - DEPARTEMEN ITSD3/SD3
echo -----------------------------------------------------

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
echo Proses selesai. Silakan baca log di atas.
pause
exit /b

:: =====================================================================
:: BLOK KODE POWERSHELL (Jangan hapus prefix #PS#)
:: =====================================================================
#PS# # --- Konfigurasi ---
#PS# $targetRoot   = 'D:\go-apps'
#PS# $goVersions   = @('1.24.6', '1.25.8')
#PS# $globalVer    = '1.25.8'
#PS# $goRootBase   = "$targetRoot\go-sdks"
#PS# $success      = $true
#PS#
#PS# Write-Host '--- Tahap 1: Persiapan Folder ---' -ForegroundColor Cyan
#PS#
#PS# if (!(Test-Path $targetRoot)) {
#PS#     try { New-Item -Path $targetRoot -ItemType Directory -Force | Out-Null }
#PS#     catch {
#PS#         Write-Host "[FAIL] Tidak bisa membuat folder $targetRoot" -ForegroundColor Red
#PS#         $success = $false
#PS#     }
#PS# }
#PS# if ($success -and !(Test-Path $goRootBase)) {
#PS#     New-Item -Path $goRootBase -ItemType Directory -Force | Out-Null
#PS# }
#PS#
#PS# if ($success) {
#PS#     Write-Host '--- Tahap 2: Download & Install Go SDKs ---' -ForegroundColor Cyan
#PS#     foreach ($ver in $goVersions) {
#PS#         $installDir = "$goRootBase\go$ver"
#PS#         if (Test-Path "$installDir\bin\go.exe") {
#PS#             Write-Host "[SKIP] Go $ver sudah terinstall di $installDir" -ForegroundColor Green
#PS#             continue
#PS#         }
#PS#
#PS#         $zipName = "go${ver}.windows-amd64.zip"
#PS#         $url     = "https://go.dev/dl/$zipName"
#PS#         $zipPath = "$env:TEMP\$zipName"
#PS#
#PS#         Write-Host "[DOWN] Mengunduh Go $ver dari $url ..." -ForegroundColor Yellow
#PS#         try {
#PS#             [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12
#PS#             Invoke-WebRequest -Uri $url -OutFile $zipPath -UseBasicParsing
#PS#         } catch {
#PS#             Write-Host "[FAIL] Gagal mengunduh Go $ver : $_" -ForegroundColor Red
#PS#             Write-Host "[HINT] Pastikan versi $ver tersedia di https://go.dev/dl/" -ForegroundColor Yellow
#PS#             $success = $false
#PS#             continue
#PS#         }
#PS#
#PS#         $fileSize = (Get-Item $zipPath).Length
#PS#         if ($fileSize -lt 1MB) {
#PS#             Write-Host "[FAIL] File download terlalu kecil ($fileSize bytes). Kemungkinan versi tidak valid." -ForegroundColor Red
#PS#             Remove-Item $zipPath -Force
#PS#             $success = $false
#PS#             continue
#PS#         }
#PS#
#PS#         Write-Host "[EXTR] Mengekstrak ke $installDir ..." -ForegroundColor Yellow
#PS#         $tempExtract = "$env:TEMP\go_extract_$ver"
#PS#         if (Test-Path $tempExtract) { Remove-Item $tempExtract -Recurse -Force }
#PS#         try {
#PS#             Expand-Archive -Path $zipPath -DestinationPath $tempExtract -Force
#PS#         } catch {
#PS#             Write-Host "[FAIL] Gagal mengekstrak: $_" -ForegroundColor Red
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
#PS#             Write-Host "[PASS] Go $ver berhasil diinstall." -ForegroundColor Green
#PS#         } else {
#PS#             Write-Host "[FAIL] Go $ver gagal diinstall." -ForegroundColor Red
#PS#             $success = $false
#PS#         }
#PS#     }
#PS# }
#PS#
#PS# if ($success) {
#PS#     Write-Host "--- Tahap 3: Set Go $globalVer sebagai Default ---" -ForegroundColor Cyan
#PS#     $activeGoRoot = "$goRootBase\go$globalVer"
#PS#     $activeGoBin  = "$activeGoRoot\bin"
#PS#     $goPath       = "$targetRoot\gopath"
#PS#
#PS#     if (!(Test-Path $goPath)) { New-Item -Path $goPath -ItemType Directory -Force | Out-Null }
#PS#
#PS#     [Environment]::SetEnvironmentVariable('GOROOT', $activeGoRoot, 'User')
#PS#     [Environment]::SetEnvironmentVariable('GOPATH', $goPath, 'User')
#PS#
#PS#     $userPath = [Environment]::GetEnvironmentVariable('Path', 'User')
#PS#     $cleanParts = ($userPath -split ';') | Where-Object {
#PS#         $_ -ne '' -and $_ -notlike "*$goRootBase*" -and $_ -notlike '*\Go\bin*'
#PS#     }
#PS#     $newParts = @($activeGoBin, "$goPath\bin") + $cleanParts
#PS#     $newPath  = ($newParts | Select-Object -Unique) -join ';'
#PS#     [Environment]::SetEnvironmentVariable('Path', $newPath, 'User')
#PS#
#PS#     Write-Host "[PASS] GOROOT = $activeGoRoot" -ForegroundColor Green
#PS#     Write-Host "[PASS] GOPATH = $goPath" -ForegroundColor Green
#PS#     Write-Host "[PASS] PATH   = Updated" -ForegroundColor Green
#PS# }
#PS#
#PS# if ($success) {
#PS#     Write-Host '--- Tahap 4: VERIFIKASI AKHIR ---' -ForegroundColor Cyan
#PS#     foreach ($ver in $goVersions) {
#PS#         $exePath = "$goRootBase\go$ver\bin\go.exe"
#PS#         if (Test-Path $exePath) {
#PS#             $out = & $exePath version
#PS#             Write-Host "[PASS] $out" -ForegroundColor Green
#PS#         } else {
#PS#             Write-Host "[FAIL] go$ver tidak ditemukan!" -ForegroundColor Red
#PS#             $success = $false
#PS#         }
#PS#     }
#PS# }
#PS#
#PS# Write-Host ''
#PS# if ($success) {
#PS#     Write-Host '---------------------------------------------------' -ForegroundColor Green
#PS#     Write-Host 'VERIFIKASI BERHASIL! Environment siap digunakan.'    -ForegroundColor White
#PS#     Write-Host "Default: Go $globalVer"                              -ForegroundColor White
#PS#     Write-Host ''
#PS#     Write-Host 'Struktur folder:'                                    -ForegroundColor Magenta
#PS#     Write-Host "  D:\go-apps\go-sdks\go1.24.6\"                     -ForegroundColor Yellow
#PS#     Write-Host "  D:\go-apps\go-sdks\go1.25.8\  (default)"          -ForegroundColor Yellow
#PS#     Write-Host "  D:\go-apps\gopath\"                                -ForegroundColor Yellow
#PS#     Write-Host ''
#PS#     Write-Host 'Untuk switch versi:'                                 -ForegroundColor Magenta
#PS#     foreach ($v in $goVersions) {
#PS#         Write-Host "  set GOROOT=$goRootBase\go$v"                   -ForegroundColor Yellow
#PS#     }
#PS#     Write-Host ''
#PS#     Write-Host 'Lalu RESTART terminal atau VS Code Anda.'            -ForegroundColor Magenta
#PS#     Write-Host '---------------------------------------------------' -ForegroundColor Green
#PS# } else {
#PS#     Write-Host '---------------------------------------------------' -ForegroundColor Red
#PS#     Write-Host 'ADA ERROR! Periksa log di atas.'                     -ForegroundColor Red
#PS#     Write-Host '---------------------------------------------------' -ForegroundColor Red
#PS# }