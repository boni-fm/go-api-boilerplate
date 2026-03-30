package middleware

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/sirupsen/logrus"
)

// LoggerMiddleware returns an HTTP request/response logger that writes to the
// provided logrus.Logger. It replaces the fiberlogrus dependency (which was
// fiber/v2-only) with a minimal custom handler that preserves the same log
// fields while remaining compatible with Fiber v3.
//
// Logged fields: method, status, path, ip, latency (matching the previous
// fiberlogrus tag set).
func LoggerMiddleware(log *logrus.Logger) fiber.Handler {
	return func(c fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		log.WithFields(logrus.Fields{
			"method":  c.Method(),
			"status":  c.Response().StatusCode(),
			"path":    c.Path(),
			"route":   c.Route().Path,
			"url":     c.OriginalURL(),
			"ip":      c.IP(),
			"latency": time.Since(start).String(),
		}).Info("http")
		return err
	}
}
