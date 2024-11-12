package utils

import (
	"math/rand"
)

func GenerateId() string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"

	b := make([]byte, 10)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}

	return string(b)
}
