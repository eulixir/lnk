package repositories

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gocql/gocql"

	"lnk/domain/entities"
)

func (r *Repository) CreateURL(ctx context.Context, url *entities.URL) error {
	url.CreatedAt = time.Now().UTC()

	err := r.session.Query(
		"INSERT INTO urls (short_code, long_url, created_at) VALUES (?, ?, ?)",
		url.ShortCode, url.LongURL, url.CreatedAt,
	).ExecContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to create URL: %w", err)
	}

	return nil
}

func (r *Repository) GetURLByShortCode(shortCode string) (*entities.URL, error) {
	var url entities.URL
	err := r.session.Query(
		"SELECT short_code, long_url, created_at FROM urls WHERE short_code = ?",
		shortCode,
	).Scan(&url.ShortCode, &url.LongURL, &url.CreatedAt)
	if err != nil {
		if errors.Is(err, gocql.ErrNotFound) {
			return nil, gocql.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get URL by short code: %w", err)
	}
	return &url, nil
}
