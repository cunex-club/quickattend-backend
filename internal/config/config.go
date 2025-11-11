package config

import (
	"os"
	"strconv"
)

type Config struct {
	AppEnv         string
	JwtKey         string
	DatabaseConfig DatabaseConfig
	LLEConfig      LLEConfig
}

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
	Schema   string
}

type LLEConfig struct {
	ClientId     string
	ClientSecret string
}

func Load() *Config {
	return &Config{
		AppEnv: getEnv("APP_ENV", "development"),
		JwtKey: getEnv("JwtKey", ""),
		DatabaseConfig: DatabaseConfig{
			Host:     getEnv("POSTGRES_HOST", "localhost"),
			Port:     getEnvAsInt("POSTGRES_PORT", 5432),
			User:     getEnv("POSTGRES_USER", "postgres"),
			Password: getEnv("POSTGRES_PASSWORD", "postgres"),
			Name:     getEnv("POSTGRES_DB", "quickattend-db"),
			Schema:   getEnv("POSTGRES_SCHEMA", "public"),
		},
		LLEConfig: LLEConfig{
			ClientId:     getEnv("ClientId", ""),
			ClientSecret: getEnv("ClientSecret", ""),
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
