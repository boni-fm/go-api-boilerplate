package middleware

import (
	"go-api-boilerplate/config"
	"go-api-boilerplate/internal/database"
	"go-api-boilerplate/internal/utility/fibererror"
	"regexp"

	pgsd3 "github.com/boni-fm/go-libsd3/pkg/db/postgres"
	"github.com/boni-fm/go-libsd3/pkg/log"
	"github.com/gofiber/fiber/v3"
)

var prefixKunciRe = regexp.MustCompile(`(?i)(g\d+(?:sim)?)(?:[^a-zA-Z0-9]|$)`)

const (
	DbLocalKey     = "dbLocal"
	KodeDcLocalKey = "kodedc"
)

// MultiDCMiddleware
// fungsi untuk dapetin atau extract kodedc (kunci)
// untuk baca kunci dan dapetin connstring db nya...
//
// prioritas baca kunci ::
// 1. parameter query "KodeDC"
// 2. request header X-kode-dc
// 3. X-Forwarded-Prefix -> transferGXXX
//
// dibuat seperti ini supaya lebih dinamis, jadi bisa menyesuaikan
// dengan kebutuhan dan gk perlu ada revisi besar jika ingin aplikasi
// ini dibuat untuk universal
func MultiDCMiddleware(logger *log.Logger, cfg *config.Config, dbAdapter *database.DcAdapter) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		kunciDc := resolveKunci(ctx)
		if kunciDc == "" {
			logger.Warn("kunci dc kosong ~~")
			return fibererror.BadRequestError(
				ctx,
				"Kunci tidak ditemukan; isi query param 'KodeDC', header 'X-kunci-dc', atau X-Forwarded-Prefix (contoh: /apixxxg001)")
		}

		db, err := dbAdapter.GetOrInit(ctx.Context(), kunciDc)
		if err != nil {
			logger.Warn("Gagal mengambil/initialize koneksi db " + kunciDc + " | ERR :: " + err.Error())
			return fibererror.InternalServerError(
				ctx,
				"Gagal mengambil/initialize koneksi database untuk Kunci "+kunciDc+" | ERR: "+err.Error())
		}

		// taruh semua di local storage requestnya
		// ini penjelasan lifecycle nya
		// disimpen buat jadi catetan ~~
		// kalau mau diubah, jadikan penjelasan dibawah sebagai acuan...
		/*
			    Incoming HTTP Request
				        │
				        ▼
				Fiber assigns a *Ctx from sync.Pool
				  Ctx {
				    fasthttp: *RequestCtx (this request's HTTP data)
				    locals:   map{}  ← fresh empty map
				  }
				        │
				        ▼
				MultiDc middleware runs:
				  1. c.Get("X-kode-dc") → "G001"
				  2. AppName from config.Config
				  3. ta.GetOrInit(ctx, "G001")
				       └── returns cached *pgsd3.Database (pointer)
				  4. c.Locals("dbLocal", db)
				       └── map["kodedc"] = 0xc000123456  ← memory address of *Database
				  5. c.Next() → passes control to handler
				        │
				        ▼
				Handler runs:
				  db, ok := middleware.DBFromLocals(c)
				       └── c.Locals("dbLocal")
				             └── returns 0xc000123456
				             └── type asserts to *pgsd3.Database
				  repo := repository.NewUserRepository(db)
				  repo.List(ctx)
				       └── db.SelectAll() → db.Pool.Query() → borrows conn from pool
				                                             → runs query
				                                             → returns conn to pool
				        │
				        ▼
				Response sent
				        │
				        ▼
				Fiber returns *Ctx to sync.Pool
				  locals map is CLEARED
				  0xc000123456 reference in locals → GONE
				  But *pgsd3.Database itself → still alive in TenantAdapter
				  (GC won't collect it — TenantAdapter holds the real reference)
		*/
		ctx.Locals(DbLocalKey, db)          //dbLocal
		ctx.Locals(KodeDcLocalKey, kunciDc) // kodedc

		return ctx.Next()
	}
}

// resolveKunci
// fungsi untuk extract kodedc nya ~~
func resolveKunci(c fiber.Ctx) string {
	if k := c.Query("KodeDC"); k != "" {
		return k
	}
	if k := c.Get("X-kode-dc"); k != "" {
		return k
	}
	if prefix := c.Get("X-Forwarded-Prefix"); prefix != "" {
		if m := prefixKunciRe.FindStringSubmatch(prefix); len(m) >= 2 {
			return m[1]
		}
	}
	return ""
}

// Kumpulan function untuk ngambil value dari local
// Kalau mau jalanin ini, pastiin middleware nya dipakai
// Jika tidak, value pasti akan error ~~~
// List yg disimpen di local
// > DB
// > KodeDC/Kunci

func DBFromLocals(c fiber.Ctx) (*pgsd3.Database, bool) {
	db, ok := c.Locals(DbLocalKey).(*pgsd3.Database)
	return db, ok
}

func KodeDCFromLocals(c fiber.Ctx) string {
	k, _ := c.Locals(KodeDcLocalKey).(string)
	return k
}
