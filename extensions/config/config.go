package config

import (
	"lnk/extensions/logger"
	"lnk/extensions/redis"
	"lnk/gateways/gocql"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	App    App
	Gocql  gocql.Config
	Redis  redis.Config
	Logger logger.Config
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
		return nil, err
	}
	config := &Config{}
	if err := envconfig.Process("", config); err != nil {
		return nil, err
	}
	if err := envconfig.Process("", &config.Gocql); err != nil {
		return nil, err
	}
	return config, nil
}
