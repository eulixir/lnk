package usecases

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
)

func (uc *UseCase) GetLongURL(ctx context.Context, shortCode string) (string, error) {
	tracer := otel.Tracer("usecases.GetLongURL")
	ctx, span := tracer.Start(ctx, "GetLongURL")
	defer span.End()

	var err error
	defer func() {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
	}()

	url, err := uc.repository.GetURLByShortCode(ctx, shortCode)

	if url == nil {
		return "", ErrURLNotFound
	}

	if err != nil {
		return "", fmt.Errorf("failed to get URL by short code: %w", err)
	}

	return url.LongURL, nil
}
