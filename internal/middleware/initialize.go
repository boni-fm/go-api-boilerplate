package middleware

import (
	"github.com/boni-fm/go-libsd3/pkg/log"
	"github.com/gofiber/fiber/v3"
)

type MiddlewareDependencies struct {
	Log           *log.Logger
	App           *fiber.App
	IsDevelopment bool
}

func NewMiddlewareDependencies(log *log.Logger, app *fiber.App, isDevelopment bool) *MiddlewareDependencies {
	return &MiddlewareDependencies{
		Log:           log,
		App:           app,
		IsDevelopment: isDevelopment,
	}
}

// InitAllMiddleware daftarin semua global middleware dengan urutan yang bener.
//
// Urutan penting banget:
//  1. RequestID   — harus paling duluan, biar semua log dan error response punya correlation ID.
//  2. Logger      — log request setelah RequestID udah ada.
//  3. Recover     — nangkep panic dan balikin error response yang rapi.
//  4. Timeout     — kasih deadline ke context tiap request, biar query DB gak nge-hang selamanya.
//  5. Favicon     — serve favicon langsung, gak perlu lewat router.
//  6. RateLimiter — lindungi handler dari request yang kebanyakan.
//
// Catatan: health check endpoint (GET /live) didaftarin sebagai route biasa di router,
// bukan middleware. RateLimiter skip path ini via fungsi Next-nya biar probe gak ikut ke-throttle.
func (md *MiddlewareDependencies) InitAllMiddleware() {
	md.App.Use(
		RequestIDMiddleware(),
		LoggerMiddleware(md.Log.Logger),
		RecoverMiddleware(md.Log),
		TimeoutMiddleware(defaultRequestTimeout),
		FaviconMiddleware(),
		RateLimiter(md.Log.Logger),
	)
}
