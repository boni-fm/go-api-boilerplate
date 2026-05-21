package middleware

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/requestid"
	"github.com/sirupsen/logrus"
)

// LoggerMiddleware
// Dengan ini per request akan di log status nya
// Jadi lebih keliatan traffic dari api nya ...
func LoggerMiddleware(log *logrus.Logger) fiber.Handler {
	return func(c fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		log.WithFields(logrus.Fields{
			"request_id": requestid.FromContext(c),
			"method":     c.Method(),
			"status":     c.Response().StatusCode(),
			"path":       c.Path(),
			"route":      c.Route().Path,
			"url":        c.OriginalURL(),
			"ip":         c.IP(),
			"latency":    time.Since(start).String(),
		}).Info("http :: request completed")
		return err
	}
}
