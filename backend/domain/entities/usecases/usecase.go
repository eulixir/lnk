package usecases

import (
	"errors"

	"go.uber.org/zap"
	"lnk/extensions/redis"
	"lnk/gateways/gocql/repositories"
)

var ErrURLNotFound = errors.New("URL not found")

type UseCase struct {
	logger     *zap.Logger
	repository *repositories.Repository
	redis      redis.Redis
	salt       string
	counterKey string
}

type NewUseCaseParams struct {
	Logger     *zap.Logger
	Repository *repositories.Repository
	Redis      redis.Redis
	Salt       string
	CounterKey string
}

func NewUseCase(params NewUseCaseParams) *UseCase {
	return &UseCase{
		logger:     params.Logger,
		repository: params.Repository,
		redis:      params.Redis,
		salt:       params.Salt,
		counterKey: params.CounterKey,
	}
}
