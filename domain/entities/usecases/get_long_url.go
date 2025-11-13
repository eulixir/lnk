package usecases

import (
	"fmt"
)

func (uc *UseCase) GetLongURL(shortCode string) (string, error) {
	url, err := uc.repository.GetURLByShortCode(shortCode)

	if url == nil {
		return "", ErrURLNotFound
	}

	if err != nil {
		return "", fmt.Errorf("failed to get URL by short code: %w", err)
	}

	return url.LongURL, nil
}
