package service

import (
	"github.com/cunex-club/quickattend-backend/internal/repository"
	"github.com/rs/zerolog"
)

type service struct {
	repo   repository.AllRepo
	logger *zerolog.Logger
}

type AllOfService struct {
	HealthCheck HealthCheckService
	Auth        AuthService
}

func NewService(repo repository.AllRepo, logger *zerolog.Logger) AllOfService {
	srv := &service{
		repo:   repo,
		logger: logger,
	}
	return AllOfService{
		HealthCheck: srv,
		Auth:        srv,
	}
}
