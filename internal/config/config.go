package config

import (
	"os"
	"strconv"
)

type Config struct {
	AppEnv         string
	DatabaseConfig DatabaseConfig
}

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
	Schema   string
}

func Load() *Config {
	return &Config{
		AppEnv: getEnv("APP_ENV", "development"),
		DatabaseConfig: DatabaseConfig{
			Host:     getEnv("POSTGRES_HOST", "localhost"),
			Port:     getEnvAsInt("POSTGRES_PORT", 5432),
			User:     getEnv("POSTGRES_USER", "postgres"),
			Password: getEnv("POSTGRES_PASSWORD", "postgres"),
			Name:     getEnv("POSTGRES_DB", "quickattend-db"),
			Schema:   getEnv("POSTGRES_SCHEMA", "public"),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if val, exists := os.LookupEnv(key); exists {
		return val
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if val, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.Atoi(val); err == nil {
			return intValue
		}
	}
	return defaultValue
}
