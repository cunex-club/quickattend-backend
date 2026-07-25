package service

import (
	"net/http"

	"github.com/cunex-club/quickattend-backend/internal/config"
	"github.com/cunex-club/quickattend-backend/internal/repository"
	"github.com/rs/zerolog"
)

type service struct {
	repo       repository.AllRepo
	cfg        *config.Config
	logger     *zerolog.Logger
	httpClient *http.Client
}

type AllOfService struct {
	HealthCheck HealthCheckService
	Auth        AuthService
	Event       EventService
	Dashboard   DashboardService
	User        UserService
}

func NewService(repo repository.AllRepo, cfg *config.Config, logger *zerolog.Logger, httpClient *http.Client) AllOfService {
	srv := &service{
		repo:       repo,
		cfg:        cfg,
		logger:     logger,
		httpClient: httpClient,
	}

	return AllOfService{
		HealthCheck: srv,
		Auth:        srv,
		Event:       srv,
		Dashboard:   srv,
		User:        srv,
	}
}
