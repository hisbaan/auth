package wellknown

import (
	"crypto/ed25519"
	"encoding/base64"
)

type WellKnownService struct {
	publicKey ed25519.PublicKey
	keyID     string
}

func NewWellKnownService(publicKey ed25519.PublicKey, keyID string) *WellKnownService {
	return &WellKnownService{
		publicKey: publicKey,
		keyID:     keyID,
	}
}

func (s *WellKnownService) GetJWKS() JWKS {
	return JWKS{
		Keys: []JWK{
			{
				Kty: "OKP",
				Use: "sig",
				Kid: s.keyID,
				Alg: "EdDSA",
				Crv: "Ed25519",
				X:   base64.RawURLEncoding.EncodeToString(s.publicKey),
			},
		},
	}
}
