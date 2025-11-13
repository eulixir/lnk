package repositories

import (
	"fmt"
	"lnk/domain/entities"
	"time"

	"github.com/google/uuid"
)

func (r *Repository) CreateURL(url *entities.URL) error {
	url.ID = uuid.New().String()
	url.CreatedAt = time.Now().UTC()

	err := r.session.Query(
		"INSERT INTO urls (id, short_code, long_url, created_at) VALUES (?, ?, ?, ?)",
		url.ID, url.ShortCode, url.LongURL, url.CreatedAt,
	).ExecContext(r.ctx)
	if err != nil {
		return fmt.Errorf("failed to create URL: %w", err)
	}

	return nil
}
