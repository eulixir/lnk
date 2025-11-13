package entities

import (
	"time"
)

type URL struct {
	ID        string
	ShortCode string
	LongURL   string
	CreatedAt time.Time
}
