package config

import (
	"strings"

	"github.com/caarlos0/env/v10"
	"github.com/rs/zerolog/log"
)

type Config struct {
	AppEnv          string `env:"APP_ENV" envDefault:"development"`
	JWTSecret       string `env:"JWT_SECRET,required"`
	FrontendHomeURL string `env:"FRONTEND_HOME_URL" envDefault:"https://quickattend.cunex.club/"`
	AllowedOrigins  string `env:"ALLOWED_ORIGINS" envDefault:"https://quickattend.cunex.club"`
	EventRetentionDays int `env:"EVENT_RETENTION_DAYS" envDefault:"90"`

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
	ClientID     string `env:"LLE_CLIENT_ID,required"`
	ClientSecret string `env:"LLE_CLIENT_SECRET,required"`

	// CU NEX issues credentials per capability, so a project can end up with a
	// separate pair for reading QR codes. We don't have one — the same pair has
	// always served both /profile and qrcodeinfo_for_all — so these are optional
	// and fall back to the pair above. Set them only if CU NEX hands over a
	// dedicated QR credential later.
	QRClientID     string `env:"LLE_QR_CLIENT_ID"`
	QRClientSecret string `env:"LLE_QR_CLIENT_SECRET"`

	// QRCodeInfoURL lets this be pointed at the UAT host for testing without
	// touching real CU NEX data. Defaults to PROD to match existing behavior.
	QRCodeInfoURL string `env:"LLE_QR_CODE_INFO_URL" envDefault:"https://culab-svc.azurewebsites.net/Service.svc/qrcodeinfo_for_all"`
}

// QRCredentials returns the credentials to use for qrcodeinfo_for_all: the
// QR-specific pair when both halves are configured, otherwise the main pair.
func (c LLEConfig) QRCredentials() (clientID, clientSecret string) {
	if strings.TrimSpace(c.QRClientID) != "" && strings.TrimSpace(c.QRClientSecret) != "" {
		return c.QRClientID, c.QRClientSecret
	}
	return c.ClientID, c.ClientSecret
}

func Load() *Config {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		log.Fatal().Err(err).Msg("failed to parse environment variables")
	}
	return cfg
}
