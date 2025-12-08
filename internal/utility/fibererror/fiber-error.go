package fibererror

import "github.com/gofiber/fiber/v2"

func InternalServerError(c *fiber.Ctx, err error) error {
	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
		"code":    fiber.StatusInternalServerError,
		"error":   "Internal Server Error",
		"message": err.Error(),
	})
}

func BadRequestError(c *fiber.Ctx, err error) error {
	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
		"code":    fiber.StatusBadRequest,
		"error":   "Bad Request - Failed to connect upstream service",
		"message": err.Error(),
	})
}

func GateawayTimeoutError(c *fiber.Ctx) error {
	return c.Status(fiber.StatusGatewayTimeout).JSON(fiber.Map{
		"code":    fiber.StatusGatewayTimeout,
		"error":   "Gateway Timeout - Upstream service timed out",
		"message": "The upstream service did not respond in time.",
	})
}

func NotFoundError(c *fiber.Ctx) error {
	return c.Status(404).SendFile("./static/public/404.html")
}
