package repository

import "gorm.io/gorm"

type repository struct {
	db *gorm.DB
}

type AllRepo struct {
	HealthCheck HealthCheckRepository
	Auth        AuthRepository
}

func NewRepository(db *gorm.DB) AllRepo {
	repo := &repository{db: db}
	return AllRepo{
		HealthCheck: repo,
		Auth:        repo,
	}
}
