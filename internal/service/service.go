package service

import "github.com/cunex-club/quickattend-backend/internal/repository"

type service struct {
	repo repository.AllRepo
}

type AllOfService struct {
	HealthCheck HealthCheckService
	Auth        AuthService
}

func NewService(repo repository.AllRepo) AllOfService {
	srv := &service{
		repo: repo,
	}
	return AllOfService{
		HealthCheck: srv,
	}
}
