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

func Test_UseCase_GetLongURL(t *testing.T) {
	t.Parallel()

	session, err := gocqltesting.NewDB(t, t.Name())
	require.NoError(t, err)
	url := "https://www.google.com"

	ctx := context.Background()
	logger := zap.NewNop()

	repository := repositories.NewRepository(ctx, logger, session)

	mockRedis := mocks.NewMockRedis(t)
	mockRedis.On("Incr", mock.Anything, mock.Anything).Return(int64(1), nil)

	params := usecases.NewUseCaseParams{
		Ctx:        ctx,
		Logger:     logger,
		Repository: repository,
		Redis:      mockRedis,
		Salt:       "test",
		CounterKey: "test",
	}

	useCase := usecases.NewUseCase(params)

	shortCode, err := useCase.CreateShortURL(url)
	require.NoError(t, err)
	require.NotEmpty(t, shortCode)

	longURL, err := useCase.GetLongURL(shortCode)
	require.NoError(t, err)
	require.NotEmpty(t, longURL)
	require.Equal(t, url, longURL)
}

func Test_UseCase_GetLongURL_NotFound(t *testing.T) {
	t.Parallel()

	session, err := gocqltesting.NewDB(t, t.Name())
	require.NoError(t, err)

	ctx := context.Background()
	logger := zap.NewNop()

	repository := repositories.NewRepository(ctx, logger, session)

	params := usecases.NewUseCaseParams{
		Ctx:        ctx,
		Logger:     logger,
		Repository: repository,
	}

	useCase := usecases.NewUseCase(params)

	shortCode := "1234567890"
	longURL, err := useCase.GetLongURL(shortCode)
	require.Error(t, err)
	require.ErrorIs(t, err, usecases.ErrURLNotFound)
	require.Empty(t, longURL)
}
