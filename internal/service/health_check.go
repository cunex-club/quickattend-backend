package service

import (
	"runtime"
	"time"

	dtoRes "github.com/cunex-club/quickattend-backend/internal/dto/response"
)

type HealthCheckService interface {
	CheckSystem() (dtoRes.HealthCheckRes, error)
}

func (s *service) CheckSystem() (dtoRes.HealthCheckRes, error) {
	dbStart := time.Now()
	dbErr := s.repo.HealthCheck.PingDatabase()
	dbDuration := time.Since(dbStart).Milliseconds()

	dbConnected := dbErr == nil

	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	memoryConsumption := float64(mem.Alloc) / 1024.0 / 1024.0 // MB

	healthStatus := dtoRes.HealthCheckRes{
		DatabaseConnection:   dbConnected,
		DatabaseResponseTime: float64(dbDuration),
		MemoryConsumption:    memoryConsumption,
	}

	return healthStatus, dbErr
}
