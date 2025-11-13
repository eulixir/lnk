package usecases

import (
	"lnk/gateways/gocql/repositories"

	"go.uber.org/zap"
)

type UseCase struct {
	logger     *zap.Logger
	repository *repositories.Repository
}

func NewUseCase(logger *zap.Logger, repository *repositories.Repository) *UseCase {
	return &UseCase{
		logger:     logger,
		repository: repository,
	}
}
