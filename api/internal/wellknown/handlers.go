package wellknown

import (
	"auth/internal/utils/httputil"
	"encoding/base64"
	"net/http"
	"strings"
)

func (s *WellKnownService) GetJWKSHandler(w http.ResponseWriter, r *http.Request) {
	jwks := JWKS{
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
	httputil.JSONResponse(w, http.StatusOK, jwks)
}

func (s *WellKnownService) GetOpenIDConfigurationHandler(w http.ResponseWriter, r *http.Request) {
	openIDConfiguration := OpenIDConfiguration{
		Issuer:                                     s.baseURL,
		AuthorizationEndpoint:                      joinURLPath(s.baseURL, "authorize"),
		TokenEndpoint:                              joinURLPath(s.baseURL, "token"),
		RevocationEndpoint:                         joinURLPath(s.baseURL, "token/revoke"),
		UserinfoEndpoint:                           joinURLPath(s.baseURL, "userinfo"),
		JWKSURI:                                    joinURLPath(s.baseURL, ".well-known/jwks.json"),
		ResponseTypesSupported:                     []string{"code"},
		SubjectTypesSupported:                      []string{"public"},
		IDTokenSigningAlgValuesSupported:           []string{"EdDSA"},
		ScopesSupported:                            []string{"openid", "profile", "email"},
		TokenEndpointAuthMethodsSupported:          []string{"none"},
		CodeChallengeMethodsSupported:              []string{"S256"},
		GrantTypesSupported:                        []string{"authorization_code", "refresh_token"},
		ClaimsSupported:                            []string{"sub", "preferred_username", "email", "email_verified"},
		AuthorizationResponseIssParameterSupported: true,
	}
	httputil.JSONResponse(w, http.StatusOK, openIDConfiguration)
}

func joinURLPath(baseURL string, path string) string {
	return strings.TrimRight(baseURL, "/") + "/" + strings.TrimLeft(path, "/")
}
