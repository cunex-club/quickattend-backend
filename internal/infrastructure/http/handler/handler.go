package handler

import (
	"github.com/cunex-club/quickattend-backend/internal/service"
	"github.com/rs/zerolog"
)

type Handler struct {
	Service *service.AllOfService
	Logger  *zerolog.Logger
}

type AllOfHandler struct {
	HealthCheckHandler HealthCheckHandler
	AuthHandler AuthHandler
}

func NewHandler(srv *service.AllOfService, logger *zerolog.Logger) *AllOfHandler {
	h := &Handler{
		Service: srv,
		Logger:  logger,
	}
	return &AllOfHandler{
		HealthCheckHandler: h,
		AuthHandler: h,
	}
}
