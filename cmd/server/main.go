package main

import (
	"github.com/cunex-club/quickattend-backend/internal/config"
	"github.com/cunex-club/quickattend-backend/internal/database"
	"github.com/cunex-club/quickattend-backend/internal/handler"
	"github.com/cunex-club/quickattend-backend/internal/logger"
	"github.com/cunex-club/quickattend-backend/internal/repository"
	"github.com/cunex-club/quickattend-backend/internal/router"
	"github.com/cunex-club/quickattend-backend/internal/service"
	"github.com/gofiber/fiber/v2"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
)

func main() {
	_ = godotenv.Load()

	// Load config
	cfg := config.Load()
	log.Logger = logger.SetupLogger(cfg.AppEnv)

	db := database.Connect(cfg.DatabaseConfig)
	if db == nil {
		log.Fatal().Msg("Database connection failed")
	}
	log.Info().Msg("Successfully connected to the database")

	repos := repository.NewRepository(db)
	services := service.NewService(repos)
	handlers := handler.NewHandler(&services, &log.Logger)

	app := fiber.New()

	router.SetupRoutes(app, handlers)
	log.Info().Msg("Starting server on :8000")
	if err := app.Listen(":8000"); err != nil {
		log.Fatal().Err(err).Msg("Server failed to start")
	}
}
