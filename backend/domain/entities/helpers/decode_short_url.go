package helpers

import (
	"strings"
)

func Base62Decode(shortURL, salt string) int64 {
	alphabet := getShuffledAlphabet(salt)

	var decoded int64

	for _, char := range shortURL {
		index := strings.IndexRune(alphabet, char)
		if index == -1 {
			return 0
		}

		decoded = decoded*base62 + int64(index)
	}

	return decoded
}
