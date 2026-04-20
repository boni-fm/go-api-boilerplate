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
//
// ARC-001 / ARC-007: error messages returned to clients are the canonical HTTP
// reason phrase — never raw Go error strings. Internal details (SQL errors,
// stack traces, etc.) must only appear in server-side logs. The previous
// handler-factory indirection (e.g. BadRequestError(err)(c)) has been inlined
// for clarity: the GlobalErrorHandler itself is already a handler, so wrapping
// the response in an additional fiber.Handler adds no value.
func GlobalErrorHandler(c fiber.Ctx, err error) error {
	if e, ok := err.(*fiber.Error); ok {
		switch e.Code {
		case fiber.StatusNotFound:
			return NotFoundError(c)
		default:
			// Preserve the exact HTTP status code and use the canonical HTTP
			// reason phrase as the error label. The Message field carries the
			// Fiber error message which, for framework-generated errors, is
			// already a safe, human-readable string.
			return c.Status(e.Code).JSON(ResponseError{
				Code:    e.Code,
				Error:   http.StatusText(e.Code),
				Message: e.Message,
			})
		}
	}

	// Non-Fiber errors (returned by handler/service code) are treated as 500.
	// ARC-001: NEVER expose err.Error() to clients — it may contain SQL
	// queries, internal paths, or stack traces. Use a generic message.
	return c.Status(fiber.StatusInternalServerError).JSON(ResponseError{
		Code:    fiber.StatusInternalServerError,
		Error:   "Internal Server Error",
		Message: "An unexpected error occurred. Please try again later.",
	})
}

// InternalServerError is a convenience helper for handlers that need to return
// a generic 500 response with a caller-specified safe message. The message
// must NOT contain raw Go error strings — log those server-side instead.
func InternalServerError(c fiber.Ctx, message string) error {
	return c.Status(fiber.StatusInternalServerError).JSON(ResponseError{
		Code:    fiber.StatusInternalServerError,
		Error:   "Internal Server Error",
		Message: message,
	})
}

// BadRequestError is a convenience helper for handlers that need to return a
// 400 response with a caller-specified safe message.
func BadRequestError(c fiber.Ctx, message string) error {
	return c.Status(fiber.StatusBadRequest).JSON(ResponseError{
		Code:    fiber.StatusBadRequest,
		Error:   "Bad Request",
		Message: message,
	})
}

// GatewayTimeoutError is a convenience helper for handlers that need to return
// a 504 response. The message is always a generic upstream-timeout string to
// avoid leaking backend topology information.
func GatewayTimeoutError(c fiber.Ctx) error {
	return c.Status(fiber.StatusGatewayTimeout).JSON(ResponseError{
		Code:    fiber.StatusGatewayTimeout,
		Error:   "Gateway Timeout",
		Message: "The upstream service did not respond in time.",
	})
}

// NotFoundError responds with HTTP 404 by serving the static 404 page.
// If the HTML file is unavailable (e.g. incorrect working directory at
// deployment), it falls back to a structured JSON response so the caller
// still receives a well-formed 404 instead of an unhandled error that
// would cause GlobalErrorHandler to recurse.
func NotFoundError(c fiber.Ctx) error {
	if err := c.Status(fiber.StatusNotFound).SendFile("./static/public/404.html"); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(ResponseError{
			Code:    fiber.StatusNotFound,
			Error:   "Not Found",
			Message: "The requested resource was not found.",
		})
	}
	return nil
}
