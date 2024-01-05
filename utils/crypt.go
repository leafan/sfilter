package utils

import (
	"math/rand"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func GenerateVerifyCode(length int) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	charset := "0123456789"

	code := make([]byte, length)
	for i := range code {
		code[i] = charset[r.Intn(len(charset))]
	}

	return string(code)
}

func GenerateAsciiCode(length int) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	charset := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"

	code := make([]byte, length)
	for i := range code {
		code[i] = charset[r.Intn(len(charset))]
	}

	return string(code)
}

func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		Errorf("[ HashPassword ] generate pass error: %v", err)
		return "", err
	}
	return string(hashedPassword), nil
}

func ComparePassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

func TEST_HASH_PASSWORD(password string) {
	resp, _ := HashPassword(password)

	Tracef("[ TEST_HASH_PASSWORD ] origin: %v, after: %v", password, resp)
}
