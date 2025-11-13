package usecases

import (
	"lnk/domain/entities/helpers"
)

func (uc *UseCase) CreateShortURL(longURL string) (string, error) {
	// the id we will get from redis using counter feature
	id, err := uc.redis.Incr(uc.ctx, "short_url_counter").Result()
	if err != nil {
		return "", err
	}

	helpers.Base62Encode(id, uc.salt)
	return "", nil
}
