package shortcode

import (
	"crypto/rand"
	"math/big"
)

const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func Generate(length int) (string, error) {
	if length <= 0 {
		length = 6
	}
	result := make([]byte, length)
	for i := range result {
		index, err := rand.Int(rand.Reader, big.NewInt(int64(len(alphabet))))
		if err != nil {
			return "", err
		}
		result[i] = alphabet[index.Int64()]
	}
	return string(result), nil
}
