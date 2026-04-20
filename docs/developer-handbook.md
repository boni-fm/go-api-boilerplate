# Boilerplate 3.0 — Developer Handbook

**Audience:** Software engineers in the ITSD3 / SD3 department who clone this
template to build a new microservice.

---

## Table of Contents

1. [Boilerplate Audit Results](#audit-results)
2. [Architecture Overview](#architecture-overview)
3. [Multi-DC / Multi-Tenant Routing](#multi-dc--multi-tenant-routing)
4. [Timezone Configuration](#timezone-configuration)
5. [Developer Setup & goswitch](#developer-setup--goswitch)
6. [Project Structure](#project-structure)
7. [Adding a New Domain Endpoint — Step-by-Step](#adding-a-new-domain-endpoint)
8. [Using the Background Worker Pool](#using-the-background-worker-pool)
9. [Production Checklist](#production-checklist)

---

## 1. Boilerplate Audit Results

### 1.1 Vulnerability Log (all fixed)

| # | Category | Finding | Fix applied |
|---|---|---|---|
| 1 | **Critical bug** | `GlobalErrorHandler` had no `default` case — any `*fiber.Error` not in the switch (401, 403, 405 …) was silently re-emitted as HTTP 500 | Added `default` case using `net/http.StatusText(e.Code)` |
| 2 | **Envelope inconsistency** | `BadRequestError` and `GatewayTimeoutError` used ad-hoc `fiber.Map{}` instead of typed `ResponseError` | Converted both helpers to `ResponseError` |
| 3 | **Fragile recover middleware** | Custom defer-based recover: silently discarded write errors, bypassed `GlobalErrorHandler`, no stack-trace capture | Replaced with Fiber built-in recover + `runtime/debug.Stack()` logging |
| 4 | **Unbounded goroutine growth** | No cap on background goroutines | Added bounded `worker.Pool` (fixed workers + buffered channel, load-shedding) |
| 5 | **Dead code** | `UpdateUserPassword` checked `userName == ""` — unreachable via Fiber route | Removed |
| 6 | **No request tracing** | No correlation ID | Added `RequestIDMiddleware` (UUIDv4, `X-Request-ID`) |
| 7 | **No context timeout propagation** | DB queries had no deadline | Added `TimeoutMiddleware` (30 s per request) |
| 8 | **No readiness probe** | Only `/live` | Added `/ready` that pings PostgreSQL with a 2 s timeout |
| 9 | **Global DB variable** | `database.Db` package-level var caused race conditions | Replaced with `database.Registry` + per-request DI via context |
| 10 | **Service-level interfaces** | `UserServiceIface` in handler layer — unnecessary abstraction | Removed; handlers use `*services.UserService`; mocking at repository level |
| 11 | **404 static-file recursion** | `NotFoundError` propagated `SendFile` error — could recurse into `GlobalErrorHandler` | Added JSON fallback when `SendFile` fails |

### 1.2 Rating

| Dimension | Score | Justification |
|---|---|---|
| **Security** | 8 / 10 | Consistent error envelopes, bcrypt passwords, rate limiting, Swagger disabled in production. Room to grow: JWT, TLS termination. |
| **Scalability** | 8 / 10 | Multi-DC routing, bounded worker pool, context-deadline propagation. Room to grow: pgx pool tuning, read-replicas. |
| **Maintainability** | 9 / 10 | Concrete-type DI, clean layered architecture, godocs on all exported symbols. |

---

## 2. Architecture Overview

```
HTTP Request
     |
     v
+--------------------------------------------------------------+
|  Fiber Middleware Stack                                       |
|  RequestID -> Logger -> Recover -> MultiTenant -> Timeout    |
|  -> Favicon -> RateLimiter                                   |
+-----------------------------+--------------------------------+
                              |  (DB injected into context here)
                              v
+--------------------------------------------------------------+
|  Handler Layer  (internal/api/handlers/)                     |
|  * Validates HTTP request / marshals response                |
|  * Delegates to Service layer                                |
|  * May submit fire-and-forget tasks to Pool                  |
+------+---------------------------------------------+--------+
       |                                             |
       v                                             v
+--------------------+                   +---------------------+
|  Service Layer     |                   |  Worker Pool        |
|  (business logic,  |                   |  (background tasks) |
|   bcrypt)          |                   +---------------------+
+------+-------------+
       |
       v
+--------------------------------------------------------------+
|  Repository Layer (internal/api/repository/)                 |
|  * Calls database.DBFromContext(ctx) for the tenant DB       |
|  * Raw SQL via pgx/scany -- no ORM                           |
+---------------------------+----------------------------------+
                            |
            +---------------+-----------------+
            v                                 v
     PostgreSQL (DC g009sim)         PostgreSQL (DC g010sim)
```

**Key design principles:**

- **Concrete types over unnecessary interfaces.** Handlers depend directly on
  `*services.UserService`. Mocking is done at the repository level
  (`repository.UserRepository` interface) where it matters for unit tests.
- **Errors wrap, never swallow.** Always use `fmt.Errorf("context: %w", err)`.
- **Handlers own only HTTP concerns.** Business rules live in the Service layer.
- **DB is request-scoped.** `MultiTenantMiddleware` injects the correct
  `*postgres.Database` into every request's `context.Context`.
  Repositories retrieve it with `database.DBFromContext(ctx)`.

---

## 3. Multi-DC / Multi-Tenant Routing

### How it works

`MultiTenantMiddleware` runs on every request (position 4 in the stack). It
resolves the tenant key and looks it up in `database.Registry`. The matching
`*postgres.Database` is stored in the request context via `database.WithDB`.
All repository calls then call `database.DBFromContext(ctx)` — no globals.

### Tenant key resolution (highest -> lowest priority)

| Priority | Source | Notes |
|----------|--------|-------|
| 1 | `?kunci=<key>` query parameter | Useful for ad-hoc debugging or SDK calls |
| 2 | `X-Kunci: <key>` request header | Standard service-to-service usage |
| 3 | `X-Forwarded-Prefix` nginx header | First path segment is the key: `/g009sim/api` -> `g009sim` |
| 4 | Default (first `Kunci` in appsettings.ini) | Used when no explicit key is found |

### Nginx configuration example

```nginx
location /g009sim/ {
    proxy_pass         http://go-api:8080/;
    proxy_set_header   Host               $host;
    proxy_set_header   X-Real-IP          $remote_addr;
    proxy_set_header   X-Forwarded-Prefix /g009sim;
}
```

### appsettings.ini for multi-DC

```ini
[CONFIG]
Kunci = g009sim,g010sim   # connects to both DCs at startup
```

The first key (`g009sim`) is the default fallback.

---

## 4. Timezone Configuration

### Resolution order (highest -> lowest)

| Priority | Source | Recommendation |
|----------|--------|---------------|
| 1 | `TZ` environment variable | **Use in production / Docker / K8s** |
| 2 | `Timezone` key in `appsettings.ini` | Use for local development convenience |
| 3 | UTC | Safe default |

### Development (`appsettings.ini`)

```ini
[CONFIG]
Timezone = Asia/Jakarta
```

### Production (Docker / K8s)

```yaml
# docker-compose
environment:
  - TZ=Asia/Jakarta

# Kubernetes Pod spec
env:
  - name: TZ
    value: Asia/Jakarta
```

---

## 5. Developer Setup & goswitch

Run `./setup.sh` (Linux/macOS) or `setup.bat` (Windows) once after cloning.

### What the scripts do

1. **Go version check (goswitch)** — reads the required version from `go.mod`.
   If your system Go differs, it installs the exact version via
   `golang.org/dl/go<version>` **without touching your system Go**.
2. **Module download & tidy** — `go mod download && go mod tidy`.
3. **swag CLI** — installs the Swagger code-generator if not already present.
4. **Doc generation** — `swag init -g main.go -o docs`.
5. **Build** — `go build ./...` to confirm the project compiles cleanly.
6. **Tests** — `go test ./...`.

### goswitch — under the hood

```
System Go : go1.23.0
Required  : go1.25.8
WARNING Version mismatch. Installing go1.25.8 via golang.org/dl...
        (Your system go1.23.0 will NOT be modified.)
OK Using go1.25.8 for this setup run.
   Tip: add $(go env GOPATH)/bin to PATH and run: go1.25.8 run main.go
```

The versioned binary (`go1.25.8` / `go1.25.8.exe`) lives in
`$(go env GOPATH)/bin/` and is completely independent of the system Go. You
can have multiple project-local Go versions on the same machine without any
conflict.

---

## 6. Project Structure

```
.
+-- config/
|   +-- initialize.go           # Config struct + LoadConfigIni()
|   +-- initialize_test.go
+-- docs/                       # Auto-generated swagger + this handbook
|   +-- developer-handbook.md
+-- internal/
|   +-- api/
|   |   +-- handlers/           # HTTP layer
|   |   +-- models/             # Request / response DTOs
|   |   +-- repository/         # Data-access layer: UserRepository interface + Postgres impl
|   |   +-- router/             # Route registration
|   |   +-- services/           # Business logic layer
|   +-- database/               # Registry, context helpers, InitDatabases
|   |   +-- db.go
|   +-- middleware/
|   |   +-- initialize.go       # Middleware dependency graph + registration order
|   |   +-- multitenant.go      # Multi-DC tenant key resolution
|   |   +-- requestid.go        # X-Request-ID generation (UUIDv4)
|   |   +-- timeout.go          # Context deadline propagation (30 s)
|   |   +-- health-check.go     # /live + /ready (DB ping) probes
|   |   +-- recover.go          # Panic recovery with stack traces
|   |   +-- ratelimiter.go      # 100 req/min per IP
|   |   +-- logger.go           # HTTP request logging
|   |   +-- favicon.go          # Static favicon serving
|   +-- server/                 # Fiber app construction + lifecycle
|   +-- utility/
|   |   +-- fibererror/         # ResponseError envelope + GlobalErrorHandler
|   |   +-- swagger/            # Swagger doc serving + proxy-path handling
|   +-- worker/                 # Bounded background worker pool
+-- build/
|   +-- build.sh                # Cross-platform build (Linux/macOS)
|   +-- build.bat               # Cross-platform build (Windows)
+-- static/public/              # Static files (favicon, 404 page)
+-- appsettings.ini             # Non-secret runtime config
+-- setup.sh / setup.bat        # Developer init scripts (goswitch + tool install)
+-- Dockerfile                  # Multi-stage production build
+-- go.mod / go.sum
+-- main.go                     # Entry point: wire -> start -> graceful shutdown
```

---

## 7. Adding a New Domain Endpoint

### Worked example: User Profile

Suppose you need a `GET /api/users/:user_name/profile` endpoint.

---

### 7.1 Model — `internal/api/models/ProfileModels.go`

```go
package models

type ProfileResponse struct {
    UserName    string `json:"user_name"    db:"user_name"`
    DisplayName string `json:"display_name" db:"display_name"`
    Email       string `json:"email"        db:"email"`
}
```

---

### 7.2 Repository — `internal/api/repository/ProfileRepo.go`

```go
package repository

import (
    "context"
    "fmt"

    "go-api-boilerplate/internal/api/models"
    "go-api-boilerplate/internal/database"
)

type ProfileRepository interface {
    GetProfile(ctx context.Context, userName string) (*models.ProfileResponse, error)
}

type PostgresProfileRepository struct{}

func NewPostgresProfileRepository() ProfileRepository {
    return &PostgresProfileRepository{}
}

func (r *PostgresProfileRepository) GetProfile(
    ctx context.Context, userName string,
) (*models.ProfileResponse, error) {
    db := database.DBFromContext(ctx)
    if db == nil {
        return nil, database.ErrNoDB
    }
    var p models.ProfileResponse
    q := `SELECT user_name, display_name, email
          FROM dc_user_profile_t WHERE user_name = $1`
    if err := db.SelectOne(ctx, &p, q, userName); err != nil {
        return nil, fmt.Errorf("GetProfile %q: %w", userName, err)
    }
    return &p, nil
}
```

---

### 7.3 Service — `internal/api/services/ProfileService.go`

```go
package services

import (
    "context"
    "fmt"

    "go-api-boilerplate/internal/api/models"
    "go-api-boilerplate/internal/api/repository"

    "github.com/boni-fm/go-libsd3/pkg/log"
)

type ProfileService struct {
    log_ *log.Logger
    repo repository.ProfileRepository
}

func NewProfileService(log_ *log.Logger, repo repository.ProfileRepository) *ProfileService {
    return &ProfileService{log_: log_, repo: repo}
}

func (s *ProfileService) GetProfile(
    ctx context.Context, userName string,
) (*models.ProfileResponse, error) {
    profile, err := s.repo.GetProfile(ctx, userName)
    if err != nil {
        return nil, fmt.Errorf("ProfileService.GetProfile: %w", err)
    }
    return profile, nil
}
```

---

### 7.4 Handler — `internal/api/handlers/ProfileHandler.go`

```go
package handlers

import (
    "go-api-boilerplate/internal/utility/fibererror"
    "github.com/gofiber/fiber/v3"
)

// GetProfile godoc
// @Summary      Get user profile
// @Tags         profile
// @Produce      json
// @Param        user_name  path      string  true  "Username"
// @Success      200        {object}  models.ProfileResponse
// @Failure      500        {object}  fibererror.ResponseError
// @Router       /api/users/{user_name}/profile [get]
func (hr *HandlersRegistry) GetProfile(c fiber.Ctx) error {
    userName := c.Params("user_name")
    profile, err := hr.ProfileService.GetProfile(c.Context(), userName)
    if err != nil {
        hr.log_.Errorf("GetProfile error: %v", err)
        return fibererror.InternalServerError(c, "Failed to fetch profile")
    }
    return c.Status(fiber.StatusOK).JSON(fiber.Map{
        "success": true,
        "data":    profile,
    })
}
```

---

### 7.5 Wire into HandlersRegistry — `BaseHandler.go`

```go
type HandlersRegistry struct {
    log_           *log.Logger
    SwaggerDoc     *swagger.DocumentModifier
    UserService    *services.UserService
    ProfileService *services.ProfileService   // add
    Pool           *worker.Pool
}

func NewHandlersRegistry(log_ *log.Logger, pool *worker.Pool) *HandlersRegistry {
    userRepo    := repository.NewPostgresUserRepository()
    profileRepo := repository.NewPostgresProfileRepository()
    return &HandlersRegistry{
        log_:           log_,
        SwaggerDoc:     swagger.NewDocumentModifier(),
        UserService:    services.NewUserService(log_, userRepo),
        ProfileService: services.NewProfileService(log_, profileRepo),
        Pool:           pool,
    }
}
```

---

### 7.6 Register route — `router.go`

```go
app.Get("/api/users/:user_name/profile", handlers.GetProfile)
```

---

### 7.7 Write tests — `profile_test.go`

```go
package handlers_test

import (
    "context"
    "net/http"
    "net/http/httptest"
    "testing"
    "time"

    "go-api-boilerplate/internal/api/models"
    "go-api-boilerplate/internal/api/repository"
    "go-api-boilerplate/internal/api/services"
    "go-api-boilerplate/internal/api/handlers"

    "github.com/boni-fm/go-libsd3/pkg/log"
    "github.com/gofiber/fiber/v3"
)

type mockProfileRepo struct {
    profile *models.ProfileResponse
    err     error
}

func (m *mockProfileRepo) GetProfile(_ context.Context, _ string) (*models.ProfileResponse, error) {
    return m.profile, m.err
}

var _ repository.ProfileRepository = (*mockProfileRepo)(nil)

func TestGetProfile_Success(t *testing.T) {
    l := log.NewLoggerWithFilename("test")
    repo := &mockProfileRepo{
        profile: &models.ProfileResponse{UserName: "alice"},
    }
    svc := services.NewProfileService(l, repo)
    hr := handlers.NewHandlersRegistryForTest(l, nil)
    hr.ProfileService = svc

    app := fiber.New()
    app.Get("/api/users/:user_name/profile", hr.GetProfile)

    req := httptest.NewRequest(http.MethodGet, "/api/users/alice/profile", nil)
    resp, err := app.Test(req, fiber.TestConfig{Timeout: 5 * time.Second})
    if err != nil {
        t.Fatal(err)
    }
    if resp.StatusCode != http.StatusOK {
        t.Errorf("status: got %d, want %d", resp.StatusCode, http.StatusOK)
    }
}
```

---

## 8. Using the Background Worker Pool

The worker pool (`internal/worker.Pool`) is injected into every `HandlersRegistry`
at startup. Handlers access it via `hr.Pool`.

### When to use it

- Audit logging to a secondary store
- Publishing events to a message queue
- Flushing in-memory metrics batches
- Sending confirmation emails

### Pattern

```go
func (hr *HandlersRegistry) CreateOrder(c fiber.Ctx) error {
    // ... validate, call service, write DB record ...

    if hr.Pool != nil {
        orderID := order.ID
        ok := hr.Pool.Submit(func(ctx context.Context) {
            if err := auditClient.Record(ctx, "order.created", orderID); err != nil {
                hr.log_.Errorf("audit record failed: %v", err)
            }
        })
        if !ok {
            hr.log_.Warnf("worker pool saturated -- audit event for order %s dropped", orderID)
        }
    }

    return c.Status(fiber.StatusCreated).JSON(...)
}
```

### Capacity tuning

| Constant | Default | Guidance |
|---|---|---|
| `defaultWorkerCount` | 4 | ~2x the number of background task types |
| `defaultWorkerCapacity` | 128 | (peak task/sec) x (max acceptable latency in seconds) |

---

## 9. Production Checklist

- [ ] **Secrets** -- DB credentials injected via env vars, **not** in `appsettings.ini`.
- [ ] **Swagger disabled** -- Set `IsDevelopment = false`.
- [ ] **Timezone** -- Set `TZ=<IANA timezone>` as an environment variable.
- [ ] **Multi-DC** -- Verify `Kunci` lists all required DC keys; confirm nginx `X-Forwarded-Prefix` header is set.
- [ ] **TLS** -- Terminate at load balancer or configure `app.ListenTLS`.
- [ ] **Rate limits** -- Tune `RateLimiter` defaults for expected peak RPS.
- [ ] **DB pool size** -- Tune pgx `MinConns` / `MaxConns` vs Postgres `max_connections`.
- [ ] **Worker pool size** -- Tune from load testing results.
- [ ] **Graceful shutdown** -- Verify `ShutdownWithTimeout` + `pool.Stop()` in `main.go`.
- [ ] **Logging** -- Switch to structured JSON logging (zerolog / zap) for production.
- [ ] **Health probes** -- Wire `/live` -> `livenessProbe`, `/ready` -> `readinessProbe`.
- [ ] **Tests** -- Run `go test -race ./...` before every merge.
