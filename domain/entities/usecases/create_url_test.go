package usecases

import (
	"context"

	"lnk/extensions/gocqltesting"
	"lnk/gateways/gocql/repositories"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestCreateURL(t *testing.T) {
	session, err := gocqltesting.NewDB(t, t.Name())
	require.NoError(t, err)

	ctx := context.Background()
	logger := zap.NewNop()

	repository := repositories.NewRepository(ctx, logger, session)

	params := NewUseCaseParams{
		Ctx:        ctx,
		Logger:     logger,
		Repository: repository,
		Redis:      &mockRedisAdapter{},
		Salt:       "test",
		CounterKey: "test",
	}

	useCase := NewUseCase(params)

	longURL := "https://www.google.com"
	shortCode, err := useCase.CreateShortURL(longURL)
	require.NoError(t, err)
	require.NotEmpty(t, shortCode)
}

type mockRedisAdapter struct{}

func (m *mockRedisAdapter) Incr(ctx context.Context, key string) (int64, error) {
	return 1, nil
}
