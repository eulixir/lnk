package usecases

import (
	"context"

	"lnk/extensions/gcqltesting"
	"lnk/gateways/gocql"
	"lnk/gateways/gocql/repositories"
	"testing"

	redis "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestCreateURL(t *testing.T) {
	rawSession, err := gcqltesting.NewDB(t, t.Name())
	require.NoError(t, err)

	ctx := context.Background()
	logger := zap.NewNop()

	session := &gocql.Session{Session: rawSession}
	repository := repositories.NewRepository(ctx, logger, session)

	params := NewUseCaseParams{
		Ctx:        ctx,
		Logger:     logger,
		Repository: repository,
		Redis:      &redis.Client{},
		Salt:       "test",
		CounterKey: "test",
	}

	_ = NewUseCase(params)
}
