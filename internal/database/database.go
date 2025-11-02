package database

import (
	"fmt"

	"github.com/cunex-club/quickattend-backend/internal/config"
	"github.com/rs/zerolog/log"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

var logger = log.With().Str("module", "database").Logger()

func Connect(config config.DatabaseConfig) (*gorm.DB, error) {

	// TODO: Change GORM to preferred library
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=Asia/Bangkok",
		config.Host, config.User, config.Password, config.Name, config.Port)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{TablePrefix: config.Schema + "."},
	})
	if err != nil {
		logger.Error().Err(err).Msg("Failed to connect to database")
		return nil, err
	}

	return db, nil
}
