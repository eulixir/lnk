package repositories

import (
	"context"

	gocql "github.com/apache/cassandra-gocql-driver/v2"
	"go.uber.org/zap"
)

type Repository struct {
	ctx     context.Context
	logger  *zap.Logger
	session *gocql.Session
}

func NewRepository(ctx context.Context, logger *zap.Logger, session *gocql.Session) *Repository {
	return &Repository{ctx: ctx, logger: logger, session: session}
}
