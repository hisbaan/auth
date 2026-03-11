package wellknown

import (
	"crypto/ed25519"
)

type WellKnownService struct {
	baseURL   string
	publicKey ed25519.PublicKey
	keyID     string
}

func NewWellKnownService(baseURL string, publicKey ed25519.PublicKey, keyID string) *WellKnownService {
	return &WellKnownService{
		baseURL:   baseURL,
		publicKey: publicKey,
		keyID:     keyID,
	}
}
