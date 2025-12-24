package config

import (
	"github.com/caarlos0/env/v10"
	"github.com/rs/zerolog/log"
)

type Config struct {
	AppEnv    string `env:"APP_ENV" envDefault:"development"`
	JWTSecret string `env:"JWT_SECRET,required"`

	DatabaseConfig DatabaseConfig
	LLEConfig      LLEConfig
}

type DatabaseConfig struct {
	Host     string `env:"POSTGRES_HOST,required"`
	Port     int    `env:"POSTGRES_PORT" envDefault:"5432"`
	User     string `env:"POSTGRES_USER,required"`
	Password string `env:"POSTGRES_PASS,required"`
	Name     string `env:"POSTGRES_DB,required"`
	Schema   string `env:"POSTGRES_SCHEMA" envDefault:"public"`
}

type LLEConfig struct {
	ClientId     string `env:"LLEClientId,required"`
	ClientSecret string `env:"LLEClientSecret,required"`
}

func Load() *Config {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		log.Fatal().Err(err).Msg("failed to parse environment variables")
	}
	return cfg
}
