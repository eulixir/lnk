package usecases

import (
	"fmt"

	"lnk/domain/entities"
	"lnk/domain/entities/helpers"
)

func (uc *UseCase) CreateShortURL(longURL string) (string, error) {
	id, err := uc.redis.Incr(uc.ctx, uc.counterKey)
	if err != nil {
		return "", fmt.Errorf("failed to increment counter: %w", err)
	}

	shortCode := helpers.Base62Encode(id, uc.salt)

	url := &entities.URL{
		ShortCode: shortCode,
		LongURL:   longURL,
	}

	err = uc.repository.CreateURL(url)
	if err != nil {
		return "", err
	}

	return shortCode, nil
}
