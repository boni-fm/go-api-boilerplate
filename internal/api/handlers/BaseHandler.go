package handlers

import (
	"context"
	"go-api-boilerplate/internal/api/services"

	"github.com/boni-fm/go-libsd3/pkg/log"
)

// ini untuk kebutuhan depedency inject di dalam all handlers
type HandlersRegistry struct {
	log_        *log.Logger
	ctx         context.Context
	UserService *services.UserService
}

func NewHandlersRegistry(log_ *log.Logger, ctx context.Context) *HandlersRegistry {
	return &HandlersRegistry{
		log_:        log_,
		ctx:         ctx,
		UserService: services.NewUserService(log_, ctx),
	}
}
