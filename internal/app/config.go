package app

import (
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"

	aphttp "order-pack-calculator/internal/api/http"
	postgres "order-pack-calculator/internal/repository/pg"
)

// AppName is the envconfig prefix for all settings (e.g. ORDER_PACK_CALCULATOR_SERVER_PORT).
const AppName = "ORDER_PACK_CALCULATOR"

// Config holds HTTP server and database settings.
type Config struct {
	Server   aphttp.ServerConfig `envconfig:"SERVER"`
	Database postgres.Config      `envconfig:"DATABASE"`
}

// Load merges optional `.env` in the current working directory with process environment (missing file is ignored).
func Load() (Config, error) {
	_ = godotenv.Load(".env")

	var cfg Config
	if err := envconfig.Process(AppName, &cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}
