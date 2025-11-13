package usecases

import (
	"fmt"

	"lnk/domain/entities/helpers"
)

func (uc *UseCase) CreateShortURL(longURL string) (string, error) {
	id, err := uc.redis.Incr(uc.ctx, uc.counterKey).Result()
	if err != nil {
		return "", fmt.Errorf("failed to increment counter: %w", err)
	}

	shortCode := helpers.Base62Encode(id, uc.salt)
	return shortCode, nil
}
