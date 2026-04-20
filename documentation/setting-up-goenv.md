# Setup Go Environment — ITSD3/SD3

Guide for standardising the Go development environment across the ITSD3/SD3 department.

---

## Table of Contents

- [Overview](#overview)
- [Requirements](#requirements)
- [Quick Start (Recommended)](#quick-start-recommended)
- [Windows Multi-SDK Setup](#windows-multi-sdk-setup)
  - [Installation](#installation)
  - [Folder Structure](#folder-structure)
  - [Verification](#verification)
- [Switching Go Versions](#switching-go-versions)
  - [Method 1 — goswitch (Windows, session-only)](#method-1--goswitch-windows-session-only)
  - [Method 2 — golang.org/dl (cross-platform, project-local)](#method-2--golangorgdl-cross-platform-project-local)
- [Why Session-Only?](#why-session-only)
- [VS Code Configuration](#vs-code-configuration)
- [Troubleshooting](#troubleshooting)
- [FAQ](#faq)

---

## Overview

This project requires a specific Go version (declared in `go.mod`). Two mechanisms are available to manage Go versions **without interfering with other projects**:

| Approach | Platform | Scope | How it works |
|----------|----------|-------|--------------|
| **`setup.sh` / `setup.bat`** | All | Project-local | Uses `golang.org/dl` to install a versioned wrapper binary (e.g. `go1.25.8`) alongside the system Go. No permanent env changes. |
| **`goswitch.bat`** | Windows | **Current terminal only** | Switches `GOROOT` for the current terminal session. Does **not** use `setx` — other terminals and projects are not affected. |

> **Key principle:** Your system Go is never modified. Each terminal session can use a different Go version independently.

---

## Requirements

| Requirement | Details |
|-------------|---------|
| OS | Windows 10/11 (64-bit), macOS, or Linux |
| Disk | At least 2 GB free space |
| Internet | Required for initial setup |
| Go | Any version installed (the setup scripts handle version switching) |

---

## Quick Start (Recommended)

The easiest way to get started is to run the project's setup script. It handles everything automatically:

```bash
# Linux / macOS
chmod +x setup.sh && ./setup.sh

# Windows
setup.bat
```

The setup script will:
1. ✅ Check your Go version against `go.mod`
2. ✅ Install the correct Go version via `golang.org/dl` if needed (without touching your system Go)
3. ✅ Download Go modules
4. ✅ Install the `swag` CLI (Swagger doc generator)
5. ✅ Generate Swagger documentation
6. ✅ Build and test the project

After setup, run the server:
```bash
go run main.go
# or, if your system Go version differs:
go1.25.8 run main.go
```

---

## Windows Multi-SDK Setup

For Windows developers who prefer to install multiple Go SDKs side-by-side (rather than using the `golang.org/dl` wrapper approach), use `setup-goenv.bat`.

### Installation

#### Step 1 — Configure the SDK root (optional)

By default, SDKs are installed to `D:\go-apps\go-sdks\`. To use a different location, set the `GOSWITCH_SDK_ROOT` environment variable **before** running the setup:

```bat
:: Example: install SDKs to E:\dev\go-sdks instead of D:\go-apps\go-sdks
set GOSWITCH_SDK_ROOT=E:\dev\go-sdks
```

#### Step 2 — Run the setup script

**Double-click** `documentation\setup-goenv.bat`. The script will:

1. Request Administrator access automatically
2. Create the SDK folders at the configured root
3. Download Go **1.24.6** and Go **1.25.8** from [go.dev/dl](https://go.dev/dl/)
4. Extract both SDKs to separate folders
5. Set default environment variables (`GOROOT`, `GOPATH`, `PATH`, `GOSWITCH_SDK_ROOT`)
6. Verify the installation

> ⏳ Download time depends on your internet speed. Do not close the terminal during the process.

> ⚠ **Note:** Step 5 permanently modifies your user-level environment variables to set a default Go version. Use `goswitch` (session-only) for day-to-day version switching — it will **not** interfere with other projects.

#### Step 3 — Restart terminal

After the script completes, **close the terminal and open a new one**, then verify:

```bat
go version
```

Expected output:
```
go version go1.25.8 windows/amd64
```

#### Step 4 — Install goswitch (optional, recommended)

Copy `documentation\goswitch.bat` to a folder in your PATH (e.g. `D:\go-apps\goswitch.bat`), then run once:

```bat
setx PATH "D:\go-apps;%PATH%"
```

Restart your terminal. You can now use `goswitch` from any directory.

### Folder Structure

After installation:

```
D:\go-apps\                        (or your custom GOSWITCH_SDK_ROOT parent)
├── go-sdks\
│   ├── go1.24.6\                  ← Go SDK 1.24.6
│   │   └── bin\go.exe
│   └── go1.25.8\                  ← Go SDK 1.25.8 (default)
│       └── bin\go.exe
├── gopath\                        ← GOPATH (modules, binaries)
│   ├── bin\
│   ├── pkg\
│   └── src\
└── goswitch.bat                   ← Version switcher (session-only)
```

### Environment Variables

| Variable | Value | Notes |
|----------|-------|-------|
| `GOROOT` | `<SDK_ROOT>\go1.25.8` | Changed by `goswitch` (session-only) |
| `GOPATH` | `<parent>\gopath` | Does not change between versions |
| `GOSWITCH_SDK_ROOT` | `<SDK_ROOT>` | Used by both `goswitch.bat` and `setup-goenv.bat` |
| `PATH` | Includes `%GOROOT%\bin` & `%GOPATH%\bin` | Updated automatically |

### Verification

Open a **new terminal** and run:

```bat
:: Check active Go version
go version

:: Check environment
go env GOROOT
go env GOPATH
```

---

## Switching Go Versions

### Method 1 — goswitch (Windows, session-only)

`goswitch` changes the Go version **only in the current terminal session**. Other terminals, other projects, and the permanent environment are **not affected**.

#### List available versions

```bat
goswitch
```

Output:
```
  GOSWITCH - Go Version Switcher  [session-only]
  ==============================================

  Usage:    goswitch <version>
  Example:  goswitch 1.24.6

  SDK root: D:\go-apps\go-sdks
  (Override with: set GOSWITCH_SDK_ROOT=<path>)

  Installed versions:
       1.24.6
     *  1.25.8

  * = active version in this terminal

  NOTE: goswitch only changes the CURRENT terminal session.
        Other terminals and projects are NOT affected.
```

#### Switch version

```bat
goswitch 1.24.6
```

Output:
```
  [PASS] Switched to Go 1.24.6  [this terminal only]
  ================================================
  go version go1.24.6 windows/amd64

  GOROOT = D:\go-apps\go-sdks\go1.24.6
  GOPATH = D:\go-apps\gopath

  NOTE: This change applies to the CURRENT terminal only.
        Other terminals and projects are NOT affected.
        To make this the default, run: setup-goenv.bat
```

#### Custom SDK root

If your SDKs are not in `D:\go-apps\go-sdks`, set the `GOSWITCH_SDK_ROOT` environment variable:

```bat
set GOSWITCH_SDK_ROOT=E:\dev\go-sdks
goswitch 1.25.8
```

### Method 2 — golang.org/dl (cross-platform, project-local)

This is the approach used by `setup.sh` / `setup.bat`. It installs a versioned wrapper binary (e.g. `go1.25.8`) into `$(go env GOPATH)/bin/` without modifying your system Go.

```bash
# Install the wrapper (one-time)
go install golang.org/dl/go1.25.8@latest

# Download the toolchain (one-time)
go1.25.8 download

# Use it
go1.25.8 run main.go
go1.25.8 build ./...
go1.25.8 test ./...
```

The wrapper binary is a thin shim that delegates to the real toolchain stored in `$(go env GOPATH)/sdk/go1.25.8/`. Your system `go` binary is untouched.

To make the wrapper binary available in any terminal, add `$(go env GOPATH)/bin` to your PATH:

```bash
# Linux / macOS (add to ~/.bashrc or ~/.zshrc)
export PATH="$(go env GOPATH)/bin:$PATH"
```

```bat
:: Windows (run once)
for /f %%P in ('go env GOPATH') do setx PATH "%%P\bin;%PATH%"
```

---

## Why Session-Only?

Previous versions of `goswitch.bat` used `setx` to permanently modify `GOROOT`, `GOPATH`, and `PATH`. This caused issues:

| Problem | Impact |
|---------|--------|
| **Interfered with other projects** | Switching to Go 1.24.6 for project A also changed the Go version for project B in another terminal |
| **Required terminal restart** | `setx` changes only take effect in new terminals |
| **Hard to diagnose** | Developers didn't realize their global environment had been changed |

The current approach (`set` only, no `setx`) avoids all of these problems. Each terminal is independent.

---

## VS Code Configuration

### 1. Install the Go extension

Open VS Code → Extensions (`Ctrl+Shift+X`) → Search **"Go"** by Go Team at Google → Install.

### 2. Configure settings

Open Settings (`Ctrl+,`) → Search `go.goroot`, or add to `settings.json`:

```json
{
    "go.goroot": "D:\\go-apps\\go-sdks\\go1.25.8",
    "go.gopath": "D:\\go-apps\\gopath"
}
```

> **Tip:** After running `goswitch`, restart VS Code to pick up the new `GOROOT`, or update `go.goroot` in settings manually.

### 3. Install Go tools

Open Command Palette (`Ctrl+Shift+P`) → type **"Go: Install/Update Tools"** → select all → OK.

---

## Troubleshooting

### `'go' is not recognized as an internal or external command`

**Cause:** Terminal has not picked up the new environment variables.

**Fix:**
1. Close **all** open terminals / command prompts
2. Open a new terminal
3. Run `go version`

If still failing, check PATH manually:
```bat
echo %GOROOT%
echo %PATH%
```

### `'goswitch' is not recognized`

**Cause:** The folder containing `goswitch.bat` is not in PATH.

**Fix:**
```bat
:: Add the folder containing goswitch.bat to PATH
setx PATH "D:\go-apps;%PATH%"
```
Restart the terminal.

### `GOROOT is empty` or `The system cannot find the path specified`

**Cause:** The requested Go version is not installed.

**Fix:**
```bat
:: List installed versions
goswitch

:: Switch to an available version
goswitch 1.25.8
```

### Setup script fails to download Go SDK

**Cause:** Internet connection issue or Go version not available.

**Fix:**
1. Check that [https://go.dev/dl/](https://go.dev/dl/) is accessible from your browser
2. Ensure no proxy/firewall is blocking the download
3. Re-run `setup-goenv.bat` — it skips already-installed versions

### Module download error / `GOPROXY` timeout

**Cause:** Default proxy `https://proxy.golang.org` is unreachable.

**Fix:**
```bat
go env -w GOPROXY=https://proxy.golang.org,direct
```

Or behind a corporate proxy:
```bat
set HTTP_PROXY=http://proxy-server:port
set HTTPS_PROXY=http://proxy-server:port
go mod tidy
```

---

## FAQ

### Do I have to use drive D:?

No. Set `GOSWITCH_SDK_ROOT` to any directory before running `setup-goenv.bat` and `goswitch.bat`. For example:

```bat
set GOSWITCH_SDK_ROOT=E:\dev\go-sdks
setup-goenv.bat
```

### Can I install additional Go versions?

Yes. Download the zip from [go.dev/dl](https://go.dev/dl/), extract to `<SDK_ROOT>\go<VERSION>\`, then use `goswitch <VERSION>`.

Example for Go 1.23.8:
1. Download `go1.23.8.windows-amd64.zip`
2. Extract the `go` folder inside the zip to `D:\go-apps\go-sdks\go1.23.8\`
3. Verify: `D:\go-apps\go-sdks\go1.23.8\bin\go.exe` must exist
4. Switch: `goswitch 1.23.8`

### Does goswitch change the Go version permanently?

**No.** `goswitch` only affects the current terminal session. When you close the terminal, the change is lost. Other open terminals are not affected. To set a new permanent default, re-run `setup-goenv.bat`.

### Is it safe to run setup-goenv.bat multiple times?

Yes. The script skips downloads for versions that are already installed.

### How do I uninstall?

1. Delete the SDK folder (e.g. `D:\go-apps\`)
2. Remove environment variables:
   ```bat
   setx GOROOT ""
   setx GOPATH ""
   setx GOSWITCH_SDK_ROOT ""
   ```
3. Remove SDK paths from PATH via **System Properties → Environment Variables**

---

## File Reference

| File | Location | Purpose |
|------|----------|---------|
| `setup.sh` | Repository root | Project setup (Linux/macOS) — uses golang.org/dl |
| `setup.bat` | Repository root | Project setup (Windows) — uses golang.org/dl |
| `setup-goenv.bat` | `documentation/` | Multi-SDK install for Windows — downloads Go zips |
| `goswitch.bat` | `documentation/` (copy to PATH) | Switch Go version per terminal session (Windows) |

---

> **Maintainer:** Tim ITSD3/SD3
> **Last Updated:** 2026-04-20
