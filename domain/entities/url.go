package entities

import (
	"time"
)

type URL struct {
	ShortCode string
	LongURL   string
	CreatedAt time.Time
}
