package main

import (
	"net/http"
	"time"

	"github.com/cunex-club/quickattend-backend/internal/config"
	"github.com/cunex-club/quickattend-backend/internal/database"
	"github.com/cunex-club/quickattend-backend/internal/infrastructure/http/handler"
	gql "github.com/cunex-club/quickattend-backend/internal/infrastructure/http/handler/graphql"
	"github.com/cunex-club/quickattend-backend/internal/infrastructure/http/middleware"
	"github.com/cunex-club/quickattend-backend/internal/infrastructure/http/router"
	"github.com/cunex-club/quickattend-backend/internal/infrastructure/logger"
	"github.com/cunex-club/quickattend-backend/internal/repository"
	"github.com/cunex-club/quickattend-backend/internal/service"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
)

func main() {
	_ = godotenv.Load()

	// Load config
	cfg := config.Load()
	log.Logger = logger.SetupLogger(cfg.AppEnv)

	db, err := database.Connect(cfg.DatabaseConfig)
	if err != nil {
		log.Fatal().Err(err).Msg("Database connection failed")
	}
	log.Info().Msg("Successfully connected to the database")

	repos := repository.NewRepository(db)
	services := service.NewService(repos, cfg, &log.Logger, &http.Client{Timeout: 10 * time.Second})
	handlers := handler.NewHandler(&services, &log.Logger, validator.New())

	app := fiber.New(fiber.Config{
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
	})

	mw := middleware.NewMiddleware(cfg)
	app.Use(
		mw.Recover(),
		mw.RequestID(),
		mw.CORS(),
		mw.RequestLogger(),
		mw.Timeout(),
	)

	gqlResolver := &gql.Resolver{Service: &services}

	router.SetupRoutes(app, handlers, mw, gqlResolver)
	log.Info().Msg("Starting server on :8000")
	if err := app.Listen(":8000"); err != nil {
		log.Fatal().Err(err).Msg("Server failed to start")
	}
}
