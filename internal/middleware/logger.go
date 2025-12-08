package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/mikhail-bigun/fiberlogrus"
	"github.com/sirupsen/logrus"
)

func LoggerMiddleware(log *logrus.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		fiberlogrus.New(
			fiberlogrus.Config{
				Logger: log,
				Tags: []string{
					fiberlogrus.TagMethod,
					fiberlogrus.TagStatus,
					fiberlogrus.TagPath,
					fiberlogrus.TagRoute,
					fiberlogrus.TagURL,
					fiberlogrus.TagIP,
				},
			},
		)
		return c.Next()
	}
}
