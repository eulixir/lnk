package config

import (
	"fmt"

	"lnk/extensions/logger"
	"lnk/extensions/opentelemetry"
	"lnk/extensions/redis"
	"lnk/gateways/gocql"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	App    App
	OTel   opentelemetry.Config
	Logger logger.Config
	Gocql  gocql.Config
	Redis  redis.Config
}

type App struct {
	ENV        string `envconfig:"ENV" default:"development"`
	Port       string `envconfig:"PORT" default:"8080"`
	GinMode    string `envconfig:"GIN_MODE" default:"debug"`
	Base62Salt string `envconfig:"BASE62_SALT" required:"true"`
}

func LoadConfig() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load .env file: %w", err)
	}

	config := &Config{}
	if err := envconfig.Process("", config); err != nil {
		return nil, fmt.Errorf("failed to process app config: %w", err)
	}

	if err := envconfig.Process("", &config.Gocql); err != nil {
		return nil, fmt.Errorf("failed to process gocql config: %w", err)
	}

	return config, nil
}
