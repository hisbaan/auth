package tokenutil

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
)

func Generate() ([]byte, []byte) {
	token := make([]byte, 32)
	rand.Read(token)

	hashedToken := Hash(token)

	return token, hashedToken
}

func Hash(token []byte) []byte {
	h := sha256.New()
	h.Write([]byte(token))
	byteSlice := h.Sum(nil)
	return byteSlice
}

func URLEncode(token []byte) string {
	return base64.URLEncoding.EncodeToString(token)
}

func URLDecode(token string) ([]byte, error) {
	return base64.URLEncoding.DecodeString(token)
}

func PKCES256Challenge(codeVerifier string) string {
	hash := Hash([]byte(codeVerifier))
	return base64.RawURLEncoding.EncodeToString(hash)
}
