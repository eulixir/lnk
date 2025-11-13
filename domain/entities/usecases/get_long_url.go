package usecases

import "fmt"

func (uc *UseCase) GetLongURL(shortCode string) (string, error) {
	url, err := uc.repository.GetURLByShortCode(shortCode)

	if url == nil {
		return "", fmt.Errorf("URL not found")
	}

	if err != nil {
		return "", err
	}

	return url.LongURL, nil
}
