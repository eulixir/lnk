package entities

import (
	"time"

	"github.com/google/uuid"
)

type URL struct {
	ID        uuid.UUID
	RawURL    string
	ShortURL  string
	CreatedAt time.Time
}
