# Go API Boilerplate

A production-ready Go REST API boilerplate built on top of the [Fiber](https://gofiber.io/) framework. This template gives you a solid starting point with pre-wired middleware, PostgreSQL integration, auto-generated Swagger documentation, and a clean layered architecture.

---

## Features

- **[Fiber v3](https://gofiber.io/)** – Fast, Express-inspired HTTP framework
- **PostgreSQL** – Database integration via `pgx` / `go-libsd3`
- **Swagger / OpenAPI** – Auto-generated docs with proxy-path support *(dev/staging only)*
- **Structured logging** – File-based rotating logs via `logrus` + `go-libsd3`
- **Rate limiting** – 100 requests per minute per client
- **Panic recovery** – Catches panics and returns a clean 500 JSON response
- **Request tracing** – UUIDv4 `X-Request-ID` header added to every request for log correlation
- **Context timeouts** – 30-second deadline propagated to every request's context so DB queries and I/O never block indefinitely
- **Health probes** – Liveness at `/live`, readiness (with DB ping) at `/ready`
- **Background worker pool** – Bounded goroutine pool for non-critical fire-and-forget tasks (audit logs, events, emails)
- **Docker support** – Multi-stage `Dockerfile` produces a minimal, non-root Alpine image
- **Reverse-proxy aware** – Reads `X-Forwarded-Prefix` for Swagger base path *and* as a tenant-key source
- **INI-based config** – Simple `appsettings.ini` configuration
- **Multi-DC / Multi-Tenant** – Route each request to the correct database via the `X-Kunci` header, `?kunci=` query param, or nginx `X-Forwarded-Prefix`
- **Timezone configuration** – Override `time.Local` via `TZ` env var (production) or `Timezone` in `appsettings.ini` (development)
- **Layered architecture** – Handler → Service → Repository separation
- **Developer init scripts** – `setup.sh` / `setup.bat` handle goswitch (exact Go version) without touching the system Go

---

## Prerequisites

| Tool | Version |
|------|---------|
| Go   | 1.25+   |
| PostgreSQL | 13+ |
| [swag CLI](https://github.com/swaggo/swag) | latest |

Install the Swagger code-generator once:

```bash
go install github.com/swaggo/swag/cmd/swag@latest
```

---

## Getting Started

### 1. Clone the repository

```bash
git clone https://github.com/boni-fm/go-api-boilerplate.git
cd go-api-boilerplate
```

### 2. Run the developer setup script

The setup script verifies your Go version, installs the exact version via
`golang.org/dl` if needed (**without** touching your system Go), downloads
modules, installs `swag`, generates docs, and confirms a clean build.

```bash
# Linux / macOS
chmod +x setup.sh && ./setup.sh

# Windows
setup.bat
```

### 3. Configure the application

Edit `appsettings.ini`:

```ini
[CONFIG]
AppName       = Go API Boilerplate
IsDevelopment = true
Port          = 8080
Kunci         = g009sim             # comma-separate for multi-DC: g009sim,g010sim
Timezone      = Asia/Jakarta        # IANA timezone; overridden by TZ env var
```

| Key | Default | Description |
|-----|---------|-------------|
| `AppName` | `Go API Boilerplate` | Application name shown in logs and Swagger UI |
| `IsDevelopment` | `false` | Set `true` to enable Swagger UI; **disable in production** |
| `Port` | `8080` | HTTP port the server listens on |
| `Kunci` | *(required)* | Comma-separated database credential keys for `go-libsd3` |
| `Timezone` | `UTC` | IANA timezone; overridden by `TZ` environment variable |

### 4. Run the application

```bash
go run main.go
```

The server starts on `http://localhost:8080`.  
Open `http://localhost:8080/swagger` to view the interactive API documentation.

---

## Project Structure

```
go-api-boilerplate/
├── main.go                        # Application entry point
├── appsettings.ini                # Runtime configuration
├── setup.sh / setup.bat           # Developer init scripts (goswitch + tool install)
├── build/
│   ├── build.sh                   # Cross-platform build (Linux / macOS)
│   └── build.bat                  # Cross-platform build (Windows)
├── Dockerfile                     # Multi-stage production Docker build
├── .dockerignore
├── go.mod / go.sum
│
├── config/
│   └── initialize.go              # Config struct & INI loader
│
├── docs/                          # Auto-generated Swagger files (do not edit)
│   ├── developer-handbook.md      # Architecture guide for contributors
│   ├── docs.go
│   ├── swagger.json
│   └── swagger.yaml
│
├── internal/
│   ├── api/
│   │   ├── handlers/              # HTTP handler functions
│   │   │   ├── BaseHandler.go     # HandlersRegistry (dependency container)
│   │   │   ├── interfaces.go      # Package doc (interfaces removed — concrete types used)
│   │   │   ├── PingHandler.go
│   │   │   ├── UserHandler.go
│   │   │   └── DocumentationSwaggerHandler.go
│   │   ├── models/                # Request / response structs
│   │   ├── repository/            # Database access layer
│   │   ├── router/
│   │   │   └── router.go          # Route registration
│   │   └── services/              # Business logic
│   │
│   ├── database/
│   │   └── db.go                  # Registry, context helpers, InitDatabases
│   │
│   ├── middleware/
│   │   ├── initialize.go          # Middleware wiring (registration order)
│   │   ├── multitenant.go         # Multi-DC tenant key resolution
│   │   ├── requestid.go
│   │   ├── timeout.go
│   │   ├── favicon.go
│   │   ├── health-check.go
│   │   ├── logger.go
│   │   ├── ratelimiter.go
│   │   └── recover.go
│   │
│   ├── server/
│   │   └── server.go
│   │
│   ├── worker/
│   │   └── pool.go
│   │
│   └── utility/
│       ├── fibererror/
│       └── swagger/
│
└── static/
    └── public/                    # Static assets (favicon, 404 page, …)
```

---

## API Endpoints

### General

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/` | Redirects to Swagger UI *(dev only)* |
| `GET` | `/ping` | Liveness check – returns `{ "message": "Pong" }` |
| `GET` | `/live` | Liveness probe (returns `200 OK` when the process is running) |
| `GET` | `/ready` | Readiness probe (returns `200 OK` only when PostgreSQL is reachable) |
| `GET` | `/swagger` | Interactive Swagger UI *(dev only)* |
| `GET` | `/swagger/doc.json` | Raw OpenAPI JSON *(dev only)* |

### Users (example CRUD)

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/api/users` | Create a new user |
| `GET` | `/api/users` | List all users |
| `PUT` | `/api/users/:user_name/password` | Update a user's password |
| `DELETE` | `/api/users/:user_name` | Delete a user |

---

## Multi-DC / Multi-Tenant Routing

The `MultiTenantMiddleware` resolves the database connection per-request.  
Tenant key resolution priority (**highest → lowest**):

| Priority | Source | Example |
|----------|--------|---------|
| 1 | Query parameter `?kunci=` | `GET /api/users?kunci=g010sim` |
| 2 | Request header `X-Kunci` | `X-Kunci: g010sim` |
| 3 | Nginx `X-Forwarded-Prefix` first segment | `X-Forwarded-Prefix: /g010sim/api` → `g010sim` |
| 4 | Default (first key in `appsettings.ini`) | — |

**nginx example** (routes `/g009sim/*` to the service):

```nginx
location /g009sim/ {
    proxy_pass         http://go-api:8080/;
    proxy_set_header   X-Forwarded-Prefix /g009sim;
}
```

Configure multiple connections in `appsettings.ini`:

```ini
Kunci = g009sim,g010sim
```

---

## Timezone Configuration

Timezone is resolved at process startup. Priority:

| Priority | Source | Example |
|----------|--------|---------|
| 1 | `TZ` environment variable | `TZ=Asia/Jakarta ./myapp` |
| 2 | `Timezone` in `appsettings.ini` | `Timezone = Asia/Jakarta` |
| 3 | UTC (default) | — |

Using the `TZ` env var is the recommended approach for production/Docker deployments:

```bash
# Docker run
docker run -p 8080:8080 -e TZ=Asia/Jakarta go-api-boilerplate

# docker-compose
environment:
  - TZ=Asia/Jakarta
```

---

## Developer Setup (goswitch)

`setup.sh` / `setup.bat` automatically install the exact Go version declared in
`go.mod` if your system Go differs — **without modifying your system Go**:

```
System Go : go1.23.0
Required  : go1.25.8
⚠ Installing go1.25.8 via golang.org/dl...
✓ Using go1.25.8 for this setup run.

  How to use go1.25.8 in future terminal sessions:
  ┌─────────────────────────────────────────────────────────────┐
  │ Option A — Add GOPATH/bin to PATH (recommended):            │
  │   export PATH="$(go env GOPATH)/bin:$PATH"                  │
  │   go1.25.8 run main.go                                      │
  │                                                             │
  │ Option B — Use the full path:                               │
  │   /home/user/go/bin/go1.25.8 run main.go                   │
  │                                                             │
  │ NOTE: This does NOT change your system Go. Other projects   │
  │ using a different Go version are not affected.              │
  └─────────────────────────────────────────────────────────────┘
```

The wrapper binary (`go1.25.8`) lives in `$(go env GOPATH)/bin/` and is
completely independent of your system Go installation.

### Switching Go versions

| Method | Platform | Scope | Command |
|--------|----------|-------|---------|
| `golang.org/dl` wrapper | All | Project-local | `go1.25.8 run main.go` |
| `goswitch.bat` | Windows | Current terminal only | `goswitch 1.25.8` |

**`golang.org/dl` (used by `setup.sh` / `setup.bat`)**

```bash
go install golang.org/dl/go1.25.8@latest   # install wrapper (one-time)
go1.25.8 download                           # download toolchain (one-time)
go1.25.8 run main.go                        # use it
```

**`goswitch.bat` (Windows only — session-only, no permanent env changes)**

```bat
goswitch 1.25.8     :: switch for this terminal only
go run main.go      :: uses the switched version
```

> `goswitch` does **not** use `setx`. Other terminals and projects are not affected.

For full details, see [`documentation/setting-up-goenv.md`](documentation/setting-up-goenv.md).

---

## Adding a New Feature

Follow the existing layer pattern:

1. **Model** – add request/response structs in `internal/api/models/`
2. **Repository** – add database queries in `internal/api/repository/`
3. **Service** – add business logic in `internal/api/services/`
4. **Handler** – add HTTP handler methods on `*HandlersRegistry` in `internal/api/handlers/`
5. **Router** – register routes in `internal/api/router/router.go`

---

## Swagger Documentation

Swagger docs are **auto-generated at startup** when `IsDevelopment = true`.  
Annotate your handlers with `swag` comments:

```go
// CreateProduct godoc
// @Summary Create a new product
// @Tags products
// @Accept json
// @Produce json
// @Param request body models.CreateProductRequest true "Product data"
// @Success 201 {object} map[string]interface{}
// @Router /api/products [post]
func (hr *HandlersRegistry) CreateProduct(c fiber.Ctx) error { ... }
```

To regenerate docs manually:

```bash
swag init -g main.go -o docs
```

---

## Middleware

All middleware is registered in `internal/middleware/initialize.go` in the following order:

| # | Middleware | Description |
|---|------------|-------------|
| 1 | **RequestID** | Generates a UUIDv4 `X-Request-ID` header for log correlation |
| 2 | **Logger** | Structured HTTP request/response logging (logrus) |
| 3 | **Recover** | Catches panics, logs stack traces, returns `500` JSON |
| 4 | **MultiTenant** | Resolves tenant DB from `?kunci`, `X-Kunci`, or `X-Forwarded-Prefix` |
| 5 | **Timeout** | Wraps each request's context with a 30 s deadline |
| 6 | **Favicon** | Serves `/domar.ico` from `static/public/favicon.ico` |
| 7 | **RateLimiter** | 100 requests / 60 seconds per IP |

---

## Running with Docker

```bash
# Build the image
docker build -t go-api-boilerplate .

# Run (pass timezone and database key as environment variables)
docker run -p 8080:8080 \
  -e TZ=Asia/Jakarta \
  go-api-boilerplate
```

The `Dockerfile` uses a two-stage build:

1. **Builder stage** (`golang:1.25-alpine`) – compiles a fully-static binary with `CGO_ENABLED=0`.
2. **Runtime stage** (`alpine:3.21`) – copies only the binary and static assets; runs as a non-root user (`appuser`).

---

## Background Worker Pool

```go
if hr.Pool != nil {
    ok := hr.Pool.Submit(func(ctx context.Context) {
        // e.g. write audit log, publish event, flush metric batch
    })
    if !ok {
        hr.log_.Warn("worker pool saturated — task dropped")
    }
}
```

Tune the pool in `internal/server/server.go`:

| Constant | Default | Guidance |
|---|---|---|
| `defaultWorkerCount` | 4 | ≈ 2× the number of background task types |
| `defaultWorkerCapacity` | 128 | `(peak tasks/sec) × (max acceptable latency in seconds)` |

---

## Configuration Reference

`appsettings.ini` (all keys live under `[CONFIG]`):

| Key | Default | Description |
|-----|---------|-------------|
| `AppName` | `Go API Boilerplate` | Application name |
| `IsDevelopment` | `false` | Enables Swagger UI; disable in production |
| `Port` | `8080` | HTTP listen port |
| `Kunci` | *(required)* | Comma-separated database credential keys |
| `Timezone` | `UTC` | IANA timezone; overridden by `TZ` env var |

---

## License

This project is provided as an open boilerplate template. Feel free to use, modify, and distribute it for your own projects.
