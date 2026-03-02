package wellknown

import (
	"crypto/ed25519"
	"encoding/base64"
)

type WellKnownService struct {
	accessPublicKey  ed25519.PublicKey
	refreshPublicKey ed25519.PublicKey
	accessKeyID      string
	refreshKeyID     string
}

func NewWellKnownService(accessPublicKey ed25519.PublicKey, refreshPublicKey ed25519.PublicKey, accessKeyID string, refreshKeyID string) *WellKnownService {
	return &WellKnownService{
		accessPublicKey:  accessPublicKey,
		refreshPublicKey: refreshPublicKey,
		accessKeyID:      accessKeyID,
		refreshKeyID:     refreshKeyID,
	}
}

func (s *WellKnownService) GetJWKS() JWKS {
	return JWKS{
		Keys: []JWK{
			{
				Kty: "OKP",
				Use: "sig",
				Kid: s.accessKeyID,
				Alg: "EdDSA",
				Crv: "Ed25519",
				X:   base64.RawURLEncoding.EncodeToString(s.accessPublicKey),
			},
			{
				Kty: "OKP",
				Use: "sig",
				Kid: s.refreshKeyID,
				Alg: "EdDSA",
				Crv: "Ed25519",
				X:   base64.RawURLEncoding.EncodeToString(s.refreshPublicKey),
			},
		},
	}
}
