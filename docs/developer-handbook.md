# Boilerplate 2.0 — Developer Handbook

**Audience:** Software engineers in the ITSD3 / SD3 department who clone this
template to build a new microservice.

---

## Table of Contents

1. [Boilerplate 2.0 Audit Results](#audit-results)
2. [Architecture Overview](#architecture-overview)
3. [Project Structure](#project-structure)
4. [Adding a New Domain Endpoint — Step-by-Step](#adding-a-new-domain-endpoint)
5. [Using the Background Worker Pool](#using-the-background-worker-pool)
6. [Production Checklist](#production-checklist)

---

## 1. Boilerplate 2.0 Audit Results

### 1.1 Vulnerability Log (Boilerplate 1.0 issues — all fixed)

| # | Category | Finding | Fix applied |
|---|---|---|---|
| 1 | **Critical bug** | `GlobalErrorHandler` had no `default` case — any `*fiber.Error` not in the switch (401, 403, 405, 408, 413, 429 …) was silently re-emitted as HTTP 500, masking every client-side error as a server crash in production dashboards | Added `default` case using `net/http.StatusText(e.Code)` |
| 2 | **Envelope inconsistency** | `BadRequestError` and `GatewayTimeoutError` used ad-hoc `fiber.Map{}` literals instead of the typed `ResponseError` struct — any field rename diverged silently | Converted both helpers to `ResponseError` |
| 3 | **Fragile recover middleware** | Custom `defer`-based recover: (a) silently discarded write errors (`_ = c.JSON(…)`), (b) bypassed `GlobalErrorHandler`, (c) had no stack-trace capture | Replaced with Fiber's built-in `middleware/recover` + `runtime/debug.Stack()` logging |
| 4 | **Unbounded goroutine growth** | No mechanism to cap background goroutines — a sudden traffic spike creating background tasks would grow goroutine count without bound | Added bounded `worker.Pool` (fixed workers + buffered channel with load-shedding) |
| 5 | **Dead code** | `UpdateUserPassword` checked `userName == ""` — a path Fiber's `:user_name` route pattern can never produce | Removed |
| 6 | **No request tracing** | No correlation ID on requests — impossible to trace a single request through logs across middleware, handler, and service layers | Added `RequestIDMiddleware` (UUIDv4, sets `X-Request-ID` header + context locals) |
| 7 | **No context timeout propagation** | DB queries used raw context without deadlines — a slow query could block a handler goroutine indefinitely | Added `TimeoutMiddleware` (30 s default deadline on every request's `UserContext`) |
| 8 | **No readiness probe** | Only a liveness probe (`/live`) — Kubernetes/load-balancer had no way to detect a service that was up but couldn't reach the database | Added `/ready` probe that pings PostgreSQL with a 2 s timeout |

### 1.2 Rating

| Dimension | Score | Justification |
|---|---|---|
| **Security** | 7 / 10 | Error envelopes are now consistent and leak no internal details. Passwords are bcrypt-hashed at the service layer. Rate limiting is in place. Room to grow: add JWT middleware, request-body signing, TLS termination config. |
| **Scalability** | 7 / 10 | Fiber/fasthttp provides excellent per-core throughput. The new bounded worker pool prevents unbounded goroutine growth under load. Room to grow: connection-pool tuning (`pgx` pool size), read-replicas, horizontal sharding. |
| **Maintainability** | 8 / 10 | Clean layered architecture (handler → service → repository) with interface-based DI, making every layer unit-testable in isolation. Godocs on all exported symbols. Room to grow: adopt structured logging (zerolog/zap), add OpenTelemetry traces. |

---

## 2. Architecture Overview

```
HTTP Request
     │
     ▼
┌──────────────────────────────────────────────────┐
│  Fiber Middleware Stack                           │
│  (RequestID → Logger → Recover → Timeout         │
│   → HealthCheck → Favicon → RateLimiter)          │
└────────────────┬─────────────────────────────────┘
                 │
                 ▼
┌──────────────────────────────────────────────────┐
│  Handler Layer  (internal/api/handlers/)         │
│  • Validates HTTP request / marshals response    │
│  • Delegates to Service layer                    │
│  • May submit fire-and-forget tasks to Pool      │
└──────┬───────────────────────────┬───────────────┘
       │                           │
       ▼                           ▼
┌────────────┐            ┌─────────────────┐
│  Service   │            │  Worker Pool    │
│  Layer     │            │  (background    │
│  (business │            │   tasks)        │
│   logic,   │            └─────────────────┘
│   bcrypt)  │
└──────┬─────┘
       │
       ▼
┌──────────────────────────────────────────────────┐
│  Repository Layer (internal/api/repository/)     │
│  • Implements Repository interfaces              │
│  • Raw SQL via pgx/scany — no ORM                │
└──────────────────────┬───────────────────────────┘
                       │
                       ▼
                   PostgreSQL
```

**Key design principles:**

- **Accept interfaces, return structs.** Every service depends on a repository
  *interface*, not a concrete struct. Swap in a mock for tests — zero real DB
  required.
- **Errors wrap, never swallow.** Always use `fmt.Errorf("context: %w", err)`
  so callers can use `errors.Is`/`errors.As`.
- **Handlers own only HTTP concerns.** Business rules live in the Service layer
  exclusively.

---

## 3. Project Structure

```
.
├── cmd/                        # (future) additional entry points (CLI, migration runner)
├── config/
│   ├── initialize.go           # Config struct + LoadConfigIni()
│   └── initialize_test.go
├── docs/                       # Auto-generated swagger + this handbook
│   └── developer-handbook.md
├── internal/
│   ├── api/
│   │   ├── handlers/           # HTTP layer: parse request → call service → write response
│   │   ├── models/             # Request / response DTOs
│   │   ├── repository/         # Data-access layer: interfaces + Postgres impls
│   │   ├── router/             # Route registration
│   │   └── services/           # Business logic layer
│   ├── database/               # DB pool initialisation
│   ├── middleware/             # Fiber middleware registrations
│   │   ├── initialize.go      # Middleware dependency graph + registration order
│   │   ├── requestid.go       # X-Request-ID generation (UUIDv4) ← NEW in 2.0
│   │   ├── timeout.go         # Context deadline propagation (30 s) ← NEW in 2.0
│   │   ├── health-check.go    # /live + /ready (DB ping) probes ← UPDATED in 2.0
│   │   ├── recover.go         # Panic recovery with stack traces
│   │   ├── ratelimiter.go     # 100 req/min per IP
│   │   ├── logger.go          # HTTP request logging (fiberlogrus)
│   │   └── favicon.go         # Static favicon serving
│   ├── server/                 # Fiber app construction + lifecycle
│   ├── utility/
│   │   ├── fibererror/         # ResponseError envelope + GlobalErrorHandler
│   │   └── swagger/            # Swagger doc serving + proxy-path handling
│   └── worker/                 # Bounded background worker pool ← NEW in 2.0
│       ├── pool.go
│       └── pool_test.go
├── static/public/              # Static files (favicon, 404 page)
├── appsettings.ini             # Non-secret runtime config
├── Dockerfile                  # Multi-stage production build ← NEW in 2.0
├── .dockerignore               # Files excluded from Docker context
├── go.mod / go.sum
└── main.go                     # Entry point: wire → start → graceful shutdown
```

---

## 4. Adding a New Domain Endpoint

### Worked example: User Profile

Suppose you need a `GET /api/users/:user_name/profile` endpoint that returns
extended profile information stored in a `dc_user_profile_t` table.

You need to create **four files** and modify **one existing file**.

---

### 4.1 Model — `internal/api/models/ProfileModels.go`

```go
package models

// ProfileResponse is the response DTO for the User Profile endpoint.
type ProfileResponse struct {
    UserName    string `json:"user_name"    db:"user_name"`
    DisplayName string `json:"display_name" db:"display_name"`
    Email       string `json:"email"        db:"email"`
}

// UpdateProfileRequest is the request DTO for updating a user profile.
type UpdateProfileRequest struct {
    DisplayName string `json:"display_name" example:"Alice Smith"`
    Email       string `json:"email"        example:"alice@example.com"`
}
```

---

### 4.2 Repository interface + implementation — `internal/api/repository/ProfileRepo.go`

```go
package repository

import (
    "context"
    "fmt"

    "go-api-boilerplate/internal/api/models"
    "go-api-boilerplate/internal/database"
)

// ProfileRepository defines the persistence contract for user profiles.
type ProfileRepository interface {
    GetProfile(ctx context.Context, userName string) (*models.ProfileResponse, error)
    UpsertProfile(ctx context.Context, userName, displayName, email string) error
}

// PostgresProfileRepository is the production implementation.
type PostgresProfileRepository struct{}

func NewPostgresProfileRepository() ProfileRepository {
    return &PostgresProfileRepository{}
}

func (r *PostgresProfileRepository) GetProfile(
    ctx context.Context, userName string,
) (*models.ProfileResponse, error) {
    var p models.ProfileResponse
    q := `SELECT user_name, display_name, email
          FROM dc_user_profile_t WHERE user_name = $1`
    if err := database.Db.SelectOne(ctx, &p, q, userName); err != nil {
        return nil, fmt.Errorf("GetProfile %q: %w", userName, err)
    }
    return &p, nil
}

func (r *PostgresProfileRepository) UpsertProfile(
    ctx context.Context, userName, displayName, email string,
) error {
    q := `INSERT INTO dc_user_profile_t (user_name, display_name, email)
          VALUES ($1, $2, $3)
          ON CONFLICT (user_name) DO UPDATE
          SET display_name = $2, email = $3`
    if _, err := database.Db.Exec(ctx, q, userName, displayName, email); err != nil {
        return fmt.Errorf("UpsertProfile %q: %w", userName, err)
    }
    return nil
}
```

---

### 4.3 Service — `internal/api/services/ProfileService.go`

```go
package services

import (
    "context"
    "fmt"

    "go-api-boilerplate/internal/api/models"
    "go-api-boilerplate/internal/api/repository"

    "github.com/boni-fm/go-libsd3/pkg/log"
)

// ProfileService owns all business rules for user profile operations.
type ProfileService struct {
    log_ *log.Logger
    repo repository.ProfileRepository
}

func NewProfileService(log_ *log.Logger, repo repository.ProfileRepository) *ProfileService {
    return &ProfileService{log_: log_, repo: repo}
}

// GetProfile returns the profile for the given user.
func (s *ProfileService) GetProfile(
    ctx context.Context, userName string,
) (*models.ProfileResponse, error) {
    profile, err := s.repo.GetProfile(ctx, userName)
    if err != nil {
        return nil, fmt.Errorf("ProfileService.GetProfile: %w", err)
    }
    return profile, nil
}

// UpsertProfile creates or updates the profile for the given user.
func (s *ProfileService) UpsertProfile(
    ctx context.Context, userName, displayName, email string,
) error {
    if userName == "" {
        return fmt.Errorf("ProfileService.UpsertProfile: userName is required")
    }
    return s.repo.UpsertProfile(ctx, userName, displayName, email)
}
```

---

### 4.4 Handler interface — add to `internal/api/handlers/interfaces.go`

```go
// ProfileServiceIface defines the profile-domain operations required by the
// handler layer.
type ProfileServiceIface interface {
    GetProfile(ctx context.Context, userName string) (*models.ProfileResponse, error)
    UpsertProfile(ctx context.Context, userName, displayName, email string) error
}
```

---

### 4.5 Handler — `internal/api/handlers/ProfileHandler.go`

```go
package handlers

import (
    "context"

    "go-api-boilerplate/internal/utility/fibererror"

    "github.com/gofiber/fiber/v2"
)

// GetProfile godoc
// @Summary      Get user profile
// @Description  Returns extended profile information for a user
// @Tags         profile
// @Produce      json
// @Param        user_name  path      string  true  "Username"
// @Success      200        {object}  models.ProfileResponse
// @Failure      404        {object}  fibererror.ResponseError
// @Failure      500        {object}  fibererror.ResponseError
// @Router       /api/users/{user_name}/profile [get]
func (hr *HandlersRegistry) GetProfile(c *fiber.Ctx) error {
    userName := c.Params("user_name")
    profile, err := hr.ProfileService.GetProfile(c.Context(), userName)
    if err != nil {
        hr.log_.Errorf("GetProfile error: %v", err)
        return c.Status(fiber.StatusInternalServerError).JSON(fibererror.ResponseError{
            Code:    fiber.StatusInternalServerError,
            Error:   "Internal Server Error",
            Message: "Failed to fetch profile",
        })
    }

    // Optional: dispatch a non-critical "profile viewed" event to the worker
    // pool so it is recorded asynchronously without blocking the response.
    if hr.Pool != nil {
        user := userName
        if ok := hr.Pool.Submit(func(_ context.Context) {
            hr.log_.Infof("[audit] profile viewed: %s", user)
        }); !ok {
            hr.log_.Warn("worker pool saturated — profile-view audit event dropped")
        }
    }

    return c.Status(fiber.StatusOK).JSON(fiber.Map{
        "success": true,
        "data":    profile,
    })
}
```

---

### 4.6 Add ProfileService to HandlersRegistry — update `BaseHandler.go`

Add the new service field and wire it in `NewHandlersRegistry`:

```go
type HandlersRegistry struct {
    log_           *log.Logger
    SwaggerDoc     *swagger.DocumentModifier
    UserService    UserServiceIface
    ProfileService ProfileServiceIface   // ← add this
    Pool           *worker.Pool
}

func NewHandlersRegistry(log_ *log.Logger, pool *worker.Pool) *HandlersRegistry {
    userRepo    := repository.NewPostgresUserRepository()
    profileRepo := repository.NewPostgresProfileRepository()   // ← add this
    return &HandlersRegistry{
        log_:           log_,
        SwaggerDoc:     swagger.NewDocumentModifier(),
        UserService:    services.NewUserService(log_, userRepo),
        ProfileService: services.NewProfileService(log_, profileRepo),  // ← add this
        Pool:           pool,
    }
}
```

---

### 4.7 Register the route — update `internal/api/router/router.go`

```go
app.Get("/api/users/:user_name/profile", handlers.GetProfile)
```

---

### 4.8 Write tests — `internal/api/handlers/profile_test.go`

```go
package handlers_test

import (
    "context"
    "net/http"
    "net/http/httptest"
    "testing"

    "go-api-boilerplate/internal/api/handlers"
    "go-api-boilerplate/internal/api/models"

    "github.com/boni-fm/go-libsd3/pkg/log"
    "github.com/gofiber/fiber/v2"
)

type mockProfileSvc struct {
    profile  *models.ProfileResponse
    getErr   error
    upsertErr error
}

func (m *mockProfileSvc) GetProfile(_ context.Context, _ string) (*models.ProfileResponse, error) {
    return m.profile, m.getErr
}

func (m *mockProfileSvc) UpsertProfile(_ context.Context, _, _, _ string) error {
    return m.upsertErr
}

var _ handlers.ProfileServiceIface = (*mockProfileSvc)(nil)

func TestGetProfile_Success(t *testing.T) {
    l := log.NewLoggerWithFilename("test")
    svc := &mockProfileSvc{
        profile: &models.ProfileResponse{
            UserName: "alice", DisplayName: "Alice Smith", Email: "alice@example.com",
        },
    }
    hr := handlers.NewHandlersRegistryForTest(l, nil)
    hr.ProfileService = svc   // inject mock

    app := fiber.New()
    app.Get("/api/users/:user_name/profile", hr.GetProfile)

    req := httptest.NewRequest(http.MethodGet, "/api/users/alice/profile", nil)
    resp, err := app.Test(req, 5000)
    if err != nil {
        t.Fatal(err)
    }
    if resp.StatusCode != http.StatusOK {
        t.Errorf("status: got %d, want %d", resp.StatusCode, http.StatusOK)
    }
}
```

---

## 5. Using the Background Worker Pool

The worker pool (`internal/worker.Pool`) is injected into every `HandlersRegistry`
at startup. Handlers access it via `hr.Pool`.

### When to use it

Use the pool for **non-critical side-effects** that should not block the HTTP
response:

- Audit logging to a secondary store
- Publishing events to a message queue
- Flushing in-memory metrics batches
- Sending confirmation emails

### Pattern

```go
func (hr *HandlersRegistry) CreateOrder(c *fiber.Ctx) error {
    // ... validate, call service, write DB record ...

    // Fire-and-forget: record audit event without delaying the 201 response.
    if hr.Pool != nil {
        orderID := order.ID
        ok := hr.Pool.Submit(func(ctx context.Context) {
            if err := auditClient.Record(ctx, "order.created", orderID); err != nil {
                hr.log_.Errorf("audit record failed: %v", err)
            }
        })
        if !ok {
            // Pool is saturated — log the drop so operators can alert on it.
            hr.log_.Warnf("worker pool saturated — audit event for order %s dropped", orderID)
        }
    }

    return c.Status(fiber.StatusCreated).JSON(...)
}
```

### Capacity tuning (in `internal/server/server.go`)

| Constant | Default | Guidance |
|---|---|---|
| `defaultWorkerCount` | 4 | Set to ≈ 2× the number of background task types |
| `defaultWorkerCapacity` | 128 | Set to (peak task/sec) × (max acceptable latency in seconds) |

Monitor `pool.Stats()` in your health/metrics endpoint:

```go
processed, dropped := srv.Pool.Stats()
// Expose as Prometheus gauges or include in /health response.
```

A non-zero `dropped` counter under steady load means the pool is under-provisioned.

---

## 6. Production Checklist

Before deploying a service built from this boilerplate:

- [ ] **Secrets** — `Kunci` and DB credentials are injected via environment variables
  (e.g. `APP_CONFIG_KUNCI`), **not** stored in `appsettings.ini`.
- [ ] **TLS** — terminate TLS at the load balancer or configure Fiber's
  `app.ListenMutualTLS` / `app.ListenTLS`.
- [ ] **Rate limits** — adjust `RateLimiter` defaults in `middleware/ratelimiter.go`
  for your service's expected peak RPS.
- [ ] **DB pool size** — tune `pgx` pool `MinConns` / `MaxConns` to match your
  Postgres server's `max_connections` divided by replica count.
- [ ] **Worker pool size** — adjust `defaultWorkerCount` and `defaultWorkerCapacity`
  based on load testing results.
- [ ] **Graceful shutdown** — verify `ShutdownWithTimeout` + `pool.Stop()` are
  wired correctly (see `main.go`) so no in-flight work is abandoned on `SIGTERM`.
- [ ] **Logging** — switch to structured JSON logging (zerolog or zap) for
  production deployments so log-aggregation pipelines (Elastic, Loki) can index
  fields.
- [ ] **Health probe** — wire `/live` into your Kubernetes `livenessProbe` and add
  a `/ready` endpoint that checks DB connectivity before marking the pod ready.
- [ ] **Tests** — run `go test -race ./...` before every merge to main.
