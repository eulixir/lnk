package repositories

import (
	"errors"
	"fmt"
	"lnk/domain/entities"
	"time"

	"github.com/gocql/gocql"
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

func (r *Repository) GetURLByShortCode(shortCode string) (*entities.URL, error) {
	var url entities.URL
	err := r.session.Query(
		"SELECT id, short_code, long_url, created_at FROM urls WHERE short_code = ?",
		shortCode,
	).Scan(&url.ID, &url.ShortCode, &url.LongURL, &url.CreatedAt)
	if err != nil {
		if errors.Is(err, gocql.ErrNotFound) {
			return nil, fmt.Errorf("URL not found")
		}
		return nil, fmt.Errorf("failed to get URL by short code: %w", err)
	}
	return &url, nil
}
