package utils

import "math/rand"

func GenerateIdText(n int) string {
	char := []byte("abcdefghijklmnopqrstuxyz")
	result := make([]byte, n)

	for i := range result {
		result[i] = char[rand.Intn(len(char))]
	}

	return string(result)
}
