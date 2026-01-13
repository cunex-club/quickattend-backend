package service

import (
	"github.com/cunex-club/quickattend-backend/internal/config"
	"github.com/cunex-club/quickattend-backend/internal/repository"
	"github.com/rs/zerolog"
)

type service struct {
	repo   repository.AllRepo
	cfg    *config.Config
	logger *zerolog.Logger
}

type AllOfService struct {
	HealthCheck HealthCheckService
	Auth        AuthService
	Event       EventService
}

func NewService(repo repository.AllRepo, cfg *config.Config, logger *zerolog.Logger) AllOfService {
	srv := &service{
		repo:   repo,
		cfg:    cfg,
		logger: logger,
	}

	return AllOfService{
		HealthCheck: srv,
		Auth:        srv,
		Event:       srv,
	}
}
