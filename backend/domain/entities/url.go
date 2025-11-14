package entities

import (
	"time"
)

type URL struct {
	CreatedAt time.Time
	ShortCode string
	LongURL   string
}
