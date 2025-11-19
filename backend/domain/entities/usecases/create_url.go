package usecases

import (
	"context"
	"fmt"
	"sync"

	"lnk/domain/entities"
	"lnk/domain/entities/helpers"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
)

var (
	urlShortenedCounter metric.Int64Counter
	counterOnce         sync.Once
)

func (uc *UseCase) CreateShortURL(ctx context.Context, longURL string) (string, error) {
	tracer := otel.Tracer("usecases.CreateShortURL")
	ctx, span := tracer.Start(ctx, "CreateShortURLUsecase")
	var err error

	defer func() {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
	}()
	defer span.End()

	id, err := uc.redis.Incr(ctx, uc.counterKey)
	if err != nil {
		return "", fmt.Errorf("failed to increment counter: %w", err)
	}

	shortCode := helpers.Base62Encode(id, uc.salt)

	url := &entities.URL{
		ShortCode: shortCode,
		LongURL:   longURL,
	}

	err = uc.repository.CreateURL(ctx, url)
	if err != nil {
		return "", fmt.Errorf("failed to create URL in repository: %w", err)
	}

	uc.incrementURLShortenedMetric(ctx)

	return shortCode, nil
}

func (uc *UseCase) incrementURLShortenedMetric(ctx context.Context) {
	counterOnce.Do(func() {
		meter := otel.Meter("lnk-backend", metric.WithInstrumentationVersion("1.0.0"))
		var err error
		urlShortenedCounter, err = meter.Int64Counter(
			"urls_shortened_total",
			metric.WithDescription("Total number of URLs shortened"),
			metric.WithUnit("1"),
		)
		if err != nil {
			return
		}
	})

	if urlShortenedCounter == nil {
		return
	}

	urlShortenedCounter.Add(ctx, 1,
		metric.WithAttributes(
			attribute.String("service", "lnk-backend"),
		),
	)
}
