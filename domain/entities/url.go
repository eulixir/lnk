package entities

import (
	"time"

	"github.com/google/uuid"
)

type URL struct {
	ID        uuid.UUID
	ShortCode string
	LongURL   string
	CreatedAt time.Time
}
