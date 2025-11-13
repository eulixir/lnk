package helpers

import (
	"crypto/sha256"
	"strings"
)

const base62Alphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func getShuffledAlphabet(salt string) string {
	alphabet := []rune(base62Alphabet)
	if salt == "" {
		return base62Alphabet
	}

	hash := sha256.New()
	hash.Write([]byte(salt))
	hashBytes := hash.Sum(nil)

	shuffled := make([]rune, len(alphabet))
	copy(shuffled, alphabet)

	for i := len(alphabet) - 1; i > 0; i-- {
		j := int(hashBytes[i%len(hashBytes)]) % (i + 1)
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	}
	return string(shuffled)
}

func Base62Encode(id int64, salt string) string {
	alphabet := getShuffledAlphabet(salt)
	var encoded string
	num := id

	if num == 0 {
		encoded = string(alphabet[0])
	} else {
		for num > 0 {
			encoded = string(alphabet[num%62]) + encoded
			num = num / 62
		}
	}
	if len(encoded) < 4 {
		encoded = strings.Repeat(string(alphabet[0]), 4-len(encoded)) + encoded
	} else if len(encoded) > 4 {
		encoded = encoded[:4]
	}
	return encoded
}
