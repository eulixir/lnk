package usecases

import (
	"context"

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

	repository := repositories.NewRepository(ctx, logger, session)

	mockRedis := mocks.NewMockRedis(t)
	mockRedis.On("Incr", mock.Anything, mock.Anything).Return(int64(1), nil)

	params := NewUseCaseParams{
		Ctx:        ctx,
		Logger:     logger,
		Repository: repository,
		Redis:      mockRedis,
		Salt:       "test",
		CounterKey: "test",
	}

	useCase := NewUseCase(params)

	longURL := "https://www.google.com"
	shortCode, err := useCase.CreateShortURL(longURL)
	require.NoError(t, err)
	require.NotEmpty(t, shortCode)
}
