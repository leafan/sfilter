package utils

import (
	"math/rand"
	"time"
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
