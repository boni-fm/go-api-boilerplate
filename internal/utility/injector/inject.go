package injector

import (
	"fmt"
	"go-api-boilerplate/config"
	"go-api-boilerplate/internal/database"
	"go-api-boilerplate/internal/middleware"

	pgsd3 "github.com/boni-fm/go-libsd3/pkg/db/postgres"
	logger "github.com/boni-fm/go-libsd3/pkg/log"
	"github.com/gofiber/fiber/v3"
)

// DBInjector abstracts where *pgsd3.Database comes from.
// Handler calls GetDB(c) ...
type DBInjector interface {
	GetDB(c fiber.Ctx) (*pgsd3.Database, error)
}

// ── StaticInjector ────────────────────────────────────────────────

type StaticInjector struct {
	ta     *database.DcAdapter
	kodeDc string
}

func NewStaticInjector(ta *database.DcAdapter, kodeDc string) *StaticInjector {
	return &StaticInjector{ta: ta, kodeDc: kodeDc}
}

func (p *StaticInjector) GetDB(c fiber.Ctx) (*pgsd3.Database, error) {
	return p.ta.GetOrInit(c.Context(), p.kodeDc)
}

// ── LocalsInjector ────────────────────────────────────────────────

type LocalsInjector struct {
	localKey string
}

func NewLocalsInjector() *LocalsInjector {
	return &LocalsInjector{localKey: middleware.DbLocalKey}
}

func (p *LocalsInjector) GetDB(c fiber.Ctx) (*pgsd3.Database, error) {
	db, ok := c.Locals(p.localKey).(*pgsd3.Database)
	if !ok || db == nil {
		return nil, fmt.Errorf("no db in locals[%s] — DBResolver is not registered", p.localKey)
	}
	return db, nil
}

// ServiceFactory ~ biar dynamic, gk init terus2an ...
// kalau mau inject baru lagi, bisa ditambahin disini ~
//
// WARN : tapi semua service disini harus direvisi dengan tambahan yg baru
type ServiceFactory[S any] struct {
	dbInjector  DBInjector
	log         *logger.Logger
	cfg         *config.Config
	constructor func(*pgsd3.Database, *logger.Logger, *config.Config) *S // service constructor function
}

func NewServiceFactory[S any](
	p DBInjector,
	log *logger.Logger,
	cfg *config.Config,
	constructor func(db *pgsd3.Database, log_ *logger.Logger, cfg *config.Config) *S,
) *ServiceFactory[S] {
	if p == nil {
		panic("ServiceFactory: DBInjector must not be nil")
	}
	if constructor == nil {
		panic("ServiceFactory: constructor must not be nil")
	}
	return &ServiceFactory[S]{
		dbInjector:  p,
		log:         log,
		cfg:         cfg,
		constructor: constructor,
	}
}

// Build resolves the DB and builds the service in one call.
// Dipaksa menggunakan service ~
//
// Note ::
//   - untuk service yang baru, bisa disamakan constructor nya dengan contoh (dipakai atau tidak)
//     > db, logger, config
func (f *ServiceFactory[S]) Build(c fiber.Ctx) (*S, error) {
	db, err := f.dbInjector.GetDB(c)
	if err != nil {
		return nil, err
	}
	return f.constructor(db, f.log, f.cfg), nil
}
