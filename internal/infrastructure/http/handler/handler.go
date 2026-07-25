package handler

import (
	"github.com/cunex-club/quickattend-backend/internal/config"
	"github.com/cunex-club/quickattend-backend/internal/service"
	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog"
)

type Handler struct {
	Service   *service.AllOfService
	Logger    *zerolog.Logger
	Validator *validator.Validate
	Config    *config.Config
}

type AllOfHandler struct {
	HealthCheckHandler HealthCheckHandler
	AuthHandler        AuthHandler
	EventHandler       EventHandler
	UserHandler        UserHandler
}

func NewHandler(srv *service.AllOfService, logger *zerolog.Logger, validator *validator.Validate, cfg *config.Config) *AllOfHandler {
	h := &Handler{
		Service:   srv,
		Logger:    logger,
		Validator: validator,
		Config:    cfg,
	}
	return &AllOfHandler{
		HealthCheckHandler: h,
		AuthHandler:        h,
		EventHandler:       h,
		UserHandler:        h,
	}
}
