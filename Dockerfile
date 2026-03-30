# ── Stage 1: Build ───────────────────────────────────────────────────────────
FROM golang:1.25-alpine AS builder

# Install ca-certs and timezone data so the final image can validate TLS
# certificates and use the Asia/Jakarta timezone correctly.
RUN apk add --no-cache ca-certificates tzdata

WORKDIR /src

# Cache dependency downloads: copy go.mod/go.sum first, then download.
COPY go.mod go.sum ./
RUN go mod download

# Copy everything else and build a fully-static binary (CGO disabled).
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" -o /app .

# ── Stage 2: Runtime ────────────────────────────────────────────────────────
FROM alpine:3.21

# Non-root user for security.
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# TLS root certs + timezone data (for Asia/Jakarta).
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Application binary and static assets.
COPY --from=builder /app /app
COPY --from=builder /src/appsettings.ini /appsettings.ini
COPY --from=builder /src/static /static
COPY --from=builder /src/docs /docs

# Set timezone to WIB (Asia/Jakarta) as required by ITSD3/SD3 standards.
ENV TZ=Asia/Jakarta

WORKDIR /
USER appuser

EXPOSE 8080

# Health checks: liveness at /live, readiness at /ready.
HEALTHCHECK --interval=10s --timeout=3s --retries=3 \
    CMD wget -qO- http://localhost:8080/live || exit 1

ENTRYPOINT ["/app"]
