// Package fibererror nyediain tipe error standar dan handler error buat Fiber.
// Semua response error harus pakai ResponseError biar formatnya konsisten.
package fibererror

import (
	"net/http"

	"github.com/gofiber/fiber/v3"
)

// ResponseError itu amplop JSON yang dikembalikan ke client kalau ada error.
// Client bisa baca field Code (ngikutin HTTP status) buat nentuin gimana nanganinya.
type ResponseError struct {
	Code    int    `json:"code"`
	Error   string `json:"error"`
	Message string `json:"message"`
}

// GlobalErrorHandler itu error handler terpusat Fiber, didaftarin lewat fiber.Config.ErrorHandler.
// Tugasnya ngubah error apapun dari handler/middleware jadi response JSON yang rapi.
//
// Penting: switch harus ada default case biar status code-nya dipreserve dengan bener.
// Tanpa itu, error kayak 401, 403, 429, dll bakal kereturn sebagai 500 — susah debugnya.
//
// Pesan error yang dikasih ke client harus aman — jangan pernah expose raw error Go
// (misalnya error SQL atau stack trace). Log aja di server.
func GlobalErrorHandler(c fiber.Ctx, err error) error {
	if e, ok := err.(*fiber.Error); ok {
		switch e.Code {
		case fiber.StatusNotFound:
			return NotFoundError(c)
		default:
			// Pakai status code aslinya dan teks HTTP standar sebagai label error.
			return c.Status(e.Code).JSON(ResponseError{
				Code:    e.Code,
				Error:   http.StatusText(e.Code),
				Message: e.Message,
			})
		}
	}

	// Error non-Fiber (dari handler/service) dianggap 500.
	// Jangan expose err.Error() ke client — bisa bocor info internal.
	return c.Status(fiber.StatusInternalServerError).JSON(ResponseError{
		Code:    fiber.StatusInternalServerError,
		Error:   "Internal Server Error",
		Message: "An unexpected error occurred. Please try again later.",
	})
}

// InternalServerError balikin response 500 dengan pesan yang bisa diatur.
// Pesannya harus aman — jangan masukin raw error Go, log aja di server.
func InternalServerError(c fiber.Ctx, message string) error {
	return c.Status(fiber.StatusInternalServerError).JSON(ResponseError{
		Code:    fiber.StatusInternalServerError,
		Error:   "Internal Server Error",
		Message: message,
	})
}

// BadRequestError balikin response 400 dengan pesan yang bisa diatur.
func BadRequestError(c fiber.Ctx, message string) error {
	return c.Status(fiber.StatusBadRequest).JSON(ResponseError{
		Code:    fiber.StatusBadRequest,
		Error:   "Bad Request",
		Message: message,
	})
}

// GatewayTimeoutError balikin response 504.
// Pesannya selalu generic — biar info soal backend gak bocor ke client.
func GatewayTimeoutError(c fiber.Ctx) error {
	return c.Status(fiber.StatusGatewayTimeout).JSON(ResponseError{
		Code:    fiber.StatusGatewayTimeout,
		Error:   "Gateway Timeout",
		Message: "The upstream service did not respond in time.",
	})
}

// NotFoundError serve halaman 404 statis.
// Kalau file HTML-nya gak ketemu (misal working directory salah pas deploy),
// fallback ke JSON biar client tetep dapet response yang bener.
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
