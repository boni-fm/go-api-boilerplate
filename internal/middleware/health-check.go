package middleware

import (
	"context"

	"go-api-boilerplate/internal/database"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/healthcheck"
)

// LivenessHandler cek apakah proses masih hidup.
// Selalu return true — kalau endpoint ini bisa dipanggil, berarti app nyala.
func LivenessHandler() fiber.Handler {
	return healthcheck.New(healthcheck.Config{
		Probe: func(_ fiber.Ctx) bool {
			return true
		},
	})
}

// ReadinessHandler cek apakah app siap nerima traffic.
// Ping semua koneksi DB yang aktif — kalau ada yang gagal, return false (503).
func ReadinessHandler(adapter *database.DcAdapter) fiber.Handler {
	return healthcheck.New(healthcheck.Config{
		Probe: func(c fiber.Ctx) bool {
			results := adapter.HealthCheck(context.Background())
			for _, err := range results {
				if err != nil {
					return false
				}
			}
			return true
		},
	})
}
