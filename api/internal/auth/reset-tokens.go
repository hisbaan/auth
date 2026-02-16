package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
)

func GenerateResetToken() ([]byte, []byte) {
	token := make([]byte, 32)
	rand.Read(token)

	hashedToken := HashToken(token)

	return token, hashedToken
}

func HashToken(token []byte) []byte {
	h := sha256.New()
	h.Write([]byte(token))
	byteSlice := h.Sum(nil)
	return byteSlice
}

func URLEncodeToken(token []byte) string {
	return base64.URLEncoding.EncodeToString(token)
}

func URLDecodeToken(token string) ([]byte, error) {
	return base64.URLEncoding.DecodeString(token)
}
