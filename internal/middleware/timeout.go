package middleware

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v3"
)

// defaultRequestTimeout batas waktu maksimal tiap request handler boleh jalan.
// Kode downstream (query DB, HTTP call ke service lain) harus respect ctx.Done()
// dan balik duluan kalau deadline udah lewat.
//
// Route yang butuh waktu lebih lama (misal generate laporan)
// bisa override ini di handler masing-masing dengan bikin derived context sendiri.
const defaultRequestTimeout = 120 * time.Second

// TimeoutMiddleware wrap context tiap request dengan deadline.
// Supaya query DB dan I/O lain otomatis di-cancel kalau handler kebanyakan makan waktu.
//
// Kode downstream (service, repo) harus nerima dan respect context ini
// biar deadline-nya nge-propagate dengan bener.
func TimeoutMiddleware(timeout time.Duration) fiber.Handler {
	if timeout <= 0 {
		timeout = defaultRequestTimeout
	}
	return func(c fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.Context(), timeout)
		defer cancel()

		c.SetContext(ctx)
		return c.Next()
	}
}
