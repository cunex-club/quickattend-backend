package database

import (
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"

	"github.com/cunex-club/quickattend-backend/internal/config"
)

var logger = log.With().Str("module", "database").Logger()

func Connect(config config.DatabaseConfig) (*gorm.DB, error) {

	// TODO: Change GORM to preferred library
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=Asia/Bangkok",
		config.Host, config.User, config.Password, config.Name, config.Port, config.SSLMode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{TablePrefix: config.Schema + "."},
	})
	if err != nil {
		logger.Error().Err(err).Msg("Failed to connect to database")
		return nil, err
	}

	err = db.Exec("CREATE EXTENSION IF NOT EXISTS \"pg_trgm\";").Error
	if err != nil {
		logger.Error().Err(err).Msg("Failed to create pg_trgm extension")
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get underlying sql.DB")
		return nil, err
	}
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)

	return db, nil
}
