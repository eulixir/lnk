package config

import "github.com/joho/godotenv"

type Config struct {
	Port string `envconfig:"PORT" default:"8080"`
}

func LoadConfig() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, err
	}
	config := &Config{}
	return config, nil
}
