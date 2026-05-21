# Boilerplate 4.0 — Developer Handbook

**Audience:** Engineer software di departemen ITSD3 / SD3 yang menggunakan template ini
sebagai fondasi microservice baru.

---

## Daftar Isi

1. [Architecture Overview](#architecture-overview)
2. [Memilih Mode Injector](#memilih-mode-injector)
3. [Multi-DC / Multi-Tenant Routing](#multi-dc--multi-tenant-routing)
4. [Timezone Configuration](#timezone-configuration)
5. [Developer Setup & goswitch](#developer-setup--goswitch)
6. [Project Structure](#project-structure)
7. [Menambahkan Domain Endpoint Baru](#menambahkan-domain-endpoint-baru)
8. [Background Worker Pool](#background-worker-pool)
9. [Production Checklist](#production-checklist)

---

## 1. Architecture Overview

```
HTTP Request
     |
     v
+--------------------------------------------------------------+
|  Global Middleware Stack                                      |
|  RequestID → Logger → Recover → Timeout → Favicon →          |
|  RateLimiter                                                  |
+-----------------------------+--------------------------------+
                              |
                              v
+--------------------------------------------------------------+
|  Router Group /api  (+  MultiDCMiddleware)                   |
|  * Resolve KodeDC dari request → GetOrInit DB di DcAdapter   |
|  * Simpan *postgres.Database ke c.Locals("dbLocal")          |
+-----------------------------+--------------------------------+
                              |  (DB siap diambil via DBInjector)
                              v
+--------------------------------------------------------------+
|  Handler Layer  (internal/api/handlers/)                     |
|  * Validasi HTTP request / marshal response                  |
|  * Panggil hr.XxxService.Build(c) untuk build service        |
|  * Submit task ke Pool jika diperlukan                       |
+------+---------------------------------------------+--------+
       |                                             |
       v                                             v
+--------------------+                   +---------------------+
|  Service Layer     |                   |  Worker Pool        |
|  (business logic,  |                   |  (background tasks) |
|   bcrypt, dll)     |                   +---------------------+
+------+-------------+
       |
       v
+--------------------------------------------------------------+
|  Repository Layer (internal/api/repository/)                 |
|  * Menerima *postgres.Database via constructor               |
|  * Raw SQL via pgx/scany — tanpa ORM                        |
+---------------------------+----------------------------------+
                            |
            +---------------+-----------------+
            v                                 v
     PostgreSQL (DC g009sim)         PostgreSQL (DC g010sim)
```

**Prinsip desain:**

- **Concrete types over unnecessary interfaces.** Handler menggunakan
  `*injector.ServiceFactory[T]`. Mock untuk unit test dilakukan di level
  repository — di situlah ketidakpastian implementasi benar-benar ada.
- **Errors wrap, never swallow.** Selalu gunakan `fmt.Errorf("context: %w", err)`.
- **Handler hanya urusan HTTP.** Business rules ada di Service layer.
- **DB adalah request-scoped** (via `LocalsInjector`) atau pre-configured
  (via `StaticInjector`). Lihat section [Memilih Mode Injector](#memilih-mode-injector).
- **ServiceFactory build service per-request.** `hr.XxxService.Build(c)` menyelesaikan
  koneksi DB dan membangun service secara atomik untuk setiap request.

---

## 2. Memilih Mode Injector

Boilerplate mendukung dua mode injeksi database yang dipilih di `main.go`:

### StaticInjector — Single-DC

Gunakan ini ketika aplikasi hanya terhubung ke **satu** database dan KodeDC
sudah diketahui saat startup.

```go
// main.go
dbInjector := injector.NewStaticInjector(dbManager, cfg.Kunci)
```

- Kunci diambil dari `appsettings.ini` (`Kunci = g009sim`)
- DB di-resolve sekali saat pertama diakses, lalu di-cache di `DcAdapter`
- `MultiDCMiddleware` **tidak perlu** didaftarkan di route group
- Cocok untuk: layanan yang hanya mengakses satu DC / satu tenant

### LocalsInjector — Multi-DC

Gunakan ini ketika KodeDC berbeda per request (multi-tenant).

```go
// main.go
dbInjector := injector.NewLocalsInjector()
```

- `MultiDCMiddleware` **wajib** didaftarkan di route group yang membutuhkan DB:

```go
// router.go
api := app.Group("/api", middleware.MultiDCMiddleware(log, cfg, dcAdapter))
```

- Middleware membaca kunci dari request (query param / header / nginx prefix)
  dan menyimpan `*postgres.Database` ke `c.Locals("dbLocal")`
- `LocalsInjector.GetDB(c)` membaca nilai tersebut
- Cocok untuk: layanan yang melayani multiple datacenter

### Konsekuensi Pilihan

| Mode | MultiDCMiddleware | Kunci | Skenario |
|------|-------------------|-------|----------|
| `StaticInjector` | Tidak diperlukan | Hardcode di `appsettings.ini` | Single-DC |
| `LocalsInjector` | **Wajib** di route group | Dari request header/query | Multi-DC |

> **Perhatian:** Jika `LocalsInjector` dipakai tanpa `MultiDCMiddleware`, setiap
> request ke endpoint yang butuh DB akan gagal dengan error
> `"no db in locals[dbLocal]"`.

---

## 3. Multi-DC / Multi-Tenant Routing

### Cara Kerja

`MultiDCMiddleware` berjalan di level route group `/api`. Ia me-resolve kunci tenant
dan memanggil `DcAdapter.GetOrInit()` — yang mengembalikan koneksi dari LRU cache
atau membuat koneksi baru jika belum ada. Koneksi disimpan ke `c.Locals("dbLocal")`.

### Prioritas Kunci (tertinggi → terendah)

| Prioritas | Sumber | Catatan |
|-----------|--------|---------|
| 1 | `?KodeDC=<key>` query parameter | Berguna untuk debug ad-hoc |
| 2 | `X-kode-dc: <key>` request header | Penggunaan service-to-service |
| 3 | `X-Forwarded-Prefix` nginx header | Segmen pertama path = kunci: `/g009sim/api` → `g009sim` |

### Contoh Konfigurasi Nginx

```nginx
location /g009sim/ {
    proxy_pass         http://go-api:8080/;
    proxy_set_header   Host               $host;
    proxy_set_header   X-Real-IP          $remote_addr;
    proxy_set_header   X-Forwarded-Prefix /g009sim;
}
```

### appsettings.ini untuk Multi-DC

```ini
[CONFIG]
Kunci = g009sim,g010sim   # connects to both DCs at startup
```

---

## 4. Timezone Configuration

### Urutan Prioritas (tertinggi → terendah)

| Prioritas | Sumber | Rekomendasi |
|-----------|--------|-------------|
| 1 | `TZ` environment variable | **Gunakan di production / Docker / K8s** |
| 2 | `Timezone` di `appsettings.ini` | Gunakan untuk development lokal |
| 3 | UTC | Default aman |

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

Jalankan `./setup.sh` (Linux/macOS) atau `setup.bat` (Windows) sekali setelah clone.

### Yang Dilakukan Script

1. **Go version check (goswitch)** — baca versi yang dibutuhkan dari `go.mod`.
   Jika system Go berbeda, install versi tepat via `golang.org/dl/go<version>`
   **tanpa menyentuh system Go**.
2. **Module download & tidy** — `go mod download && go mod tidy`.
3. **swag CLI** — install Swagger code-generator jika belum ada.
4. **Doc generation** — `swag init -g main.go -o docs`.
5. **Build** — `go build ./...` untuk konfirmasi project compile bersih.
6. **Tests** — `go test ./...`.

### goswitch — Cara Kerja

```
System Go : go1.23.0
Required  : go1.25.8
WARNING Version mismatch. Installing go1.25.8 via golang.org/dl...
        (Your system go1.23.0 will NOT be modified.)
OK Using go1.25.8 for this setup run.
   Tip: add $(go env GOPATH)/bin to PATH and run: go1.25.8 run main.go
```

Binary `go1.25.8` tersimpan di `$(go env GOPATH)/bin/` dan independen dari system Go.
Bisa ada banyak versi Go di mesin yang sama tanpa konflik.

---

## 6. Project Structure

```
.
+-- config/
|   +-- initialize.go           # Config struct + LoadConfigIni()
+-- docs/                       # Auto-generated swagger (jangan diedit)
+-- documentation/
|   +-- developer-handbook.md   # File ini
+-- internal/
|   +-- api/
|   |   +-- handlers/           # HTTP layer — HandlersRegistry, UserHandler, dst.
|   |   +-- models/             # Request / response DTOs
|   |   +-- repository/         # Data-access layer: struct dengan *postgres.Database
|   |   +-- router/             # Registrasi route
|   |   +-- services/           # Business logic layer
|   +-- database/
|   |   +-- adapter.go          # DcAdapter — LRU+TTL cache koneksi multi-DC
|   |   +-- db.go               # Registry, context helpers, InitDatabases
|   +-- middleware/
|   |   +-- initialize.go       # Urutan middleware global
|   |   +-- multidc.go          # MultiDCMiddleware + resolveKunci
|   |   +-- requestid.go        # X-Request-ID (UUIDv4)
|   |   +-- timeout.go          # Context deadline propagation (30 s)
|   |   +-- health-check.go     # /live + /ready (DB ping)
|   |   +-- recover.go          # Panic recovery + stack trace
|   |   +-- ratelimiter.go      # 100 req/min per IP
|   |   +-- logger.go           # HTTP request logging
|   |   +-- favicon.go          # Static favicon
|   +-- server/                 # Fiber app construction + lifecycle
|   +-- utility/
|   |   +-- fibererror/         # ResponseError envelope + GlobalErrorHandler
|   |   +-- injector/           # DBInjector, StaticInjector, LocalsInjector, ServiceFactory
|   |   +-- swagger/            # Swagger doc serving + proxy-path handling
|   |   +-- tztime/             # Timezone setup
|   +-- worker/                 # Bounded background worker pool
+-- build/
|   +-- build.sh / build.bat    # Cross-platform build
+-- static/public/              # Static files (favicon, halaman 404)
+-- appsettings.ini             # Konfigurasi runtime non-secret
+-- setup.sh / setup.bat        # Developer init scripts
+-- Dockerfile                  # Multi-stage production build
+-- main.go                     # Entry point: wire → start → graceful shutdown
```

---

## 7. Menambahkan Domain Endpoint Baru

### Contoh: User Profile

Misalnya kamu butuh endpoint `GET /api/users/:user_name/profile`.

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

    "github.com/boni-fm/go-libsd3/pkg/db/postgres"
)

type ProfileRepository struct {
    db *postgres.Database
}

func NewProfileRepository(db *postgres.Database) *ProfileRepository {
    return &ProfileRepository{db: db}
}

func (r *ProfileRepository) GetProfile(
    ctx context.Context, userName string,
) (*models.ProfileResponse, error) {
    if r.db == nil {
        return nil, database.ErrNoDB
    }
    var p models.ProfileResponse
    q := `SELECT user_name, display_name, email
          FROM dc_user_profile_t WHERE user_name = $1`
    if err := r.db.SelectOne(ctx, &p, q, userName); err != nil {
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

    "go-api-boilerplate/config"
    "go-api-boilerplate/internal/api/models"
    "go-api-boilerplate/internal/api/repository"

    "github.com/boni-fm/go-libsd3/pkg/db/postgres"
    "github.com/boni-fm/go-libsd3/pkg/log"
)

type ProfileService struct {
    db   *postgres.Database
    log_ *log.Logger
    cfg  *config.Config
    repo *repository.ProfileRepository
}

func NewProfileService(db *postgres.Database, log_ *log.Logger, cfg *config.Config) *ProfileService {
    return &ProfileService{
        db:   db,
        log_: log_,
        cfg:  cfg,
        repo: repository.NewProfileRepository(db),
    }
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
    svc, err := hr.ProfileService.Build(c)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fibererror.ResponseError{
            Code:    fiber.StatusInternalServerError,
            Error:   "Internal Server Error",
            Message: "Gagal initialize profile service",
        })
    }

    userName := c.Params("user_name")
    profile, err := svc.GetProfile(c.Context(), userName)
    if err != nil {
        hr.log_.Errorf("GetProfile error: %v", err)
        return c.Status(fiber.StatusInternalServerError).JSON(fibererror.ResponseError{
            Code:    fiber.StatusInternalServerError,
            Error:   "Internal Server Error",
            Message: "Failed to fetch profile",
        })
    }

    return c.Status(fiber.StatusOK).JSON(fiber.Map{
        "success": true,
        "data":    profile,
    })
}
```

---

### 7.5 Wire ke HandlersRegistry — `BaseHandler.go`

```go
type HandlersRegistry struct {
    log_       *log.Logger
    dbManager  *database.DcAdapter
    cfg        *config.Config
    dbInjector injector.DBInjector

    SwaggerDoc     *swagger.DocumentModifier
    UserService    *injector.ServiceFactory[services.UserService]
    ProfileService *injector.ServiceFactory[services.ProfileService]  // tambahkan ini
    Pool           *worker.Pool
}

func NewHandlersRegistry(log_ *log.Logger, pool *worker.Pool, manager *database.DcAdapter, cfg *config.Config, dbInject injector.DBInjector) *HandlersRegistry {
    return &HandlersRegistry{
        log_:       log_,
        dbManager:  manager,
        cfg:        cfg,
        SwaggerDoc: swagger.NewDocumentModifier(),
        Pool:       pool,
        UserService: injector.NewServiceFactory[services.UserService](
            dbInject, log_, cfg, services.NewUserService,
        ),
        ProfileService: injector.NewServiceFactory[services.ProfileService](  // tambahkan ini
            dbInject, log_, cfg, services.NewProfileService,
        ),
    }
}
```

---

### 7.6 Daftarkan Route — `router.go`

```go
api.Get("/users/:user_name/profile", handlersRegistry.GetProfile)
```

---

### 7.7 Tulis Tests — `profile_test.go`

```go
package handlers_test

import (
    "fmt"
    "net/http"
    "net/http/httptest"
    "testing"
    "time"

    "go-api-boilerplate/config"
    "go-api-boilerplate/internal/api/handlers"
    "go-api-boilerplate/internal/utility/injector"

    pgsd3 "github.com/boni-fm/go-libsd3/pkg/db/postgres"
    "github.com/boni-fm/go-libsd3/pkg/log"
    "github.com/gofiber/fiber/v3"
)

// mockInjector untuk unit test tanpa DB sungguhan.
type mockInjector struct{ err error }

func (m *mockInjector) GetDB(_ fiber.Ctx) (*pgsd3.Database, error) {
    return nil, m.err
}

var _ injector.DBInjector = (*mockInjector)(nil)

func TestGetProfile_DBError(t *testing.T) {
    l := log.NewLoggerWithFilename("test")
    inject := &mockInjector{err: fmt.Errorf("no db")}
    hr := handlers.NewHandlersRegistry(l, nil, nil, &config.Config{}, inject)

    app := fiber.New()
    app.Get("/api/users/:user_name/profile", hr.GetProfile)

    req := httptest.NewRequest(http.MethodGet, "/api/users/alice/profile", nil)
    resp, err := app.Test(req, fiber.TestConfig{Timeout: 5 * time.Second})
    if err != nil {
        t.Fatal(err)
    }
    if resp.StatusCode != http.StatusInternalServerError {
        t.Errorf("status: got %d, want %d", resp.StatusCode, http.StatusInternalServerError)
    }
}
```

---

## 8. Background Worker Pool

Pool (`internal/worker.Pool`) di-inject ke setiap `HandlersRegistry` saat startup.
Handler mengaksesnya via `hr.Pool`.

### Kapan Digunakan

- Audit logging ke secondary store
- Publish event ke message queue
- Flush in-memory metrics batch
- Kirim email konfirmasi

### Pattern

```go
func (hr *HandlersRegistry) CreateOrder(c fiber.Ctx) error {
    // ... validasi, panggil service, tulis ke DB ...

    if hr.Pool != nil {
        orderID := order.ID
        ok := hr.Pool.Submit(func(ctx context.Context) {
            if err := auditClient.Record(ctx, "order.created", orderID); err != nil {
                hr.log_.Errorf("audit record failed: %v", err)
            }
        })
        if !ok {
            hr.log_.Warnf("worker pool penuh -- audit event untuk order %s dibuang", orderID)
        }
    }

    return c.Status(fiber.StatusCreated).JSON(...)
}
```

### Tuning Kapasitas

| Konstanta | Default | Panduan |
|-----------|---------|---------|
| `defaultWorkerCount` | 4 | ~2x jumlah tipe background task |
| `defaultWorkerCapacity` | 128 | (peak task/detik) × (latensi maksimal yang bisa ditoleransi dalam detik) |

---

## 9. Production Checklist

- [ ] **Secrets** — kredensial DB diinjeksi via env var, **bukan** di `appsettings.ini`.
- [ ] **Swagger dimatikan** — set `IsDevelopment = false`.
- [ ] **Timezone** — set `TZ=<IANA timezone>` sebagai environment variable.
- [ ] **Mode Injector** — pilih `StaticInjector` (single-DC) atau `LocalsInjector` (multi-DC);
      pastikan `MultiDCMiddleware` aktif di route group jika menggunakan `LocalsInjector`.
- [ ] **Multi-DC** — verifikasi `Kunci` berisi semua DC yang diperlukan;
      konfirmasi nginx header `X-Forwarded-Prefix` sudah di-set.
- [ ] **TLS** — terminasi di load balancer atau konfigurasi `app.ListenTLS`.
- [ ] **Rate limits** — tune `RateLimiter` sesuai peak RPS yang diharapkan.
- [ ] **DB pool size** — tune pgx `MinConns` / `MaxConns` vs `max_connections` PostgreSQL.
- [ ] **Worker pool size** — tune dari hasil load testing.
- [ ] **Graceful shutdown** — verifikasi urutan `ShutdownWithTimeout` → `pool.Stop()` →
      `dbManager.CloseAll()` ada di `main.go`.
- [ ] **Error messages** — konfirmasi tidak ada internal error string yang bocor ke client.
- [ ] **Logging** — pertimbangkan migrasi ke structured JSON logging (zerolog / zap) untuk production.
- [ ] **Health probes** — hubungkan `/live` → `livenessProbe`, `/ready` → `readinessProbe`.
- [ ] **Tests** — jalankan `go test -race ./...` sebelum setiap merge.
