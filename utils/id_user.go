package utils

import "math/rand"

func GenerateIDUser(n int) string {
	char := []byte("123456789")
	result := make([]byte, n)

	for i := range result {
		result[i] = char[rand.Intn(len(char))]
	}
	return string(result)
}
