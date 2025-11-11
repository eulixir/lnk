package logger

import (
	"fmt"

	"go.uber.org/zap"
)

func NewLogger(config Config) (*zap.Logger, error) {

	if config.Level == "debug" {
		logger, err := zap.NewDevelopment()
		if err != nil {
			return nil, fmt.Errorf("failed to create development logger: %w", err)
		}
		return logger, nil
	}
	logger, err := zap.NewProduction()
	if err != nil {
		return nil, fmt.Errorf("failed to create production logger: %w", err)
	}
	return logger, nil
}
