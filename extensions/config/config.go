package config

import (
	"lnk/gateways/gocql"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Port  string       `envconfig:"PORT" default:"8080"`
	Gocql gocql.Config
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
