# Go API Boilerplate

A production-ready Go REST API boilerplate built on top of the [Fiber](https://gofiber.io/) framework. This template gives you a solid starting point with pre-wired middleware, PostgreSQL integration, auto-generated Swagger documentation, and a clean layered architecture.

---

## Features

- **[Fiber v2](https://gofiber.io/)** – Fast, Express-inspired HTTP framework
- **PostgreSQL** – Database integration via `pgx` / `go-libsd3`
- **Swagger / OpenAPI** – Auto-generated docs with proxy-path support
- **Structured logging** – File-based rotating logs via `logrus` + `go-libsd3`
- **Rate limiting** – 100 requests per minute per client
- **Panic recovery** – Catches panics and returns a clean 500 JSON response
- **Health check** – Liveness probe at `/live`
- **Reverse-proxy aware** – Reads `X-Forwarded-Prefix` to adjust Swagger base path
- **INI-based config** – Simple `appsettings.ini` configuration
- **Layered architecture** – Handler → Service → Repository separation

---

## Prerequisites

| Tool | Version |
|------|---------|
| Go   | 1.24+   |
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

### 2. Install dependencies

```bash
go mod download
```

### 3. Configure the application

Edit `appsettings.ini`:

```ini
[CONFIG]
AppName       = Go API Boilerplate
IsDevelopment = true
Port          = 8080
Kunci         = <your-database-key>
```

| Key | Description |
|-----|-------------|
| `AppName` | Application name shown in logs and Swagger UI |
| `IsDevelopment` | Set `true` to enable Swagger auto-generation on startup |
| `Port` | HTTP port the server listens on |
| `Kunci` | Database credential key passed to `go-libsd3` |

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
├── go.mod / go.sum
│
├── config/
│   └── initialize.go              # Config struct & INI loader
│
├── docs/                          # Auto-generated Swagger files (do not edit)
│   ├── docs.go
│   ├── swagger.json
│   └── swagger.yaml
│
├── internal/
│   ├── api/
│   │   ├── handlers/              # HTTP handler functions
│   │   │   ├── BaseHandler.go     # HandlersRegistry (dependency container)
│   │   │   ├── PingHandler.go
│   │   │   ├── UserHandler.go
│   │   │   └── DocumentationSwaggerHandler.go
│   │   ├── models/                # Request / response structs
│   │   │   ├── PingPongModels.go
│   │   │   └── UserModels.go
│   │   ├── repository/            # Database access layer
│   │   │   └── UserRepo.go
│   │   ├── router/
│   │   │   └── router.go          # Route registration
│   │   └── services/              # Business logic
│   │       ├── PingPongService.go
│   │       └── UserService.go
│   │
│   ├── database/
│   │   └── db.go                  # PostgreSQL connection initializer
│   │
│   ├── middleware/
│   │   ├── initialize.go          # Middleware wiring
│   │   ├── favicon.go
│   │   ├── health-check.go
│   │   ├── logger.go
│   │   ├── ratelimiter.go
│   │   └── recover.go
│   │
│   └── utility/
│       ├── fibererror/            # Global error handler + helpers
│       │   └── fiber-error.go
│       └── swagger/               # Swagger generation & proxy utilities
│           ├── modifier.go
│           ├── proxypass.go
│           └── swagger.go
│
└── static/
    └── public/                    # Static assets (favicon, 404 page, …)
```

---

## API Endpoints

### General

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/` | Redirects to Swagger UI |
| `GET` | `/ping` | Liveness check – returns `{ "message": "Pong" }` |
| `GET` | `/live` | Health-check probe (returns `200 OK`) |
| `GET` | `/swagger` | Interactive Swagger UI |
| `GET` | `/swagger/doc.json` | Raw OpenAPI JSON |

### Users (example CRUD)

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/api/users` | Create a new user |
| `GET` | `/api/users` | List all users |
| `PUT` | `/api/users/:user_name/password` | Update a user's password |
| `DELETE` | `/api/users/:user_name` | Delete a user |

#### Create User – `POST /api/users`

```json
// Request
{ "user_name": "john_doe", "password": "secret123" }

// Response 201
{ "success": true, "message": "User created", "user": "john_doe" }
```

#### Get All Users – `GET /api/users`

```json
// Response 200
{ "success": true, "data": [ { "user_name": "john_doe" } ] }
```

#### Update Password – `PUT /api/users/:user_name/password`

```json
// Request
{ "new_password": "newSecret456" }

// Response 200
{ "success": true, "message": "Password updated" }
```

#### Delete User – `DELETE /api/users/:user_name`

```json
// Response 200
{ "success": true, "message": "User deleted" }
```

---

## Adding a New Feature

Follow the existing layer pattern:

1. **Model** – add request/response structs in `internal/api/models/`
2. **Repository** – add database queries in `internal/api/repository/`
3. **Service** – add business logic in `internal/api/services/`
4. **Handler** – add HTTP handler methods on `*HandlersRegistry` in `internal/api/handlers/`
5. **Router** – register routes in `internal/api/router/router.go`

### Example: Adding a "product" resource

```bash
# Create files following the naming convention
touch internal/api/models/ProductModels.go
touch internal/api/repository/ProductRepo.go
touch internal/api/services/ProductService.go
touch internal/api/handlers/ProductHandler.go
```

Inject the new service into `HandlersRegistry` inside `BaseHandler.go`:

```go
type HandlersRegistry struct {
    // ...existing fields...
    ProductService *services.ProductService
}

func NewHandlersRegistry(log_ *log.Logger, ctx context.Context) *HandlersRegistry {
    return &HandlersRegistry{
        // ...existing fields...
        ProductService: services.NewProductService(log_, ctx),
    }
}
```

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
func (hr *HandlersRegistry) CreateProduct(c *fiber.Ctx) error { ... }
```

To regenerate docs manually:

```bash
swag init -g main.go -o docs
```

---

## Middleware

All middleware is registered in `internal/middleware/initialize.go`:

| Middleware | Description |
|------------|-------------|
| **Logger** | Structured HTTP request/response logging (logrus) |
| **Recover** | Catches panics, logs them, returns `500` JSON |
| **HealthCheck** | `GET /live` → `200 OK` liveness probe |
| **Favicon** | Serves `/domar.ico` from `static/public/favicon.ico` |
| **RateLimiter** | 100 requests / 60 seconds per IP |

---

## Configuration Reference

`appsettings.ini` (all keys live under `[CONFIG]`):

| Key | Default | Description |
|-----|---------|-------------|
| `AppName` | `Go API Boilerplate` | Application name |
| `IsDevelopment` | `false` | Enables Swagger auto-gen on startup |
| `Port` | `8080` | HTTP listen port |
| `Kunci` | _(required)_ | Database credential key for `go-libsd3` |

---

## License

This project is provided as an open boilerplate template. Feel free to use, modify, and distribute it for your own projects.
