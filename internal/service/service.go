package service

import (
	"github.com/cunex-club/quickattend-backend/internal/config"
	"github.com/cunex-club/quickattend-backend/internal/repository"
)

type service struct {
	repo repository.AllRepo
	cfg  *config.Config
}

type AllOfService struct {
	HealthCheck HealthCheckService
	Auth        AuthService
}

func NewService(repo repository.AllRepo, cfg *config.Config) AllOfService {
	srv := &service{
		repo: repo,
		cfg:  cfg,
	}

	return AllOfService{
		HealthCheck: srv,
		Auth:        srv,
	}
}
