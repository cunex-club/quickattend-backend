package repository

import "gorm.io/gorm"

type repository struct {
	db *gorm.DB
}

type AllRepo struct {
	HealthCheck HealthCheckRepository
}

func NewRepository(db *gorm.DB) AllRepo {
	repo := &repository{db: db}
	return AllRepo{
		HealthCheck: repo,
	}
}
