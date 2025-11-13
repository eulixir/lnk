package repositories

import (
	"context"
	"lnk/gateways/gocql"

	"go.uber.org/zap"
)

type Repository struct {
	ctx     context.Context
	logger  *zap.Logger
	session *gocql.Session
}

func NewRepository(ctx context.Context, logger *zap.Logger, session *gocql.Session) *Repository {
	return &Repository{
		ctx:     ctx,
		logger:  logger,
		session: session,
	}
}
