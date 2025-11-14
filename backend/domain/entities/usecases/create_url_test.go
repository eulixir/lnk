package usecases_test

import (
	"context"

	"lnk/domain/entities/usecases"
	"lnk/extensions/gocqltesting"
	"lnk/extensions/redis/mocks"
	"lnk/gateways/gocql/repositories"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func Test_UseCase_CreateURL(t *testing.T) {
	t.Parallel()

	session, err := gocqltesting.NewDB(t, t.Name())
	require.NoError(t, err)

	ctx := context.Background()
	logger := zap.NewNop()

	repository := repositories.NewRepository(logger, session)

	mockRedis := mocks.NewMockRedis(t)
	mockRedis.On("Incr", mock.Anything, mock.Anything).Return(int64(1), nil)

	params := usecases.NewUseCaseParams{
		Logger:     logger,
		Repository: repository,
		Redis:      mockRedis,
		Salt:       "test",
		CounterKey: "test",
	}

	useCase := usecases.NewUseCase(params)

	longURL := "https://www.google.com"
	shortCode, err := useCase.CreateShortURL(ctx, longURL)
	require.NoError(t, err)
	require.NotEmpty(t, shortCode)
}
