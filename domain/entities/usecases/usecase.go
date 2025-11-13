package usecases

import (
	"lnk/gateways/gocql/repositories"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type UseCase struct {
	logger     *zap.Logger
	repository *repositories.Repository
	redis      *redis.Client
}

func NewUseCase(logger *zap.Logger, repository *repositories.Repository, redis *redis.Client) *UseCase {
	return &UseCase{
		logger:     logger,
		repository: repository,
		redis:      redis,
	}
}
