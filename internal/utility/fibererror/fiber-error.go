// Package fibererror provides the standard error types and Fiber error handlers
// used across all API endpoints. Every error response must use ResponseError so
// that clients can rely on a consistent JSON envelope.
package fibererror

import (
	"net/http"

	"github.com/gofiber/fiber/v3"
)

// ResponseError is the canonical error envelope returned by every API endpoint.
// Clients should key off the Code field (mirrors the HTTP status) to branch their
// error-handling logic.
type ResponseError struct {
	Code    int    `json:"code"`
	Error   string `json:"error"`
	Message string `json:"message"`
}

// GlobalErrorHandler is Fiber's central error handler, registered via
// fiber.Config.ErrorHandler. It converts any error returned by a route handler
// or middleware into a structured JSON response using ResponseError.
//
// Critical: the switch must have a default case that preserves the original HTTP
// status code. Without it, any Fiber error not explicitly listed (e.g. 401, 403,
// 405, 408, 413, 429) would be incorrectly returned as HTTP 500, making real
// client errors impossible to diagnose in production.
func GlobalErrorHandler(c fiber.Ctx, err error) error {
	if e, ok := err.(*fiber.Error); ok {
		switch e.Code {
		case fiber.StatusBadRequest:
			return BadRequestError(err)(c)
		case fiber.StatusGatewayTimeout:
			return GatewayTimeoutError(err)(c)
		case fiber.StatusNotFound:
			return NotFoundError(c)
		case fiber.StatusInternalServerError:
			return InternalServerError(err)(c)
		default:
			// Preserve the exact HTTP status code for all other Fiber errors.
			// http.StatusText maps code → canonical reason phrase (e.g. 401 →
			// "Unauthorized", 403 → "Forbidden", 413 → "Request Entity Too Large").
			return c.Status(e.Code).JSON(ResponseError{
				Code:    e.Code,
				Error:   http.StatusText(e.Code),
				Message: e.Message,
			})
		}
	}
	return InternalServerError(err)(c)
}

// InternalServerError returns a handler that responds with HTTP 500 and a
// ResponseError body containing the error detail.
func InternalServerError(err error) fiber.Handler {
	return func(c fiber.Ctx) error {
		return c.Status(fiber.StatusInternalServerError).JSON(ResponseError{
			Code:    fiber.StatusInternalServerError,
			Error:   "Internal Server Error",
			Message: err.Error(),
		})
	}
}

// BadRequestError returns a handler that responds with HTTP 400 and a
// ResponseError body containing the error detail.
func BadRequestError(err error) fiber.Handler {
	return func(c fiber.Ctx) error {
		return c.Status(fiber.StatusBadRequest).JSON(ResponseError{
			Code:    fiber.StatusBadRequest,
			Error:   "Bad Request",
			Message: err.Error(),
		})
	}
}

// GatewayTimeoutError returns a handler that responds with HTTP 504 and a
// ResponseError body with a standard upstream-timeout message.
//
// The Error field uses the canonical HTTP reason phrase "Gateway Timeout"
// (matching net/http.StatusText(504)) for consistency with the default case
// in GlobalErrorHandler. Descriptive context is carried by the Message field.
func GatewayTimeoutError(err error) fiber.Handler {
	return func(c fiber.Ctx) error {
		return c.Status(fiber.StatusGatewayTimeout).JSON(ResponseError{
			Code:    fiber.StatusGatewayTimeout,
			Error:   "Gateway Timeout",
			Message: "The upstream service did not respond in time.",
		})
	}
}

// NotFoundError responds with HTTP 404 by serving the static 404 page.
func NotFoundError(c fiber.Ctx) error {
	return c.Status(fiber.StatusNotFound).SendFile("./static/public/404.html")
}
