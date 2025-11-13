package usecases

import (
	"context"
	"lnk/gateways/gocql/repositories"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type UseCase struct {
	ctx        context.Context
	logger     *zap.Logger
	repository *repositories.Repository
	redis      *redis.Client
	salt       string
	counterKey string
}

type NewUseCaseParams struct {
	Ctx        context.Context
	Logger     *zap.Logger
	Repository *repositories.Repository
	Redis      *redis.Client
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
