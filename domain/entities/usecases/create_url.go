package usecases

import (
	"context"
	"fmt"

	"lnk/domain/entities"
	"lnk/domain/entities/helpers"
)

func (uc *UseCase) CreateShortURL(ctx context.Context, longURL string) (string, error) {
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

	return shortCode, nil
}
