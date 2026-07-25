package config

import (
	"github.com/caarlos0/env/v10"
	"github.com/rs/zerolog/log"
)

type Config struct {
	AppEnv          string `env:"APP_ENV" envDefault:"development"`
	JWTSecret       string `env:"JWT_SECRET,required"`
	FrontendHomeURL string `env:"FRONTEND_HOME_URL" envDefault:"https://quickattend.cunex.club/"`
	AllowedOrigins  string `env:"ALLOWED_ORIGINS" envDefault:"https://quickattend.cunex.club"`

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
	SSLMode  string `env:"POSTGRES_SSLMODE" envDefault:"disable"`
}

type LLEConfig struct {
	ProfileClientID     string `env:"LLE_PROFILE_CLIENT_ID,required"`
	ProfileClientSecret string `env:"LLE_PROFILE_CLIENT_SECRET,required"`
	QRClientID          string `env:"LLE_QR_CLIENT_ID,required"`
	QRClientSecret      string `env:"LLE_QR_CLIENT_SECRET,required"`
	// QRCodeInfoURL lets this be pointed at the UAT host for testing without
	// touching real CU NEX data. Defaults to PROD to match existing behavior.
	QRCodeInfoURL string `env:"LLE_QR_CODE_INFO_URL" envDefault:"https://culab-svc.azurewebsites.net/Service.svc/qrcodeinfo_for_all"`
}

func Load() *Config {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		log.Fatal().Err(err).Msg("failed to parse environment variables")
	}
	return cfg
}
