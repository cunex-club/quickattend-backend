package main

import (
	"github.com/cunex-club/quickattend-backend/internal/config"
	"github.com/cunex-club/quickattend-backend/internal/database"
	"github.com/cunex-club/quickattend-backend/internal/logger"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
)

func main() {
	_ = godotenv.Load()

	cfg := config.Load()
	log.Logger = logger.SetupLogger(cfg.AppEnv)

	db := database.Connect(cfg.DatabaseConfig)

	log.Info().Msg("Successfully connected to the database")

	// Application logic here...
	i, _ := db.DB()

	i.Ping()
	log.Debug().Msg("Pinged database")
}
