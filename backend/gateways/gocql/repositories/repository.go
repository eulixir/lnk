package repositories

import (
	gocql "github.com/apache/cassandra-gocql-driver/v2"
	"go.uber.org/zap"
)

type Repository struct {
	logger  *zap.Logger
	session *gocql.Session
}

func NewRepository(logger *zap.Logger, session *gocql.Session) *Repository {
	return &Repository{logger: logger, session: session}
}
