package usecases

import (
	"context"
	"errors"
	"lnk/extensions/redis"
	"lnk/gateways/gocql/repositories"

	"go.uber.org/zap"
)

// ErrURLNotFound is returned when a URL is not found
var ErrURLNotFound = errors.New("URL not found")

type UseCase struct {
	ctx        context.Context
	logger     *zap.Logger
	repository *repositories.Repository
	redis      redis.Redis
	salt       string
	counterKey string
}

type NewUseCaseParams struct {
	Ctx        context.Context
	Logger     *zap.Logger
	Repository *repositories.Repository
	Redis      redis.Redis
	Salt       string
	CounterKey string
}

func NewUseCase(params NewUseCaseParams) *UseCase {
	return &UseCase{
		ctx:        params.Ctx,
		logger:     params.Logger,
		repository: params.Repository,
		redis:      params.Redis,
		salt:       params.Salt,
		counterKey: params.CounterKey,
	}
}
