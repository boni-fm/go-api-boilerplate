# Go API Boilerplate

Template REST API berbasis [Go Fiber v3](https://gofiber.io/) untuk departemen ITSD3/SD3.
Middleware, PostgreSQL multi-tenant, Swagger, dan arsitektur berlapis sudah terpasang.

---

## Fitur

- **Fiber v3** — framework HTTP cepat berbasis fasthttp
- **PostgreSQL** — koneksi via `pgx` / `go-libsd3`, tanpa ORM
- **Multi-DC / Multi-Tenant** — `DcAdapter` (LRU+TTL) mengelola koneksi per-KodeDC
- **Swagger / OpenAPI** — auto-generate saat startup, aktif hanya di mode development
- **Structured logging** — logrus + rotating file via `go-libsd3`
- **Rate limiting** — 100 req/menit per IP
- **Panic recovery** — tangkap panic, kembalikan JSON 500 yang rapi
- **Request tracing** — `X-Request-ID` UUIDv4 di setiap request
- **Context timeout** — deadline 30 detik per request, propagate ke query DB
- **Health probes** — `/live` (proses berjalan) dan `/ready` (DB ping)
- **Background worker pool** — goroutine pool terbatas untuk task fire-and-forget
- **Docker support** — multi-stage build, image minimal Alpine, non-root user
- **INI config** — `appsettings.ini`, sederhana tanpa enkripsi
- **Timezone** — override via `TZ` env var atau `appsettings.ini`
- **Developer init scripts** — `setup.sh` / `setup.bat` install Go versi tepat tanpa ubah system Go

---

## Mulai

### 1. Clone repositori

```bash
git clone https://github.com/boni-fm/go-api-boilerplate.git
cd go-api-boilerplate
```

### 2. Jalankan setup script

```bash
# Linux / macOS
chmod +x setup.sh && ./setup.sh

# Windows
setup.bat
```

Script memeriksa versi Go, menginstall versi tepat via `golang.org/dl` (tanpa menyentuh system Go),
mengunduh modul, menginstall `swag`, generate docs, build, dan run tests.
Detail goswitch → [`documentation/developer-handbook.md`](documentation/developer-handbook.md).

### 3. Konfigurasi

Edit `appsettings.ini`:

```ini
[CONFIG]
AppName       = Go API Boilerplate
IsDevelopment = true
Port          = 8080
Kunci         = g009sim          # pisahkan koma untuk multi-DC: g009sim,g010sim
Timezone      = Asia/Jakarta     # IANA timezone; di-override oleh TZ env var
```

| Key | Default | Keterangan |
|-----|---------|------------|
| `AppName` | `Go API Boilerplate` | Nama aplikasi di log dan Swagger UI |
| `IsDevelopment` | `false` | `true` untuk aktifkan Swagger UI; **matikan di production** |
| `Port` | `8080` | Port HTTP server |
| `Kunci` | *(wajib)* | Kunci lookup koneksi database di `go-libsd3` |
| `Timezone` | `Asia/Jakarta` | IANA timezone; di-override oleh `TZ` env var |

### 4. Jalankan aplikasi

```bash
go run main.go
```

Server berjalan di `http://localhost:8080`. Swagger UI di `http://localhost:8080/swagger`.

---

## Struktur Proyek

```
go-api-boilerplate/
├── main.go                        # Entry point: wire → start → graceful shutdown
├── appsettings.ini                # Konfigurasi runtime
├── setup.sh / setup.bat           # Developer init scripts
├── build/
│   ├── build.sh                   # Cross-platform build (Linux / macOS)
│   └── build.bat                  # Cross-platform build (Windows)
├── Dockerfile                     # Multi-stage production Docker build
├── config/
│   └── initialize.go              # Config struct & INI loader
├── docs/                          # File Swagger auto-generated (jangan diedit)
├── documentation/
│   └── developer-handbook.md      # Panduan arsitektur & pengembangan
├── internal/
│   ├── api/
│   │   ├── handlers/              # HTTP handler — BaseHandler.go, UserHandler.go, dst.
│   │   ├── models/                # Request / response structs
│   │   ├── repository/            # Database access layer
│   │   ├── router/router.go       # Registrasi route
│   │   └── services/              # Business logic
│   ├── database/
│   │   ├── adapter.go             # DcAdapter — LRU+TTL cache koneksi per-KodeDC
│   │   └── db.go                  # Registry, context helpers, InitDatabases
│   ├── middleware/
│   │   ├── initialize.go          # Urutan registrasi middleware global
│   │   ├── multidc.go             # MultiDCMiddleware — resolve KodeDC per request
│   │   └── ...                    # requestid, timeout, logger, ratelimiter, dst.
│   ├── server/server.go
│   ├── worker/pool.go
│   └── utility/
│       ├── fibererror/            # ResponseError envelope + GlobalErrorHandler
│       ├── injector/              # DBInjector, StaticInjector, LocalsInjector, ServiceFactory
│       ├── swagger/               # Swagger doc serving + proxy-path handling
│       └── tztime/                # Timezone setup
└── static/public/                 # Static assets (favicon, halaman 404)
```

---

## Endpoint API

### Umum

| Method | Path | Keterangan |
|--------|------|------------|
| `GET` | `/` | Redirect ke Swagger UI *(dev only)* |
| `GET` | `/ping` | Returns `{ "message": "Pong" }` |
| `GET` | `/live` | Liveness probe — selalu `200 OK` |
| `GET` | `/ready` | Readiness probe — `200 OK` jika DB reachable |
| `GET` | `/swagger` | Swagger UI *(dev only)* |
| `GET` | `/swagger/doc.json` | Raw OpenAPI JSON *(dev only)* |

### Pengguna (contoh CRUD)

| Method | Path | Keterangan |
|--------|------|------------|
| `POST` | `/api/users` | Buat user baru |
| `GET` | `/api/users` | Ambil semua user |
| `PUT` | `/api/users/:user_name/password` | Update password |
| `DELETE` | `/api/users/:user_name` | Hapus user |

---

## Multi-DC / Multi-Tenant

`MultiDCMiddleware` di-register di route group `/api`. Prioritas kunci (**tertinggi → terendah**):

| Prioritas | Sumber | Contoh |
|-----------|--------|--------|
| 1 | Query parameter `?KodeDC=` | `GET /api/users?KodeDC=g010sim` |
| 2 | Request header `X-kode-dc` | `X-kode-dc: g010sim` |
| 3 | Nginx `X-Forwarded-Prefix` (segmen pertama) | `/g010sim/api` → `g010sim` |

**Contoh nginx:**

```nginx
location /g009sim/ {
    proxy_pass         http://go-api:8080/;
    proxy_set_header   X-Forwarded-Prefix /g009sim;
}
```

Multi-DC di `appsettings.ini`: `Kunci = g009sim,g010sim`

---

## Middleware

Urutan middleware global (`internal/middleware/initialize.go`):

| # | Middleware | Keterangan |
|---|------------|------------|
| 1 | **RequestID** | Generate `X-Request-ID` UUIDv4 |
| 2 | **Logger** | Logging HTTP request/response terstruktur |
| 3 | **Recover** | Tangkap panic, log stack trace, kembalikan JSON 500 |
| 4 | **Timeout** | Deadline 30 detik per request |
| 5 | **Favicon** | Serve favicon dari `static/public/` |
| 6 | **RateLimiter** | 100 req / 60 detik per IP |
| * | **MultiDCMiddleware** | Di route group `/api` — resolve DB per request |

---

## Menambahkan Fitur Baru

1. **Model** — tambah request/response struct di `internal/api/models/`
2. **Repository** — tambah query database di `internal/api/repository/`
3. **Service** — tambah business logic di `internal/api/services/`
4. **Handler** — tambah method di `internal/api/handlers/`, gunakan `hr.XxxService.Build(c)`
5. **Router** — daftarkan route di `internal/api/router/router.go`

Panduan lengkap dengan contoh kode → [`documentation/developer-handbook.md`](documentation/developer-handbook.md).

---

## Docker

```bash
docker build -t go-api-boilerplate .
docker run -p 8080:8080 -e TZ=Asia/Jakarta go-api-boilerplate
```

Build dua tahap: **builder** (`golang:1.25-alpine`) → **runtime** (`alpine:3.21`, non-root user).

---

## Background Worker Pool

```go
if hr.Pool != nil {
    ok := hr.Pool.Submit(func(ctx context.Context) {
        // audit log, publish event, flush metrik, kirim email, dll.
    })
    if !ok {
        hr.log_.Warn("worker pool penuh — task dibuang")
    }
}
```

Tune di `internal/server/server.go`: `defaultWorkerCount = 4`, `defaultWorkerCapacity = 128`.
