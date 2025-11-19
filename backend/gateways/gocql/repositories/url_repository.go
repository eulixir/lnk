package repositories

import (
	"context"
	"errors"
	"fmt"
	"time"

	"lnk/domain/entities"

	"github.com/gocql/gocql"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
)

func (r *Repository) CreateURL(ctx context.Context, url *entities.URL) error {
	tracer := otel.Tracer("repositories.CreateURL")
	ctx, span := tracer.Start(ctx, "CreateURL")
	defer span.End()

	var err error
	defer func() {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
	}()

	url.CreatedAt = time.Now().UTC()

	err = r.session.Query(
		"INSERT INTO urls (short_code, long_url, created_at) VALUES (?, ?, ?)",
		url.ShortCode, url.LongURL, url.CreatedAt,
	).ExecContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to create URL: %w", err)
	}

	return nil
}

func (r *Repository) GetURLByShortCode(ctx context.Context, shortCode string) (*entities.URL, error) {
	tracer := otel.Tracer("repositories.GetURLByShortCode")
	ctx, span := tracer.Start(ctx, "GetURLByShortCode")
	defer span.End()

	var err error
	defer func() {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
	}()

	var url entities.URL

	err = r.session.Query(
		"SELECT short_code, long_url, created_at FROM urls WHERE short_code = ?",
		shortCode,
	).ScanContext(ctx, &url.ShortCode, &url.LongURL, &url.CreatedAt)
	if err != nil {
		if errors.Is(err, gocql.ErrNotFound) {
			return nil, gocql.ErrNotFound
		}

		return nil, fmt.Errorf("failed to get URL by short code: %w", err)
	}

	return &url, nil
}
