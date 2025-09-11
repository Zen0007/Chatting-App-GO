package utils

import "golang.org/x/crypto/bcrypt"


func CompareHashPassword(password,hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash),[]byte(password))
}