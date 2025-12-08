package fibererror

import "github.com/gofiber/fiber/v2"

func GlobalErrorHandler(c *fiber.Ctx, err error) error {
	if e, ok := err.(*fiber.Error); ok {
		switch e.Code {
		case fiber.StatusBadRequest:
			return BadRequestError(err)(c)
		case fiber.StatusGatewayTimeout:
			return GateawayTimeoutError(err)(c)
		case fiber.StatusNotFound:
			return NotFoundError(c)
		case fiber.StatusInternalServerError:
			return InternalServerError(err)(c)
		}
	}
	return InternalServerError(err)(c)
}

func InternalServerError(err error) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    fiber.StatusInternalServerError,
			"error":   "Internal Server Error",
			"message": err.Error(),
		})
	}
}

func BadRequestError(err error) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    fiber.StatusBadRequest,
			"error":   "Bad Request",
			"message": err.Error(),
		})
	}
}

func GateawayTimeoutError(err error) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusGatewayTimeout).JSON(fiber.Map{
			"code":    fiber.StatusGatewayTimeout,
			"error":   "Gateway Timeout - Upstream service timed out",
			"message": "The upstream service did not respond in time.",
		})
	}
}

func NotFoundError(c *fiber.Ctx) error {
	return c.Status(404).SendFile("./static/public/404.html")
}
